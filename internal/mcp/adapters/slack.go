// Package adapters provides MCP server adapters.
// This file implements the Slack MCP server adapter.
package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SlackConfig configures the Slack adapter.
type SlackConfig struct {
	BotToken string        `json:"bot_token"`
	AppToken string        `json:"app_token,omitempty"`
	Timeout  time.Duration `json:"timeout"`
	TeamID   string        `json:"team_id,omitempty"`
}

// DefaultSlackConfig returns default configuration.
func DefaultSlackConfig() SlackConfig {
	return SlackConfig{
		Timeout: 30 * time.Second,
	}
}

// SlackAdapter implements the Slack MCP server.
type SlackAdapter struct {
	config SlackConfig
	client SlackClient
}

// SlackClient interface for Slack operations.
type SlackClient interface {
	PostMessage(ctx context.Context, channel, text string, options MessageOptions) (string, error)
	UpdateMessage(ctx context.Context, channel, ts, text string) error
	DeleteMessage(ctx context.Context, channel, ts string) error
	GetMessage(ctx context.Context, channel, ts string) (*SlackMessage, error)
	ListChannels(ctx context.Context, types string, limit int) ([]SlackChannel, error)
	GetChannel(ctx context.Context, channelID string) (*SlackChannel, error)
	CreateChannel(ctx context.Context, name string, isPrivate bool) (*SlackChannel, error)
	ArchiveChannel(ctx context.Context, channelID string) error
	JoinChannel(ctx context.Context, channelID string) error
	LeaveChannel(ctx context.Context, channelID string) error
	InviteToChannel(ctx context.Context, channelID, userID string) error
	ListUsers(ctx context.Context, limit int) ([]SlackUser, error)
	GetUser(ctx context.Context, userID string) (*SlackUser, error)
	AddReaction(ctx context.Context, channel, ts, emoji string) error
	RemoveReaction(ctx context.Context, channel, ts, emoji string) error
	UploadFile(ctx context.Context, channels []string, filename, content string) (*SlackFile, error)
	SearchMessages(ctx context.Context, query string, count int) ([]SlackMessage, error)
	GetConversationHistory(ctx context.Context, channel string, limit int) ([]SlackMessage, error)
}

// MessageOptions represents message options.
type MessageOptions struct {
	ThreadTS       string       `json:"thread_ts,omitempty"`
	ReplyBroadcast bool         `json:"reply_broadcast,omitempty"`
	Blocks         []Block      `json:"blocks,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty"`
	Unfurl         bool         `json:"unfurl,omitempty"`
}

// Block represents a Slack block.
type Block struct {
	Type string      `json:"type"`
	Text *TextObject `json:"text,omitempty"`
}

// TextObject represents a Slack text object.
type TextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Attachment represents a Slack attachment.
type Attachment struct {
	Color  string `json:"color,omitempty"`
	Title  string `json:"title,omitempty"`
	Text   string `json:"text,omitempty"`
	Footer string `json:"footer,omitempty"`
}

// SlackMessage represents a Slack message.
type SlackMessage struct {
	TS        string     `json:"ts"`
	Text      string     `json:"text"`
	User      string     `json:"user"`
	Channel   string     `json:"channel"`
	ThreadTS  string     `json:"thread_ts,omitempty"`
	Reactions []Reaction `json:"reactions,omitempty"`
}

// Reaction represents a message reaction.
type Reaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

// SlackChannel represents a Slack channel.
type SlackChannel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsPrivate  bool   `json:"is_private"`
	IsArchived bool   `json:"is_archived"`
	IsMember   bool   `json:"is_member"`
	NumMembers int    `json:"num_members"`
	Topic      string `json:"topic"`
	Purpose    string `json:"purpose"`
}

// SlackUser represents a Slack user.
type SlackUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
	Email    string `json:"email,omitempty"`
	IsBot    bool   `json:"is_bot"`
	IsAdmin  bool   `json:"is_admin"`
	Status   string `json:"status,omitempty"`
}

// SlackFile represents an uploaded file.
type SlackFile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url_private"`
	Size     int    `json:"size"`
	Mimetype string `json:"mimetype"`
}

