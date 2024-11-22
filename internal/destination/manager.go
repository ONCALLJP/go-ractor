package destination

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	configPath   string
	destinations map[string]Destination
}

func NewManager(configPath string) *Manager {
	return &Manager{
		configPath:   configPath,
		destinations: make(map[string]Destination),
	}
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error reading destinations: %w", err)
	}

	if err := yaml.Unmarshal(data, &m.destinations); err != nil {
		return fmt.Errorf("error parsing destinations: %w", err)
	}

	return nil
}

func (m *Manager) Save() error {
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	data, err := yaml.Marshal(m.destinations)
	if err != nil {
		return fmt.Errorf("error marshaling destinations: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing destinations: %w", err)
	}

	return nil
}

func (m *Manager) List() []string {
	var names []string
	for name := range m.destinations {
		names = append(names, name)
	}
	return names
}

func (m *Manager) Get(name string) (Destination, bool) {
	dest, exists := m.destinations[name]
	return dest, exists
}

func (m *Manager) Add(name string, dest Destination) error {
	if _, exists := m.destinations[name]; exists {
		return fmt.Errorf("destination %s already exists", name)
	}

	m.destinations[name] = dest
	return m.Save()
}

func (m *Manager) Update(name string, dest Destination) error {
	if _, exists := m.destinations[name]; !exists {
		return fmt.Errorf("destination %s not found", name)
	}

	m.destinations[name] = dest
	return m.Save()
}

func (m *Manager) Remove(name string) error {
	if _, exists := m.destinations[name]; !exists {
		return fmt.Errorf("destination %s not found", name)
	}

	delete(m.destinations, name)
	return m.Save()
}
