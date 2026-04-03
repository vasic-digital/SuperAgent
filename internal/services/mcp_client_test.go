// Package services provides comprehensive MCP client tests
package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPClientCreation tests creating a new MCP client
func TestMCPClientCreation(t *testing.T) {
	logger := logrus.New()
	client := NewMCPClient(logger)
	
	assert.NotNil(t, client)
	assert.NotNil(t, client.servers)
	assert.NotNil(t, client.tools)
	assert.Equal(t, int64(0), client.messageID.Load()) // Starts at 0, first nextMessageID returns 1
}

// TestMCPClientConnectServer tests connecting to an MCP server
func TestMCPClientConnectServer(t *testing.T) {
	logger := logrus.New()
	client := NewMCPClient(logger)
	
	ctx := context.Background()
	
	// Note: This requires an actual MCP server to be running
	// For unit tests, we'll test the connection logic without actual server
	
	err := client.ConnectServer(ctx, "test-server", "Test Server", "echo", []string{"hello"})
	// This will fail because echo doesn't implement MCP protocol
	// But it tests the connection logic
	assert.Error(t, err) // Expected to fail
}

// TestMCPClientDisconnectServer tests disconnecting from a server
func TestMCPClientDisconnectServer(t *testing.T) {
	logger := logrus.New()
	client := NewMCPClient(logger)
	
	// Try to disconnect from non-existent server
	err := client.DisconnectServer("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// TestMCPClientListServers tests listing connected servers
func TestMCPClientListServers(t *testing.T) {
	logger := logrus.New()
	client := NewMCPClient(logger)
	
	// Initially no servers
	servers := client.ListServers()
	assert.Empty(t, servers)
	
	// Add a mock server manually
	client.mu.Lock()
	client.servers["test"] = &MCPServerConnection{
		ID:        "test",
		Name:      "Test",
		Connected: true,
	}
	client.mu.Unlock()
	
	servers = client.ListServers()
	assert.Len(t, servers, 1)
	assert.Equal(t, "test", servers[0].ID)
}

// TestMCPClientHealthCheck tests health checking
func TestMCPClientHealthCheck(t *testing.T) {
	logger := logrus.New()
	client := NewMCPClient(logger)
	
	ctx := context.Background()
	
	// Add mock servers
	client.mu.Lock()
	client.servers["connected"] = &MCPServerConnection{
		ID:        "connected",
		Connected: true,
		Transport: &mockTransport{connected: true},
	}
	client.servers["disconnected"] = &MCPServerConnection{
		ID:        "disconnected",
		Connected: false,
		Transport: &mockTransport{connected: false},
	}
	client.mu.Unlock()
	
	results := client.HealthCheck(ctx)
	assert.Len(t, results, 2)
	assert.True(t, results["connected"])
	assert.False(t, results["disconnected"])
}

// TestMCPRequestCreation tests creating MCP requests
func TestMCPRequestCreation(t *testing.T) {
	logger := logrus.New()
	client := NewMCPClient(logger)
	
	// Test message ID generation
	id1 := client.nextMessageID()
	id2 := client.nextMessageID()
	
	assert.Equal(t, 1, id1)
	assert.Equal(t, 2, id2)
}

// TestMCPJSONRPCError tests JSON-RPC error handling
func TestMCPJSONRPCError(t *testing.T) {
	err := &JSONRPCError{
		Code:    JSONRPCInvalidParams,
		Message: "Invalid parameters",
		Data:    json.RawMessage(`{"field": "name"}`),
	}
	
	assert.Equal(t, "JSON-RPC Error -32602: Invalid parameters", err.Error())
}

// TestMCPToolTypes tests MCP tool types
func TestMCPToolTypes(t *testing.T) {
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"param1": map[string]string{
					"type": "string",
				},
			},
			Required: []string{"param1"},
		},
	}
	
	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "object", tool.InputSchema.Type)
	assert.Contains(t, tool.InputSchema.Required, "param1")
}

