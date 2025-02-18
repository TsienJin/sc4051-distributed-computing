package tests

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// PrefixedLoggerHandler is a custom slog.Handler that adds a prefix to each log message.
type PrefixedLoggerHandler struct {
	handler slog.Handler // The underlying handler.
	prefix  string       // The prefix string.
}

// NewPrefixedLoggerHandler creates a new PrefixedLoggerHandler.
func NewPrefixedLoggerHandler(handler slog.Handler, prefix string) *PrefixedLoggerHandler {
	return &PrefixedLoggerHandler{
		handler: handler,
		prefix:  fmt.Sprintf("[%s]", prefix),
	}
}

// Enabled reports whether the handler handles records at the given level.
// It simply delegates to the underlying handler.
func (h *PrefixedLoggerHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle handles the slog.Record by adding the prefix and then calling the underlying handler.
func (h *PrefixedLoggerHandler) Handle(ctx context.Context, record slog.Record) error {
	// Add the prefix attribute to the record.  This ensures the prefix appears in the log output.
	record.Add("prefix", h.prefix)

	// IMPORTANT: Modify Record.Message to prepend. Alternative: use an  Attribute for prefix.
	record.Message = h.prefix + " " + record.Message // Prepend prefix to message.

	// Delegate to the wrapped handler.
	return h.handler.Handle(ctx, record)
}

// WithAttrs returns a new Handler whose attributes consist of the
// attributes of the receiver, followed by attrs.
// It simply delegates to the underlying handler.
func (h *PrefixedLoggerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewPrefixedLoggerHandler(h.handler.WithAttrs(attrs), h.prefix)
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// It simply delegates to the underlying handler.
func (h *PrefixedLoggerHandler) WithGroup(name string) slog.Handler {
	return NewPrefixedLoggerHandler(h.handler.WithGroup(name), h.prefix)
}

func NewTestLogger() *slog.Logger {
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	prefixedHandler := NewPrefixedLoggerHandler(textHandler, "TEST")
	return slog.New(prefixedHandler)
}

func NewNamedTestLogger(name string) *slog.Logger {
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	prefixedHandler := NewPrefixedLoggerHandler(textHandler, fmt.Sprintf("TEST|%s", name))
	return slog.New(prefixedHandler)
}

func NewNamedTestLoggerWithOutputShim(name string, shim io.Writer) *slog.Logger {
	textHandler := slog.NewTextHandler(shim, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	prefixedHandler := NewPrefixedLoggerHandler(textHandler, fmt.Sprintf("TEST|%s", name))
	return slog.New(prefixedHandler)
}
