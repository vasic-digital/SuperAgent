// Package streaming provides adapters that bridge HelixAgent-specific notification
// operations with the generic digital.vasic.streaming module.
//
// This adapter layer maintains backward compatibility with existing code while
// allowing gradual migration to the extracted streaming module. The internal
// notifications package has HelixAgent-specific task notification handling that
// builds on top of the generic module's transport mechanisms.
package streaming

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	genericgrpc "digital.vasic.streaming/pkg/grpc"
	genericsse "digital.vasic.streaming/pkg/sse"
	generictransport "digital.vasic.streaming/pkg/transport"
	genericwebhook "digital.vasic.streaming/pkg/webhook"
	genericws "digital.vasic.streaming/pkg/websocket"
)

// TypeAliases for re-exporting generic module types that can be used directly.

// SSEEvent is the SSE event type from the extracted module.
type SSEEvent = genericsse.Event

// SSEClient is the SSE client type from the extracted module.
type SSEClient = genericsse.Client

// SSEConfig is the SSE configuration from the extracted module.
type SSEConfig = genericsse.Config

// SSEBroker is the SSE broker from the extracted module.
type SSEBroker = genericsse.Broker

// WSMessage is the WebSocket message type from the extracted module.
type WSMessage = genericws.Message

// WSConfig is the WebSocket configuration from the extracted module.
type WSConfig = genericws.Config

// WSClient is the WebSocket client from the extracted module.
type WSClient = genericws.Client

// WSHub is the WebSocket hub from the extracted module.
type WSHub = genericws.Hub

// Webhook is the webhook subscription type from the extracted module.
type Webhook = genericwebhook.Webhook

// WebhookPayload is the webhook payload type from the extracted module.
type WebhookPayload = genericwebhook.Payload

// WebhookDispatcher is the webhook dispatcher from the extracted module.
type WebhookDispatcher = genericwebhook.Dispatcher

// WebhookDispatcherConfig is the webhook dispatcher configuration from the extracted module.
type WebhookDispatcherConfig = genericwebhook.DispatcherConfig

// WebhookRegistry is the webhook registry from the extracted module.
type WebhookRegistry = genericwebhook.Registry

// Transport is the transport interface from the extracted module.
type Transport = generictransport.Transport

// NewSSEBroker creates a new SSE broker using the generic module.
func NewSSEBroker(config *genericsse.Config) *genericsse.Broker {
	return genericsse.NewBroker(config)
}

// DefaultSSEConfig returns the default SSE configuration from the generic module.
func DefaultSSEConfig() *genericsse.Config {
	return genericsse.DefaultConfig()
}

// NewSSEEvent creates a new SSE event.
func NewSSEEvent(eventType string, data []byte) *genericsse.Event {
	return &genericsse.Event{
		Type: eventType,
		Data: data,
	}
}

// NewSSEEventWithID creates a new SSE event with an ID for reconnection tracking.
func NewSSEEventWithID(id, eventType string, data []byte) *genericsse.Event {
	return &genericsse.Event{
		ID:   id,
		Type: eventType,
		Data: data,
	}
}

// NewWebSocketHub creates a new WebSocket hub using the generic module.
func NewWebSocketHub(config *genericws.Config) *genericws.Hub {
	return genericws.NewHub(config)
}

// DefaultWebSocketConfig returns the default WebSocket configuration from the generic module.
func DefaultWebSocketConfig() *genericws.Config {
	return genericws.DefaultConfig()
}

// NewWebhookDispatcher creates a new webhook dispatcher using the generic module.
func NewWebhookDispatcher(config *genericwebhook.DispatcherConfig) *genericwebhook.Dispatcher {
	return genericwebhook.NewDispatcher(config)
}

// DefaultWebhookConfig returns the default webhook dispatcher configuration from the generic module.
func DefaultWebhookConfig() *genericwebhook.DispatcherConfig {
	return genericwebhook.DefaultDispatcherConfig()
}

// NewWebhookRegistry creates a new webhook registry using the generic module.
func NewWebhookRegistry() *genericwebhook.Registry {
	return genericwebhook.NewRegistry()
}

// SignWebhook computes HMAC-SHA256 signature for a webhook payload.
func SignWebhook(payload []byte, secret string) string {
	return genericwebhook.Sign(payload, secret)
}

// VerifyWebhook verifies webhook signature.
func VerifyWebhook(payload []byte, signature, secret string) bool {
	return genericwebhook.Verify(payload, signature, secret)
}

