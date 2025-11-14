package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps zerolog with rotation support
type Logger struct {
	logger zerolog.Logger
	writer io.Writer
	mu     sync.Mutex
}

var (
	globalLogger *Logger
	once         sync.Once
)

// Config holds logger configuration
type Config struct {
	Path       string
	Level      string
	Format     string // json or console
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

// Init initializes the global logger
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		globalLogger, err = New(cfg)
	})
	return err
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	// Create log directory if it doesn't exist
	if cfg.Path != "" {
		dir := filepath.Dir(cfg.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Set up log rotation
	var writer io.Writer
	if cfg.Path != "" {
		writer = &lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
		}
	} else {
		writer = os.Stdout
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var zlog zerolog.Logger
	if cfg.Format == "console" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
		}
		zlog = zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()
	} else {
		zlog = zerolog.New(writer).With().Timestamp().Caller().Logger()
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zlog = zlog.Level(level)

	return &Logger{
		logger: zlog,
		writer: writer,
	}, nil
}

// Get returns the global logger
func Get() *Logger {
	if globalLogger == nil {
		// Fallback to console logger
		globalLogger = &Logger{
			logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
			writer: os.Stdout,
		}
	}
	return globalLogger
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...interface{}) {
	event := l.logger.Debug()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...interface{}) {
	event := l.logger.Info()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...interface{}) {
	event := l.logger.Warn()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields ...interface{}) {
	event := l.logger.Error().Err(err)
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error, fields ...interface{}) {
	event := l.logger.Fatal().Err(err)
	l.addFields(event, fields...)
	event.Msg(msg)
}

// addFields adds key-value pairs to a log event
func (l *Logger) addFields(event *zerolog.Event, fields ...interface{}) {
	if len(fields)%2 != 0 {
		event.Interface("invalid_fields", fields)
		return
	}

	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		event.Interface(key, fields[i+1])
	}
}

// WithComponent returns a new logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("component", component).Logger(),
		writer: l.writer,
	}
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{
		logger: ctx.Logger(),
		writer: l.writer,
	}
}

// Global convenience functions
func Debug(msg string, fields ...interface{}) {
	Get().Debug(msg, fields...)
}

func Info(msg string, fields ...interface{}) {
	Get().Info(msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
	Get().Warn(msg, fields...)
}

func Error(msg string, err error, fields ...interface{}) {
	Get().Error(msg, err, fields...)
}

func Fatal(msg string, err error, fields ...interface{}) {
	Get().Fatal(msg, err, fields...)
}
