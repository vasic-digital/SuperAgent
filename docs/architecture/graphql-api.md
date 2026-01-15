# GraphQL API

HelixAgent provides a GraphQL API for efficient, flexible data querying and real-time subscriptions.

## Overview

The GraphQL API offers:
- **Queries** - Read providers, debates, tasks, verification results
- **Mutations** - Create debates, tasks, refresh providers
- **Subscriptions** - Real-time updates via WebSocket

## Endpoint

```
POST /v1/graphql
WS   /v1/graphql/subscriptions
```

## Schema

### Query Types

```graphql
type Query {
    # Provider queries
    providers(filter: ProviderFilter, limit: Int, offset: Int): ProviderConnection!
    provider(id: ID!): Provider

    # Debate queries
    debates(filter: DebateFilter, limit: Int, offset: Int): DebateConnection!
    debate(id: ID!): Debate
    debateRounds(debateId: ID!): [DebateRound!]!

    # Task queries
    tasks(filter: TaskFilter, limit: Int, offset: Int): TaskConnection!
    task(id: ID!): Task

    # Verification queries
    verificationResults: VerificationResults!
    providerScores: [ProviderScore!]!
    modelRankings: [ModelRanking!]!

    # System queries
    health: HealthStatus!
    metrics: SystemMetrics!
}
```

### Mutation Types

```graphql
type Mutation {
    # Debate mutations
    createDebate(input: CreateDebateInput!): Debate!
    submitDebateResponse(debateId: ID!, input: DebateResponseInput!): DebateRound!
    closeDebate(id: ID!): Debate!

    # Task mutations
    createTask(input: CreateTaskInput!): Task!
    cancelTask(id: ID!): Task!
    retryTask(id: ID!): Task!

    # Provider mutations
    refreshProvider(id: ID!): Provider!
    disableProvider(id: ID!): Provider!
    enableProvider(id: ID!): Provider!
}
```

### Subscription Types

```graphql
type Subscription {
    # Real-time debate updates
    debateUpdates(debateId: ID!): DebateUpdate!
    debateRoundAdded(debateId: ID!): DebateRound!

    # Task progress
    taskProgress(taskId: ID!): TaskProgress!
    taskCompleted(taskId: ID!): Task!

    # Provider health
    providerHealth: ProviderHealthUpdate!
    providerStatusChanged(providerId: ID!): Provider!

    # Token streaming
    tokenStream(requestId: ID!): TokenStreamEvent!
}
```

## Core Types

### Provider

```graphql
type Provider {
    id: ID!
    name: String!
    type: ProviderType!
    status: ProviderStatus!
    models: [Model!]!
    score: Float
    capabilities: ProviderCapabilities!
    healthCheck: HealthCheckResult
    createdAt: DateTime!
    updatedAt: DateTime!
}

enum ProviderType {
    API_KEY
    OAUTH
    FREE
}

enum ProviderStatus {
    HEALTHY
    UNHEALTHY
    DEGRADED
    UNKNOWN
}

type ProviderCapabilities {
    supportsStreaming: Boolean!
    supportsTools: Boolean!
    supportsVision: Boolean!
    maxTokens: Int!
    supportedModels: [String!]!
}
```

### Debate

```graphql
type Debate {
    id: ID!
    topic: String!
    participants: [DebateParticipant!]!
    rounds: [DebateRound!]!
    status: DebateStatus!
    result: DebateResult
    multiPassValidation: MultiPassResult
    createdAt: DateTime!
    completedAt: DateTime
}

type DebateParticipant {
    id: ID!
    provider: Provider!
    model: Model!
    position: DebatePosition!
}

type DebateRound {
    number: Int!
    participant: DebateParticipant!
    content: String!
    confidence: Float!
    timestamp: DateTime!
}

enum DebateStatus {
    PENDING
    IN_PROGRESS
    VALIDATING
    POLISHING
    COMPLETED
    FAILED
}

type DebateResult {
    consensus: String
    winner: DebateParticipant
    confidence: Float!
    summary: String!
}
```

