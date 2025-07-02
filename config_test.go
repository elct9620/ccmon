package main

import (
	"strings"
	"testing"
	"time"
)

func TestServer_ValidateRetention(t *testing.T) {
	tests := []struct {
		name      string
		retention string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty retention (valid)",
			retention: "",
			wantErr:   false,
		},
		{
			name:      "never retention (valid)",
			retention: "never",
			wantErr:   false,
		},
		{
			name:      "valid 24h retention",
			retention: "24h",
			wantErr:   false,
		},
		{
			name:      "valid 1d retention",
			retention: "1d",
			wantErr:   false,
		},
		{
			name:      "valid 7d retention",
			retention: "7d",
			wantErr:   false,
		},
		{
			name:      "valid 30d retention",
			retention: "30d",
			wantErr:   false,
		},
		{
			name:      "valid 168h retention",
			retention: "168h",
			wantErr:   false,
		},
		{
			name:      "invalid less than 24h",
			retention: "12h",
			wantErr:   true,
			errMsg:    "retention period must be at least 24h",
		},
		{
			name:      "invalid less than 24h (23h59m)",
			retention: "23h59m",
			wantErr:   true,
			errMsg:    "retention period must be at least 24h",
		},
		{
			name:      "invalid format",
			retention: "invalid",
			wantErr:   true,
			errMsg:    "invalid duration format",
		},
		{
			name:      "invalid negative duration",
			retention: "-24h",
			wantErr:   true,
			errMsg:    "retention period must be at least 24h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Retention: tt.retention,
			}

			err := server.ValidateRetention()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateRetention() expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateRetention() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateRetention() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestServer_IsRetentionEnabled(t *testing.T) {
	tests := []struct {
		name      string
		retention string
		want      bool
	}{
		{
			name:      "empty retention",
			retention: "",
			want:      false,
		},
		{
			name:      "never retention",
			retention: "never",
			want:      false,
		},
		{
			name:      "1d retention",
			retention: "1d",
			want:      true,
		},
		{
			name:      "7d retention",
			retention: "7d",
			want:      true,
		},
		{
			name:      "30d retention",
			retention: "30d",
			want:      true,
		},
		{
			name:      "24h retention",
			retention: "24h",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Retention: tt.retention,
			}

			got := server.IsRetentionEnabled()
			if got != tt.want {
				t.Errorf("IsRetentionEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_GetRetentionDuration(t *testing.T) {
	tests := []struct {
		name      string
		retention string
		want      time.Duration
	}{
		{
			name:      "empty retention",
			retention: "",
			want:      0,
		},
		{
			name:      "never retention",
			retention: "never",
			want:      0,
		},
		{
			name:      "1d retention",
			retention: "1d",
			want:      1 * 24 * time.Hour,
		},
		{
			name:      "7d retention",
			retention: "7d",
			want:      7 * 24 * time.Hour,
		},
		{
			name:      "30d retention",
			retention: "30d",
			want:      30 * 24 * time.Hour,
		},
		{
			name:      "24h retention",
			retention: "24h",
			want:      24 * time.Hour,
		},
		{
			name:      "168h retention (7 days)",
			retention: "168h",
			want:      168 * time.Hour,
		},
		{
			name:      "invalid format returns 0",
			retention: "invalid",
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Retention: tt.retention,
			}

			got := server.GetRetentionDuration()
			if got != tt.want {
				t.Errorf("GetRetentionDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_ValidateRetentionIntegration(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with retention",
			config: Config{
				Server: Server{
					Address:   "127.0.0.1:4317",
					Retention: "7d",
				},
				Claude: Claude{
					Plan:      "pro",
					MaxTokens: 0,
				},
				Monitor: Monitor{
					Timezone: "UTC",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config with bad retention",
			config: Config{
				Server: Server{
					Address:   "127.0.0.1:4317",
					Retention: "12h",
				},
				Claude: Claude{
					Plan:      "pro",
					MaxTokens: 0,
				},
				Monitor: Monitor{
					Timezone: "UTC",
				},
			},
			wantErr: true,
			errMsg:  "invalid server.retention",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

