package reports

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hybridgroup/go-haystack/lib/device"
)

type Report struct {
	DatePublished time.Time
	Description   string
	Data          PayloadData
	StatusCode    int64
	// RawPayload    string
}

type Reports map[string][]Report

type Get func([]device.Device) (Reports, error)

func GetFn(url string, days int) Get {
	return func(devices []device.Device) (Reports, error) {
		contentType := "application/json; charset=utf-8"

		mappedDevices := make(map[string]device.Device, len(devices))
		ids := make([]string, 0, len(devices))
		for _, d := range devices {
			ids = append(ids, d.ID)
			mappedDevices[d.ID] = d
		}

		jsonData, err := json.Marshal(map[string]interface{}{
			"ids":  ids,
			"days": days,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}

		resp, err := http.Post(url, contentType, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to make POST request: %w", err)
		}
		defer resp.Body.Close()

		response := struct {
			Results []struct {
				DatePublished int64  `json:"datePublished"`
				Payload       string `json:"payload"`
				Description   string `json:"description"`
				ID            string `json:"id"`
				StatusCode    int64  `json:"statusCode"`
			} `json:"results"`
			StatusCode string `json:"statusCode"`
		}{}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}

		if response.StatusCode != "200" {
			return nil, fmt.Errorf("failed to get reports: %s", response.StatusCode)
		}

		reports := make(Reports)
		for _, report := range response.Results {
			// Decode Base64-encoded values
			rawPayload, err := base64.StdEncoding.DecodeString(report.Payload)
			if err != nil {
				fmt.Printf("failed to decode payload %v\n", err)
				continue
			}
			decrypted, err := decrypt(rawPayload, mappedDevices[report.ID].PrivateKey)
			if err != nil {
				fmt.Printf("failed to decrypt payload: %v\n", err)
				continue
			}

			reports[report.ID] = append(reports[report.ID], Report{
				DatePublished: time.UnixMilli(report.DatePublished).Local(),
				Description:   report.Description,
				Data:          parse(rawPayload, decrypted),
				StatusCode:    report.StatusCode,
				//RawPayload:    report.Payload,
			})
		}
		return reports, nil
	}
}