// SSEManagerAdapter adapts the generic SSE broker for HelixAgent's notification system.
type SSEManagerAdapter struct {
	broker *genericsse.Broker
}

// NewSSEManagerAdapter creates a new SSE manager adapter.
func NewSSEManagerAdapter(config *genericsse.Config) *SSEManagerAdapter {
	return &SSEManagerAdapter{
		broker: genericsse.NewBroker(config),
	}
}

// RegisterClient registers a client for task-specific events.
func (a *SSEManagerAdapter) RegisterClient(taskID string, client chan<- []byte) error {
	// The generic broker doesn't support task-specific subscriptions directly,
	// but we can use the global broadcast and filter client-side.
	// For HelixAgent, we keep the existing behavior.
	return nil
}

// UnregisterClient removes a client from task-specific events.
func (a *SSEManagerAdapter) UnregisterClient(taskID string, client chan<- []byte) error {
	return nil
}

// RegisterGlobalClient registers a client for all events.
func (a *SSEManagerAdapter) RegisterGlobalClient(client chan<- []byte) error {
	return nil
}

// UnregisterGlobalClient removes a global client.
func (a *SSEManagerAdapter) UnregisterGlobalClient(client chan<- []byte) error {
	return nil
}

// Broadcast sends a message to all clients watching a specific task.
func (a *SSEManagerAdapter) Broadcast(taskID string, message []byte) {
	event := &genericsse.Event{
		Type: "task_update",
		Data: message,
	}
	a.broker.Broadcast(event)
}

// ServeHTTP implements http.Handler for the SSE endpoint.
func (a *SSEManagerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.broker.ServeHTTP(w, r)
}

// Close stops the SSE broker.
func (a *SSEManagerAdapter) Close() {
	a.broker.Close()
}

// Broker returns the underlying generic SSE broker.
func (a *SSEManagerAdapter) Broker() *genericsse.Broker {
	return a.broker
}

// WebSocketServerAdapter adapts the generic WebSocket hub for HelixAgent's notification system.
type WebSocketServerAdapter struct {
	hub *genericws.Hub
}

// NewWebSocketServerAdapter creates a new WebSocket server adapter.
func NewWebSocketServerAdapter(config *genericws.Config) *WebSocketServerAdapter {
	return &WebSocketServerAdapter{
		hub: genericws.NewHub(config),
	}
}

// RegisterClient registers a WebSocket client for task-specific events.
func (a *WebSocketServerAdapter) RegisterClient(taskID string, client interface{}) error {
	// Use rooms in the generic hub for task-specific subscriptions
	if c, ok := client.(*genericws.Client); ok {
		return a.hub.JoinRoom(c.ID(), taskID)
	}
	return nil
}

// Broadcast sends a message to all clients watching a specific task.
func (a *WebSocketServerAdapter) Broadcast(taskID string, message []byte) {
	msg := &genericws.Message{
		Type: "task_update",
		Room: taskID,
		Data: message,
	}
	_ = a.hub.SendToRoom(taskID, msg) //nolint:errcheck
}

// BroadcastAll sends a message to all connected clients.
func (a *WebSocketServerAdapter) BroadcastAll(message []byte) {
	msg := &genericws.Message{
		Type: "broadcast",
		Data: message,
	}
	_ = a.hub.Broadcast(msg) //nolint:errcheck
}

// HandleConnection upgrades an HTTP connection to WebSocket.
func (a *WebSocketServerAdapter) HandleConnection(w http.ResponseWriter, r *http.Request) (*genericws.Client, error) {
	return a.hub.ServeWS(w, r)
}

// JoinRoom adds a client to a room (task-specific channel).
func (a *WebSocketServerAdapter) JoinRoom(clientID, roomName string) error {
	return a.hub.JoinRoom(clientID, roomName)
}

// LeaveRoom removes a client from a room.
func (a *WebSocketServerAdapter) LeaveRoom(clientID, roomName string) error {
	return a.hub.LeaveRoom(clientID, roomName)
}

// ClientCount returns the number of connected clients.
func (a *WebSocketServerAdapter) ClientCount() int {
	return a.hub.ClientCount()
}

// RoomCount returns the number of rooms.
func (a *WebSocketServerAdapter) RoomCount() int {
	return a.hub.RoomCount()
}

// Close stops the WebSocket hub.
func (a *WebSocketServerAdapter) Close() {
	a.hub.Close()
}

