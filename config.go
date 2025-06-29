package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Database Database `mapstructure:"database"`
	Server   Server   `mapstructure:"server"`
	Monitor  Monitor  `mapstructure:"monitor"`
	Claude   Claude   `mapstructure:"claude"`
}

// Database configuration
type Database struct {
	Path string `mapstructure:"path"`
}

// Server configuration
type Server struct {
	Address string `mapstructure:"address"`
}

// Monitor configuration
type Monitor struct {
	Server string `mapstructure:"server"`
}

// Claude configuration
type Claude struct {
	Plan string `mapstructure:"plan"` // enum: unset, pro, max, max20
}

// LoadConfig loads configuration from files
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("database.path", "~/.ccmon/ccmon.db")
	v.SetDefault("server.address", "127.0.0.1:4317")
	v.SetDefault("monitor.server", "127.0.0.1:4317")
	v.SetDefault("claude.plan", "unset")

	// Set config name (without extension)
	v.SetConfigName("config")

	// Add config paths (first found wins)
	v.AddConfigPath(".") // Current directory (highest priority)
	if homeDir, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(homeDir, ".ccmon")) // User config directory
	}

	// Read config file (if exists)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		// No config file found is OK - use defaults
	}

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Expand home directory in database path
	config.Database.Path = expandPath(config.Database.Path)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Claude plan
	validPlans := map[string]bool{
		"unset": true,
		"pro":   true,
		"max":   true,
		"max20": true,
	}

	if !validPlans[c.Claude.Plan] {
		return fmt.Errorf("invalid claude plan: %s (must be one of: unset, pro, max, max20)", c.Claude.Plan)
	}

	return nil
}