// NewSlackAdapter creates a new Slack adapter.
func NewSlackAdapter(config SlackConfig, client SlackClient) *SlackAdapter {
	return &SlackAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *SlackAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "slack",
		Version:     "1.0.0",
		Description: "Slack workspace integration for messaging, channels, and user management",
		Capabilities: []string{
			"messaging",
			"channels",
			"users",
			"files",
			"reactions",
			"search",
		},
	}
}

// ListTools returns available tools.
func (a *SlackAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "slack_post_message",
			Description: "Post a message to a Slack channel",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{
						"type":        "string",
						"description": "Channel ID or name",
					},
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Message text",
					},
					"thread_ts": map[string]interface{}{
						"type":        "string",
						"description": "Thread timestamp for reply",
					},
				},
				"required": []string{"channel", "text"},
			},
		},
		{
			Name:        "slack_update_message",
			Description: "Update an existing message",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{
						"type":        "string",
						"description": "Channel ID",
					},
					"ts": map[string]interface{}{
						"type":        "string",
						"description": "Message timestamp",
					},
					"text": map[string]interface{}{
						"type":        "string",
						"description": "New message text",
					},
				},
				"required": []string{"channel", "ts", "text"},
			},
		},
		{
			Name:        "slack_delete_message",
			Description: "Delete a message",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{
						"type":        "string",
						"description": "Channel ID",
					},
					"ts": map[string]interface{}{
						"type":        "string",
						"description": "Message timestamp",
					},
				},
				"required": []string{"channel", "ts"},
			},
		},
		{
			Name:        "slack_list_channels",
			Description: "List Slack channels",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"types": map[string]interface{}{
						"type":        "string",
						"description": "Channel types (public_channel, private_channel, mpim, im)",
						"default":     "public_channel",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of channels to return",
						"default":     100,
					},
				},
			},
		},
		{
			Name:        "slack_create_channel",
			Description: "Create a new channel",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Channel name",
					},
					"is_private": map[string]interface{}{
						"type":        "boolean",
						"description": "Create as private channel",
						"default":     false,
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "slack_archive_channel",
			Description: "Archive a channel",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{
						"type":        "string",
						"description": "Channel ID",
					},
				},
				"required": []string{"channel"},
			},
		},
		{
			Name:        "slack_invite_to_channel",
			Description: "Invite a user to a channel",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{
						"type":        "string",
						"description": "Channel ID",
					},
					"user": map[string]interface{}{
						"type":        "string",
						"description": "User ID",
					},
				},
				"required": []string{"channel", "user"},
			},
		},
		{
			Name:        "slack_list_users",
			Description: "List workspace users",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of users to return",
						"default":     100,
					},
				},
			},
		},
		{
			Name:        "slack_add_reaction",
			Description: "Add a reaction to a message",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{
						"type":        "string",
						"description": "Channel ID",
					},
					"ts": map[string]interface{}{
						"type":        "string",
						"description": "Message timestamp",
					},
					"emoji": map[string]interface{}{
						"type":        "string",
						"description": "Emoji name (without colons)",
					},
				},
				"required": []string{"channel", "ts", "emoji"},
			},
		},
		{
			Name:        "slack_upload_file",
			Description: "Upload a file to Slack",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Channel IDs to share to",
					},
					"filename": map[string]interface{}{
						"type":        "string",
						"description": "Filename",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "File content",
					},
				},
				"required": []string{"channels", "filename", "content"},
			},
		},
		{
			Name:        "slack_search_messages",
			Description: "Search for messages",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results",
						"default":     20,
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "slack_get_history",
			Description: "Get channel message history",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{
						"type":        "string",
						"description": "Channel ID",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of messages",
						"default":     50,
					},
				},
				"required": []string{"channel"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *SlackAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "slack_post_message":
		return a.postMessage(ctx, args)
	case "slack_update_message":
		return a.updateMessage(ctx, args)
	case "slack_delete_message":
		return a.deleteMessage(ctx, args)
	case "slack_list_channels":
		return a.listChannels(ctx, args)
	case "slack_create_channel":
		return a.createChannel(ctx, args)
	case "slack_archive_channel":
		return a.archiveChannel(ctx, args)
	case "slack_invite_to_channel":
		return a.inviteToChannel(ctx, args)
	case "slack_list_users":
		return a.listUsers(ctx, args)
	case "slack_add_reaction":
		return a.addReaction(ctx, args)
	case "slack_upload_file":
		return a.uploadFile(ctx, args)
	case "slack_search_messages":
		return a.searchMessages(ctx, args)
	case "slack_get_history":
		return a.getHistory(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *SlackAdapter) postMessage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	channel, _ := args["channel"].(string)
	text, _ := args["text"].(string)
	threadTS, _ := args["thread_ts"].(string)

	options := MessageOptions{
		ThreadTS: threadTS,
	}

	ts, err := a.client.PostMessage(ctx, channel, text, options)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Message posted (ts: %s)", ts)}},
	}, nil
}

