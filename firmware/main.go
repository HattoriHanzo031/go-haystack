// Firmware to advertise a FindMy compatible device aka AirTag
// see https://github.com/biemster/FindMy for more information.
//
// To build:
// tinygo flash -target nano-rp2040 -ldflags="-X main.AdvertisingKey='SGVsbG8sIFdvcmxkIQ=='" .
//
// For Linux:
// go run . SGVsbG8sIFdvcmxkIQ==
package main

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/HattoriHanzo031/go-haystack/lib/findmy"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	// wait for USB serial to be available
	time.Sleep(2 * time.Second)

	key, err := getKeyData()
	if err != nil {
		fail("failed to get key data: " + err.Error())
	}
	println("key is", AdvertisingKey, "(", len(key), "bytes)")

	opts := bluetooth.AdvertisementOptions{
		AdvertisementType: bluetooth.AdvertisingTypeNonConnInd,
		Interval:          bluetooth.NewDuration(1285000 * time.Microsecond), // 1285ms
		ManufacturerData:  []bluetooth.ManufacturerDataElement{findmy.NewData(key)},
	}

	must("enable BLE stack", adapter.Enable())

	// Set the address to the first 6 bytes of the public key.
	adapter.SetRandomAddress(bluetooth.MAC{key[5], key[4], key[3], key[2], key[1], key[0] | 0xC0})

	println("configure advertising...")
	adv := adapter.DefaultAdvertisement()
	must("config adv", adv.Configure(opts))

	println("start advertising...")
	must("start adv", adv.Start())

	address, _ := adapter.Address()
	for {
		println("FindMy device using", address.MAC.String())
		time.Sleep(time.Second)
	}
}

// getKeyData returns the public key data from the base64 encoded string.
func getKeyData() ([]byte, error) {
	val, err := base64.StdEncoding.DecodeString(AdvertisingKey)
	if err != nil {
		return nil, err
	}
	if len(val) != 28 {
		return nil, errors.New("public key must be 28 bytes long")
	}

	return val, nil
}

// must calls a function and fails if an error occurs.
func must(action string, err error) {
	if err != nil {
		fail("failed to " + action + ": " + err.Error())
	}
}

// fail prints a message over and over forever.
func fail(msg string) {
	for {
		println(msg)
		time.Sleep(time.Second)
	}
}
