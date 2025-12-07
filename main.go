package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net"
	"os"
	"regexp"
	"register-bot/tasks"
	"strings"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// loadCredentials reads username, password, and webhook from .credentials file
func loadCredentials() (string, string, string) {
	file, err := os.Open(".credentials")
	if err != nil {
		return "", "", ""
	}
	defer file.Close()

	var username, password, webhook string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "username=") {
			username = strings.TrimPrefix(line, "username=")
		} else if strings.HasPrefix(line, "password=") {
			password = strings.TrimPrefix(line, "password=")
		} else if strings.HasPrefix(line, "webhook=") {
			webhook = strings.TrimPrefix(line, "webhook=")
		}
	}

	return username, password, webhook
}

func main() {
	var dnsServers = []string{"8.8.8.8", "8.8.4.4", "1.1.1.1", "1.0.0.1"}
	randomIndex := rand.Intn(len(dnsServers))

	dnsServer := dnsServers[randomIndex]

	t := &tasks.Task{}
	jar := tls_client.NewCookieJar()
	dialer := net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(context context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(5) * time.Second,
				}
				return d.DialContext(context, "udp", net.JoinHostPort(dnsServer, "53"))
			},
		},
	}

	client_options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_117),
		tls_client.WithCookieJar(jar),
		tls_client.WithDialer(dialer),
	}
	t.Client, _ = tls_client.NewHttpClient(tls_client.NewLogger(), client_options...)

	file, err := os.Open("settings.csv")
	if err != nil {
		fmt.Println("Error Opening File:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.Read()
	if err != nil {
		fmt.Println("Error Reading Header:", err)
		return
	}

	// Load credentials once (priority: env vars > credentials file)
	credUsername, credPassword, credWebhook := loadCredentials()
	for {
		row, err := reader.Read()
		if err != nil {
			if err != csv.ErrFieldCount {
				fmt.Println("Finished Reading Configuration File")
			} else {
				fmt.Println("Error Reading Row: ", err)
			}
			break
		}
		if len(row) < 5 {
			fmt.Println("Invalid Configuration File")
			continue
		}
		// Read username and password with priority: env vars > credentials file
		if envUsername := os.Getenv("REGISTER_BOT_USERNAME"); envUsername != "" {
			t.Username = envUsername
		} else if credUsername != "" {
			t.Username = credUsername
		} else {
			fmt.Println("Error: Username not found in environment variables or .credentials file")
			continue
		}

		if envPassword := os.Getenv("REGISTER_BOT_PASSWORD"); envPassword != "" {
			t.Password = envPassword
		} else if credPassword != "" {
			t.Password = credPassword
		} else {
			fmt.Println("Error: Password not found in environment variables or .credentials file")
			continue
		}

		// Read webhook from credentials file or environment variable
		if envWebhook := os.Getenv("REGISTER_BOT_WEBHOOK"); envWebhook != "" {
			t.WebhookURL = envWebhook
		} else if credWebhook != "" {
			t.WebhookURL = credWebhook
		} else {
			t.WebhookURL = "" // Webhook is optional
		}

		// Read other settings from CSV (Term, Subject, Mode, CRNs, SavedRegistrationTime)
		t.GetTermByName(row[0])
		t.Subject = row[1]
		t.Mode = row[2]
		t.CRNs = strings.Split(row[3], ",")
		var registrationTime = row[4]

		if t.Mode == "Release" {
			t.Mode = "Signup"
			pattern := regexp.MustCompile(`\d{2}/\d{2}/\d{4} \d{2}:\d{2} [APM]{2}`)
			matches := pattern.FindAllString(registrationTime, -1)
			if len(matches) == 0 {
				fmt.Printf("Invalid Registration Time Format")
			}

			location, _ := time.LoadLocation("America/Los_Angeles")
			targetTime, _ := time.ParseInLocation("01/02/2006 03:04 PM", matches[0], location)
			now := time.Now().In(location)
			timeToWait := targetTime.Sub(now) - 5*time.Minute

			if now.Before(targetTime) {
				fmt.Printf("Will continue in: %s\n", timeToWait.String())
				time.Sleep(timeToWait)
			}
		}
	}

	t.Run()
}
