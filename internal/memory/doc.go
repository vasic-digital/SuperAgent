// Package memory provides Mem0-style memory management for HelixAgent.
//
// This package implements persistent memory with entity graphs, enabling
// HelixAgent to remember context across conversations and sessions.
//
// # Memory Architecture
//
// The memory system consists of:
//
//   - Short-term memory: Current conversation context
//   - Long-term memory: Persistent facts and knowledge
//   - Entity graph: Relationships between entities
//   - Memory index: Efficient retrieval
//
// # Memory Manager
//
// Central interface for memory operations:
//
//	manager := memory.NewManager(config)
//
//	// Store memory
//	if err := manager.Store(ctx, memoryItem); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Retrieve memories
//	memories, err := manager.Retrieve(ctx, query, limit)
//
//	// Search by entity
//	memories, err := manager.SearchByEntity(ctx, entityID)
//
// # Memory Item
//
// Basic unit of memory storage:
//
//	item := &memory.Item{
//	    ID:        uuid.New().String(),
//	    Content:   "User prefers Python for scripting",
//	    Type:      memory.TypeFact,
//	    Entities:  []string{"user", "python"},
//	    Metadata:  map[string]interface{}{"source": "conversation"},
//	    CreatedAt: time.Now(),
//	}
//
// Memory types:
//   - Fact: Specific pieces of information
//   - Preference: User preferences
//   - Event: Time-based occurrences
//   - Relationship: Entity relationships
//
// # Entity Graph
//
// Track relationships between entities:
//
//	graph := memory.NewEntityGraph()
//
//	// Add relationship
//	graph.AddRelationship("user_123", "prefers", "python")
//
//	// Query relationships
//	relations := graph.GetRelationships("user_123")
//
//	// Find connected entities
//	connected := graph.FindConnected("user_123", 2) // depth 2
//
// # Memory Retrieval
//
// Semantic search for relevant memories:
//
//	retriever := memory.NewSemanticRetriever(embedder, vectorStore)
//
//	// Retrieve by similarity
//	memories, err := retriever.Retrieve(ctx, query, topK)
//
//	// Retrieve with filters
//	memories, err := retriever.RetrieveFiltered(ctx, query, &Filter{
//	    Type:      memory.TypeFact,
//	    Since:     time.Now().Add(-7 * 24 * time.Hour),
//	    Entities:  []string{"user_123"},
//	})
//
// # Memory Consolidation
//
// Merge and summarize memories:
//
//	consolidator := memory.NewConsolidator(llmProvider)
//
//	// Consolidate related memories
//	consolidated, err := consolidator.Consolidate(ctx, memories)
//
//	// Summarize memory set
//	summary, err := consolidator.Summarize(ctx, memories)
//
// # Persistence
//
// Memory persistence backends:
//
//   - PostgreSQL: Relational storage with pgvector
//   - Qdrant: Vector database
//   - Redis: Fast caching layer
//
// # Key Files
//
//   - manager.go: Memory manager
//   - item.go: Memory item definitions
//   - entity_graph.go: Entity relationship graph
//   - retriever.go: Memory retrieval
//   - consolidator.go: Memory consolidation
//   - store.go: Persistence layer
//
// # Configuration
//
//	config := &memory.Config{
//	    MaxShortTermItems:  100,
//	    MaxLongTermItems:   10000,
//	    ConsolidationThreshold: 50,
//	    RetentionDays:      365,
//	    EnableEntityGraph:  true,
//	}
//
// # Example: Conversation Memory
//
//	manager := memory.NewManager(config)
//
//	// Store user preference
//	item := &memory.Item{
//	    Content:  "User prefers concise responses",
//	    Type:     memory.TypePreference,
//	    Entities: []string{userID},
//	}
//	manager.Store(ctx, item)
//
//	// Later: Retrieve for context
//	memories, _ := manager.Retrieve(ctx, "user preferences", 5)
//	context := memory.FormatForLLM(memories)
package memory
