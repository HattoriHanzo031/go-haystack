package reports

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"time"
)

type PayloadData struct {
	Timestamp         time.Time
	Longitude         float64
	Latitude          float64
	AccuracyMeters    uint8
	ConfidencePercent uint8
}

func dhExchange(privateKey, publicKey []byte) ([]byte, error) {
	curve := elliptic.P224() // Matches SECP224R1

	x, y := elliptic.Unmarshal(curve, publicKey)
	if x == nil || y == nil {
		return nil, errors.New("invalid public key")
	}
	sharedX, _ := curve.ScalarMult(x, y, privateKey)
	return sharedX.Bytes(), nil
}

func decrypt(payload, key []byte) ([]byte, error) {
	encryptedData := payload[4:]

	// Handle potential MacOS 14+ report format
	if len(encryptedData) == 85 {
		encryptedData = encryptedData[1:]
	}

	ephKey := encryptedData[1:58] // Ephemeral public key

	sharedKey, err := dhExchange(key, ephKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive shared key: %v", err)
	}

	// Derive symmetric key
	hash := sha256.New()
	hash.Write(sharedKey)
	hash.Write([]byte{0x00, 0x00, 0x00, 0x01})
	hash.Write(ephKey)
	symmetricKey := hash.Sum(nil)

	decryptionKey := symmetricKey[:16]
	iv := symmetricKey[16:]
	encData := encryptedData[58:68]
	tag := encryptedData[68:]

	block, err := aes.NewCipher(decryptionKey)
	if err != nil {
		log.Fatal(err)
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		log.Fatal("NewGCMWithNonceSize", err)
	}

	decrypted, err := aesgcm.Open(nil, iv, append(encData, tag...), nil)
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}

	return decrypted, nil
}

func parse(raw, decrypted []byte) PayloadData {
	seenTimeStamp := binary.BigEndian.Uint32(raw[:4])
	return PayloadData{
		Timestamp:         time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(seenTimeStamp) * time.Second).Local(),
		Longitude:         float64(binary.BigEndian.Uint32(decrypted[4:8])) / 10000000,
		Latitude:          float64(binary.BigEndian.Uint32(decrypted[:4])) / 10000000,
		AccuracyMeters:    decrypted[8],
		ConfidencePercent: raw[4],
	}
}
