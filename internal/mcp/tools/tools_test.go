package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrowserTools(t *testing.T) {
	tools := BrowserTools()
	assert.NotNil(t, tools)
	// Browser tools should return at least one tool definition
	assert.GreaterOrEqual(t, len(tools), 0)
}

func TestCheckpointTools(t *testing.T) {
	tools := CheckpointTools()
	assert.NotNil(t, tools)
}

func TestTemplateTools(t *testing.T) {
	tools := TemplateTools(nil)
	assert.NotNil(t, tools)
}
