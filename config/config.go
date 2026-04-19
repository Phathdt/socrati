package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Env string `mapstructure:"env"` // "production" or "development"

	Server   ServerConfig   `mapstructure:"server"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Embedder EmbedderConfig `mapstructure:"embedder"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	URI string `mapstructure:"uri"`
}

type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type EmbedderConfig struct {
	Provider   string `mapstructure:"provider"`    // e.g. "voyage"
	Model      string `mapstructure:"model"`       // e.g. "voyage-4-lite"
	BaseURL    string `mapstructure:"base_url"`    // e.g. "https://api.voyageai.com/v1"
	APIKey     string `mapstructure:"api_key"`     // override via EMBEDDER__API_KEY or VOYAGE_API_KEY
	TimeoutMS  int    `mapstructure:"timeout_ms"`  // per-request HTTP timeout
	MaxRetries int    `mapstructure:"max_retries"` // retry attempts for transient errors
	MaxChars   int    `mapstructure:"max_chars"`   // truncate input longer than this
}

// LoadConfig reads config from YAML file and environment variables.
// Env vars use __ as separator (e.g. SERVER__PORT overrides server.port).
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Bind known env vars explicitly for nested keys.
	envBindings := map[string]string{
		"env":                  "ENV",
		"server.host":          "SERVER__HOST",
		"server.port":          "SERVER__PORT",
		"logger.level":         "LOGGER__LEVEL",
		"logger.format":        "LOGGER__FORMAT",
		"embedder.provider":    "EMBEDDER__PROVIDER",
		"embedder.model":       "EMBEDDER__MODEL",
		"embedder.base_url":    "EMBEDDER__BASE_URL",
		"embedder.api_key":     "EMBEDDER__API_KEY",
		"embedder.timeout_ms":  "EMBEDDER__TIMEOUT_MS",
		"embedder.max_retries": "EMBEDDER__MAX_RETRIES",
		"embedder.max_chars":   "EMBEDDER__MAX_CHARS",
		"database.uri":         "DATABASE__URI",
	}
	for key, env := range envBindings {
		if err := v.BindEnv(key, env); err != nil {
			log.Printf("warn: failed to bind env var key=%s env=%s error=%s", key, env, err.Error())
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Fallback for the standard provider env var name (kept outside config.yml
	// so operators can inject secrets via environment alone).
	if cfg.Embedder.APIKey == "" {
		cfg.Embedder.APIKey = os.Getenv("VOYAGE_API_KEY")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that required configuration fields are set.
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be > 0")
	}
	return nil
}
