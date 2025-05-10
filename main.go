package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/estavadormir/gomonitor/config"
)

const (
	PORT        = 8080
	CONFIG_FILE = "config.json"
)

var appConfig *config.Config

func main() {
	fmt.Println("Starting Service Monitor...")

	var err error
	appConfig, err = config.Load(CONFIG_FILE)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	//log the loaded configs
	fmt.Printf("Loaded config with %d services to monitor", len(appConfig.Services))

	//register the handler for the diff routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthCheckHandler)

	//config the server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", PORT),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("Server starting on port %d...\n", PORT)

	//serving or stopping and checking for errors
	err = server.ListenAndServe()

	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// root handler
func homeHandler(w http.ResponseWriter, r *http.Request) {
	//lets log the request
	log.Printf("Request receive: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "text/html")
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
			</style>
		</head>
		<body>
			<h1>%s</h1>
			<p>Monitoring %d services. Dashboard refreshes every %d seconds.</p>
			<p><a href='/health'>Check overall health status</a></p>

			<div class="services">
				<h2>Configured Services</h2>`,
		appConfig.Dashboard.Title,
		appConfig.Dashboard.Title,
		len(appConfig.Services),
		appConfig.Dashboard.RefreshInterval)

	for _, service := range appConfig.Services {
		html += fmt.Sprintf(`
				<div class="service">
					<h3>%s</h3>
					<p>URL: %s</p>
					<p>Check Interval: %d seconds</p>
					<p>Timeout: %d seconds</p>
				</div>`,
			service.Name, service.URL, service.Interval, service.Timeout)
	}

	// Close the HTML tags
	html += `
			</div>
		</body>
		</html>`

	// Write the response
	fmt.Fprint(w, html)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	//log the request
	log.Printf("Request recevied: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-type", "application/json")

	response := fmt.Sprintf(`{
		"status": "up",
		"timestamp": "%s",
		"monitor": {
			"services_configured": %d,
			"dashboard_title": "%s"
		}
	}`, time.Now().Format(time.RFC3339), len(appConfig.Services), appConfig.Dashboard.Title)

	w.Write([]byte(response))
}
