package device

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Device struct {
	Name             string
	ID               string
	AdvertisementKey string
	PrivateKey       []byte
}

func LoadFromFile(fileName string) (*Device, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file (%s): %w", fileName, err)
	}
	defer f.Close()

	baseName := filepath.Base(fileName)
	device := Device{
		Name: strings.TrimSuffix(baseName, ".keys"),
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		key, val, found := strings.Cut(line, ": ")
		if !found {
			fmt.Println("invalid line:", line)
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch key {
		case "Private key":
			pk, err := base64.StdEncoding.DecodeString(val)
			if err != nil {
				return nil, fmt.Errorf("failed to decode private key: %w", err)
			}
			device.PrivateKey = pk
		case "Hashed adv key":
			device.ID = val
		case "Advertisement key":
			device.AdvertisementKey = val
		default:
			fmt.Println("unknown key:", key, "on line:", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file (%s): %w", fileName, err)
	}
	return &device, nil
}

func (d *Device) SaveToFile() error {
	f, err := os.Create(d.Name + ".keys")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("Private key: %s\n", d.PrivateKey))
	b.WriteString(fmt.Sprintf("Advertisement key: %s\n", d.AdvertisementKey))
	b.WriteString(fmt.Sprintf("Hashed adv key: %s\n", base64.StdEncoding.EncodeToString(d.PrivateKey)))
	if _, err = f.WriteString(b.String()); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func Generate(name string) (*Device, error) {
	// Generate ECDSA private key using P-224 curve
	pk, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// Extract the raw private key bytes
	privateKeyBytes := pk.D.Bytes()

	// Ensure the private key is 28 bytes long (P-224 curve)
	if len(privateKeyBytes) != 28 {
		return nil, errors.New("Private key is not 28 bytes long")
	}

	// extract raw public key bytes
	publicKeyBytes := pk.PublicKey.X.Bytes()

	// Encode the public key to Base64
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	// Hash the public key using SHA-256
	hash := sha256.Sum256(publicKeyBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])

	// make sure not '/' in the base64 string
	if strings.Contains(hashBase64, "/") {
		return nil, fmt.Errorf("hash contains '/'")
	}

	return &Device{
		Name:             name,
		ID:               hashBase64,
		AdvertisementKey: publicKeyBase64,
		PrivateKey:       privateKeyBytes,
	}, nil
}
