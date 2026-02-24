// Package debate provides the AI Debate Orchestrator Framework for HelixAgent.
//
// This package implements a sophisticated multi-agent debate system with learning
// and knowledge management capabilities, enabling consensus-building across
// multiple LLM providers. The framework implements full specification compliance
// including Reflexion, adversarial dynamics, dehallucination, and self-evolvement.
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
//	│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌─────────────┐ │
//	│  │ Reflexion │ │  Gates    │ │Evaluation │ │   Audit     │ │
//	│  └───────────┘ └───────────┘ └───────────┘ └─────────────┘ │
//	└─────────────────────────────────────────────────────────────┘
//
// # Components
//
// The framework consists of 13 packages:
//
//   - debate: Core types and interfaces
//   - debate/agents: Agent factory, pool, 23 templates, 21 roles, adversarial dynamics
//   - debate/topology: Graph mesh, star, chain, tree topologies; CPDE/DPDE planning styles
//   - debate/protocol: 8-phase debate execution with dehallucination and self-evolvement
//   - debate/voting: 6 voting methods (Weighted, Majority, Borda, Condorcet, Plurality, Unanimous)
//   - debate/cognitive: Reasoning and analysis patterns
//   - debate/knowledge: Repository, lessons, patterns
//   - debate/reflexion: Episodic memory, reflection generator, retry-and-learn loop, accumulated wisdom
//   - debate/gates: Configurable approval gates with REST API
//   - debate/evaluation: Benchmark bridge (SWE-bench, HumanEval, MMLU), static code analysis
//   - debate/audit: Provenance tracking with 14 event types for full reproducibility
//   - debate/tools: Git worktree tool, CI/CD hooks for validation pipelines
//   - debate/orchestrator: Main orchestrator and integration
//
// # Debate Topologies
//
// Supported topologies for agent communication:
//
//   - Mesh: All agents communicate with each other (parallel)
//   - Star: Hub-spoke pattern with central coordinator
//   - Chain: Sequential discussion (one after another)
//   - Tree: Hierarchical parent-child with subtree parallelism and rebalancing
//
// # Protocol Phases (8-Phase)
//
// Debates proceed through 8 defined phases:
//
//  1. Dehallucination: ChatDev-inspired proactive clarification before generation
//  2. SelfEvolvement: Agents self-test and refine solutions
//  3. Proposal: Agents submit initial positions
//  4. Critique: Agents critique other positions
//  5. Review: Self-review and refinement
//  6. Optimization: Performance and quality optimization
//  7. Adversarial: Red/Blue team attack-defend cycles
//  8. Convergence: Final consensus building
//
// # Voting Methods
//
// Six methods for determining consensus:
//
//   - Weighted (MiniMax): Score by agent confidence with weighted aggregation
//   - Majority: Simple majority wins
//   - Borda Count: Ranked preference scoring
//   - Condorcet: Pairwise comparison with cycle detection and Borda fallback
//   - Plurality: Most first-choice votes wins
//   - Unanimous: Requires complete agreement
//   - AutoSelectMethod: Selects optimal method based on candidate/voter count
//
// # Reflexion Framework
//
// Learn from execution failures:
//
//	memory := reflexion.NewEpisodicMemoryBuffer(100)
//	generator := reflexion.NewReflectionGenerator(llmClient)
//	loop := reflexion.NewReflexionLoop(memory, generator, codeGen, testExec)
//
//	// Run retry-and-learn cycle
//	result, err := loop.Run(ctx, task, maxAttempts)
//
// # Approval Gates
//
// Human-in-the-loop decision points (disabled by default):
//
//	gate := gates.NewApprovalGate(gates.GateConfig{Enabled: true})
//	gate.CheckGate(ctx, "proposal_review", debateData)
//	// REST API: POST /v1/debates/:id/approve, POST /v1/debates/:id/reject
//
// # Agent Roles (21)
//
// Original: Analyst, Critic, Synthesizer, Optimizer, Validator, Architect,
// Developer, Tester, SecurityExpert, DomainExpert, Moderator, Researcher,
// EthicsAdvisor, PerformanceEngineer, UserAdvocate, DevilsAdvocate
// New: Compiler, Executor, Judge, Implementer, Designer
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
// # Provenance & Audit
//
// Track all debate events for reproducibility:
//
//	tracker := audit.NewProvenanceTracker()
//	tracker.Record(sessionID, &audit.AuditEntry{
//	    EventType: audit.EventPhaseStarted,
//	    Phase:     "proposal",
//	})
//	trail := tracker.GetAuditTrail(sessionID)
//
// # Configuration
//
//	config := orchestrator.DefaultServiceIntegrationConfig()
//	config.EnableNewFramework = true       // Enable new system
//	config.FallbackToLegacy = true         // Fall back on failure
//	config.EnableLearning = true           // Enable learning
//	config.MinAgentsForNewFramework = 3    // Minimum agents required
//
// # Key Files
//
//   - orchestrator/orchestrator.go: Main orchestrator
//   - orchestrator/service_integration.go: Services bridge
//   - agents/factory.go: Agent creation and pooling
//   - agents/adversarial.go: Red/Blue team adversarial dynamics
//   - knowledge/repository.go: Knowledge management
//   - protocol/protocol.go: 8-phase debate protocol execution
//   - protocol/dehallucination.go: Proactive clarification phase
//   - protocol/self_evolvement.go: Self-test and refinement phase
//   - topology/tree.go: Tree topology with subtree parallelism
//   - topology/planning.go: CPDE/DPDE planning styles
//   - voting/weighted_voting.go: 6 voting method implementations
//   - reflexion/episodic_memory.go: Episode buffer with FIFO eviction
//   - reflexion/reflection_generator.go: LLM-based reflection
//   - reflexion/reflexion_loop.go: Retry-and-learn execution loop
//   - reflexion/accumulated_wisdom.go: Cross-session wisdom
//   - gates/approval_gate.go: Human-in-the-loop gates
//   - evaluation/benchmark_bridge.go: SWE-bench/HumanEval/MMLU bridge
//   - audit/provenance.go: Full audit trail with 14 event types
//   - tools/git_tool.go: Git worktree management
//   - tools/cicd_hook.go: CI/CD validation hooks
//
// # Database Tables
//
//   - debate_sessions: Session lifecycle tracking (id, status, config, timestamps)
//   - debate_turns: Turn-level state for replay/recovery
//   - code_versions: Code snapshots at debate milestones
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
