package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Env string `mapstructure:"env"` // "production" or "development"

	Server ServerConfig `mapstructure:"server"`
	Logger LoggerConfig `mapstructure:"logger"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// LoadConfig reads config from YAML file and environment variables.
// Env vars use __ as separator (e.g. DATABASE__URI overrides database.uri).
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	v.SetDefault("env", "development")
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 4000)
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "text")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Bind known env vars explicitly for nested keys
	envBindings := map[string]string{
		"env":           "ENV",
		"server.host":   "SERVER__HOST",
		"server.port":   "SERVER__PORT",
		"logger.level":  "LOGGER__LEVEL",
		"logger.format": "LOGGER__FORMAT",
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
