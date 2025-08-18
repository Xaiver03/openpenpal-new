package config

import (
	"os"
	"testing"
)

func TestSSLConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *SSLConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid disable mode",
			config: &SSLConfig{
				Mode: SSLModeDisable,
			},
			wantErr: false,
		},
		{
			name: "valid require mode",
			config: &SSLConfig{
				Mode: SSLModeRequire,
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			config: &SSLConfig{
				Mode: "invalid",
			},
			wantErr:     true,
			errContains: "invalid SSL mode",
		},
		{
			name: "verify-ca without CA file",
			config: &SSLConfig{
				Mode: SSLModeVerifyCA,
			},
			wantErr:     true,
			errContains: "CA file is required",
		},
		{
			name: "cert without key",
			config: &SSLConfig{
				Mode:     SSLModeRequire,
				CertFile: "/path/to/cert",
			},
			wantErr:     true,
			errContains: "both cert file and key file must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %v, want containing %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestSSLConfigLoadFromEnv(t *testing.T) {
	// Save original env
	origMode := os.Getenv("DB_SSL_MODE")
	origCA := os.Getenv("DB_SSL_CA_FILE")
	defer func() {
		os.Setenv("DB_SSL_MODE", origMode)
		os.Setenv("DB_SSL_CA_FILE", origCA)
	}()

	// Test loading from env
	os.Setenv("DB_SSL_MODE", "verify-full")
	os.Setenv("DB_SSL_CA_FILE", "/test/ca.pem")

	config := &SSLConfig{}
	config.LoadFromEnv()

	if config.Mode != "verify-full" {
		t.Errorf("LoadFromEnv() Mode = %v, want verify-full", config.Mode)
	}
	if config.CAFile != "/test/ca.pem" {
		t.Errorf("LoadFromEnv() CAFile = %v, want /test/ca.pem", config.CAFile)
	}
}

func TestBuildDSNParams(t *testing.T) {
	tests := []struct {
		name   string
		config *SSLConfig
		want   string
	}{
		{
			name: "disable mode",
			config: &SSLConfig{
				Mode: SSLModeDisable,
			},
			want: "sslmode=disable",
		},
		{
			name: "require mode with certificates",
			config: &SSLConfig{
				Mode:     SSLModeRequire,
				CAFile:   "/path/to/ca.pem",
				CertFile: "/path/to/cert.pem",
				KeyFile:  "/path/to/key.pem",
			},
			want: "sslmode=require sslrootcert=/path/to/ca.pem sslcert=/path/to/cert.pem sslkey=/path/to/key.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.BuildDSNParams()
			if got != tt.want {
				t.Errorf("BuildDSNParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRecommendedSSLMode(t *testing.T) {
	tests := []struct {
		environment string
		want        string
	}{
		{"production", SSLModeVerifyFull},
		{"staging", SSLModeRequire},
		{"test", SSLModeDisable},
		{"development", SSLModeDisable},
		{"unknown", SSLModeDisable},
	}

	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			got := GetRecommendedSSLMode(tt.environment)
			if got != tt.want {
				t.Errorf("GetRecommendedSSLMode(%v) = %v, want %v", tt.environment, got, tt.want)
			}
		})
	}
}

func TestDefaultSSLConfigs(t *testing.T) {
	// Test that default configs are properly set
	for env, config := range DefaultSSLConfigs {
		t.Run(env, func(t *testing.T) {
			if config.Environment != env {
				t.Errorf("DefaultSSLConfigs[%v].Environment = %v, want %v", env, config.Environment, env)
			}
			
			// Validate default configs
			if env == "development" || env == "test" {
				if config.Mode != SSLModeDisable {
					t.Errorf("DefaultSSLConfigs[%v].Mode = %v, want %v", env, config.Mode, SSLModeDisable)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}