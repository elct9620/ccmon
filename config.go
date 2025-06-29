package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
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
	Server   string `mapstructure:"server"`
	Timezone string `mapstructure:"timezone"`
}

// Claude configuration
type Claude struct {
	Plan      string `mapstructure:"plan"`       // enum: unset, pro, max, max20
	MaxTokens int    `mapstructure:"max_tokens"` // override default token limits
}

// LoadConfig loads configuration from files and command-line flags
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("database.path", "~/.ccmon/ccmon.db")
	v.SetDefault("server.address", "127.0.0.1:4317")
	v.SetDefault("monitor.server", "127.0.0.1:4317")
	v.SetDefault("monitor.timezone", "UTC")
	v.SetDefault("claude.plan", "unset")
	v.SetDefault("claude.max_tokens", 0) // 0 means use plan defaults

	// Define command-line flags using pflag (if not already defined)
	if pflag.Lookup("database-path") == nil {
		pflag.String("database-path", "", "Path to the BoltDB database file")
	}
	if pflag.Lookup("server-address") == nil {
		pflag.String("server-address", "", "gRPC server address for OTLP receiver + Query service")
	}
	if pflag.Lookup("monitor-server") == nil {
		pflag.String("monitor-server", "", "gRPC server address for query service")
	}
	if pflag.Lookup("monitor-timezone") == nil {
		pflag.String("monitor-timezone", "", "Timezone for time filtering and display")
	}
	if pflag.Lookup("claude-plan") == nil {
		pflag.String("claude-plan", "", "Claude subscription plan (unset, pro, max, max20)")
	}
	if pflag.Lookup("claude-max-tokens") == nil {
		pflag.Int("claude-max-tokens", 0, "Custom token limit override (0 means use plan defaults)")
	}

	// Parse flags if not already parsed
	if !pflag.Parsed() {
		pflag.Parse()
	}

	// Bind flags to viper
	if err := v.BindPFlag("database.path", pflag.Lookup("database-path")); err != nil {
		log.Printf("Warning: failed to bind database-path flag: %v", err)
	}
	if err := v.BindPFlag("server.address", pflag.Lookup("server-address")); err != nil {
		log.Printf("Warning: failed to bind server-address flag: %v", err)
	}
	if err := v.BindPFlag("monitor.server", pflag.Lookup("monitor-server")); err != nil {
		log.Printf("Warning: failed to bind monitor-server flag: %v", err)
	}
	if err := v.BindPFlag("monitor.timezone", pflag.Lookup("monitor-timezone")); err != nil {
		log.Printf("Warning: failed to bind monitor-timezone flag: %v", err)
	}
	if err := v.BindPFlag("claude.plan", pflag.Lookup("claude-plan")); err != nil {
		log.Printf("Warning: failed to bind claude-plan flag: %v", err)
	}
	if err := v.BindPFlag("claude.max_tokens", pflag.Lookup("claude-max-tokens")); err != nil {
		log.Printf("Warning: failed to bind claude-max-tokens flag: %v", err)
	}

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

	// Validate timezone
	if c.Monitor.Timezone != "" {
		_, err := time.LoadLocation(c.Monitor.Timezone)
		if err != nil {
			return fmt.Errorf("invalid timezone: %s (%w)", c.Monitor.Timezone, err)
		}
	}

	// Validate max_tokens
	if c.Claude.MaxTokens < 0 {
		return fmt.Errorf("claude.max_tokens must be >= 0, got: %d", c.Claude.MaxTokens)
	}

	return nil
}

// GetTokenLimit returns the effective token limit based on plan and config
func (c *Claude) GetTokenLimit() int {
	// If max_tokens is explicitly set, use it
	if c.MaxTokens > 0 {
		return c.MaxTokens
	}

	// Otherwise, use plan defaults
	switch c.Plan {
	case "pro":
		return 7000
	case "max":
		return 35000
	case "max20":
		return 140000
	default:
		return 0 // No limit for unset plan
	}
}
