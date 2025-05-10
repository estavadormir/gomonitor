package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	PORT = 8080
)

func main() {
	fmt.Println("Starting Service Monitor...")

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
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// root handler
func homeHandler(w http.ResponseWriter, r *http.Request) {
	//lets log the request
	log.Printf("Request receive: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "text/html")

	//might change it later to smt else
	fmt.Fprintf(w, "<html><body><h1>Service Monitor</h1><p>Welcome to the Service Monitor!</p><p><a href='/health'>Check health status</a></p></body></html>")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	//log the request
	log.Printf("Request recevied: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-type", "application/json")

	response := `{"status": "up", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`

	w.Write([]byte(response))
}
