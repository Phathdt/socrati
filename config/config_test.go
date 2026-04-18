package config

import (
	"os"
	"testing"
)

func TestLoadConfigWithEnvOverrides(t *testing.T) {
	envVars := map[string]string{
		"SERVER__PORT":   "3000",
		"LOGGER__LEVEL":  "error",
		"LOGGER__FORMAT": "json",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			os.Unsetenv(key)
		}
	}()

	cfg, err := LoadConfig("../config.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("Expected server port 3000, got %d", cfg.Server.Port)
	}
	if cfg.Logger.Level != "error" {
		t.Errorf("Expected logger level 'error', got '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.Format != "json" {
		t.Errorf("Expected logger format 'json', got '%s'", cfg.Logger.Format)
	}
}

func TestLoadConfigFromYAMLOnly(t *testing.T) {
	envKeys := []string{"SERVER__PORT", "LOGGER__LEVEL", "LOGGER__FORMAT"}
	for _, key := range envKeys {
		os.Unsetenv(key)
	}

	cfg, err := LoadConfig("../config.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 4000 {
		t.Errorf("Expected server port 4000, got %d", cfg.Server.Port)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("Expected logger level 'debug', got '%s'", cfg.Logger.Level)
	}
}