func (a *SlackAdapter) updateMessage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	channel, _ := args["channel"].(string)
	ts, _ := args["ts"].(string)
	text, _ := args["text"].(string)

	err := a.client.UpdateMessage(ctx, channel, ts, text)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Message updated"}},
	}, nil
}

func (a *SlackAdapter) deleteMessage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	channel, _ := args["channel"].(string)
	ts, _ := args["ts"].(string)

	err := a.client.DeleteMessage(ctx, channel, ts)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Message deleted"}},
	}, nil
}

func (a *SlackAdapter) listChannels(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	types, _ := args["types"].(string)
	if types == "" {
		types = "public_channel"
	}
	limit := getIntArg(args, "limit", 100)

	channels, err := a.client.ListChannels(ctx, types, limit)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d channels:\n\n", len(channels)))

	for _, ch := range channels {
		icon := "#"
		if ch.IsPrivate {
			icon = "🔒"
		}
		sb.WriteString(fmt.Sprintf("%s %s (%s) - %d members\n", icon, ch.Name, ch.ID, ch.NumMembers))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SlackAdapter) createChannel(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	isPrivate, _ := args["is_private"].(bool)

	channel, err := a.client.CreateChannel(ctx, name, isPrivate)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created channel #%s (ID: %s)", channel.Name, channel.ID)}},
	}, nil
}

func (a *SlackAdapter) archiveChannel(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	channel, _ := args["channel"].(string)

	err := a.client.ArchiveChannel(ctx, channel)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Archived channel %s", channel)}},
	}, nil
}

func (a *SlackAdapter) inviteToChannel(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	channel, _ := args["channel"].(string)
	user, _ := args["user"].(string)

	err := a.client.InviteToChannel(ctx, channel, user)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invited %s to channel %s", user, channel)}},
	}, nil
}

func (a *SlackAdapter) listUsers(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	limit := getIntArg(args, "limit", 100)

	users, err := a.client.ListUsers(ctx, limit)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d users:\n\n", len(users)))

	for _, u := range users {
		icon := "👤"
		if u.IsBot {
			icon = "🤖"
		} else if u.IsAdmin {
			icon = "⭐"
		}
		sb.WriteString(fmt.Sprintf("%s %s (%s) - %s\n", icon, u.RealName, u.Name, u.ID))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SlackAdapter) addReaction(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	channel, _ := args["channel"].(string)
	ts, _ := args["ts"].(string)
	emoji, _ := args["emoji"].(string)

	err := a.client.AddReaction(ctx, channel, ts, emoji)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Added :%s: reaction", emoji)}},
	}, nil
}

func (a *SlackAdapter) uploadFile(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	channelsRaw, _ := args["channels"].([]interface{})
	filename, _ := args["filename"].(string)
	content, _ := args["content"].(string)

	var channels []string
	for _, c := range channelsRaw {
		if s, ok := c.(string); ok {
			channels = append(channels, s)
		}
	}

	file, err := a.client.UploadFile(ctx, channels, filename, content)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Uploaded file: %s (ID: %s)", file.Name, file.ID)}},
	}, nil
}

func (a *SlackAdapter) searchMessages(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	count := getIntArg(args, "count", 20)

	messages, err := a.client.SearchMessages(ctx, query, count)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d messages matching '%s':\n\n", len(messages), query))

	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", m.Channel, m.User, truncate(m.Text, 100)))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *SlackAdapter) getHistory(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	channel, _ := args["channel"].(string)
	limit := getIntArg(args, "limit", 50)

	messages, err := a.client.GetConversationHistory(ctx, channel, limit)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Last %d messages in channel:\n\n", len(messages)))

	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", m.TS, m.User, truncate(m.Text, 200)))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