// TestMCPResourceTypes tests MCP resource types
func TestMCPResourceTypes(t *testing.T) {
	resource := Resource{
		URI:         "file:///test.txt",
		Name:        "test.txt",
		Description: "Test file",
		MimeType:    "text/plain",
	}
	
	assert.Equal(t, "file:///test.txt", resource.URI)
	assert.Equal(t, "text/plain", resource.MimeType)
}

// TestMCPPromptTypes tests MCP prompt types
func TestMCPPromptTypes(t *testing.T) {
	prompt := Prompt{
		Name:        "test_prompt",
		Description: "A test prompt",
		Arguments: []PromptArgument{
			{
				Name:     "name",
				Required: true,
			},
		},
	}
	
	assert.Equal(t, "test_prompt", prompt.Name)
	assert.Len(t, prompt.Arguments, 1)
	assert.True(t, prompt.Arguments[0].Required)
}

// TestMCPSamplingTypes tests MCP sampling types
func TestMCPSamplingTypes(t *testing.T) {
	request := CreateMessageRequest{
		Messages: []SamplingMessage{
			{
				Role: "user",
				Content: Content{
					Type: "text",
					Text: "Hello",
				},
			},
		},
		MaxTokens: 100,
	}
	
	assert.Len(t, request.Messages, 1)
	assert.Equal(t, 100, request.MaxTokens)
}

// TestMCPLoggingTypes tests MCP logging types
func TestMCPLoggingTypes(t *testing.T) {
	notification := LoggingMessageNotification{
		Level:  LoggingLevelInfo,
		Logger: "test",
		Data:   "Test message",
	}
	
	assert.Equal(t, LoggingLevelInfo, notification.Level)
	assert.Equal(t, "test", notification.Logger)
}

// TestMCPCapabilities tests MCP capabilities
func TestMCPCapabilities(t *testing.T) {
	caps := Capabilities{
		Tools: &ToolCapabilities{
			ListChanged: true,
		},
		Resources: &ResourceCapabilities{
			Subscribe:   true,
			ListChanged: true,
		},
		Prompts: &PromptCapabilities{
			ListChanged: true,
		},
	}
	
	assert.True(t, caps.Tools.ListChanged)
	assert.True(t, caps.Resources.Subscribe)
	assert.True(t, caps.Prompts.ListChanged)
	
	// Test simplified capabilities for backward compatibility
	simpleCaps := map[string]interface{}{
		"tools": map[string]interface{}{
			"listChanged": true,
		},
	}
	assert.NotNil(t, simpleCaps["tools"])
}

// TestMCPContentTypes tests MCP content types
func TestMCPContentTypes(t *testing.T) {
	// Text content
	textContent := Content{
		Type: "text",
		Text: "Hello, World!",
	}
	assert.Equal(t, "text", textContent.Type)
	assert.Equal(t, "Hello, World!", textContent.Text)
	
	// Image content
	imageContent := Content{
		Type: "image",
		Image: &ImageContent{
			Data:     "base64encoded",
			MimeType: "image/png",
		},
	}
	assert.Equal(t, "image", imageContent.Type)
	assert.Equal(t, "image/png", imageContent.Image.MimeType)
	
	// Resource content
	resourceContent := Content{
		Type: "resource",
		Resource: &EmbeddedResource{
			Type: "text",
			Text: "Resource text",
			URI:  "file:///test.txt",
		},
	}
	assert.Equal(t, "resource", resourceContent.Type)
	assert.Equal(t, "file:///test.txt", resourceContent.Resource.URI)
}

// TestMCPInitialization tests MCP initialization flow
func TestMCPInitialization(t *testing.T) {
	initRequest := InitializeRequest{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
		ClientInfo: map[string]string{
			"name":    "helixagent-test",
			"version": "1.0.0",
		},
	}
	
	assert.Equal(t, MCPProtocolVersion, initRequest.ProtocolVersion)
	assert.Equal(t, "helixagent-test", initRequest.ClientInfo["name"])
	
	initResult := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
		ServerInfo: map[string]string{
			"name":    "test-server",
			"version": "1.0.0",
		},
		Instructions: "Test instructions",
	}
	
	assert.Equal(t, "Test instructions", initResult.Instructions)
}

