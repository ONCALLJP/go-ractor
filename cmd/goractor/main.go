package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ONCALLJP/goractor/internal/config"
	"github.com/ONCALLJP/goractor/internal/executor"
	"github.com/ONCALLJP/goractor/internal/prompt"
	"github.com/ONCALLJP/goractor/internal/scheduler"
	"github.com/ONCALLJP/goractor/internal/task"
	"gopkg.in/yaml.v3"
)

var (
	taskManager   *task.Manager
	configManager *config.Manager
	configPrompt  *config.Prompt
	mainScheduler *scheduler.Scheduler
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting home directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize config manager
	configPath := filepath.Join(homeDir, ".goractor", "config.yaml")
	configManager = config.NewManager(configPath)
	if err := configManager.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	configPrompt = config.NewPrompt()

	// Initialize task manager
	taskPath := filepath.Join(homeDir, ".goractor", "tasks.yaml")
	taskManager = task.NewManager(taskPath)
	if err := taskManager.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "error loading tasks: %v\n", err)
		os.Exit(1)
	}

	// Initialize executor and scheduler
	exec := executor.NewExecutor(configManager.GetDatabases())
	mainScheduler = scheduler.NewScheduler(taskManager, exec)
}

func main() {
	if err := rootCommand(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func rootCommand() error {
	// If no arguments provided, show help
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: goractor [task|config] [command]")
	}

	switch os.Args[1] {
	case "task":
		return handleTaskCommand(os.Args[2:])
	case "config":
		return handleConfigCommand(os.Args[2:])
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func handleTaskCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: goractor task [list|add|remove|edit]")
	}

	switch args[0] {
	case "list":
		return listTasks()
	case "show":
		if len(args) != 2 {
			return fmt.Errorf("usage: goractor task show [task-name]")
		}
		return showTask(args[1])
	case "add":
		return addTask()
	case "remove":
		if len(args) != 2 {
			return fmt.Errorf("usage: goractor task remove [task-name]")
		}
		return removeTask(args[1])
	case "edit":
		if len(args) != 2 {
			return fmt.Errorf("usage: goractor task edit [task-name]")
		}
		return editTask(args[1])
	case "start":
		if len(args) == 2 {
			return startTask(args[1])
		}
		return startAllTasks()
	case "stop":
		if len(args) == 2 {
			return stopTask(args[1])
		}
		return stopAllTasks()
	case "test":
		if len(args) != 2 {
			return fmt.Errorf("usage: goractor task test [task-name]")
		}
		return testTask(args[1])
	default:
		return fmt.Errorf("unknown task command: %s", args[0])
	}
}

func handleConfigCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: goractor config [list|show|add|remove|edit] [database-name]")
	}

	switch args[0] {
	case "list":
		dbs := configManager.ListDatabases()
		if len(dbs) == 0 {
			fmt.Println("No databases configured")
			return nil
		}
		fmt.Println("Configured Databases:")
		for name := range configManager.GetDatabases() {
			fmt.Printf("- %s\n", name)
		}
		return nil

	case "show":
		if len(args) != 2 {
			return fmt.Errorf("usage: goractor config show [database-name]")
		}
		db, exists := configManager.GetDatabase(args[1])
		if !exists {
			return fmt.Errorf("database %s not found", args[1])
		}
		dbCopy := *db
		dbCopy.Password = "********"
		data, _ := yaml.Marshal(dbCopy)
		fmt.Printf("Database: %s\n%s", args[1], string(data))
		return nil

	case "add":
		name, db, err := configPrompt.PromptDatabase(nil)
		if err != nil {
			return err
		}
		return configManager.AddDatabase(name, db)

	case "edit":
		if len(args) != 2 {
			return fmt.Errorf("usage: goractor config edit [database-name]")
		}
		db, exists := configManager.GetDatabase(args[1])
		if !exists {
			return fmt.Errorf("database %s not found", args[1])
		}
		name, updatedDB, err := configPrompt.PromptDatabase(db)
		if err != nil {
			return err
		}
		return configManager.UpdateDatabase(name, updatedDB)

	case "remove":
		if len(args) != 2 {
			return fmt.Errorf("usage: goractor config remove [database-name]")
		}
		return configManager.RemoveDatabase(args[1])

	default:
		return fmt.Errorf("unknown config command: %s", args[0])
	}
}

// Placeholder functions - we'll implement these next
func listTasks() error {
	tasks := taskManager.List()
	if len(tasks) == 0 {
		fmt.Println("No tasks configured")
		return nil
	}

	fmt.Println("Tasks:")
	fmt.Println("---")
	for _, t := range tasks {
		fmt.Println(t.String())
		fmt.Println("---")
	}
	return nil
}

func showTask(name string) error {
	task, err := taskManager.Get(name)
	if err != nil {
		return err
	}

	fmt.Println(task.String())
	return nil
}

func addTask() error {
	taskPrompt := prompt.NewTaskPrompt()
	newTask, err := taskPrompt.CreateTask()
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	if err := taskManager.Add(*newTask); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	fmt.Printf("Successfully created task: %s\n", newTask.Name)
	return nil
}

func editTask(name string) error {
	currentTask, err := taskManager.Get(name)
	if err != nil {
		return err
	}

	taskPrompt := prompt.NewTaskPrompt()
	if err := taskPrompt.EditTask(&currentTask); err != nil {
		return fmt.Errorf("failed to edit task: %w", err)
	}

	if err := taskManager.Update(currentTask); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	fmt.Printf("Successfully updated task: %s\n", name)
	return nil
}

func removeTask(name string) error {
	return taskManager.Remove(name)
}

func startTask(name string) error {
	task, err := taskManager.Get(name)
	if err != nil {
		return err
	}
	return mainScheduler.StartTask(&task)
}

func startAllTasks() error {
	return mainScheduler.Start()
}

func stopTask(name string) error {
	mainScheduler.StopTask(name)
	return nil
}

func stopAllTasks() error {
	mainScheduler.Stop()
	return nil
}

func testTask(name string) error {
	task, err := taskManager.Get(name)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Testing task '%s'...\n\n", name)
	if err := mainScheduler.GetExecutor().Test(ctx, &task); err != nil {
		fmt.Printf("\n❌ Test failed: %v\n", err)
		return err
	}
	return nil
}

func showConfig() error {
	fmt.Println("Showing config... (to be implemented)")
	return nil
}

func editConfig() error {
	fmt.Println("Editing config... (to be implemented)")
	return nil
}
