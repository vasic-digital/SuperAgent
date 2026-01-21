# GraphQL Package

The graphql package provides a GraphQL API for HelixAgent, enabling flexible and efficient querying of providers, debates, tasks, and verification data.

## Overview

This package implements a complete GraphQL schema with queries and mutations for interacting with HelixAgent's core functionality. It provides type-safe access to provider information, debate sessions, background tasks, and verification results.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      GraphQL Server                          │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                       Schema                             ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ ││
│  │  │   Queries   │  │  Mutations  │  │  Subscriptions  │ ││
│  │  └─────────────┘  └─────────────┘  └─────────────────┘ ││
│  └─────────────────────────────────────────────────────────┘│
│                           │                                  │
│  ┌────────────────────────▼─────────────────────────────┐  │
│  │                     Resolvers                          │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────┐ │  │
│  │  │Provider │ │ Debate  │ │  Task   │ │Verification │ │  │
│  │  │Resolver │ │Resolver │ │Resolver │ │  Resolver   │ │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│                           │                                  │
│  ┌────────────────────────▼─────────────────────────────┐  │
│  │                   Data Sources                         │  │
│  │  Provider Registry | Debate Service | Task Queue       │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Schema Types

### Provider Type

```graphql
type Provider {
    id: ID!
    name: String!
    type: String!
    status: String!
    healthStatus: String!
    models: [Model!]!
    lastChecked: String
    score: Float
}
```

### Model Type

```graphql
type Model {
    id: ID!
    name: String!
    provider: String!
    contextWindow: Int!
    maxTokens: Int!
    capabilities: [String!]!
    pricing: ModelPricing
}

type ModelPricing {
    inputPerMillion: Float!
    outputPerMillion: Float!
    currency: String!
}
```

### Debate Type

```graphql
type Debate {
    id: ID!
    topic: String!
    status: String!
    participants: [DebateParticipant!]!
    rounds: [DebateRound!]!
    consensus: String
    confidence: Float
    startedAt: String!
    completedAt: String
    duration: Int
}

type DebateParticipant {
    id: ID!
    provider: String!
    model: String!
    role: String!
}

type DebateRound {
    number: Int!
    responses: [DebateResponse!]!
    timestamp: String!
}
```

### Task Type

```graphql
type Task {
    id: ID!
    type: String!
    status: String!
    progress: Float!
    result: String
    error: String
    createdAt: String!
    startedAt: String
    completedAt: String
    metadata: String
}
```

### Verification Types

```graphql
type VerificationResults {
    timestamp: String!
    totalProviders: Int!
    verifiedProviders: Int!
    failedProviders: Int!
    providers: [ProviderVerification!]!
}

type ProviderVerification {
    providerId: String!
    status: String!
    score: Float!
    tests: [TestResult!]!
    timestamp: String!
}

type ProviderScore {
    providerId: String!
    overallScore: Float!
    components: ScoreComponents!
    rank: Int!
}

type ScoreComponents {
    responseSpeed: Float!
    modelEfficiency: Float!
    costEffectiveness: Float!
    capability: Float!
    recency: Float!
}
```

## Queries

### Available Queries

```graphql
type Query {
    # Get all registered providers
    providers: [Provider!]!

    # Get a specific provider by ID
    provider(id: ID!): Provider

    # Get all debates with optional filtering
    debates(status: String, limit: Int): [Debate!]!

    # Get a specific debate by ID
    debate(id: ID!): Debate

    # Get all tasks with optional filtering
    tasks(status: String, type: String, limit: Int): [Task!]!

    # Get a specific task by ID
    task(id: ID!): Task

    # Get verification results
    verificationResults: VerificationResults!

    # Get provider scores sorted by rank
    providerScores: [ProviderScore!]!
}
```

## Mutations

### Available Mutations

```graphql
type Mutation {
    # Create a new debate session
    createDebate(input: CreateDebateInput!): Debate!

    # Submit a response in a debate
    submitDebateResponse(debateId: ID!, participantId: ID!, content: String!): DebateResponse!

    # Create a background task
    createTask(input: CreateTaskInput!): Task!

    # Cancel a running task
    cancelTask(id: ID!): Task!

    # Refresh provider status
    refreshProvider(id: ID!): Provider!

    # Run verification for all providers
    runVerification: VerificationResults!
}

input CreateDebateInput {
    topic: String!
    participantCount: Int
    maxRounds: Int
    timeout: Int
    enableMultiPass: Boolean
}

input CreateTaskInput {
    type: String!
    payload: String!
    priority: Int
}
```

## Usage Examples

### Query Providers

