// Package debate provides the AI Debate Orchestrator Framework for HelixAgent.
//
// This package implements a sophisticated multi-agent debate system with learning
// and knowledge management capabilities, enabling consensus-building across
// multiple LLM providers.
//
// # Architecture Overview
//
//	┌─────────────────────────────────────────────────────────────┐
//	│                     DebateHandler                            │
//	│  ┌─────────────────┐  ┌──────────────────────────────────┐ │
//	│  │ Legacy Services │  │ ServiceIntegration (new)         │ │
//	│  │  DebateService  │  │  ├─ orchestrator                 │ │
//	│  │  AdvancedDebate │  │  ├─ providerRegistry            │ │
//	│  └────────┬────────┘  │  └─ config (feature flags)      │ │
//	│           ↓           └─────────────┬────────────────────┘ │
//	│      Fallback                       ↓                       │
//	└───────────────────────┬─────────────┴───────────────────────┘
//	                        │
//	┌───────────────────────▼─────────────────────────────────────┐
//	│                    Orchestrator                              │
//	│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌─────────────┐ │
//	│  │Agent Pool │ │Team Build │ │ Protocol  │ │  Knowledge  │ │
//	│  └───────────┘ └───────────┘ └───────────┘ └─────────────┘ │
//	└─────────────────────────────────────────────────────────────┘
//
// # Components
//
// The framework consists of 8 packages (~16,650 lines, 500+ tests):
//
//   - debate: Core types and interfaces
//   - debate/agents: Agent factory, pool, templates, specialization
//   - debate/topology: Graph mesh, star, chain topologies
//   - debate/protocol: Phase-based debate execution
//   - debate/voting: Weighted confidence voting
//   - debate/cognitive: Reasoning and analysis patterns
//   - debate/knowledge: Repository, lessons, patterns
//   - debate/orchestrator: Main orchestrator and integration
//
// # Debate Topologies
//
// Supported topologies for agent communication:
//
//   - Mesh: All agents communicate with each other (parallel)
//   - Star: Hub-spoke pattern with central coordinator
//   - Chain: Sequential discussion (one after another)
//
// # Protocol Phases
//
// Debates proceed through defined phases:
//
//  1. Proposal: Agents submit initial positions
//  2. Critique: Agents critique other positions
//  3. Review: Self-review and refinement
//  4. Synthesis: Final consensus building
//
// # Agent Pool
//
// Manage debate participants:
//
//	pool := agents.NewPool(config)
//
//	// Create agents from templates
//	agent := pool.CreateFromTemplate("analyst", providerConfig)
//
//	// Acquire agent for debate
//	agent, err := pool.Acquire(ctx, requirements)
//	defer pool.Release(agent)
//
// # Knowledge Management
//
// Learn from debate outcomes:
//
//	knowledge := knowledge.NewRepository(store)
//
//	// Extract lessons from debate
//	lessons := knowledge.ExtractLessons(ctx, debateResult)
//
//	// Apply learnings to future debates
//	context := knowledge.GetRelevantContext(ctx, topic)
//
// # Configuration
//
//	config := orchestrator.DefaultServiceIntegrationConfig()
//	config.EnableNewFramework = true       // Enable new system
//	config.FallbackToLegacy = true         // Fall back on failure
//	config.EnableLearning = true           // Enable learning
//	config.MinAgentsForNewFramework = 3    // Minimum agents required
//
// # Orchestrator Usage
//
//	orch := orchestrator.New(config, providerRegistry)
//
//	// Run debate
//	result, err := orch.RunDebate(ctx, &DebateRequest{
//	    Topic:       "Should AI systems be transparent?",
//	    Topology:    topology.Mesh,
//	    MaxRounds:   5,
//	    TimeLimit:   10 * time.Minute,
//	})
//
// # Voting Strategies
//
// Determine consensus from agent positions:
//
//   - Weighted confidence: Score by agent confidence
//   - Majority vote: Simple majority wins
//   - Ranked choice: Preference ranking
//   - Consensus threshold: Require minimum agreement
//
// # Key Files
//
//   - orchestrator/orchestrator.go: Main orchestrator
//   - orchestrator/service_integration.go: Services bridge
//   - agents/factory.go: Agent creation and pooling
//   - knowledge/repository.go: Knowledge management
//   - protocol/protocol.go: Debate protocol execution
//   - topology/mesh.go: Mesh topology
//   - topology/star.go: Star topology
//   - topology/chain.go: Chain topology
//   - voting/weighted.go: Voting implementations
//
// # Example: Running a Debate
//
//	// Create orchestrator
//	orch := orchestrator.New(config, registry)
//
//	// Define debate
//	request := &DebateRequest{
//	    Topic: "Best practices for API design",
//	    Agents: []AgentSpec{
//	        {Template: "architect", Provider: "claude"},
//	        {Template: "developer", Provider: "gemini"},
//	        {Template: "reviewer", Provider: "deepseek"},
//	    },
//	    Topology: topology.Mesh,
//	}
//
//	// Run debate
//	result, err := orch.RunDebate(ctx, request)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Consensus: %s\n", result.Consensus)
//	fmt.Printf("Confidence: %.2f\n", result.Confidence)
package debate
