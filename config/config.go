package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Services  []ServiceConfig `json:"services"`
	Dashboard DashboardConfig `json:"dashboard"`
}

type ServiceConfig struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	Method         string `json:"method"`
	Interval       int    `json:"interval"`
	Timeout        int    `json:"timeout"`
	ExpectedStatus int    `json:"expectedStatus"`
}

type DashboardConfig struct {
	Title           string `json:"title"`
	RefreshInterval int    `json:"refreshInterval"`
}

func Load(filePath string) (*Config, error) {

	//read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the config file: %w", err)
	}

	//parse the JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse the config file: %w", err)
	}

	if err := validate(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate each sevice
func validate(config *Config) error {

	//check if the file has at least 1 service
	if len(config.Services) == 0 {
		return fmt.Errorf("no services defined in the config")
	}

	for i, service := range config.Services {

		//check if the name is empty
		if service.Name == "" {
			return fmt.Errorf("service #%d has no name", i+1)
		}

		// validate the url
		if _, err := url.ParseRequestURI(service.URL); err != nil {
			return fmt.Errorf("service '%s' has an invalid URL: %w", service.Name, err)
		}

		//validate the method
		if service.Method == "" {
			return fmt.Errorf("service '%s' has no HTTP method defined", service.Name)
		}

		//Interval must be positive
		if service.Interval <= 0 {
			return fmt.Errorf("service '%s' has an invalid interval: must be > 0", service.Name)
		}

		//Timeout must be positive
		if service.Timeout <= 0 {
			return fmt.Errorf("service '%s' has an invalid timeout: must be > 0", service.Name)
		}

		if config.Dashboard.RefreshInterval <= 0 {
			return fmt.Errorf("dashboard has an invalid refresh interval: must be > 0")
		}

	}

	return nil
}

func (s *ServiceConfig) GetCheckInterval() time.Duration {
	return time.Duration(s.Interval) * time.Second
}

func (s *ServiceConfig) GetTimeout() time.Duration {
	return time.Duration(s.Timeout) * time.Second
}

func (d *DashboardConfig) GetRefreshInterval() time.Duration {
	return time.Duration(d.RefreshInterval) * time.Second
}
