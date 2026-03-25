package main

import (
	"flag"
	"log/slog"

	"cs-agent/internal/bootstrap"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	app, cfg, err := bootstrap.NewServer(*configPath)
	if err != nil {
		slog.Error("bootstrap server failed", "error", err)
		return
	}

	if err := app.Listen(cfg.Server.Address()); err != nil {
		slog.Error("start server failed", "error", err)
		return
	}
}
