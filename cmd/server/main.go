package main

import (
	"flag"
	"log/slog"

	"cs-agent/internal/bootstrap"
	"cs-agent/internal/pkg/config"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	if err := bootstrap.Init(*configPath); err != nil {
		slog.Error("bootstrap init failed", "error", err)
		return
	}

	cfg := config.Current()

	app, err := bootstrap.NewServer()
	if err != nil {
		slog.Error("bootstrap server failed", "error", err)
		return
	}

	if err := app.Listen(cfg.Server.Address()); err != nil {
		slog.Error("start server failed", "error", err)
		return
	}
}
