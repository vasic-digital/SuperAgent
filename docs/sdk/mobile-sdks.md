# SuperAgent Mobile SDKs

Cross-platform mobile SDKs for integrating SuperAgent AI capabilities into iOS and Android applications.

## Overview

SuperAgent provides native mobile SDKs for iOS (Swift) and Android (Kotlin) that enable seamless integration of AI-powered features into mobile applications.

## iOS SDK (Swift)

### Installation

#### CocoaPods
```ruby
pod 'SuperAgent', '~> 1.0.0'
```

#### Swift Package Manager
```swift
dependencies: [
    .package(url: "https://github.com/superagent/superagent-ios.git", .upToNextMajor(from: "1.0.0"))
]
```

### Quick Start

```swift
import SuperAgent

class ViewController: UIViewController {
    private var client: SuperAgentClient!

    override func viewDidLoad() {
        super.viewDidLoad()

        // Initialize client
        client = SuperAgentClient(baseURL: "https://api.superagent.ai", apiKey: "your-api-key")
    }

    func generateResponse() async {
        let messages = [
            ChatMessage(role: "user", content: "Explain quantum computing")
        ]

        do {
            let response = try await client.chatCompletion(
                model: "superagent-ensemble",
                messages: messages,
                maxTokens: 500
            )
            displayResponse(response.content)
        } catch {
            print("Error: \(error.localizedDescription)")
        }
    }

    // With ensemble configuration
    func generateEnsembleResponse() async {
        let messages = [
            ChatMessage(role: "user", content: "What is machine learning?")
        ]
        let ensemble = EnsembleConfig(
            strategy: "confidence_weighted",
            minProviders: 2,
            confidenceThreshold: 0.8
        )

        do {
            let response = try await client.chatCompletionWithEnsemble(
                model: "superagent-ensemble",
                messages: messages,
                ensembleConfig: ensemble
            )
            displayResponse(response.content)
        } catch {
            print("Error: \(error)")
        }
    }
}
```

### Advanced Features

```swift
// Streaming responses
let request = ChatCompletionRequest(
    model: "deepseek-chat",
    messages: [ChatMessage(role: .user, content: "Tell me a story")],
    stream: true
)

client.chat.completions.createStream(request: request) { result in
    switch result {
    case .success(let stream):
        stream.onChunk { chunk in
            // Process streaming chunk
            if let content = chunk.choices[0].delta.content {
                self.appendToUI(content)
            }
        }
        stream.onComplete { _ in
            // Handle completion
        }
    case .failure(let error):
        print("Stream error: \(error)")
    }
}
```

## Android SDK (Kotlin)

### Installation

#### Gradle
```kotlin
dependencies {
    implementation 'ai.superagent:sdk:1.0.0'
}
```

### Quick Start

```kotlin
import com.superagent.protocol.SuperAgentClient
import com.superagent.protocol.ChatMessage
import kotlinx.coroutines.launch
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers

class MainActivity : AppCompatActivity() {
    private lateinit var client: SuperAgentClient

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        // Initialize client
        client = SuperAgentClient(
            baseUrl = "https://api.superagent.ai",
            apiKey = "your-api-key"
        )
    }

    private fun generateResponse() {
        CoroutineScope(Dispatchers.IO).launch {
            val messages = listOf(
                ChatMessage(role = "user", content = "Explain quantum computing")
            )

            try {
                val response = client.chatCompletion(
                    model = "superagent-ensemble",
                    messages = messages,
                    maxTokens = 500
                )
                runOnUiThread {
                    displayResponse(response.choices.first().message.content)
                }
            } catch (e: Exception) {
                Log.e("SuperAgent", "Error: ${e.message}")
            }
        }
    }

    // With ensemble configuration
    private fun generateEnsembleResponse() {
        CoroutineScope(Dispatchers.IO).launch {
            val messages = listOf(
                ChatMessage(role = "user", content = "What is machine learning?")
            )
            val ensemble = EnsembleConfig(
                strategy = "confidence_weighted",
                minProviders = 2,
                confidenceThreshold = 0.8
            )

            try {
                val response = client.chatCompletionWithEnsemble(
                    model = "superagent-ensemble",
                    messages = messages,
                    ensembleConfig = ensemble
                )
                runOnUiThread {
                    displayResponse(response.choices.first().message.content)
                }
            } catch (e: Exception) {
                Log.e("SuperAgent", "Error: ${e.message}")
            }
        }
    }
}
```

