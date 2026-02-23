package utils

import (
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	logger     *logrus.Logger
	loggerOnce sync.Once
)

// GetLogger returns the configured logger instance, initializing it on first call.
func GetLogger() *logrus.Logger {
	loggerOnce.Do(func() {
		l := logrus.New()
		l.SetOutput(os.Stdout)
		l.SetFormatter(&logrus.JSONFormatter{})

		// Set log level from environment
		level := strings.ToLower(os.Getenv("LOG_LEVEL"))
		switch level {
		case "debug":
			l.SetLevel(logrus.DebugLevel)
		case "info":
			l.SetLevel(logrus.InfoLevel)
		case "warn", "warning":
			l.SetLevel(logrus.WarnLevel)
		case "error":
			l.SetLevel(logrus.ErrorLevel)
		default:
			l.SetLevel(logrus.InfoLevel)
		}

		logger = l
	})
	return logger
}

// Logger is a package-level variable for backward compatibility.
// Prefer using GetLogger() for new code.
var Logger = GetLogger()
