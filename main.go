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
	"register-bot/internal/tasks"
	"strings"
	"sync"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// TaskConfig represents a single task configuration from CSV
type TaskConfig struct {
	Term              string
	Subject           string
	Mode              string
	CRNs              []string
	RegistrationTime  string
	Username          string
	Password          string
	WebhookURL        string
}

// loadCredentials reads username, password, and webhook from .credentials file
func loadCredentials() (string, string, string) {
	file, err := os.Open("config/.credentials")
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

// createHTTPClient creates a new HTTP client with its own cookie jar and dialer
func createHTTPClient() (tls_client.HttpClient, error) {
	var dnsServers = []string{"8.8.8.8", "8.8.4.4", "1.1.1.1", "1.0.0.1"}
	randomIndex := rand.Intn(len(dnsServers))
	dnsServer := dnsServers[randomIndex]

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

	return tls_client.NewHttpClient(tls_client.NewLogger(), client_options...)
}

// parseCSVRow parses a CSV row into a TaskConfig
func parseCSVRow(row []string, credUsername, credPassword, credWebhook string) (*TaskConfig, error) {
	if len(row) < 5 {
		return nil, fmt.Errorf("invalid row: expected at least 5 columns, got %d", len(row))
	}

	config := &TaskConfig{
		Term:             row[0],
		Subject:          row[1],
		Mode:             strings.TrimSpace(row[2]),
		CRNs:             strings.Split(strings.Trim(row[3], "\""), ","),
		RegistrationTime: row[4],
	}

	// Clean up CRNs (remove empty strings)
	var cleanCRNs []string
	for _, crn := range config.CRNs {
		crn = strings.TrimSpace(crn)
		if crn != "" {
			cleanCRNs = append(cleanCRNs, crn)
		}
	}
	config.CRNs = cleanCRNs

	// Set default mode to Watch if empty
	if config.Mode == "" {
		config.Mode = "Watch"
	}

	// Read username and password with priority: env vars > credentials file
	if envUsername := os.Getenv("REGISTER_BOT_USERNAME"); envUsername != "" {
		config.Username = envUsername
	} else if credUsername != "" {
		config.Username = credUsername
	} else {
		return nil, fmt.Errorf("username not found in environment variables or .credentials file")
	}

	if envPassword := os.Getenv("REGISTER_BOT_PASSWORD"); envPassword != "" {
		config.Password = envPassword
	} else if credPassword != "" {
		config.Password = credPassword
	} else {
		return nil, fmt.Errorf("password not found in environment variables or .credentials file")
	}

	// Read webhook from credentials file or environment variable
	if envWebhook := os.Getenv("REGISTER_BOT_WEBHOOK"); envWebhook != "" {
		config.WebhookURL = envWebhook
	} else if credWebhook != "" {
		config.WebhookURL = credWebhook
	} else {
		config.WebhookURL = "" // Webhook is optional
	}

	return config, nil
}

// runTask runs a single task configuration
func runTask(config *TaskConfig) {
	// Create a new HTTP client for this task (each task needs its own session)
	client, err := createHTTPClient()
	if err != nil {
		fmt.Printf("[%s] Error creating HTTP client: %v\n", config.Term, err)
		return
	}
	defer client.CloseIdleConnections()

	// Create task instance
	t := &tasks.Task{
		Client:     client,
		Username:   config.Username,
		Password:   config.Password,
		WebhookURL: config.WebhookURL,
		Subject:    config.Subject,
		Mode:       config.Mode,
		CRNs:       config.CRNs,
	}

	// Get term ID
	t.GetTermByName(config.Term)

	// Handle Release mode (wait until registration time)
	if config.Mode == "Release" {
		t.Mode = "Signup"
		pattern := regexp.MustCompile(`\d{2}/\d{2}/\d{4} \d{2}:\d{2} [APM]{2}`)
		matches := pattern.FindAllString(config.RegistrationTime, -1)
		if len(matches) == 0 {
			fmt.Printf("[%s] Invalid Registration Time Format\n", config.Term)
			return
		}

		location, _ := time.LoadLocation("America/Los_Angeles")
		targetTime, _ := time.ParseInLocation("01/02/2006 03:04 PM", matches[0], location)
		now := time.Now().In(location)
		timeToWait := targetTime.Sub(now) - 5*time.Minute

		if now.Before(targetTime) {
			fmt.Printf("[%s] Will continue in: %s\n", config.Term, timeToWait.String())
			time.Sleep(timeToWait)
		}
	}

	// Log task start
	fmt.Printf("[%s] Starting task: Mode=%s, Subject=%s, CRNs=%v\n", config.Term, t.Mode, t.Subject, t.CRNs)

	// Run the task
	t.Run()
}

func main() {
	file, err := os.Open("config/settings.csv")
	if err != nil {
		fmt.Println("Error Opening File:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	_, err = reader.Read()
	if err != nil {
		fmt.Println("Error Reading Header:", err)
		return
	}

	// Load credentials once (priority: env vars > credentials file)
	credUsername, credPassword, credWebhook := loadCredentials()

	// Read all CSV rows and create task configurations
	var taskConfigs []*TaskConfig
	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("Error Reading Row: %v\n", err)
			continue
		}

		config, err := parseCSVRow(row, credUsername, credPassword, credWebhook)
		if err != nil {
			fmt.Printf("Error parsing row: %v\n", err)
			continue
		}

		taskConfigs = append(taskConfigs, config)
	}

	if len(taskConfigs) == 0 {
		fmt.Println("No valid task configurations found in settings.csv")
		return
	}

	fmt.Printf("Loaded %d task configuration(s). Starting concurrent execution...\n\n", len(taskConfigs))

	// Run all tasks concurrently
	var wg sync.WaitGroup
	for i, config := range taskConfigs {
		wg.Add(1)
		go func(idx int, cfg *TaskConfig) {
			defer wg.Done()
			runTask(cfg)
		}(i, config)
	}

	// Wait for all tasks to complete
	wg.Wait()
	fmt.Println("\nAll tasks completed.")
}
