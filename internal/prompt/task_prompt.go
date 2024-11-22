package prompt

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ONCALLJP/goractor/internal/config"
	"github.com/ONCALLJP/goractor/internal/destination"
	"github.com/ONCALLJP/goractor/internal/task"
	"github.com/manifoldco/promptui"
)

type TaskPrompt struct {
	DestinationManager *destination.Manager
	ConfigManger       *config.Manager
}

func NewTaskPrompt(destinationManager *destination.Manager, configManager *config.Manager) *TaskPrompt {
	return &TaskPrompt{DestinationManager: destinationManager, ConfigManger: configManager}
}

func (p *TaskPrompt) promptBasicInfo(defaultValues *task.Task) (*task.Task, error) {
	// Name
	defaultName := ""
	if defaultValues != nil {
		defaultName = defaultValues.Name
	}
	var name string
	namePrompt := promptui.Prompt{
		Label:     "Task Name",
		Default:   defaultName,
		AllowEdit: true,
		Validate:  validateNotEmpty,
	}
	var err error
	name, err = namePrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("task name prompt failed: %w", err)
	}

	// Database
	dbs := p.ConfigManger.ListDatabases()
	if len(dbs) == 0 {
		return nil, fmt.Errorf("no configration found. Please add a database config first")
	}
	databases := make([]string, 0)
	for _, db := range dbs {
		databases = append(databases, db.DBName)
	}

	dbPrompt := promptui.Select{
		Label: "Database Config Name",
		Items: databases,
	}
	_, database, err := dbPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("database prompt failed: %w", err)
	}

	// Schedule

	schedule, err := p.promptSchedule()
	if err != nil {
		return nil, err
	}

	var timezone string
	zones := getAvailableTimezones()

	// Timezone prompt
	tzPrompt := promptui.Select{
		Label: "Timezone",
		Items: zones,
	}
	_, timezone, err = tzPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("timezone prompt failed: %w", err)
	}

	// Query
	query, err := p.promptQuery(defaultValues)
	if err != nil {
		return nil, err
	}

	// Columns
	columns, err := p.promptColumns(defaultValues)
	if err != nil {
		return nil, err
	}

	// Message
	var defualtMessage string
	if defaultValues != nil {
		defualtMessage = defaultValues.Message
	}
	messagePrompt := promptui.Prompt{
		Label:     "Message",
		Validate:  validateNotEmpty,
		AllowEdit: true,
		Default:   defualtMessage,
	}
	message, err := messagePrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("query name prompt failed: %w", err)
	}

	// Output Format
	formatPrompt := promptui.Select{
		Label: "Output Format",
		Items: []string{"csv"}, // TODO add more formats
	}
	_, outputFormat, err := formatPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("output format prompt failed: %w", err)
	}

	// Destination
	destinations := p.DestinationManager.List()
	if len(destinations) == 0 {
		return nil, fmt.Errorf("no destinations configured. Please add a destination first")
	}

	// Destination selection
	destPrompt := promptui.Select{
		Label: "Select Destination",
		Items: destinations,
	}
	_, destName, err := destPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("destination selection failed: %w", err)
	}

	return &task.Task{
		Name:            name,
		Database:        database,
		Schedule:        schedule,
		Timezone:        timezone,
		Query:           query,
		Columns:         columns,
		Message:         message,
		DestinationName: destName,
		OutputFormat:    outputFormat,
	}, nil
}

func (p *TaskPrompt) promptQuery(defaultValues *task.Task) (string, error) {
	// Create temporary file
	tmpfile, err := ioutil.TempFile("", "goractor-sql-*.sql")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write default SQL if exists
	if defaultValues != nil && defaultValues.Query != "" {
		if _, err := tmpfile.WriteString(defaultValues.Query); err != nil {
			return "", fmt.Errorf("failed to write default SQL: %w", err)
		}
	} else {
		// Write template/instructions
		template := `-- Enter your SQL query here (query must start with WITH, SELECT, EXPLAIN, or be a subquery)
-- Example:
-- SELECT column1, column2
-- FROM table_name
-- WHERE condition
-- GROUP BY column1
-- HAVING condition
-- ORDER BY column1;

`
		if _, err := tmpfile.WriteString(template); err != nil {
			return "", fmt.Errorf("failed to write SQL template: %w", err)
		}
	}
	tmpfile.Close()

	// Get editor from environment or default to nano
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	// Open editor
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Opening %s to edit SQL query...\n", editor)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run editor: %w", err)
	}

	// Read the edited content
	content, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read SQL from file: %w", err)
	}

	sql := string(content)

	// Validate SQL
	if err := validateSQL(sql); err != nil {
		return "", err
	}

	return sql, nil
}

