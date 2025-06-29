package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/khamiruf/internal_transfers_system_go/internal/config"
)

var (
	instance *Logger
	once     sync.Once
)

// Logger wraps the standard logger with additional functionality
type Logger struct {
	*log.Logger
	level config.LogLevel
}

// Config holds the logger configuration
type Config struct {
	Level  config.LogLevel
	Output io.Writer
	Prefix string
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  config.INFO,
		Output: os.Stdout,
	}
}

// Initialize sets up the logger with the given configuration
func Initialize(cfg *Config) {
	once.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
		}
		instance = &Logger{
			Logger: log.New(cfg.Output, cfg.Prefix, log.LstdFlags|log.Lmicroseconds|log.Lshortfile),
			level:  cfg.Level,
		}
	})
}

// GetInstance returns the singleton logger instance
func GetInstance() *Logger {
	if instance == nil {
		Initialize(nil)
	}
	return instance
}

// log formats and outputs a log message if the level is sufficient
func (l *Logger) log(level config.LogLevel, format string, v ...interface{}) {
	if level >= l.level {
		msg := fmt.Sprintf(format, v...)
		l.Output(2, fmt.Sprintf("[%s] %s", level.String(), msg))
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(config.DEBUG, format, v...)
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(config.INFO, format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(config.WARN, format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(config.ERROR, format, v...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(config.FATAL, format, v...)
	os.Exit(1)
}

// Convenience functions for package-level logging
func Debug(format string, v ...interface{}) {
	GetInstance().Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	GetInstance().Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	GetInstance().Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	GetInstance().Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	GetInstance().Fatal(format, v...)
}
