package monitor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/estavadormir/gomonitor/config"
)

// the outcome from an health check
type Result struct {
	ServiceName  string    `json:"serviceName"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	ResponseTime int64     `json:"responseTimeMs"`
	StatusCode   int       `json:"statusCode"`
	Message      string    `json:"message"`
	LastChecked  time.Time `json:"lastChecked"`
}

// manages health checks for config'ed services
type Monitor struct {
	config      *config.Config
	results     map[string]*Result
	resultMutex sync.RWMutex
	client      *http.Client
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// create a new monitor instance
func New(cfg *config.Config) *Monitor {
	return &Monitor{
		config:   cfg,
		results:  make(map[string]*Result),
		stopChan: make(chan struct{}),
		//create a custom http client that cant be re-used for all req
		client: &http.Client{},
	}
}

// starts the monitoring for all config'ed services
func (m *Monitor) Start() {

	for _, svc := range m.config.Services {
		m.results[svc.Name] = &Result{
			ServiceName: svc.Name,
			URL:         svc.URL,
			Status:      "unknown",
			Message:     "Monitoring not started",
			LastChecked: time.Now(),
		}
	}

	for _, svc := range m.config.Services {
		//create a copy
		service := svc

		m.wg.Add(1)
		go m.monitorService(service)
	}
}

// stop all the monitoring routines
func (m *Monitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

func (m *Monitor) monitorService(svc config.ServiceConfig) {
	defer m.wg.Done()

	ticker := time.NewTicker(svc.GetCheckInterval())
	defer ticker.Stop()

	// init check
	m.checkService(svc)

	// then check on each tick
	for {
		select {
		case <-ticker.C:
			m.checkService(svc)
		case <-m.stopChan:
			return
		}
	}
}

func (m *Monitor) checkService(svc config.ServiceConfig) {

	ctx, cancel := context.WithTimeout(context.Background(), svc.GetTimeout())
	defer cancel()

	result := &Result{
		ServiceName: svc.Name,
		URL:         svc.URL,
		LastChecked: time.Now(),
	}

	//create the req
	req, err := http.NewRequestWithContext(ctx, svc.Method, svc.URL, nil)
	if err != nil {
		result.Status = "down"
		result.Message = fmt.Sprintf("Failed to create request: %v", err)
		m.updateResult(result)
		return
	}

	startTime := time.Now()

	resp, err := m.client.Do(req)

	responseTime := time.Since(startTime)
	result.ResponseTime = responseTime.Milliseconds()

	if err != nil {
		result.Status = "down"
		result.Message = fmt.Sprintf("Request failed: %v", err)
		m.updateResult(result)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == svc.ExpectedStatus {
		result.Status = "up"
		result.Message = "Service is healthy"
	} else {
		result.Status = "down"
		result.Message = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
	}

	m.updateResult(result)
}

func (m *Monitor) updateResult(result *Result) {
	m.resultMutex.Lock()
	defer m.resultMutex.Unlock()
	m.results[result.ServiceName] = result
}

// return a copy of all current check results
func (m *Monitor) GetResults() map[string]Result {
	m.resultMutex.RLock()
	defer m.resultMutex.RUnlock()

	//return a copy
	resultsCopy := make(map[string]Result)
	for k, v := range m.results {
		resultsCopy[k] = *v
	}

	return resultsCopy
}

// returns a slice for easier JSON marshaling

func (m *Monitor) GetResultsSlice() []Result {
	m.resultMutex.RLock()
	defer m.resultMutex.RUnlock()

	results := make([]Result, 0, len(m.results))
	for _, v := range m.results {
		results = append(results, *v)
	}

	return results
}
