package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/estavadormir/gomonitor/config"
	"github.com/estavadormir/gomonitor/monitor"
)

const (
	PORT        = 8080
	CONFIG_FILE = "config.json"
)

var (
	appConfig  *config.Config
	appMonitor *monitor.Monitor
)

func main() {
	fmt.Println("Starting Service Monitor...")

	var err error
	appConfig, err = config.Load(CONFIG_FILE)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Loaded configuration with %d services to monitor\n", len(appConfig.Services))

	appMonitor = monitor.New(appConfig)
	appMonitor.Start()

	defer appMonitor.Stop()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/api/services", servicesAPIHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", PORT),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("Server starting on port %d...\n", PORT)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "text/html")

	results := appMonitor.GetResults()

	html := fmt.Sprintf(`
	<html>
	<head>
		<title>%s</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 0; padding: 20px; line-height: 1.6; }
			h1 { color: #333; }
			.services { margin-top: 20px; }
			.service { border: 1px solid #ddd; padding: 10px; margin-bottom: 10px; border-radius: 4px; }
			.service h3 { margin-top: 0; }
			.up { background-color: #d4edda; border-color: #c3e6cb; }
			.down { background-color: #f8d7da; border-color: #f5c6cb; }
			.unknown { background-color: #fff3cd; border-color: #ffeeba; }
			.status-indicator { display: inline-block; padding: 3px 8px; border-radius: 3px; margin-left: 10px; }
			.status-up { background-color: #28a745; color: white; }
			.status-down { background-color: #dc3545; color: white; }
			.status-unknown { background-color: #ffc107; color: black; }
		</style>
		<script>
			// Refresh the page at the configured interval
			setTimeout(function() {
				window.location.reload();
			}, %d * 1000);
		</script>
	</head>
	<body>
		<h1>%s</h1>
		<p>Monitoring %d services. Dashboard refreshes every %d seconds.</p>
		<p><a href='/health'>View health API response</a> | <a href='/api/services'>JSON API</a></p>

		<div class="services">
			<h2>Service Status</h2>`,
		appConfig.Dashboard.Title,
		appConfig.Dashboard.RefreshInterval,
		appConfig.Dashboard.Title,
		len(appConfig.Services),
		appConfig.Dashboard.RefreshInterval)

	for _, service := range appConfig.Services {
		result, exists := results[service.Name]

		var statusClass, statusText, statusIndicatorClass string
		var responseTime, statusCode int64
		var lastChecked, message string

		if exists {
			statusClass = result.Status
			statusText = result.Status
			statusIndicatorClass = "status-" + result.Status
			responseTime = result.ResponseTime
			statusCode = int64(result.StatusCode)
			lastChecked = result.LastChecked.Format(time.RFC3339)
			message = result.Message
		} else {
			statusClass = "unknown"
			statusText = "unknown"
			statusIndicatorClass = "status-unknown"
			lastChecked = "Never"
			message = "Not checked yet"
		}

		html += fmt.Sprintf(`
			<div class="service %s">
				<h3>%s <span class="status-indicator %s">%s</span></h3>
				<p>URL: %s</p>
				<p>Last Checked: %s</p>
				<p>Response Time: %d ms</p>
				<p>Status Code: %d</p>
				<p>Message: %s</p>
			</div>`,
			statusClass, service.Name, statusIndicatorClass, statusText,
			service.URL, lastChecked, responseTime, statusCode, message)
	}

	html += `
		</div>
	</body>
	</html>`

	fmt.Fprint(w, html)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health check requested: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	results := appMonitor.GetResultsSlice()

	overallStatus := "up"
	for _, result := range results {
		if result.Status == "down" {
			overallStatus = "down"
			break
		} else if result.Status == "unknown" && overallStatus != "down" {
			overallStatus = "unknown"
		}
	}

	response := struct {
		Status    string           `json:"status"`
		Timestamp string           `json:"timestamp"`
		Services  []monitor.Result `json:"services"`
	}{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  results,
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

func servicesAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("API request: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	results := appMonitor.GetResultsSlice()

	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
