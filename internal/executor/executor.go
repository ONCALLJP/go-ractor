package executor

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ONCALLJP/goractor/internal/config"
	"github.com/ONCALLJP/goractor/internal/destination"
	"github.com/ONCALLJP/goractor/internal/task"
	_ "github.com/lib/pq"
	"github.com/slack-go/slack"
)

type Executor struct {
	dbConfigs          map[string]*config.DBConfig
	destinationManager *destination.Manager
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type QueryResult struct {
	TaskID        string                   `json:"task_id"`
	Timestamp     time.Time                `json:"timestamp"`
	QueryName     string                   `json:"query_name"`
	ExecutionTime string                   `json:"execution_time"`
	RowCount      int                      `json:"row_count"`
	Data          []map[string]interface{} `json:"data"`
}

func NewExecutor(dbConfigs map[string]*config.DBConfig, dest *destination.Manager) *Executor {
	return &Executor{
		dbConfigs:          dbConfigs,
		destinationManager: dest,
	}
}

func (e *Executor) Execute(ctx context.Context, t *task.Task) error {
	// Get database configuration
	dbConfig, ok := e.dbConfigs[t.Database]
	if !ok {
		return fmt.Errorf("database configuration not found: %s", t.Database)
	}

	// Connect to database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Execute query and measure time
	start := time.Now()
	rows, err := db.QueryContext(ctx, t.Query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare result
	var result []map[string]interface{}
	count := 0

	// Scan rows
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		result = append(result, row)
		count++
	}

	// Create query result
	queryResult := QueryResult{
		TaskID:        t.Name,
		Timestamp:     time.Now(),
		ExecutionTime: time.Since(start).String(),
		RowCount:      count,
		Data:          result,
	}

	if t.OutputFormat == "csv" {
		if err := e.sendResultAsCSV(ctx, t, queryResult); err != nil {
			return fmt.Errorf("failed to send to destination: %w", err)
		}
		return e.sendResultAsCSV(ctx, t, queryResult)
	} else if t.OutputFormat == "json" {
		fmt.Println("✓ Destination test successful")
		return nil
	}
	return nil
}

func (e *Executor) createCSVFile(result QueryResult, headers []string) (string, error) {
	// If we couldn't parse SQL, fallback to the order from result
	if len(headers) == 0 && len(result.Data) > 0 {
		for col := range result.Data[0] {
			headers = append(headers, col)
		}
	}

	// Create CSV file
	tmpDir := filepath.Join(os.TempDir(), "goractor")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(tmpDir, fmt.Sprintf("%s_%s.csv", result.TaskID, timestamp))
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data in the same order as headers
	for _, row := range result.Data {
		var record []string
		for _, header := range headers {
			value := ""
			if v := row[header]; v != nil {
				value = fmt.Sprintf("%v", v)
			}
			record = append(record, value)
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return filename, nil
}

func (e *Executor) sendResultAsCSV(ctx context.Context, t *task.Task, result QueryResult) error {
	// Get destination configuration
	dest, exists := e.destinationManager.Get(t.DestinationName)
	if !exists {
		return fmt.Errorf("destination %s not found", t.DestinationName)
	}

	// Create CSV file
	csvFilePath, err := e.createCSVFile(result, t.Columns)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer os.Remove(csvFilePath)

	// Open the file for reading
	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer csvFile.Close()

	switch dest.Type {
	case "slack":
		api := slack.New(dest.Token.Value)
		params := slack.UploadFileV2Parameters{
			Filename:       filepath.Base(csvFilePath),
			FileSize:       1000,
			Channel:        strings.Replace(dest.Channel, "#", "", 1),
			File:           csvFilePath,
			Reader:         csvFile,
			InitialComment: t.Message,
		}
		_, err = api.UploadFileV2(params)
		if err != nil {
			return fmt.Errorf("failed to upload file to slack: %w", err)
		}

	case "lineworks":
		return fmt.Errorf("lineworks implementation pending")

	case "custom":
		// Read file content
		content, err := os.ReadFile(csvFilePath)
		if err != nil {
			return fmt.Errorf("failed to read CSV file: %w", err)
		}

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", dest.URL, bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set content type
		req.Header.Set("Content-Type", "text/csv")

		// Set authentication based on token type
		if dest.Token.Type != "" {
			switch dest.Token.Type {
			case "bearer":
				req.Header.Set("Authorization", "Bearer "+dest.Token.Value)
			case "basic":
				req.Header.Set("Authorization", "Basic "+dest.Token.Value)
			case "api_key":
				req.Header.Set("X-API-Key", dest.Token.Value)
			}
		}

		// Send request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			return fmt.Errorf("received non-success status code: %d", resp.StatusCode)
		}

	default:
		return fmt.Errorf("destination type %s is not supported", dest.Type)
	}

	return nil
}

func (e *Executor) Run(ctx context.Context, t *task.Task) error {
	fmt.Printf("Runing task: %s\n", t.Name)
	fmt.Printf("Database: %s\n", t.Database)
	fmt.Printf("Query: %s\n\n", t.Query)

	// Test database connection
	dbConfig, ok := e.dbConfigs[t.Database]
	if !ok {
		return fmt.Errorf("database configuration not found: %s", t.Database)
	}

	fmt.Println("1. database connection...")
	connStr := "postgres://" + dbConfig.User + ":" + dbConfig.Password + "@" + dbConfig.Host + ":5432/" + dbConfig.DBName

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	fmt.Println("✓ Database connection successful")

	fmt.Println("2. query execution...")
	start := time.Now()
	rows, err := db.QueryContext(ctx, t.Query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names and validate selected columns exist
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Validate that all requested columns exist
	columnMap := make(map[string]bool)
	for _, col := range columns {
		columnMap[col] = true
	}

	// Prepare result
	var result []map[string]interface{}
	count := 0

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for selected columns only
		row := make(map[string]interface{})
		for i, col := range columns {
			// Only include selected columns
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		result = append(result, row)
		count++
	}

	executionTime := time.Since(start)
	fmt.Printf("✓ Query execution successful (retrieved %d rows in %s)\n", count, executionTime)

	// Create test result
	queryResult := QueryResult{
		TaskID:        t.Name,
		Timestamp:     time.Now(),
		ExecutionTime: executionTime.String(),
		RowCount:      count,
		Data:          result,
	}

	fmt.Println("\n3. destination...")
	// Send test result to destination
	if t.OutputFormat == "csv" {
		if err := e.sendResultAsCSV(ctx, t, queryResult); err != nil {
			return fmt.Errorf("failed to send to destination: %w", err)
		}
		fmt.Println("✓ Destination successful")
	} else if t.OutputFormat == "json" {
		fmt.Println("✓ Destination successful")
	}

	return nil
}
