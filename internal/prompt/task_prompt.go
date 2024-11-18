package prompt

import (
	"fmt"
	"strings"
	"time"

	"github.com/ONCALLJP/goractor/internal/task"
	"github.com/manifoldco/promptui"
)

type TaskPrompt struct{}

func NewTaskPrompt() *TaskPrompt {
	return &TaskPrompt{}
}

func (p *TaskPrompt) promptBasicInfo(defaultValues *task.Task) (*task.Task, error) {
	// Name
	defaultName := ""
	if defaultValues != nil {
		defaultName = defaultValues.Name
	}
	var name string
	namePrompt := promptui.Prompt{
		Label:    "Task Name",
		Default:  defaultName,
		Validate: validateNotEmpty,
	}
	var err error
	name, err = namePrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("task name prompt failed: %w", err)
	}

	// Database
	defaultDB := ""
	if defaultValues != nil {
		defaultDB = defaultValues.Database
	}
	dbPrompt := promptui.Prompt{
		Label:    "Database Name",
		Validate: validateNotEmpty,
		Default:  defaultDB,
	}
	database, err := dbPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("database prompt failed: %w", err)
	}

	// Schedule
	schedulePrompt := promptui.Select{
		Label: "Schedule Type",
		Items: []string{"hourly", "daily", "custom"},
	}
	_, scheduleType, err := schedulePrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("schedule type prompt failed: %w", err)
	}

	var schedule string
	switch scheduleType {
	case "hourly":
		schedule = "every 1h"
	case "daily":
		defaultTime := "15:00"
		if defaultValues != nil && strings.HasPrefix(defaultValues.Schedule, "daily ") {
			defaultTime = strings.TrimPrefix(defaultValues.Schedule, "daily ")
		}
		timePrompt := promptui.Prompt{
			Label:    "Time (HH:MM)",
			Validate: validateTime,
			Default:  defaultTime,
		}
		time, err := timePrompt.Run()
		if err != nil {
			return nil, fmt.Errorf("time prompt failed: %w", err)
		}
		schedule = fmt.Sprintf("daily %s", time)
	case "custom":
		defaultSchedule := ""
		if defaultValues != nil {
			defaultSchedule = defaultValues.Schedule
		}
		customPrompt := promptui.Prompt{
			Label:    "Custom Schedule (e.g., every 30m, every 2h)",
			Validate: validateSchedule,
			Default:  defaultSchedule,
		}
		schedule, err = customPrompt.Run()
		if err != nil {
			return nil, fmt.Errorf("custom schedule prompt failed: %w", err)
		}
	}

	// Query
	query, err := p.promptQuery(defaultValues)
	if err != nil {
		return nil, err
	}

	// Destination
	destination, err := p.promptDestination(defaultValues)
	if err != nil {
		return nil, err
	}

	// Output Format
	formatPrompt := promptui.Select{
		Label: "Output Format",
		Items: []string{"json", "csv"},
	}
	_, outputFormat, err := formatPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("output format prompt failed: %w", err)
	}

	return &task.Task{
		Name:         name,
		Database:     database,
		Schedule:     schedule,
		Query:        *query,
		Destination:  *destination,
		OutputFormat: outputFormat,
	}, nil
}

// Helper function to safely get default string value
func getDefaultString(defaultValues *task.Task, value string) string {
	if defaultValues == nil {
		return ""
	}
	return value
}

func (p *TaskPrompt) promptQuery(defaultValues *task.Task) (*task.Query, error) {
	var defaultName, defaultSQL string
	if defaultValues != nil {
		defaultName = defaultValues.Query.Name
		defaultSQL = defaultValues.Query.SQL
	}

	// Query Name
	namePrompt := promptui.Prompt{
		Label:    "Query Name",
		Validate: validateNotEmpty,
		Default:  defaultName,
	}
	name, err := namePrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("query name prompt failed: %w", err)
	}

	// SQL
	sqlPrompt := promptui.Prompt{
		Label:    "SQL (SELECT only)",
		Validate: validateSelectSQL,
		Default:  defaultSQL,
	}
	sql, err := sqlPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("SQL prompt failed: %w", err)
	}

	return &task.Query{
		Name: name,
		SQL:  sql,
	}, nil
}

