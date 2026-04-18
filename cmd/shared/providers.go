package shared

import (
	"socrati/config"
	"socrati/pkg/logger"
)

// InitLogger creates and initializes the application logger
func InitLogger(cfg *config.Config) logger.Logger {
	return logger.New(cfg.Logger.Format, cfg.Logger.Level)
}