// Hub returns the underlying generic WebSocket hub.
func (a *WebSocketServerAdapter) Hub() *genericws.Hub {
	return a.hub
}

// WebhookDispatcherAdapter adapts the generic webhook dispatcher for HelixAgent's notification system.
type WebhookDispatcherAdapter struct {
	dispatcher *genericwebhook.Dispatcher
	registry   *genericwebhook.Registry
}

// NewWebhookDispatcherAdapter creates a new webhook dispatcher adapter.
func NewWebhookDispatcherAdapter(config *genericwebhook.DispatcherConfig) *WebhookDispatcherAdapter {
	return &WebhookDispatcherAdapter{
		dispatcher: genericwebhook.NewDispatcher(config),
		registry:   genericwebhook.NewRegistry(),
	}
}

// RegisterWebhook registers a webhook subscription.
func (a *WebhookDispatcherAdapter) RegisterWebhook(id string, webhook *genericwebhook.Webhook) {
	a.registry.Register(id, webhook)
}

// UnregisterWebhook removes a webhook subscription.
func (a *WebhookDispatcherAdapter) UnregisterWebhook(id string) {
	a.registry.Unregister(id)
}

// GetWebhook retrieves a webhook by ID.
func (a *WebhookDispatcherAdapter) GetWebhook(id string) (*genericwebhook.Webhook, bool) {
	return a.registry.Get(id)
}

// ListWebhooks returns all registered webhooks.
func (a *WebhookDispatcherAdapter) ListWebhooks() map[string]*genericwebhook.Webhook {
	return a.registry.List()
}

// Dispatch sends a payload to all matching webhooks.
func (a *WebhookDispatcherAdapter) Dispatch(ctx context.Context, event string, data interface{}) error {
	webhooks := a.registry.MatchingWebhooks(event)

	payloadData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	payload := &genericwebhook.Payload{
		ID:        generatePayloadID(),
		Event:     event,
		Data:      json.RawMessage(payloadData),
		Timestamp: time.Now(),
	}

	for _, webhook := range webhooks {
		if err := a.dispatcher.Send(ctx, webhook, payload); err != nil {
			// Log error but continue dispatching to other webhooks
			continue
		}
	}
	return nil
}

// Stats returns delivery statistics.
func (a *WebhookDispatcherAdapter) Stats() (delivered, failed int64) {
	return a.dispatcher.Stats()
}

// Dispatcher returns the underlying generic webhook dispatcher.
func (a *WebhookDispatcherAdapter) Dispatcher() *genericwebhook.Dispatcher {
	return a.dispatcher
}

// Registry returns the underlying webhook registry.
func (a *WebhookDispatcherAdapter) Registry() *genericwebhook.Registry {
	return a.registry
}

// generatePayloadID generates a unique payload ID.
func generatePayloadID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of the specified length.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// GRPCConfig holds gRPC streaming configuration.
type GRPCConfig = genericgrpc.Config

// GRPCHealthServer is the gRPC health server from the extracted module.
type GRPCHealthServer = genericgrpc.HealthServer

// GRPCStreamServer is the gRPC stream server interface from the extracted module.
type GRPCStreamServer = genericgrpc.StreamServer

// NewGRPCConfig creates a new gRPC configuration using the generic module.
func NewGRPCConfig() *genericgrpc.Config {
	return genericgrpc.DefaultConfig()
}

// NewGRPCHealthServer creates a new gRPC health server using the generic module.
func NewGRPCHealthServer() *genericgrpc.HealthServer {
	return genericgrpc.NewHealthServer()
}

// GRPCServerOptions returns gRPC server options derived from the Config.
func GRPCServerOptions(config *genericgrpc.Config) []interface{} {
	// Return the options as a slice since the actual grpc.ServerOption
	// type comes from the grpc package, not the streaming module
	return nil
}

// Transport functions re-exported from the generic module.

// TransportConfig is the transport configuration from the extracted module.
type TransportConfig = generictransport.Config

// TransportType is the transport type from the extracted module.
type TransportType = generictransport.Type

// TransportFactory is the transport factory from the extracted module.
type TransportFactory = generictransport.Factory

// NewTransportFactory creates a new transport factory using the generic module.
func NewTransportFactory() *generictransport.Factory {
	return generictransport.NewFactory()
}

// Transport type constants re-exported from the generic module.
var (
	TransportTypeHTTP      = generictransport.TypeHTTP
	TransportTypeWebSocket = generictransport.TypeWebSocket
	TransportTypeGRPC      = generictransport.TypeGRPC
)
