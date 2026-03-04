package userflow

import (
	"context"
	"fmt"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/registry"
	"digital.vasic.challenges/pkg/runner"
	uf "digital.vasic.challenges/pkg/userflow"
)

// Orchestrator registers and executes all HelixAgent user flow
// challenges. It uses the Challenges module's runner for
// dependency ordering, parallel execution, and reporting.
type Orchestrator struct {
	registry registry.Registry
	runner   runner.Runner
	adapter  uf.APIAdapter
	baseURL  string
}

// NewOrchestrator creates an Orchestrator for HelixAgent user
// flow testing. The baseURL should point to the running
// HelixAgent server (e.g., "http://localhost:7061").
func NewOrchestrator(baseURL string) *Orchestrator {
	reg := registry.NewRegistry()
	adapter := uf.NewHTTPAPIAdapter(baseURL)

	o := &Orchestrator{
		registry: reg,
		adapter:  adapter,
		baseURL:  baseURL,
	}

	o.registerChallenges()

	r := runner.NewRunner(
		runner.WithRegistry(reg),
	)
	o.runner = r

	return o
}

// registerChallenges registers all HelixAgent user flow
// challenges with their dependency graph.
func (o *Orchestrator) registerChallenges() {
	healthDep := []challenge.ID{"helix-health-check"}

	// Phase 1: Health (no dependencies)
	_ = o.registry.Register(
		NewHealthCheckChallenge(o.adapter),
	)

	// Phase 2: Feature flags (no dependencies, public)
	_ = o.registry.Register(
		NewFeatureFlagsChallenge(o.adapter),
	)

	// Phase 3: Provider discovery (depends on health)
	_ = o.registry.Register(
		NewProviderDiscoveryChallenge(
			o.adapter, healthDep,
		),
	)

	// Phase 4: Monitoring (depends on health)
	_ = o.registry.Register(
		NewMonitoringChallenge(o.adapter, healthDep),
	)

	// Phase 5: Code formatters (depends on health)
	_ = o.registry.Register(
		NewFormattersChallenge(o.adapter, healthDep),
	)

	// Phase 6: Chat completion (depends on providers)
	providerDep := []challenge.ID{
		"helix-provider-discovery",
	}
	_ = o.registry.Register(
		NewChatCompletionChallenge(
			o.adapter, providerDep,
		),
	)

	// Phase 7: Streaming (depends on completion)
	completionDep := []challenge.ID{
		"helix-chat-completion",
	}
	_ = o.registry.Register(
		NewStreamingCompletionChallenge(
			o.adapter, completionDep,
		),
	)

	// Phase 8: Embeddings (depends on providers)
	_ = o.registry.Register(
		NewEmbeddingsChallenge(
			o.adapter, providerDep,
		),
	)

	// Phase 9: Debate (depends on completion)
	_ = o.registry.Register(
		NewDebateChallenge(o.adapter, completionDep),
	)

	// Phase 10: MCP protocol (depends on health)
	_ = o.registry.Register(
		NewMCPChallenge(o.adapter, healthDep),
	)

	// Phase 11: RAG (depends on embeddings)
	embeddingsDep := []challenge.ID{
		"helix-embeddings",
	}
	_ = o.registry.Register(
		NewRAGChallenge(o.adapter, embeddingsDep),
	)

	// Full system flow (standalone, no deps)
	_ = o.registry.Register(
		NewFullSystemChallenge(o.adapter),
	)
}

// RunAll executes all registered challenges in dependency
// order.
func (o *Orchestrator) RunAll(
	ctx context.Context,
) ([]*challenge.Result, error) {
	config := &challenge.Config{
		ResultsDir: "results/userflow",
		LogsDir:    "logs/userflow",
		Verbose:    true,
	}
	return o.runner.RunAll(ctx, config)
}

// RunByID executes a single challenge by its ID.
func (o *Orchestrator) RunByID(
	ctx context.Context, id string,
) (*challenge.Result, error) {
	config := &challenge.Config{
		ChallengeID: challenge.ID(id),
		ResultsDir:  "results/userflow",
		LogsDir:     "logs/userflow",
		Verbose:     true,
	}
	return o.runner.Run(ctx, challenge.ID(id), config)
}

// ChallengeCount returns the number of registered challenges.
func (o *Orchestrator) ChallengeCount() int {
	return o.registry.Count()
}

// ListChallenges returns all registered challenge IDs.
func (o *Orchestrator) ListChallenges() []string {
	challenges := o.registry.List()
	ids := make([]string, len(challenges))
	for i, c := range challenges {
		ids[i] = string(c.ID())
	}
	return ids
}

// Challenges returns all registered challenges for use by
// the main orchestrator.
func (o *Orchestrator) Challenges() []challenge.Challenge {
	return o.registry.List()
}

// Summary returns a summary string of the orchestrator state.
func (o *Orchestrator) Summary() string {
	return fmt.Sprintf(
		"HelixAgent UserFlow Orchestrator: %d challenges "+
			"registered, base URL: %s",
		o.ChallengeCount(), o.baseURL,
	)
}