func (p *TaskPrompt) promptDestination(defaultValues *task.Task) (*task.Destination, error) {
	var destination task.Destination

	// Always ask for URL
	defaultURL := ""
	if defaultValues != nil {
		defaultURL = defaultValues.Destination.WebhookURL
	}

	// URL prompt
	urlPrompt := promptui.Prompt{
		Label:    "Destination URL",
		Validate: validateURL,
		Default:  defaultURL,
	}
	url, err := urlPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("URL prompt failed: %w", err)
	}
	destination.WebhookURL = url

	// Channel name if URL contains "slack"
	if strings.Contains(strings.ToLower(url), "slack") {
		defaultChannel := ""
		if defaultValues != nil {
			defaultChannel = defaultValues.Destination.Channel
		}
		channelPrompt := promptui.Prompt{
			Label:    "Slack Channel",
			Validate: validateNotEmpty,
			Default:  defaultChannel,
		}
		channel, err := channelPrompt.Run()
		if err != nil {
			return nil, fmt.Errorf("channel prompt failed: %w", err)
		}
		destination.Channel = channel
	}

	// Auth Type
	authTypePrompt := promptui.Select{
		Label: "Authentication Type",
		Items: []string{"bearer", "basic", "api_key", "none"},
	}
	_, authType, err := authTypePrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("auth type prompt failed: %w", err)
	}

	if authType != "none" {
		defaultTokenValue := ""
		if defaultValues != nil {
			defaultTokenValue = defaultValues.Destination.Token.Value
		}

		// Token Value
		tokenPrompt := promptui.Prompt{
			Label:    fmt.Sprintf("%s Token", authType),
			Mask:     '*', // Hide token input
			Validate: validateNotEmpty,
			Default:  defaultTokenValue,
		}
		tokenValue, err := tokenPrompt.Run()
		if err != nil {
			return nil, fmt.Errorf("token prompt failed: %w", err)
		}

		destination.Token = task.TokenConfig{
			Type:  authType,
			Value: tokenValue,
		}
	} else {
		destination.Token = task.TokenConfig{
			Type:  "none",
			Value: "",
		}
	}

	destination.Type = "api"
	return &destination, nil
}

func (p *TaskPrompt) CreateTask() (*task.Task, error) {
	return p.promptBasicInfo(nil)
}

func (p *TaskPrompt) EditTask(t *task.Task) error {
	updatedTask, err := p.promptBasicInfo(t)
	if err != nil {
		return err
	}

	// Copy back all values except the name
	t.Database = updatedTask.Database
	t.Schedule = updatedTask.Schedule
	t.Query = updatedTask.Query
	t.Destination = updatedTask.Destination
	t.OutputFormat = updatedTask.OutputFormat

	return nil
}

// Validation functions
func validateNotEmpty(input string) error {
	if input == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func validateSelectSQL(input string) error {
	// Simple validation - just check if it starts with SELECT
	if len(input) < 6 || input[0:6] != "SELECT" {
		return fmt.Errorf("query must start with SELECT")
	}
	return nil
}

func validateURL(input string) error {
	if !strings.HasPrefix(input, "https://") {
		return fmt.Errorf("URL must start with https://")
	}
	return nil
}

func validateTime(input string) error {
	_, err := time.Parse("15:04", input)
	return err
}

func validateSchedule(input string) error {
	// Basic validation for now
	if !strings.HasPrefix(input, "every ") && !strings.HasPrefix(input, "daily ") {
		return fmt.Errorf("schedule must start with 'every ' or 'daily '")
	}
	return nil
}