### Coroutines Support

```kotlin
// Using Kotlin Coroutines
lifecycleScope.launch {
    try {
        val messages = listOf(
            ChatMessage(ChatMessage.Role.USER, "What is machine learning?")
        )

        val request = ChatCompletionRequest(
            model = "superagent-ensemble",
            messages = messages
        )

        val response = client.chat.completions.createAsync(request)
        displayResponse(response.choices[0].message.content)

    } catch (e: Exception) {
        Log.e("SuperAgent", "Error: ${e.message}")
    }
}
```

## AI Debate Integration

### iOS Debate Creation

```swift
// Create participants
let participants = [
    DebateParticipant(
        name: "UserAdvocate",
        role: "User Experience Expert",
        provider: "claude",
        model: "claude-3-haiku"
    ),
    DebateParticipant(
        name: "AIResearcher",
        role: "AI Research Scientist",
        provider: "deepseek",
        model: "deepseek-chat"
    )
]

// Create debate
Task {
    do {
        let debate = try await client.createDebate(
            topic: "Should AI assistants be more opinionated?",
            participants: participants,
            maxRounds: 3,
            strategy: "consensus"
        )
        print("Debate created: \(debate.id)")

        // Wait for completion
        let result = try await client.waitForDebateCompletion(
            debateId: debate.id,
            pollInterval: 5.0
        )
        if result.consensusReached {
            print("Consensus: \(result.finalPosition ?? "")")
        }
    } catch {
        print("Failed: \(error)")
    }
}
```

### Android Debate Creation

```kotlin
CoroutineScope(Dispatchers.IO).launch {
    val participants = listOf(
        DebateParticipant(
            name = "UserAdvocate",
            role = "User Experience Expert",
            llmProvider = "claude",
            llmModel = "claude-3-haiku"
        ),
        DebateParticipant(
            name = "AIResearcher",
            role = "AI Research Scientist",
            llmProvider = "deepseek",
            llmModel = "deepseek-chat"
        )
    )

    try {
        val debate = client.createDebate(
            topic = "Should AI assistants be more opinionated?",
            participants = participants,
            maxRounds = 3,
            strategy = "consensus"
        )
        Log.d("SuperAgent", "Debate created: ${debate.debateId}")

        // Poll for status
        var status = client.getDebateStatus(debate.debateId)
        while (status.status !in listOf("completed", "failed")) {
            delay(5000)
            status = client.getDebateStatus(debate.debateId)
        }

        // Get results
        val result = client.getDebateResults(debate.debateId)
        result.consensus?.let { consensus ->
            if (consensus.reached) {
                Log.d("SuperAgent", "Consensus: ${consensus.finalPosition}")
            }
        }
    } catch (e: Exception) {
        Log.e("SuperAgent", "Failed: ${e.message}")
    }
}
```

## Cross-Platform Features

### Offline Support

```swift
// iOS
if client.isOfflineCapable {
    // Configure offline mode
    client.offlineConfig = OfflineConfig(
        maxCacheSize: 100,
        cacheExpirationHours: 24
    )
}
```

```kotlin
// Android
if (client.isOfflineCapable) {
    client.offlineConfig = OfflineConfig(
        maxCacheSize = 100,
        cacheExpirationHours = 24
    )
}
```

### Background Processing

```swift
// iOS Background Task
let backgroundTask = client.debates.create(config: debateConfig)

backgroundTask.onProgress { progress in
    // Update UI with progress
    DispatchQueue.main.async {
        self.progressView.progress = progress.percentage
    }
}

backgroundTask.onComplete { result in
    // Handle completion
    DispatchQueue.main.async {
        self.displayResults(result)
    }
}
```

```kotlin
// Android WorkManager integration
val debateWork = OneTimeWorkRequest.Builder(DebateWorker::class.java)
    .setInputData(workDataOf("debate_config" to debateConfigJson))
    .build()

WorkManager.getInstance(context).enqueue(debateWork)
```

