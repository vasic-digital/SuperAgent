package services

import (
	"github.com/sirupsen/logrus"
)

// newTestLogger creates a logger for testing purposes.
func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return logger
}
