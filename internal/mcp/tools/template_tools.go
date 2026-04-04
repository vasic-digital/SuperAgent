// Package tools provides MCP tools for templates
package tools

import "dev.helix.agent/internal/templates"

// TemplateTools returns template-related MCP tools
func TemplateTools(manager *templates.Manager) []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "list_context_templates",
			Description: "List available context templates",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tag": map[string]interface{}{
						"type":        "string",
						"description": "Filter by tag",
					},
				},
			},
		},
		{
			Name:        "apply_context_template",
			Description: "Apply a context template to load relevant files and context",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"template_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the template to apply",
					},
					"variables": map[string]interface{}{
						"type":        "object",
						"description": "Template variables",
					},
				},
				"required": []string{"template_id"},
			},
		},
		{
			Name:        "get_template_prompt",
			Description: "Get a predefined prompt from a template",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"template_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the template",
					},
					"prompt_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the prompt",
					},
					"variables": map[string]interface{}{
						"type":        "object",
						"description": "Prompt variables",
					},
				},
				"required": []string{"template_id", "prompt_name"},
			},
		},
	}
}

// ToolDefinition represents an MCP tool
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}