func (p *TaskPrompt) promptColumns(defaultValues *task.Task) ([]string, error) {
	// Extract columns from defaultValues.Columns if available
	var defaultColumns string
	if defaultValues != nil && defaultValues.Columns != nil {
		defaultColumns = strings.Join(defaultValues.Columns, ", ")
	}

	columnPrompt := promptui.Prompt{
		Label:     "Columns (comma separated)",
		Validate:  validateNotEmpty,
		AllowEdit: true,
		Default:   defaultColumns,
	}
	columnsInput, err := columnPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("columns prompt failed: %w", err)
	}

	columns := strings.Split(columnsInput, ",")
	trimmedColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		trimmedColumn := strings.TrimSpace(column)
		if trimmedColumn != "" {
			trimmedColumns = append(trimmedColumns, trimmedColumn)
		}
	}

	return trimmedColumns, nil
}

func validateSQL(sql string) error {
	// Remove comments and merge lines
	lines := strings.Split(sql, "\n")
	var validLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			validLines = append(validLines, line)
		}
	}
	sql = strings.Join(validLines, " ")
	sql = strings.TrimSpace(sql)

	if sql == "" {
		return fmt.Errorf("SQL query cannot be empty")
	}

	// Basic validation - query should eventually have a SELECT
	sqlLower := strings.ToLower(sql)
	if !strings.Contains(sqlLower, "select") {
		return fmt.Errorf("query must contain at least one SELECT statement")
	}

	// Allow common SQL constructs
	validStartPatterns := []string{
		"with",
		"select",
		"(",
		"explain",
		"analyze",
		"explain analyze",
		"explain (analyze)",
		"explain (analyze, buffers)",
	}

	// Check if query starts with any valid pattern
	isValidStart := false
	for _, pattern := range validStartPatterns {
		if strings.HasPrefix(sqlLower, pattern) {
			isValidStart = true
			break
		}
	}

	if !isValidStart {
		return fmt.Errorf("query must start with WITH, SELECT, EXPLAIN, or be a subquery")
	}

	return nil
}
func (p *TaskPrompt) promptSchedule() (string, error) {
	schedulePrompt := promptui.Select{
		Label: "Schedule Type",
		Items: []string{"every_5min", "every_hour", "daily", "weekly", "monthly"},
	}
	_, scheduleType, err := schedulePrompt.Run()
	if err != nil {
		return "", fmt.Errorf("schedule type prompt failed: %w", err)
	}

	switch scheduleType {
	case "every_5min", "every_hour":
		return scheduleType, nil

	case "daily":
		timeStr, err := promptTime()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("daily %s", timeStr), nil

	case "weekly":
		// Select days of week
		daysPrompt := promptui.Select{
			Label: "Days of Week",
			Items: []string{
				"Sunday",
				"Monday",
				"Tuesday",
				"Wednesday",
				"Thursday",
				"Friday",
				"Saturday",
				"Monday-Friday",
				"Saturday,Sunday",
			},
		}
		_, days, err := daysPrompt.Run()
		if err != nil {
			return "", fmt.Errorf("days selection failed: %w", err)
		}

		timeStr, err := promptTime()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("weekly %s %s", days, timeStr), nil

	case "monthly":
		// Day of month prompt
		dayPrompt := promptui.Prompt{
			Label:    "Day of Month (1-31)",
			Validate: validateDayOfMonth,
			Default:  "1",
		}
		day, err := dayPrompt.Run()
		if err != nil {
			return "", fmt.Errorf("day prompt failed: %w", err)
		}

		timeStr, err := promptTime()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("monthly %s %s", day, timeStr), nil
	}

	return "", fmt.Errorf("invalid schedule type")
}

func promptTime() (string, error) {
	timePrompt := promptui.Prompt{
		Label:    "Time (HH:MM)",
		Validate: validateTime,
		Default:  "09:00",
	}
	return timePrompt.Run()
}

func validateDayOfMonth(input string) error {
	day, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if day < 1 || day > 31 {
		return fmt.Errorf("must be between 1 and 31")
	}
	return nil
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
	t.Timezone = updatedTask.Timezone
	t.Query = updatedTask.Query
	t.Columns = updatedTask.Columns
	t.Message = updatedTask.Message
	t.DestinationName = updatedTask.DestinationName
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

func validateTime(input string) error {
	_, err := time.Parse("15:04", input)
	return err
}

func getAvailableTimezones() []string {
	// Common timezones first
	commonZones := []string{
		"UTC",
		"Asia/Tokyo",
		"America/New_York",
		"Europe/London",
		"Asia/Singapore",
		"Australia/Sydney",
		// Add more as needed
	}

	return commonZones
}
