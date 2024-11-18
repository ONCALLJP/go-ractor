package config

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type Config struct {
	Databases map[string]*DBConfig `yaml:"databases"`
}

// func LoadConfig() (*Config, error) {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting home directory: %w", err)
// 	}

// 	configPath := filepath.Join(homeDir, ".goractor", "config.yaml")
// 	data, err := os.ReadFile(configPath)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			// Return empty config if file doesn't exist
// 			return &Config{
// 				Databases: make(map[string]*executor.DBConfig),
// 			}, nil
// 		}
// 		return nil, fmt.Errorf("error reading config: %w", err)
// 	}

// 	var config Config
// 	if err := yaml.Unmarshal(data, &config); err != nil {
// 		return nil, fmt.Errorf("error parsing config: %w", err)
// 	}

// 	return &config, nil
// }

// func SaveConfig(config *Config) error {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		return fmt.Errorf("error getting home directory: %w", err)
// 	}

// 	configDir := filepath.Join(homeDir, ".goractor")
// 	if err := os.MkdirAll(configDir, 0755); err != nil {
// 		return fmt.Errorf("error creating config directory: %w", err)
// 	}

// 	configPath := filepath.Join(configDir, "config.yaml")
// 	data, err := yaml.Marshal(config)
// 	if err != nil {
// 		return fmt.Errorf("error marshaling config: %w", err)
// 	}

// 	if err := os.WriteFile(configPath, data, 0644); err != nil {
// 		return fmt.Errorf("error writing config: %w", err)
// 	}

// 	return nil
// }
