package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmittmann/tint"
	"log"
	"log/slog"
	client "server/internal/client"
	"server/internal/logging"
	"sync"
	"time"
)

func StartClient(address string, port int) {

	var logs []string
	var mu sync.Mutex
	logChan := make(chan string, 32)

	go func() {
		for r := range logChan {
			mu.Lock()
			logs = append(logs, r)
			mu.Unlock()
		}
	}()

	// Create logger
	stdErrShim := logging.NewStderrCustomShim(logChan)
	logger := slog.New(
		tint.NewHandler(stdErrShim, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	)

	c, err := client.NewClient(
		client.WithTargetAsIpV4(address, port),
		client.WithClientName("TUI-Client"),
		client.WithCustomLogger(logger),
	)
	if err != nil {
		log.Fatal(err)
	}

	model := newModel(c)

	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		slog.Error(fmt.Sprintf("%v", err))
	}

	// Drain the log channel BEFORE closing it
	go func() {
		for len(logChan) > 0 { // Read everything before closing
			mu.Lock()
			logs = append(logs, <-logChan)
			mu.Unlock()
		}
		close(logChan) // Now it's safe to close
	}()

	// Print logs safely
	mu.Lock()
	for _, s := range logs {
		println(s)
	}
	mu.Unlock()

}