// TestMCPProgressNotification tests progress notifications
func TestMCPProgressNotification(t *testing.T) {
	notification := ProgressNotification{
		ProgressToken: "token-123",
		Progress:      0.5,
		Total:         1.0,
	}
	
	assert.Equal(t, "token-123", notification.ProgressToken)
	assert.Equal(t, 0.5, notification.Progress)
	assert.Equal(t, 1.0, notification.Total)
}

// TestMCPCancellation tests cancellation
func TestMCPCancellation(t *testing.T) {
	notification := CancelledNotification{
		RequestID: 123,
		Reason:    "User cancelled",
	}
	
	assert.Equal(t, 123, notification.RequestID)
	assert.Equal(t, "User cancelled", notification.Reason)
}

// TestMCPHelixAgentExtensions tests HelixAgent MCP extensions
func TestMCPHelixAgentExtensions(t *testing.T) {
	serverInfo := MCPServerInfo{
		ID:          "test-server",
		Name:        "Test Server",
		Description: "A test MCP server",
		Version:     "1.0.0",
		Package:     "@modelcontextprotocol/server-test",
		Category:    "core",
		CostModel:   "free",
		Enabled:     true,
	}
	
	assert.Equal(t, "test-server", serverInfo.ID)
	assert.Equal(t, "free", serverInfo.CostModel)
	assert.True(t, serverInfo.Enabled)
}

// TestMCPConfiguration tests MCP configuration
func TestMCPConfiguration(t *testing.T) {
	config := MCPConfiguration{
		Enabled:        true,
		Timeout:        30000,
		MaxConcurrent:  10,
		DefaultServers: []string{"filesystem", "github"},
		Servers: map[string]*MCPServerConfig{
			"filesystem": {
				Enabled: true,
				Package: "@modelcontextprotocol/server-filesystem",
				Args:    []string{"/workspace"},
			},
		},
	}
	
	assert.True(t, config.Enabled)
	assert.Equal(t, 30000, config.Timeout)
	assert.Contains(t, config.DefaultServers, "github")
}

// TestMCPConnectionStats tests connection statistics
func TestMCPConnectionStats(t *testing.T) {
	now := time.Now().Unix()
	stats := MCPConnectionStats{
		ServerID:         "test",
		State:            ConnectionStateConnected,
		ConnectedAt:      &now,
		LastActivityAt:   now,
		MessagesSent:     10,
		MessagesReceived: 20,
		Errors:           0,
		LatencyMs:        50,
	}
	
	assert.Equal(t, ConnectionStateConnected, stats.State)
	assert.Equal(t, int64(10), stats.MessagesSent)
	assert.Equal(t, int64(50), stats.LatencyMs)
}

// TestMCPValidation validates tool arguments
func TestMCPValidation(t *testing.T) {
	client := NewMCPClient(nil)
	
	tool := &MCPTool{
		Name:        "test_tool",
		Description: "Test tool",
		InputSchema: map[string]interface{}{
			"required": []interface{}{"required_field"},
		},
	}
	
	// Valid arguments
	validArgs := map[string]interface{}{
		"required_field": "value",
	}
	err := client.validateToolArguments(tool, validArgs)
	assert.NoError(t, err)
	
	// Missing required field
	invalidArgs := map[string]interface{}{
		"other_field": "value",
	}
	err = client.validateToolArguments(tool, invalidArgs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required field")
}

// TestMCPRequestResponseMarshal tests marshaling/unmarshaling
func TestMCPRequestResponseMarshal(t *testing.T) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}
	
	data, err := json.Marshal(request)
	require.NoError(t, err)
	
	var decoded MCPRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	
	assert.Equal(t, "2.0", decoded.JSONRPC)
	// JSON unmarshaling converts numbers to float64
	assert.Equal(t, float64(1), decoded.ID)
	assert.Equal(t, "tools/list", decoded.Method)
}

// mockTransport implements MCPTransport for testing
type mockTransport struct {
	connected bool
}

func (m *mockTransport) Send(ctx context.Context, message interface{}) error {
	return nil
}

