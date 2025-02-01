package main

import (
	"context"
	"log/slog"

	"github.com/charmbracelet/log"
)

type charmSLogHandler struct{}

func (csh *charmSLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (csh *charmSLogHandler) Handle(ctx context.Context, rec slog.Record) error {
	attrs := []any{}
	rec.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a.Key, a.Value.Any())
		return true
	})
	switch rec.Level {
	case slog.LevelDebug:
		log.Debug(rec.Message, attrs...)
	case slog.LevelInfo:
		log.Info(rec.Message, attrs...)
	case slog.LevelWarn:
		log.Warn(rec.Message, attrs...)
	case slog.LevelError:
		log.Error(rec.Message, attrs...)
	}
	return nil
}

func (csh *charmSLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return csh
}

func (csh *charmSLogHandler) WithGroup(name string) slog.Handler {
	return csh
}