## Error Handling

### iOS Error Handling

```swift
client.chat.completions.create(request: request) { result in
    switch result {
    case .success(let response):
        // Handle success
        break
    case .failure(let error):
        switch error {
        case .authentication:
            // Handle auth error
            self.showLoginScreen()
        case .rateLimit:
            // Handle rate limit
            self.showRateLimitWarning()
        case .network:
            // Handle network error
            self.showRetryDialog()
        default:
            // Handle other errors
            self.showGenericError()
        }
    }
}
```

### Android Error Handling

```kotlin
try {
    val response = client.chat.completions.createAsync(request)
    // Handle success
} catch (e: AuthenticationException) {
    // Handle auth error
    showLoginScreen()
} catch (e: RateLimitException) {
    // Handle rate limit
    showRateLimitWarning()
} catch (e: NetworkException) {
    // Handle network error
    showRetryDialog()
} catch (e: SuperAgentException) {
    // Handle other errors
    showGenericError()
}
```

## Performance Optimization

### iOS Performance Tips

```swift
// Configure connection pooling
let config = SuperAgentConfig(
    apiKey: "your-api-key",
    maxConnections: 5,
    timeout: 30.0
)

// Use background sessions for large requests
client.backgroundSession = URLSession(
    configuration: .background(withIdentifier: "com.superagent.background")
)
```

### Android Performance Tips

```kotlin
// Configure OkHttp client
val okHttpClient = OkHttpClient.Builder()
    .connectTimeout(30, TimeUnit.SECONDS)
    .readTimeout(30, TimeUnit.SECONDS)
    .writeTimeout(30, TimeUnit.SECONDS)
    .connectionPool(ConnectionPool(5, 5, TimeUnit.MINUTES))
    .build()

val client = SuperAgentClient.Builder()
    .apiKey("your-api-key")
    .httpClient(okHttpClient)
    .build()
```

## Security Best Practices

### API Key Management

```swift
// iOS Keychain
let keychain = KeychainSwift()
keychain.set("your-api-key", forKey: "superagent-api-key")

let config = SuperAgentConfig(
    apiKey: keychain.get("superagent-api-key")!
)
```

```kotlin
// Android EncryptedSharedPreferences
val masterKey = MasterKey.Builder(context)
    .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
    .build()

val sharedPreferences = EncryptedSharedPreferences.create(
    context,
    "superagent_prefs",
    masterKey,
    EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
    EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
)

val client = SuperAgentClient.Builder()
    .apiKey(sharedPreferences.getString("api_key", "")!!)
    .build()
```

### Certificate Pinning

```swift
// iOS Certificate Pinning
let config = SuperAgentConfig(
    apiKey: "your-api-key",
    certificatePinning: CertificatePinning(
        publicKeyHashes: ["your-public-key-hash"]
    )
)
```

```kotlin
// Android Certificate Pinning
val certificatePinner = CertificatePinner.Builder()
    .add("api.superagent.ai", "sha256/your-public-key-hash")
    .build()

val okHttpClient = OkHttpClient.Builder()
    .certificatePinner(certificatePinner)
    .build()
```

## Requirements

### iOS
- iOS 13.0+
- Xcode 13.0+
- Swift 5.5+

### Android
- Android API 21+ (Android 5.0)
- Kotlin 1.6+
- Android Gradle Plugin 7.0+

## Sample Applications

Complete sample applications are available in the SDK repositories:

- [iOS Sample App](https://github.com/superagent/superagent-ios/tree/main/Example)
- [Android Sample App](https://github.com/superagent/superagent-android/tree/main/sample)

## Contributing

We welcome contributions to the mobile SDKs:

1. Fork the respective SDK repository
2. Create a feature branch
3. Add comprehensive tests
4. Ensure all tests pass
5. Submit a pull request

## Support

- [iOS SDK Issues](https://github.com/superagent/superagent-ios/issues)
- [Android SDK Issues](https://github.com/superagent/superagent-android/issues)
- [Documentation](https://docs.superagent.ai/mobile)

## License

MIT License - see LICENSE file for details.