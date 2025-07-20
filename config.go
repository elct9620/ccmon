package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	Address   string      `mapstructure:"address"`
	Retention string      `mapstructure:"retention"`
	Cache     ServerCache `mapstructure:"cache"`
}

// ServerCache configuration
type ServerCache struct {
	Stats CacheStats `mapstructure:"stats"`
}

// CacheStats configuration
type CacheStats struct {
	Enabled bool   `mapstructure:"enabled"`
	TTL     string `mapstructure:"ttl"`
}

// Monitor configuration
type Monitor struct {
	Server          string `mapstructure:"server"`
	Timezone        string `mapstructure:"timezone"`
	RefreshInterval string `mapstructure:"refresh_interval"`
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
	v.SetDefault("server.retention", "never")
	v.SetDefault("server.cache.stats.enabled", true)
	v.SetDefault("server.cache.stats.ttl", "1m")
	v.SetDefault("monitor.server", "127.0.0.1:4317")
	v.SetDefault("monitor.timezone", "UTC")
	v.SetDefault("monitor.refresh_interval", "5s")
	v.SetDefault("claude.plan", "unset")
	v.SetDefault("claude.max_tokens", 0) // 0 means use plan defaults

	// Define command-line flags using pflag (if not already defined)
	if pflag.Lookup("database-path") == nil {
		pflag.String("database-path", "", "Path to the BoltDB database file")
	}
	if pflag.Lookup("server-address") == nil {
		pflag.String("server-address", "", "gRPC server address for OTLP receiver + Query service")
	}
	if pflag.Lookup("server-retention") == nil {
		pflag.String("server-retention", "", "Data retention period (e.g., '7d', '30d', 'never')")
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
	if pflag.Lookup("server-cache-stats-enabled") == nil {
		pflag.Bool("server-cache-stats-enabled", true, "Enable stats cache")
	}
	if pflag.Lookup("server-cache-stats-ttl") == nil {
		pflag.String("server-cache-stats-ttl", "1m", "Stats cache TTL")
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
	if err := v.BindPFlag("server.retention", pflag.Lookup("server-retention")); err != nil {
		log.Printf("Warning: failed to bind server-retention flag: %v", err)
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
	if err := v.BindPFlag("server.cache.stats.enabled", pflag.Lookup("server-cache-stats-enabled")); err != nil {
		log.Printf("Warning: failed to bind server-cache-stats-enabled flag: %v", err)
	}
	if err := v.BindPFlag("server.cache.stats.ttl", pflag.Lookup("server-cache-stats-ttl")); err != nil {
		log.Printf("Warning: failed to bind server-cache-stats-ttl flag: %v", err)
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

	// Validate retention
	if err := c.Server.ValidateRetention(); err != nil {
		return fmt.Errorf("invalid server.retention: %w", err)
	}

	// Validate cache TTL
	if c.Server.Cache.Stats.TTL != "" {
		_, err := time.ParseDuration(c.Server.Cache.Stats.TTL)
		if err != nil {
			return fmt.Errorf("invalid cache TTL format: %s (%w)", c.Server.Cache.Stats.TTL, err)
		}
	}

	return nil
}

// ValidateRetention validates the retention configuration
func (s *Server) ValidateRetention() error {
	if s.Retention == "" || s.Retention == "never" {
		return nil // No retention is valid
	}

	// Try to parse as duration (with support for days)
	duration, err := s.parseRetentionDuration(s.Retention)
	if err != nil {
		return fmt.Errorf("invalid duration format: %s", s.Retention)
	}

	// Minimum retention period: 24 hours
	if duration < 24*time.Hour {
		return fmt.Errorf("retention period must be at least 24h, got: %s", s.Retention)
	}

	return nil
}

// IsRetentionEnabled returns true if data retention is configured
func (s *Server) IsRetentionEnabled() bool {
	return s.Retention != "" && s.Retention != "never"
}

// GetRetentionDuration returns the retention duration or zero if disabled
func (s *Server) GetRetentionDuration() time.Duration {
	if !s.IsRetentionEnabled() {
		return 0
	}

	duration, err := s.parseRetentionDuration(s.Retention)
	if err != nil {
		return 0 // Should not happen after validation
	}

	return duration
}

// parseRetentionDuration parses duration strings with support for days (e.g., "7d", "30d")
func (s *Server) parseRetentionDuration(retention string) (time.Duration, error) {
	// Handle days suffix (e.g., "7d", "30d")
	if strings.HasSuffix(retention, "d") {
		daysStr := strings.TrimSuffix(retention, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, fmt.Errorf("invalid day format: %s", retention)
		}
		if days < 0 {
			return 0, fmt.Errorf("negative duration not allowed: %s", retention)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// Use standard Go duration parsing for other formats (h, m, s)
	return time.ParseDuration(retention)
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

// GetClaudePlan returns the configured Claude plan, implementing PlanConfig interface
func (c *Config) GetClaudePlan() string {
	return c.Claude.Plan
}