func (m *mockTransport) Receive(ctx context.Context) (interface{}, error) {
	return nil, nil
}

func (m *mockTransport) Close() error {
	m.connected = false
	return nil
}

func (m *mockTransport) IsConnected() bool {
	return m.connected
}

// TestMCPTransportInterface tests the transport interface
func TestMCPTransportInterface(t *testing.T) {
	transport := &mockTransport{connected: true}
	
	assert.True(t, transport.IsConnected())
	
	err := transport.Close()
	assert.NoError(t, err)
	assert.False(t, transport.IsConnected())
}

// TestMCPProtocolVersion tests protocol version constant
func TestMCPProtocolVersion(t *testing.T) {
	assert.Equal(t, "2024-11-05", MCPProtocolVersion)
}

// TestMCPLoggingLevels tests all logging levels
func TestMCPLoggingLevels(t *testing.T) {
	levels := []LoggingLevel{
		LoggingLevelDebug,
		LoggingLevelInfo,
		LoggingLevelNotice,
		LoggingLevelWarning,
		LoggingLevelError,
		LoggingLevelCritical,
		LoggingLevelAlert,
		LoggingLevelEmergency,
	}
	
	for _, level := range levels {
		assert.NotEmpty(t, level)
	}
}

// TestMCPConnectionStates tests all connection states
func TestMCPConnectionStates(t *testing.T) {
	states := []MCPConnectionState{
		ConnectionStateDisconnected,
		ConnectionStateConnecting,
		ConnectionStateConnected,
		ConnectionStateError,
	}
	
	for _, state := range states {
		assert.NotEmpty(t, state)
	}
}

// TestMCPToolResult tests tool result handling
func TestMCPToolResult(t *testing.T) {
	result := MCPToolResult{
		Content: []MCPContent{
			{Type: "text", Text: "Result 1"},
			{Type: "text", Text: "Result 2"},
		},
		IsError: false,
	}
	
	assert.Len(t, result.Content, 2)
	assert.False(t, result.IsError)
	assert.Equal(t, "Result 1", result.Content[0].Text)
}

// TestMCPEmbeddedResource tests embedded resource content
func TestMCPEmbeddedResource(t *testing.T) {
	resource := EmbeddedResource{
		Type: "text",
		Text: "Resource content",
		URI:  "file:///test.txt",
	}
	
	assert.Equal(t, "text", resource.Type)
	assert.Equal(t, "file:///test.txt", resource.URI)
}

// TestMCPRoots tests root types
func TestMCPRoots(t *testing.T) {
	root := Root{
		URI:  "file:///workspace",
		Name: "Workspace",
	}
	
	assert.Equal(t, "file:///workspace", root.URI)
	assert.Equal(t, "Workspace", root.Name)
	
	result := ListRootsResult{
		Roots: []Root{root},
	}
	
	assert.Len(t, result.Roots, 1)
}

// TestMCPModelPreferences tests model preferences
func TestMCPModelPreferences(t *testing.T) {
	prefs := ModelPreferences{
		Hints: []ModelHint{
			{Name: "claude-3-opus"},
		},
		CostPriority:         0.3,
		SpeedPriority:        0.4,
		IntelligencePriority: 0.3,
	}
	
	assert.Len(t, prefs.Hints, 1)
	assert.Equal(t, "claude-3-opus", prefs.Hints[0].Name)
	assert.Equal(t, 0.4, prefs.SpeedPriority)
}

// TestMCPListToolsRequest tests list tools request
func TestMCPListToolsRequest(t *testing.T) {
	req := ListToolsRequest{
		Cursor: "cursor-123",
	}
	
	assert.Equal(t, "cursor-123", req.Cursor)
}

// TestMCPListResourcesRequest tests list resources request
func TestMCPListResourcesRequest(t *testing.T) {
	req := ListResourcesRequest{
		Cursor: "cursor-456",
	}
	
	assert.Equal(t, "cursor-456", req.Cursor)
}

// TestMCPReadResourceRequest tests read resource request
func TestMCPReadResourceRequest(t *testing.T) {
	req := ReadResourceRequest{
		URI: "file:///test.txt",
	}
	
	assert.Equal(t, "file:///test.txt", req.URI)
}

