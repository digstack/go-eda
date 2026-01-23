package logger

import (
	"fmt"
	"log"
	"os"
)

// Logger represents a generic logger interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// StandardLogger implements a simple standard logger
type StandardLogger struct {
	logger *log.Logger
}

func NewStandardLogger() *StandardLogger {
	return &StandardLogger{
		logger: log.New(os.Stdout, "[INFO] ", log.LstdFlags),
	}
}

func NewStandardLoggerWithPrefix(prefix string) *StandardLogger {
	return &StandardLogger{
		logger: log.New(os.Stdout, "["+prefix+"] ", log.LstdFlags),
	}
}

func (l *StandardLogger) Debug(msg string, fields ...Field) {
	l.logWithLevel("DEBUG", msg, fields...)
}

func (l *StandardLogger) Info(msg string, fields ...Field) {
	l.logWithLevel("INFO", msg, fields...)
}

func (l *StandardLogger) Warn(msg string, fields ...Field) {
	l.logWithLevel("WARN", msg, fields...)
}

func (l *StandardLogger) Error(msg string, fields ...Field) {
	l.logWithLevel("ERROR", msg, fields...)
}

func (l *StandardLogger) Fatal(msg string, fields ...Field) {
	l.logWithLevel("FATAL", msg, fields...)
	os.Exit(1)
}

func (l *StandardLogger) logWithLevel(level, msg string, fields ...Field) {
	logMsg := "[" + level + "] " + msg
	if len(fields) > 0 {
		logMsg += " |"
		for _, field := range fields {
			logMsg += " " + field.Key + "=" + formatValue(field.Value)
		}
	}
	l.logger.Println(logMsg)
}

// NoOpLogger implements a logger that does nothing
type NoOpLogger struct{}

func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Debug(msg string, fields ...Field) {}
func (l *NoOpLogger) Info(msg string, fields ...Field)  {}
func (l *NoOpLogger) Warn(msg string, fields ...Field)  {}
func (l *NoOpLogger) Error(msg string, fields ...Field) {}
func (l *NoOpLogger) Fatal(msg string, fields ...Field) {}

// Helper functions
func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%+v", v)
	}
}
