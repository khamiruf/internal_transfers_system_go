package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL      string
	ServerPort       int
	MaxDBConnections int
	MaxIdleConns     int
	ConnMaxLifetime  int // in minutes
	LogLevel         string
}

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// Log levels
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO // default to INFO if invalid
	}
}

func Load() (*Config, error) {
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/transfers?sslmode=disable")
	serverPort := getEnvAsInt("PORT", 8080)
	maxDBConns := getEnvAsInt("MAX_DB_CONNECTIONS", 25)
	maxIdleConns := getEnvAsInt("MAX_IDLE_CONNECTIONS", 5)
	connMaxLifetime := getEnvAsInt("CONN_MAX_LIFETIME_MINUTES", 30)
	logLevel := getEnv("LOG_LEVEL", "info")

	return &Config{
		DatabaseURL:      databaseURL,
		ServerPort:       serverPort,
		MaxDBConnections: maxDBConns,
		MaxIdleConns:     maxIdleConns,
		ConnMaxLifetime:  connMaxLifetime,
		LogLevel:         logLevel,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