// TestMCPSamplingMessage tests sampling message
func TestMCPSamplingMessage(t *testing.T) {
	msg := SamplingMessage{
		Role: "assistant",
		Content: Content{
			Type: "text",
			Text: "Hello!",
		},
	}
	
	assert.Equal(t, "assistant", msg.Role)
	assert.Equal(t, "Hello!", msg.Content.Text)
}

// TestMCPCreateMessageResult tests create message result
func TestMCPCreateMessageResult(t *testing.T) {
	result := CreateMessageResult{
		Role:       "assistant",
		Model:      "claude-3-opus",
		StopReason: "endTurn",
		Content: Content{
			Type: "text",
			Text: "Response text",
		},
	}
	
	assert.Equal(t, "assistant", result.Role)
	assert.Equal(t, "claude-3-opus", result.Model)
	assert.Equal(t, "endTurn", result.StopReason)
}

// TestMCPSetLevelRequest tests set level request
func TestMCPSetLevelRequest(t *testing.T) {
	req := SetLevelRequest{
		Level: LoggingLevelDebug,
	}
	
	assert.Equal(t, LoggingLevelDebug, req.Level)
}

// TestMCPToolCallRequest tests tool call request
func TestMCPToolCallRequest(t *testing.T) {
	req := ToolCallRequest{
		Name: "read_file",
		Arguments: map[string]interface{}{
			"path": "/test.txt",
		},
	}
	
	assert.Equal(t, "read_file", req.Name)
	assert.Equal(t, "/test.txt", req.Arguments["path"])
}

// TestMCPToolCallResult tests tool call result
func TestMCPToolCallResult(t *testing.T) {
	result := ToolCallResult{
		Content: []Content{
			{Type: "text", Text: "File contents"},
		},
		IsError: false,
	}
	
	assert.Len(t, result.Content, 1)
	assert.False(t, result.IsError)
}

// TestMCPToolListChangedNotification tests tool list changed notification
func TestMCPToolListChangedNotification(t *testing.T) {
	notification := ToolListChangedNotification{}
	// Just verify it can be created
	_ = notification
}

// TestMCPResourceUpdatedNotification tests resource updated notification
func TestMCPResourceUpdatedNotification(t *testing.T) {
	notification := ResourceUpdatedNotification{
		URI: "file:///test.txt",
	}
	
	assert.Equal(t, "file:///test.txt", notification.URI)
}

// TestMCPPromptListChangedNotification tests prompt list changed notification
func TestMCPPromptListChangedNotification(t *testing.T) {
	notification := PromptListChangedNotification{}
	_ = notification
}

// TestMCPRootsListChangedNotification tests roots list changed notification
func TestMCPRootsListChangedNotification(t *testing.T) {
	notification := RootsListChangedNotification{}
	_ = notification
}

// TestMCPResourceListChangedNotification tests resource list changed notification
func TestMCPResourceListChangedNotification(t *testing.T) {
	notification := ResourceListChangedNotification{}
	_ = notification
}

// TestMCPGetPromptRequest tests get prompt request
func TestMCPGetPromptRequest(t *testing.T) {
	req := GetPromptRequest{
		Name: "greeting",
		Arguments: map[string]string{
			"name": "World",
		},
	}
	
	assert.Equal(t, "greeting", req.Name)
	assert.Equal(t, "World", req.Arguments["name"])
}

// TestMCPGetPromptResult tests get prompt result
func TestMCPGetPromptResult(t *testing.T) {
	result := GetPromptResult{
		Description: "A greeting prompt",
		Messages: []PromptMessage{
			{
				Role: "user",
				Content: Content{
					Type: "text",
					Text: "Hello, World!",
				},
			},
		},
	}
	
	assert.Equal(t, "A greeting prompt", result.Description)
	assert.Len(t, result.Messages, 1)
	assert.Equal(t, "user", result.Messages[0].Role)
}

