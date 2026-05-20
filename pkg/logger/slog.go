package logger

import (
	"context"
	"log/slog"
	"os"
)

// SlogLogger adapts a *slog.Logger to the Logger interface so consumers
// can plug in any slog handler (JSON, text, OTEL bridge, ...) while keeping
// the boilerplate's Logger API stable.
type SlogLogger struct {
	l *slog.Logger
}

// NewSlogLogger wraps an existing *slog.Logger. Pass slog.Default() for the
// process-wide default, or build a custom one.
func NewSlogLogger(l *slog.Logger) *SlogLogger {
	if l == nil {
		l = slog.Default()
	}
	return &SlogLogger{l: l}
}

// NewJSONSlogLogger returns a JSON slog logger writing to stdout at the
// given level.
func NewJSONSlogLogger(level slog.Level) *SlogLogger {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return &SlogLogger{l: slog.New(h)}
}

// NewTextSlogLogger returns a text slog logger writing to stdout at the
// given level.
func NewTextSlogLogger(level slog.Level) *SlogLogger {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return &SlogLogger{l: slog.New(h)}
}

// Slog returns the underlying *slog.Logger for direct use.
func (s *SlogLogger) Slog() *slog.Logger { return s.l }

func (s *SlogLogger) Debug(msg string, fields ...Field) { s.log(slog.LevelDebug, msg, fields) }
func (s *SlogLogger) Info(msg string, fields ...Field)  { s.log(slog.LevelInfo, msg, fields) }
func (s *SlogLogger) Warn(msg string, fields ...Field)  { s.log(slog.LevelWarn, msg, fields) }
func (s *SlogLogger) Error(msg string, fields ...Field) { s.log(slog.LevelError, msg, fields) }

// Fatal logs at error level then exits the process.
func (s *SlogLogger) Fatal(msg string, fields ...Field) {
	s.log(slog.LevelError, msg, fields)
	os.Exit(1)
}

func (s *SlogLogger) log(lvl slog.Level, msg string, fields []Field) {
	if !s.l.Enabled(context.Background(), lvl) {
		return
	}
	attrs := make([]slog.Attr, 0, len(fields))
	for _, f := range fields {
		attrs = append(attrs, slog.Any(f.Key, f.Value))
	}
	s.l.LogAttrs(context.Background(), lvl, msg, attrs...)
}
