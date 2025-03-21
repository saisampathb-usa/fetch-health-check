package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Endpoint struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

type DomainStats struct {
	Success int
	Total   int
}

var stats = make(map[string]*DomainStats)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <config_file>")
	}

	configFile := os.Args[1]
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", configFile, err)
	}

	var endpoints []Endpoint
	if err := yaml.Unmarshal(data, &endpoints); err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	initStats(endpoints)

	for {
		for _, endpoint := range endpoints {
			checkHealth(endpoint)
		}
		logResults()
		time.Sleep(15 * time.Second)
	}
}

// Initializes stats map for all unique domains
func initStats(endpoints []Endpoint) {
	for _, ep := range endpoints {
		domain := extractDomain(ep.URL)
		if stats[domain] == nil {
			stats[domain] = &DomainStats{}
		}
	}
}

// Executes a health check for a given endpoint
func checkHealth(endpoint Endpoint) {
	method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
	if method == "" {
		method = http.MethodGet
	}

	var body io.Reader
	if endpoint.Body != "" && (method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch) {
		body = bytes.NewBuffer([]byte(endpoint.Body))
	}

	req, err := http.NewRequest(method, endpoint.URL, body)
	if err != nil {
		log.Printf("Failed to create request for %s: %v", endpoint.Name, err)
		return
	}

	for k, v := range endpoint.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	domain := extractDomain(endpoint.URL)

	// Make sure domain exists in stats (should always be true due to initStats)
	if stats[domain] == nil {
		stats[domain] = &DomainStats{}
	}
	stats[domain].Total++

	if err == nil && resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 && elapsed < 500*time.Millisecond {
			stats[domain].Success++
		}
	}
}

// Logs current availability per domain
func logResults() {
	for domain, stat := range stats {
		if stat.Total == 0 {
			fmt.Printf("%s has 0%% availability\n", domain)
			continue
		}
		percentage := (100 * stat.Success) / stat.Total // integer division drops decimal
		fmt.Printf("%s has %d%% availability\n", domain, percentage)
	}
}

// Extracts domain (without port) from a URL
func extractDomain(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw // fallback to raw if parsing fails
	}
	host := parsed.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return host
}