### Task

```graphql
type Task {
    id: ID!
    type: TaskType!
    status: TaskStatus!
    priority: Int!
    payload: JSON
    result: JSON
    error: String
    progress: Float
    retryCount: Int!
    createdAt: DateTime!
    startedAt: DateTime
    completedAt: DateTime
}

enum TaskType {
    BACKGROUND
    LLM_REQUEST
    DEBATE
    VERIFICATION
    NOTIFICATION
}

enum TaskStatus {
    PENDING
    QUEUED
    RUNNING
    COMPLETED
    FAILED
    CANCELLED
    STUCK
}
```

## Input Types

### CreateDebateInput

```graphql
input CreateDebateInput {
    topic: String!
    participants: [DebateParticipantInput!]
    enableMultiPassValidation: Boolean = true
    validationConfig: ValidationConfigInput
    maxRounds: Int = 5
    timeout: Int = 300
}

input DebateParticipantInput {
    providerId: ID!
    modelId: ID!
    position: DebatePosition!
}

input ValidationConfigInput {
    enableValidation: Boolean = true
    enablePolish: Boolean = true
    validationTimeout: Int = 120
    polishTimeout: Int = 60
    minConfidenceToSkip: Float = 0.9
    maxValidationRounds: Int = 3
}
```

### CreateTaskInput

```graphql
input CreateTaskInput {
    type: TaskType!
    payload: JSON!
    priority: Int = 5
    maxRetries: Int = 3
    timeout: Int = 300
}
```

### Filter Types

```graphql
input ProviderFilter {
    status: ProviderStatus
    type: ProviderType
    minScore: Float
    hasCapability: String
}

input DebateFilter {
    status: DebateStatus
    topic: String
    participantId: ID
    fromDate: DateTime
    toDate: DateTime
}

input TaskFilter {
    type: TaskType
    status: TaskStatus
    minPriority: Int
    fromDate: DateTime
}
```

## Examples

### Query Providers

```graphql
query GetProviders {
    providers(filter: { status: HEALTHY, minScore: 7.0 }, limit: 10) {
        edges {
            node {
                id
                name
                score
                status
                models {
                    id
                    name
                }
            }
        }
        pageInfo {
            hasNextPage
            endCursor
        }
    }
}
```

### Create Debate

```graphql
mutation CreateDebate {
    createDebate(input: {
        topic: "Should AI systems have consciousness?"
        enableMultiPassValidation: true
        validationConfig: {
            enableValidation: true
            enablePolish: true
            minConfidenceToSkip: 0.9
        }
    }) {
        id
        status
        participants {
            provider {
                name
            }
            position
        }
    }
}
```

### Subscribe to Debate Updates

```graphql
subscription WatchDebate($debateId: ID!) {
    debateUpdates(debateId: $debateId) {
        type
        round {
            number
            participant {
                provider {
                    name
                }
            }
            content
            confidence
        }
        phase
        progress
    }
}
```

### Get Task with Progress

```graphql
query GetTask($taskId: ID!) {
    task(id: $taskId) {
        id
        type
        status
        progress
        result
        error
        createdAt
        completedAt
    }
}
```

## Pagination

GraphQL uses cursor-based pagination:

```graphql
type ProviderConnection {
    edges: [ProviderEdge!]!
    pageInfo: PageInfo!
    totalCount: Int!
}

type ProviderEdge {
    node: Provider!
    cursor: String!
}

type PageInfo {
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
    startCursor: String
    endCursor: String
}
```

### Usage

```graphql
query PaginatedProviders($after: String) {
    providers(limit: 10, after: $after) {
        edges {
            node {
                id
                name
            }
            cursor
        }
        pageInfo {
            hasNextPage
            endCursor
        }
    }
}
```

## Error Handling

### Error Format