```graphql
query GetProviders {
    providers {
        id
        name
        status
        healthStatus
        score
        models {
            id
            name
            contextWindow
        }
    }
}
```

### Create a Debate

```graphql
mutation StartDebate {
    createDebate(input: {
        topic: "What is the best approach for implementing microservices?"
        participantCount: 5
        maxRounds: 3
        enableMultiPass: true
    }) {
        id
        status
        participants {
            id
            provider
            model
        }
    }
}
```

### Get Debate Results

```graphql
query GetDebateResults($id: ID!) {
    debate(id: $id) {
        id
        topic
        status
        rounds {
            number
            responses {
                participantId
                content
                confidence
            }
        }
        consensus
        confidence
        duration
    }
}
```

### Query Provider Scores

```graphql
query GetProviderRankings {
    providerScores {
        providerId
        overallScore
        rank
        components {
            responseSpeed
            modelEfficiency
            costEffectiveness
            capability
            recency
        }
    }
}
```

### Create Background Task

```graphql
mutation CreateVerificationTask {
    createTask(input: {
        type: "verification"
        payload: "{\"providers\": [\"claude\", \"deepseek\", \"gemini\"]}"
        priority: 2
    }) {
        id
        status
    }
}
```

### Monitor Task Progress

```graphql
query TaskStatus($id: ID!) {
    task(id: $id) {
        id
        status
        progress
        result
        error
    }
}
```

## Go Integration

### Setting Up the Server

```go
import "dev.helix.agent/internal/graphql"

// Create schema
schema := graphql.NewSchema(graphql.SchemaConfig{
    ProviderRegistry: providerRegistry,
    DebateService:    debateService,
    TaskQueue:        taskQueue,
    Verifier:         verifier,
})

// Create handler
handler := graphql.NewHandler(schema, graphql.HandlerConfig{
    Playground: true,  // Enable GraphiQL in dev
    Tracing:    true,  // Enable query tracing
})

// Mount on router
router.POST("/graphql", handler.ServeHTTP)
router.GET("/graphql", handler.PlaygroundHandler)  // GraphiQL
```

### Custom Resolvers

```go
// Add custom resolver
schema.AddResolver("customQuery", func(p graphql.ResolveParams) (interface{}, error) {
    // Custom resolution logic
    return result, nil
})
```

### Middleware Integration

```go
// Add authentication middleware
router.POST("/graphql",
    authMiddleware,
    rateLimitMiddleware,
    handler.ServeHTTP,
)
```

## Configuration

```go
type SchemaConfig struct {
    // Data sources
    ProviderRegistry ProviderRegistry
    DebateService    DebateService
    TaskQueue        TaskQueue
    Verifier         Verifier

    // Options
    MaxQueryDepth    int           // Default: 10
    MaxQueryCost     int           // Default: 1000
    EnableIntrospection bool       // Default: true in dev
    CacheEnabled     bool          // Enable response caching
    CacheTTL         time.Duration // Cache TTL
}
```

## Feature Flags

GraphQL can be enabled/disabled via feature flags:

```go
if features.IsEnabled("graphql.enabled") {
    router.POST("/graphql", graphqlHandler)
}
```

## Error Handling

```graphql
# Errors are returned in standard GraphQL format
{
    "data": null,
    "errors": [
        {
            "message": "Provider not found",
            "locations": [{"line": 2, "column": 3}],
            "path": ["provider"],
            "extensions": {
                "code": "NOT_FOUND",
                "providerId": "unknown-provider"
            }
        }
    ]
}
```

## Testing

```bash
go test -v ./internal/graphql/...
```

### Testing Queries

```go
func TestProvidersQuery(t *testing.T) {
    schema := graphql.NewTestSchema()

    query := `{ providers { id name status } }`
    result := graphql.Do(graphql.Params{
        Schema:        schema,
        RequestString: query,
    })

    require.Empty(t, result.Errors)
    providers := result.Data.(map[string]interface{})["providers"]
    assert.NotEmpty(t, providers)
}
```

## Performance

| Operation | Typical Latency | Notes |
|-----------|----------------|-------|
| Simple Query | < 10ms | Single resolver |
| Complex Query | < 100ms | Multiple resolvers |
| Mutation | < 200ms | Includes side effects |
| Subscription | Streaming | WebSocket-based |

## Security Considerations

1. **Query Depth Limiting**: Prevents deeply nested queries
2. **Query Cost Analysis**: Limits expensive operations
3. **Rate Limiting**: Per-user query limits
4. **Authentication**: Required for mutations
5. **Input Validation**: All inputs validated
