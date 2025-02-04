package main

import (
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
	"server/internal/server"
	"time"
)

func main() {
	println("Starting server!")

	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	server.Serve()
}
