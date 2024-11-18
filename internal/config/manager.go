package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	configPath string
	config     *Config
}

func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
		config: &Config{
			Databases: make(map[string]*DBConfig),
		},
	}
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return nil
		}
		return fmt.Errorf("error reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	return nil
}

func (m *Manager) Save() error {
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	return nil
}

func (m *Manager) ListDatabases() []*DBConfig {
	dbs := make([]*DBConfig, 0, len(m.config.Databases))
	for _, db := range m.config.Databases {
		dbs = append(dbs, db)
	}
	return dbs
}

func (m *Manager) GetDatabase(name string) (*DBConfig, bool) {
	db, exists := m.config.Databases[name]
	return db, exists
}

func (m *Manager) GetDatabases() map[string]*DBConfig {
	return m.config.Databases
}

func (m *Manager) AddDatabase(name string, config *DBConfig) error {
	if _, exists := m.config.Databases[name]; exists {
		return fmt.Errorf("database %s already exists", name)
	}

	m.config.Databases[name] = config
	return m.Save()
}

func (m *Manager) UpdateDatabase(name string, config *DBConfig) error {
	if _, exists := m.config.Databases[name]; !exists {
		return fmt.Errorf("database %s not found", name)
	}

	m.config.Databases[name] = config
	return m.Save()
}

func (m *Manager) RemoveDatabase(name string) error {
	if _, exists := m.config.Databases[name]; !exists {
		return fmt.Errorf("database %s not found", name)
	}

	delete(m.config.Databases, name)
	return m.Save()
}
