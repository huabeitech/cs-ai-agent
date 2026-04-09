package logx

import (
	"log/slog"
	"os"
	"strings"
)

// Config 定义日志初始化参数。
type Config struct {
	Level     string `yaml:"level"`
	Format    string `yaml:"format"`
	AddSource bool   `yaml:"addSource"`
}

// Init 初始化全局 slog logger，并设置为默认 logger。
func Init(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "json") {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
