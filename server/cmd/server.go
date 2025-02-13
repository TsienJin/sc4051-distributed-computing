package main

import (
	"github.com/lmittmann/tint"
	"log/slog"
	"server/internal/logging"
	"server/internal/server"
	"server/internal/vars"
	"sync"
	"time"
)

func main() {

	// Create logger
	w := logging.NewStderrShim()
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen,
		}),
	))

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		go logging.Serve(vars.GetStaticEnv().ServerLogPort)
		wg.Done()
	}()

	wg.Wait()

	slog.Info("Starting server!")
	server.Serve(vars.GetStaticEnv().ServerPort)
}