// TestMCPSubscribeRequest tests subscribe request
func TestMCPSubscribeRequest(t *testing.T) {
	req := SubscribeRequest{
		URI: "file:///test.txt",
	}
	
	assert.Equal(t, "file:///test.txt", req.URI)
}

// TestMCPUnsubscribeRequest tests unsubscribe request
func TestMCPUnsubscribeRequest(t *testing.T) {
	req := UnsubscribeRequest{
		URI: "file:///test.txt",
	}
	
	assert.Equal(t, "file:///test.txt", req.URI)
}

// TestMCPInitializedNotification tests initialized notification
func TestMCPInitializedNotification(t *testing.T) {
	notification := InitializedNotification{}
	_ = notification
}

// TestMCPResourceContents tests resource contents
func TestMCPResourceContents(t *testing.T) {
	// Text content
	textContent := ResourceContents{
		URI:      "file:///test.txt",
		MimeType: "text/plain",
		Text:     "Hello, World!",
	}
	assert.Equal(t, "Hello, World!", textContent.Text)
	
	// Binary content
	binaryContent := ResourceContents{
		URI:      "file:///test.bin",
		MimeType: "application/octet-stream",
		Blob:     "base64encodeddata",
	}
	assert.Equal(t, "base64encodeddata", binaryContent.Blob)
}

// TestMCPListToolsResult tests list tools result
func TestMCPListToolsResult(t *testing.T) {
	result := ListToolsResult{
		Tools: []Tool{
			{
				Name:        "tool1",
				Description: "First tool",
			},
			{
				Name:        "tool2",
				Description: "Second tool",
			},
		},
		NextCursor: "next-page",
	}
	
	assert.Len(t, result.Tools, 2)
	assert.Equal(t, "next-page", result.NextCursor)
}

// TestMCPListResourcesResult tests list resources result
func TestMCPListResourcesResult(t *testing.T) {
	result := ListResourcesResult{
		Resources: []Resource{
			{
				URI:  "file:///test1.txt",
				Name: "test1.txt",
			},
			{
				URI:  "file:///test2.txt",
				Name: "test2.txt",
			},
		},
		NextCursor: "next-page",
	}
	
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "next-page", result.NextCursor)
}

// TestMCPListPromptsResult tests list prompts result
func TestMCPListPromptsResult(t *testing.T) {
	result := ListPromptsResult{
		Prompts: []Prompt{
			{
				Name:        "prompt1",
				Description: "First prompt",
			},
			{
				Name:        "prompt2",
				Description: "Second prompt",
			},
		},
		NextCursor: "next-page",
	}
	
	assert.Len(t, result.Prompts, 2)
	assert.Equal(t, "next-page", result.NextCursor)
}

// TestMCPReadResourceResult tests read resource result
func TestMCPReadResourceResult(t *testing.T) {
	result := ReadResourceResult{
		Contents: []ResourceContents{
			{
				URI:  "file:///test.txt",
				Text: "File contents",
			},
		},
	}
	
	assert.Len(t, result.Contents, 1)
	assert.Equal(t, "File contents", result.Contents[0].Text)
}

// TestMCPToolInputSchema tests tool input schema
func TestMCPToolInputSchema(t *testing.T) {
	schema := ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name": map[string]string{
				"type": "string",
			},
			"count": map[string]string{
				"type": "integer",
			},
		},
		Required: []string{"name"},
	}
	
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Required, "name")
}

// TestMCPImplementation tests implementation info
func TestMCPImplementation(t *testing.T) {
	impl := Implementation{
		Name:    "helixagent",
		Version: "1.0.0",
	}
	
	assert.Equal(t, "helixagent", impl.Name)
	assert.Equal(t, "1.0.0", impl.Version)
}

// TestMCPClientConcurrency tests concurrent operations
func TestMCPClientConcurrency(t *testing.T) {
	logger := logrus.New()
	client := NewMCPClient(logger)
	
	// Simulate concurrent operations
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			_ = client.nextMessageID()
			done <- true
		}()
	}
	
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
	
	// All message IDs should be unique
	assert.GreaterOrEqual(t, client.messageID.Load(), int64(10))
}
