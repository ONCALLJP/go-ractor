package config

import (
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
)

type Prompt struct{}

func NewPrompt() *Prompt {
	return &Prompt{}
}

func (p *Prompt) PromptDatabase(defaultConfig *DBConfig) (string, *DBConfig, error) {
	// Database name
	namePrompt := promptui.Prompt{
		Label:     "Database Name",
		Validate:  validateNotEmpty,
		AllowEdit: true,
	}
	name, err := namePrompt.Run()
	if err != nil {
		return "", nil, fmt.Errorf("database name prompt failed: %w", err)
	}

	defaultHost := "localhost"
	if defaultConfig != nil {
		defaultHost = defaultConfig.Host
	}
	hostPrompt := promptui.Prompt{
		Label:     "Host",
		Validate:  validateNotEmpty,
		AllowEdit: true,
		Default:   defaultHost,
	}
	host, err := hostPrompt.Run()
	if err != nil {
		return "", nil, fmt.Errorf("host prompt failed: %w", err)
	}

	defaultPort := "5432"
	if defaultConfig != nil {
		defaultPort = fmt.Sprintf("%d", defaultConfig.Port)
	}
	portPrompt := promptui.Prompt{
		Label:     "Port",
		Validate:  validatePort,
		AllowEdit: true,
		Default:   defaultPort,
	}
	portStr, err := portPrompt.Run()
	if err != nil {
		return "", nil, fmt.Errorf("port prompt failed: %w", err)
	}
	port, _ := strconv.Atoi(portStr)

	defaultUser := "postgres"
	if defaultConfig != nil {
		defaultUser = defaultConfig.User
	}
	userPrompt := promptui.Prompt{
		Label:     "User",
		Validate:  validateNotEmpty,
		AllowEdit: true,
		Default:   defaultUser,
	}
	user, err := userPrompt.Run()
	if err != nil {
		return "", nil, fmt.Errorf("user prompt failed: %w", err)
	}

	var defaultPass string
	if defaultConfig != nil {
		defaultPass = defaultConfig.Password
	}
	passPrompt := promptui.Prompt{
		Label:     "Password",
		Mask:      '*',
		Validate:  validateNotEmpty,
		AllowEdit: true,
		Default:   defaultPass,
	}
	pass, err := passPrompt.Run()
	if err != nil {
		return "", nil, fmt.Errorf("password prompt failed: %w", err)
	}

	defaultDBName := ""
	if defaultConfig != nil {
		defaultDBName = defaultConfig.DBName
	}
	dbNamePrompt := promptui.Prompt{
		Label:     "Database Name",
		Validate:  validateNotEmpty,
		AllowEdit: true,
		Default:   defaultDBName,
	}
	dbName, err := dbNamePrompt.Run()
	if err != nil {
		return "", nil, fmt.Errorf("database name prompt failed: %w", err)
	}

	return name, &DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		DBName:   dbName,
	}, nil
}

// Validation functions
func validateNotEmpty(input string) error {
	if input == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func validatePort(input string) error {
	port, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}
