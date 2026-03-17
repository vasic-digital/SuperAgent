# HelixAgent iOS SDK

Swift client library for the HelixAgent Protocol Enhancement API, providing async/await access to chat completions, streaming via AsyncThrowingStream, MCP tools, AI debates, embeddings, vision, and RAG endpoints.

## Platform Requirements

- iOS 15.0+ / macOS 12.0+
- Swift 5.5+ (for async/await and AsyncThrowingStream)
- No external dependencies (uses Foundation URLSession)

## Installation

Copy `SuperAgent.swift` into your Xcode project. No package manager setup needed -- the SDK uses only Foundation framework.

## Quick Start

```swift
import Foundation

let client = HelixAgentClient(
    baseURL: "http://your-server:8080",
    apiKey: "your-api-key"
)

// Simple chat completion
let response = try await client.chatCompletion(
    model: "helixagent-ensemble",
    messages: [ChatMessage(role: "user", content: "Hello!")]
)
print(response.choices.first?.message.content ?? "")
```

## API Reference

### HelixAgentClient

```swift
public class HelixAgentClient {
    public init(baseURL: String = "http://localhost:8080", apiKey: String? = nil, timeout: TimeInterval = 30)
}
```

### Chat Completions

```swift
public func chatCompletion(
    model: String,
    messages: [ChatMessage],
    temperature: Double = 0.7,
    maxTokens: Int = 1000,
    topP: Double = 1.0,
    stop: [String]? = nil
) async throws -> ChatCompletionResponse

public func chatCompletionWithEnsemble(
    model: String,
    messages: [ChatMessage],
    ensembleConfig: EnsembleConfig,
    temperature: Double = 0.7,
    maxTokens: Int = 1000
) async throws -> ChatCompletionResponse
```

### Streaming

Returns an `AsyncThrowingStream` for real-time SSE processing:

```swift
public func chatCompletionStream(
    model: String,
    messages: [ChatMessage],
    temperature: Double = 0.7,
    maxTokens: Int = 1000,
    topP: Double = 1.0,
    stop: [String]? = nil
) -> AsyncThrowingStream<ChatCompletionChunk, Error>

public func chatCompletionStreamWithEnsemble(
    model: String,
    messages: [ChatMessage],
    ensembleConfig: EnsembleConfig,
    temperature: Double = 0.7,
    maxTokens: Int = 1000
) -> AsyncThrowingStream<ChatCompletionChunk, Error>
```

### MCP Protocol

```swift
public func mcpCallTool(serverId: String, toolName: String, parameters: [String: Any] = [:]) async throws -> [String: Any]
public func mcpListTools(serverId: String? = nil) async throws -> [String: Any]
public func mcpListServers() async throws -> [String: Any]
```

### AI Debates

```swift
public func createDebate(
    topic: String,
    participants: [DebateParticipant],
    maxRounds: Int = 3,
    timeout: Int = 300,
    strategy: String = "consensus"
) async throws -> DebateResponse

public func getDebate(debateId: String) async throws -> DebateResponse
```

### Additional APIs

Embeddings, Vision, RAG, LSP, ACP, Analytics, Plugins, Templates, Health, and Status endpoints are all available as async functions.

## Data Types

- **ChatMessage** -- `role` (String), `content` (String), `name` (String?)
- **ChatCompletionResponse** -- `id`, `model`, `choices`, `usage`
- **ChatCompletionChunk** -- Streaming chunk with delta content
- **EnsembleConfig** -- Ensemble strategy, provider list, voting method
- **DebateParticipant** -- `name`, `provider`, `model`, `position`
- **DebateResponse** -- `id`, `status`, `consensus`, `rounds`

## Error Handling

```swift
do {
    let response = try await client.chatCompletion(model: "gpt-4", messages: messages)
} catch HelixAgentError.httpError(let statusCode, let message) {
    print("HTTP \(statusCode): \(message)")
} catch HelixAgentError.invalidResponse {
    print("Invalid response from server")
} catch {
    print("Error: \(error.localizedDescription)")
}
```

### Error Types

```swift
public enum HelixAgentError: Error {
    case invalidResponse
    case httpError(Int, String)
}
```

## Streaming Example

```swift
let stream = client.chatCompletionStream(
    model: "helixagent-ensemble",
    messages: [ChatMessage(role: "user", content: "Tell me a story")]
)

for try await chunk in stream {
    if let content = chunk.choices.first?.delta.content {
        print(content, terminator: "")
    }
}
```

## Configuration

| Parameter | Default | Description |
|-----------|---------|-------------|
| `baseURL` | `http://localhost:8080` | HelixAgent server URL |
| `apiKey` | `nil` | Bearer token for authentication |
| `timeout` | `30` | Request timeout in seconds |

## Troubleshooting

- **App Transport Security**: For HTTP (non-HTTPS) connections, add an ATS exception in `Info.plist`
- **Streaming on background threads**: Ensure streaming is consumed on a background task to avoid blocking the main thread
- **SSE parsing warnings**: Non-fatal; logged to console and processing continues
- **Connection issues**: Verify network permissions in your app's entitlements

## License

Proprietary.
