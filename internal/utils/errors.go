package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AppError represents an application error with context.
type AppError struct {
	Code    string
	Message string
	Status  int
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// NewAppError creates a new application error.
func NewAppError(code, message string, status int, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Cause:   cause,
	}
}

// HandleError handles errors in Gin handlers.
func HandleError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		Logger.WithFields(logrus.Fields{
			"code":  appErr.Code,
			"cause": appErr.Cause,
		}).Error(appErr.Message)
		c.JSON(appErr.Status, gin.H{
			"error": appErr.Message,
			"code":  appErr.Code,
		})
		return
	}

	Logger.Error(err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
	})
}
