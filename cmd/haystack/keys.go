package main

import (
	"encoding/base64"
	"errors"

	"github.com/HattoriHanzo031/go-haystack/lib/device"
)

var errInvalidHash = errors.New("Hash contains '/'")

func generateKey() (string, string, string, error) {
	d, err := device.Generate("")
	if err != nil {
		return "", "", "", err
	}
	return base64.StdEncoding.EncodeToString(d.PrivateKey), d.AdvertisementKey, d.ID, nil
}
