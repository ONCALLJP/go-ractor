package task

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	configPath string
	tasks      map[string]Task
}

func NewManager(configPath string) *Manager {
	if configPath == "" {
		userHome := getRealUserHome()
		if userHome != "" {
			configPath = filepath.Join(userHome, ".goractor", "tasks.yaml")
		}
	}

	return &Manager{
		configPath: configPath,
		tasks:      make(map[string]Task),
	}
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No config file yet, start empty
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	return yaml.Unmarshal(data, &m.tasks)
}

func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(m.configPath, data, 0644)
}

func (m *Manager) List() []Task {
	tasks := make([]Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (m *Manager) Add(task Task) error {
	if _, exists := m.tasks[task.Name]; exists {
		return fmt.Errorf("task %s already exists", task.Name)
	}
	m.tasks[task.Name] = task
	return m.Save()
}

func (m *Manager) Remove(name string) error {
	if _, exists := m.tasks[name]; !exists {
		return fmt.Errorf("task %s does not exist", name)
	}
	delete(m.tasks, name)
	return m.Save()
}

func (m *Manager) Get(name string) (Task, error) {
	task, exists := m.tasks[name]
	if !exists {
		return Task{}, fmt.Errorf("task '%s' not found in %s", name, m.configPath)
	}
	return task, nil
}

func (m *Manager) Update(task Task) error {
	if _, exists := m.tasks[task.Name]; !exists {
		return fmt.Errorf("task %s does not exist", task.Name)
	}
	m.tasks[task.Name] = task
	return m.Save()
}

func getRealUserHome() string {

	// If not running with sudo, use regular HOME
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// Last resort: get current user's home
	if user, err := user.Current(); err == nil {
		return user.HomeDir
	}

	return ""
}
