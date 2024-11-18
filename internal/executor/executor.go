package executor

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ONCALLJP/goractor/internal/config"
	"github.com/ONCALLJP/goractor/internal/task"
	_ "github.com/lib/pq"
)

type Executor struct {
	dbConfigs map[string]*config.DBConfig
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

func NewExecutor(dbConfigs map[string]*config.DBConfig) *Executor {
	return &Executor{
		dbConfigs: dbConfigs,
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
	rows, err := db.QueryContext(ctx, t.Query.SQL)
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
		QueryName:     t.Query.Name,
		ExecutionTime: time.Since(start).String(),
		RowCount:      count,
		Data:          result,
	}

	// Send result to destination
	return e.sendResult(ctx, t, queryResult)
}

func (e *Executor) sendResult(ctx context.Context, t *task.Task, result QueryResult) error {
	// Convert result to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", t.Destination.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers based on token type
	switch t.Destination.Token.Type {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+t.Destination.Token.Value)
	case "basic":
		req.Header.Set("Authorization", "Basic "+t.Destination.Token.Value)
	case "api_key":
		req.Header.Set("X-API-Key", t.Destination.Token.Value)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response: %d", resp.StatusCode)
	}

	return nil
}

func (e *Executor) Test(ctx context.Context, t *task.Task) error {
	fmt.Printf("Testing task: %s\n", t.Name)
	fmt.Printf("Database: %s\n", t.Database)
	fmt.Printf("Query: %s\n\n", t.Query.SQL)

	// Test database connection
	dbConfig, ok := e.dbConfigs[t.Database]
	if !ok {
		return fmt.Errorf("database configuration not found: %s", t.Database)
	}

	fmt.Println("1. Testing database connection...")
	connStr := buildDSN(dbConfig)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	fmt.Println("✓ Database connection successful")

	// Test query execution
	fmt.Println("\n2. Testing query execution...")
	start := time.Now()
	rows, err := db.QueryContext(ctx, t.Query.SQL)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Count rows
	rowCount := 0
	for rows.Next() {
		rowCount++
		if rowCount == 1 {
			// Get column names for first row
			columns, err := rows.Columns()
			if err != nil {
				return fmt.Errorf("failed to get columns: %w", err)
			}
			fmt.Printf("✓ Query returns columns: %v\n", columns)
		}
		if rowCount > 5 {
			break // Don't process all rows for test
		}
	}
	executionTime := time.Since(start)
	fmt.Printf("✓ Query execution successful (first %d rows in %s)\n", rowCount, executionTime)

	// Test destination
	fmt.Println("\n3. Testing destination...")
	if t.Destination.Channel != "" {
		fmt.Printf("Target Slack channel: %s\n", t.Destination.Channel)
	}

	// Prepare test message
	testResult := QueryResult{
		TaskID:        t.Name,
		Timestamp:     time.Now(),
		QueryName:     t.Query.Name,
		ExecutionTime: executionTime.String(),
		RowCount:      rowCount,
		Data:          []map[string]interface{}{{"test": "data"}},
	}

	fmt.Println("Sending test message to destination...")
	if err := e.sendResult(ctx, t, testResult); err != nil {
		return fmt.Errorf("failed to send to destination: %w", err)
	}
	fmt.Println("✓ Destination test successful")

	fmt.Println("\n✓ All tests passed successfully!")
	return nil
}

func buildDSN(config *config.DBConfig) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Tokyo",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
	)
}
