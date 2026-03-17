# HelixAgent Android SDK

Kotlin client library for the HelixAgent Protocol Enhancement API, providing coroutine-based async access to chat completions, streaming, MCP tools, AI debates, embeddings, vision, and RAG endpoints.

## Platform Requirements

- Android API 21+ (Android 5.0 Lollipop)
- Kotlin 1.6+
- Kotlin Coroutines
- OkHttp 4.x

## Installation

Add the SDK file to your project and include OkHttp as a dependency:

```kotlin
// build.gradle.kts
dependencies {
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.7.3")
}
```

Copy `SuperAgent.kt` into your project under `com.helixagent.protocol`.

## Quick Start

```kotlin
import com.helixagent.protocol.HelixAgentClient
import com.helixagent.protocol.ChatMessage

val client = HelixAgentClient(
    baseUrl = "http://your-server:8080",
    apiKey = "your-api-key"
)

// Simple chat completion
val response = client.chatCompletion(
    model = "helixagent-ensemble",
    messages = listOf(ChatMessage("user", "Hello, how are you?"))
)
println(response.choices.first().message.content)
```

## API Reference

### HelixAgentClient

```kotlin
class HelixAgentClient(
    baseUrl: String = "http://localhost:8080",
    apiKey: String? = null,
    timeoutSeconds: Long = 30
)
```

### Chat Completions

```kotlin
// Standard completion
suspend fun chatCompletion(
    model: String,
    messages: List<ChatMessage>,
    temperature: Double = 0.7,
    maxTokens: Int = 1000,
    topP: Double = 1.0,
    stop: List<String>? = null
): ChatCompletionResponse

// With ensemble configuration
suspend fun chatCompletionWithEnsemble(
    model: String,
    messages: List<ChatMessage>,
    ensembleConfig: EnsembleConfig,
    temperature: Double = 0.7,
    maxTokens: Int = 1000
): ChatCompletionResponse
```

### Streaming

Returns a Kotlin `Flow` of SSE chunks:

```kotlin
fun chatCompletionStream(
    model: String,
    messages: List<ChatMessage>,
    temperature: Double = 0.7,
    maxTokens: Int = 1000,
    topP: Double = 1.0,
    stop: List<String>? = null
): Flow<ChatCompletionChunk>

fun chatCompletionStreamWithEnsemble(
    model: String,
    messages: List<ChatMessage>,
    ensembleConfig: EnsembleConfig,
    temperature: Double = 0.7,
    maxTokens: Int = 1000
): Flow<ChatCompletionChunk>
```

### MCP Protocol

```kotlin
suspend fun mcpCallTool(serverId: String, toolName: String, parameters: JSONObject = JSONObject()): JSONObject
suspend fun mcpListTools(serverId: String? = null): JSONObject
suspend fun mcpListServers(): JSONObject
```

### AI Debates

```kotlin
suspend fun createDebate(topic: String, participants: List<DebateParticipant>, ...): DebateResponse
suspend fun getDebate(debateId: String): DebateResponse
```

### Additional APIs

Embeddings, Vision, RAG, LSP, ACP, Analytics, Plugins, Templates, Health, and Status endpoints are all available as suspend functions.

## Data Types

- **ChatMessage** -- `role` (String), `content` (String), `name` (String?)
- **ChatCompletionResponse** -- `id`, `model`, `choices`, `usage`
- **ChatCompletionChunk** -- Streaming chunk with delta content
- **EnsembleConfig** -- Ensemble strategy, provider list, voting method
- **DebateParticipant** -- `name`, `provider`, `model`, `position`
- **DebateResponse** -- `id`, `status`, `consensus`, `rounds`

## Error Handling

```kotlin
try {
    val response = client.chatCompletion(model = "gpt-4", messages = messages)
} catch (e: HelixAgentException) {
    println("HTTP ${e.statusCode}: ${e.message}")
} catch (e: IOException) {
    println("Network error: ${e.message}")
}
```

## Streaming Example

```kotlin
client.chatCompletionStream(
    model = "helixagent-ensemble",
    messages = listOf(ChatMessage("user", "Tell me a story"))
).collect { chunk ->
    chunk.choices.firstOrNull()?.delta?.content?.let { content ->
        print(content)
    }
}
```

## Configuration

| Parameter | Default | Description |
|-----------|---------|-------------|
| `baseUrl` | `http://localhost:8080` | HelixAgent server URL |
| `apiKey` | `null` | Bearer token for authentication |
| `timeoutSeconds` | `30` | HTTP request timeout (streaming uses 5 min) |

## Troubleshooting

- **Connection refused**: Verify the HelixAgent server is running at the configured URL
- **401 Unauthorized**: Check your API key is correct
- **Streaming timeout**: Streaming uses a 5-minute timeout; adjust if needed for long responses
- **SSE parsing warnings**: Non-fatal; the SDK logs and continues processing the stream

## License

Proprietary.
