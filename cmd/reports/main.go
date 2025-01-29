package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/HattoriHanzo031/go-haystack/lib/device"
	"github.com/HattoriHanzo031/go-haystack/lib/reports"
)

func main() {
	endpoint := flag.String("endpoint", "http://localhost:6176", "Address of the macless-haystack server")
	days := flag.Int("days", 7, "Number of days to retrieve reports for")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	devices := []device.Device{}
	for _, deviceFile := range flag.Args() {
		device, err := device.LoadFromFile(deviceFile)
		if err != nil {
			fmt.Println("failed to load device:", err)
			os.Exit(1)
		}
		devices = append(devices, *device)
	}

	if err := run(reports.GetFn(*endpoint, *days), devices); err != nil {
		fmt.Println("failed to run:", err)
		os.Exit(1)
	}
}

func run(getReports reports.Get, devices []device.Device) error {
	deviceReports, err := getReports(devices)
	if err != nil {
		return fmt.Errorf("failed to get reports: %w", err)
	}

	out, _ := json.MarshalIndent(deviceReports, "", "\t")
	fmt.Println(string(out))
	return nil
}
