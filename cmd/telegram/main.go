package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/HattoriHanzo031/go-haystack/lib/device"
	"github.com/HattoriHanzo031/go-haystack/lib/reports"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	chatID := int64(1799333354)
	botApiToken, ok := os.LookupEnv("TELEGRAM_API_TOKEN")
	if !ok {
		log.Fatalf("TELEGRAM_API_TOKEN environment variable not set")
	}

	endpoint := flag.String("endpoint", "http://localhost:6176", "Address of the macless-haystack server")
	days := flag.Int("days", 7, "Number of days to retrieve reports for")
	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	getReports := reports.GetFn(*endpoint, *days)
	devices := []device.Device{}
	for _, deviceFile := range flag.Args() {
		device, err := device.LoadFromFile(deviceFile)
		if err != nil {
			log.Fatalf("failed to load device: %v", err)
		}
		devices = append(devices, *device)
	}

	bot, err := tgbotapi.NewBotAPI(botApiToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Set debug mode for development
	bot.Debug = *debug
	run(bot, chatID, getReports, devices)
}

func run(bot *tgbotapi.BotAPI, chatID int64, getReports reports.Get, devices []device.Device) {
	deviceNames := make(map[string]device.Device, len(devices))
	deviceIDs := make(map[string]device.Device, len(devices))
	for _, d := range devices {
		deviceNames[d.Name] = d
		deviceIDs[d.ID] = d
	}

	handleMessages(bot, chatID, map[string]func(bot *tgbotapi.BotAPI, args string) error{
		"locate": func(bot *tgbotapi.BotAPI, args string) error {
			locateDevices := []device.Device{}
			if args == "" || args == "all" {
				locateDevices = devices
			} else if device, ok := deviceNames[args]; ok {
				locateDevices = append(locateDevices, device)
			} else {
				return fmt.Errorf("device(s) %s not found", args)
			}

			deviceReports, err := getReports(locateDevices)
			if err != nil {
				e := device.NonFatalError{}
				if errors.As(err, &e) {
					fmt.Println("reports retrieved with errors:", e)
				} else {
					return fmt.Errorf("failed to get reports: %w", err)
				}
			}
			if len(deviceReports) == 0 {
				return fmt.Errorf("No reports found")
			}

			for id, deviceReport := range deviceReports {
				if len(deviceReport) == 0 {
					fmt.Println("No reports found for device", deviceIDs[id].Name)
					continue
				}
				slices.SortFunc(deviceReport, func(i, j reports.Report) int {
					return i.Data.Timestamp.Compare(j.Data.Timestamp)
				})

				report := deviceReport[len(deviceReport)-1]
				msg := tgbotapi.NewLocation(chatID, report.Data.Latitude, report.Data.Longitude)
				if _, err = bot.Send(msg); err != nil {
					return fmt.Errorf("Failed to send location: %w", err)
				}
				if _, err = bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("[%s]: %s", deviceIDs[id].Name, report.Data.Timestamp))); err != nil {
					return fmt.Errorf("Failed to send timestamp: %w", err)
				}
			}
			return nil
		},
	})
}

func handleMessages(bot *tgbotapi.BotAPI /*TODO: logger*/, chatID int64, handlers map[string]func(bot *tgbotapi.BotAPI, args string) error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		command, args, _ := strings.Cut(update.Message.Text, " ")
		if handler, ok := handlers[strings.ToLower(strings.TrimSpace(command))]; ok {
			err := handler(bot, strings.ToLower(strings.TrimSpace(args)))
			if err != nil {
				fmt.Printf("Failed to handle command %s: %v", command, err) // TODO: logger
				_, _ = bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Failed to handle command %s: %v", command, err)))
			}
		}
	}
}