```json
{
    "errors": [
        {
            "message": "Provider not found",
            "locations": [{ "line": 2, "column": 3 }],
            "path": ["provider"],
            "extensions": {
                "code": "NOT_FOUND",
                "providerId": "invalid-id"
            }
        }
    ],
    "data": null
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `NOT_FOUND` | Resource not found |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Permission denied |
| `BAD_REQUEST` | Invalid input |
| `RATE_LIMITED` | Too many requests |
| `INTERNAL_ERROR` | Server error |

## Authentication

### JWT Token

```http
POST /v1/graphql
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

### API Key

```http
POST /v1/graphql
X-API-Key: <api-key>
Content-Type: application/json
```

## Rate Limiting

| Operation | Limit |
|-----------|-------|
| Queries | 100/minute |
| Mutations | 30/minute |
| Subscriptions | 10 concurrent |

Headers returned:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642032000
```

## TOON Integration

Request TOON-encoded responses:

```http
POST /v1/graphql
Accept: application/toon+json
Content-Type: application/json

{
    "query": "{ providers { id name status } }"
}
```

Response:
```http
Content-Type: application/toon+json

{
    "da": {
        "providers": [
            {"i": "p1", "n": "Claude", "s": "H"},
            {"i": "p2", "n": "GPT-4", "s": "H"}
        ]
    }
}
```

## WebSocket Subscriptions

### Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/v1/graphql/subscriptions');

ws.onopen = () => {
    // Initialize connection
    ws.send(JSON.stringify({
        type: 'connection_init',
        payload: { authorization: 'Bearer <token>' }
    }));
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    switch (message.type) {
        case 'connection_ack':
            // Start subscription
            ws.send(JSON.stringify({
                id: '1',
                type: 'subscribe',
                payload: {
                    query: `subscription { debateUpdates(debateId: "123") { type phase } }`
                }
            }));
            break;
        case 'next':
            console.log('Update:', message.payload.data);
            break;
        case 'error':
            console.error('Error:', message.payload);
            break;
    }
};
```

### Protocol Messages

| Type | Direction | Description |
|------|-----------|-------------|
| `connection_init` | Client → Server | Initialize connection |
| `connection_ack` | Server → Client | Connection accepted |
| `subscribe` | Client → Server | Start subscription |
| `next` | Server → Client | Subscription data |
| `error` | Server → Client | Subscription error |
| `complete` | Server → Client | Subscription complete |
| `ping` | Both | Keep-alive |
| `pong` | Both | Keep-alive response |

## Query Complexity

Complex queries are limited to prevent abuse:

```graphql
# This query has complexity 25 (under limit of 100)
query {
    providers(limit: 10) {      # 1 + 10 = 11
        edges {
            node {
                models {         # 11 * 1 = 11
                    id
                    name
                }
            }
        }
    }
    debates(limit: 3) {          # 1 + 3 = 4
        edges {
            node {
                id
            }
        }
    }
}
```

Queries exceeding complexity limit return:
```json
{
    "errors": [{
        "message": "Query complexity 150 exceeds limit 100",
        "extensions": { "code": "COMPLEXITY_LIMIT" }
    }]
}
```

## Introspection

Query the schema:

```graphql
query IntrospectionQuery {
    __schema {
        types {
            name
            fields {
                name
                type {
                    name
                }
            }
        }
    }
}
```

Note: Introspection can be disabled in production.

## Testing

```bash
# Run GraphQL unit tests
go test ./internal/graphql/... -v

# Run GraphQL challenge
./challenges/scripts/graphql_integration_challenge.sh

# Interactive testing (GraphQL Playground)
open http://localhost:8080/v1/graphql/playground
```

## Configuration

```yaml
graphql:
  enabled: true
  endpoint: /v1/graphql
  playground: true  # Disable in production
  introspection: true  # Disable in production
  max_complexity: 100
  max_depth: 10
  rate_limit:
    queries: 100
    mutations: 30
    subscriptions: 10
  toon:
    enabled: true
    compression_level: standard
```
