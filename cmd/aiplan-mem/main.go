package main

import (
	"log/slog"
	"os"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/aisa-it/aiplan-mem/internal/db"
	"github.com/aisa-it/aiplan-mem/internal/server"
)

func main() {
	cfg := config.FromENV()
	ds, err := db.OpenDB(cfg)
	if err != nil {
		slog.Error("Open bolt db", "err", err)
		os.Exit(1)
	}

	server.RunServer(cfg, ds)
}
