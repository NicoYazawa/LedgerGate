package log

import (
	"io"
	"log/slog"
	"os"
)

func New(service, level string) *slog.Logger { return newWithWriter(service, level, os.Stderr) }

func newWithWriter(service, level string, writer io.Writer) *slog.Logger {
	levels := map[string]slog.Level{
		"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError,
	}
	options := &slog.HandlerOptions{Level: levels[level]}
	return slog.New(slog.NewJSONHandler(writer, options)).With("service", service)
}
