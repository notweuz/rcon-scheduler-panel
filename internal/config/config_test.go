package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	valid := Config{
		DatabaseDSN: "host=localhost",
		JWTSecret:   defaultJWTSecret,
		LogLevel:    "info",
		AppPort:     8080,
		Timezone:    "Europe/Moscow",
		GinMode:     "debug",
	}

	tests := []struct {
		name    string
		modify  func(c *Config)
		wantErr string
	}{
		{"default secret in debug", func(c *Config) {}, ""},
		{"strong secret in release", func(c *Config) {
			c.GinMode = "release"
			c.JWTSecret = strings.Repeat("a", minJWTSecretLen)
		}, ""},
		{"default secret in release", func(c *Config) { c.GinMode = "release" }, "JWT_SECRET"},
		{"short secret in release", func(c *Config) {
			c.GinMode = "release"
			c.JWTSecret = "123"
		}, "JWT_SECRET"},
		{"empty dsn", func(c *Config) { c.DatabaseDSN = "" }, "DATABASE_DSN"},
		{"invalid port", func(c *Config) { c.AppPort = 0 }, "APP_PORT"},
		{"invalid log level", func(c *Config) { c.LogLevel = "verbose" }, "LOG_LEVEL"},
		{"invalid gin mode", func(c *Config) { c.GinMode = "prod" }, "GIN_MODE"},
		{"invalid timezone", func(c *Config) { c.Timezone = "Mars/Olympus" }, "TIMEZONE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.modify(&cfg)

			err := cfg.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
