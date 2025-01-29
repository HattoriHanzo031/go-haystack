package main

import (
	"encoding/base64"
	"math/rand"
	"os"
	"strconv"
	"text/template"

	"github.com/HattoriHanzo031/go-haystack/lib/device"
)

func saveKeys(name string, priv string, pub string, hash string) error {
	pk, err := base64.StdEncoding.DecodeString(priv)
	if err != nil {
		return err
	}

	device := device.Device{
		Name:             name,
		ID:               hash,
		AdvertisementKey: pub,
		PrivateKey:       pk,
	}
	return device.SaveToFile()
}

const deviceTemplate = `[
    {
        "id": {{.ID}},
        "colorComponents": [
            0,
            1,
            0,
            1
        ],
        "name": "{{.Name}}",
        "privateKey": "{{.PrivateKey}}",
        "icon": "",
        "isDeployed": true,
        "colorSpaceName": "kCGColorSpaceExtendedSRGB",
        "usesDerivation": false,
        "isActive": false,
        "additionalKeys": []
    }
]
`

func saveDevice(name string, priv string) error {
	t, err := template.New("device").Parse(deviceTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(name + ".json")
	if err != nil {
		return err
	}

	defer f.Close()

	err = t.Execute(f, map[string]string{
		"ID":         randomInt(1000, 999999),
		"Name":       name,
		"PrivateKey": priv,
	})
	if err != nil {
		return err
	}

	return nil
}

// Returns an int >= min, < max
func randomInt(min, max int) string {
	return strconv.Itoa(min + rand.Intn(max-min))
}
