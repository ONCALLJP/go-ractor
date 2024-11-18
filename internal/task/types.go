package task

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Destination struct {
	Type       string      `yaml:"type"` // e.g., "slack"
	WebhookURL string      `yaml:"webhook_url"`
	Token      TokenConfig `yaml:"token"`
	Channel    string      `yaml:"channel"`
}

type TokenConfig struct {
	Type  string `yaml:"type"`  // e.g., "bearer", "basic", "api_key"
	Value string `yaml:"value"` // The actual token
}

type Query struct {
	Name string `yaml:"name"`
	SQL  string `yaml:"sql"`
}

type Task struct {
	Name         string      `yaml:"name"`
	Database     string      `yaml:"database"` // reference to database config
	Schedule     string      `yaml:"schedule"` // e.g., "every 1h", "daily 15:00"
	Query        Query       `yaml:"query"`
	Destination  Destination `yaml:"destination"`
	OutputFormat string      `yaml:"output_format"` // "json" or "csv"
}

func (t Task) String() string {
	data, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Sprintf("error marshaling task: %v", err)
	}
	return string(data)
}
