package userflow

import (
	"testing"

	"digital.vasic.challenges/pkg/challenge"
)

// BenchmarkNewOrchestrator measures the cost of creating an
// Orchestrator, including registry setup and registration of
// all 22 challenges.
func BenchmarkNewOrchestrator(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		o := NewOrchestrator(
			"http://localhost:7061",
		)
		if o == nil {
			b.Fatal("orchestrator must not be nil")
		}
	}
}

// BenchmarkHealthCheckFlow measures construction of the
// HealthCheckFlow API flow struct.
func BenchmarkHealthCheckFlow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		flow := HealthCheckFlow()
		if len(flow.Steps) == 0 {
			b.Fatal("flow must have steps")
		}
	}
}

// BenchmarkAllFlowConstruction measures the aggregate cost
// of constructing every flow definition in sequence.
func BenchmarkAllFlowConstruction(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = HealthCheckFlow()
		_ = ProviderDiscoveryFlow("")
		_ = ChatCompletionFlow()
		_ = StreamingCompletionFlow()
		_ = EmbeddingsFlow()
		_ = FormattersFlow()
		_ = DebateFlow()
		_ = MonitoringFlow()
		_ = MCPProtocolFlow()
		_ = RAGFlow()
		_ = FeatureFlagsFlow()
		_ = FullSystemFlow()
		_ = AuthenticationFlow()
		_ = ErrorHandlingFlow()
		_ = ConcurrentUsersFlow()
		_ = MultiTurnConversationFlow()
		_ = ToolCallingFlow()
		_ = ProviderFailoverFlow()
		_ = WebSocketStreamingFlow()
		_ = GRPCServiceFlow()
		_ = RateLimitingFlow()
		_ = PaginationFlow()
	}
}

// BenchmarkChallengeConstructors measures the cost of
// creating all 22 challenge objects via their constructors,
// using a mock adapter.
func BenchmarkChallengeConstructors(b *testing.B) {
	var adapter mockAPIAdapter
	healthDep := []challenge.ID{
		"helix-health-check",
	}
	providerDep := []challenge.ID{
		"helix-provider-discovery",
	}
	completionDep := []challenge.ID{
		"helix-chat-completion",
	}
	embeddingsDep := []challenge.ID{
		"helix-embeddings",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewHealthCheckChallenge(&adapter)
		_ = NewFeatureFlagsChallenge(&adapter)
		_ = NewProviderDiscoveryChallenge(
			&adapter, healthDep,
		)
		_ = NewMonitoringChallenge(
			&adapter, healthDep,
		)
		_ = NewFormattersChallenge(
			&adapter, healthDep,
		)
		_ = NewChatCompletionChallenge(
			&adapter, providerDep,
		)
		_ = NewStreamingCompletionChallenge(
			&adapter, completionDep,
		)
		_ = NewEmbeddingsChallenge(
			&adapter, providerDep,
		)
		_ = NewDebateChallenge(
			&adapter, completionDep,
		)
		_ = NewMCPChallenge(
			&adapter, healthDep,
		)
		_ = NewRAGChallenge(
			&adapter, embeddingsDep,
		)
		_ = NewAuthenticationChallenge(
			&adapter, healthDep,
		)
		_ = NewErrorHandlingChallenge(
			&adapter, healthDep,
		)
		_ = NewConcurrentUsersChallenge(
			&adapter, healthDep,
		)
		_ = NewMultiTurnConversationChallenge(
			&adapter, completionDep,
		)
		_ = NewToolCallingChallenge(
			&adapter, completionDep,
		)
		_ = NewProviderFailoverChallenge(
			&adapter, providerDep,
		)
		_ = NewWebSocketStreamingChallenge(
			&adapter, healthDep,
		)
		_ = NewGRPCServiceChallenge(
			&adapter, healthDep,
		)
		_ = NewRateLimitingChallenge(
			&adapter, healthDep,
		)
		_ = NewPaginationChallenge(
			&adapter, healthDep,
		)
		_ = NewFullSystemChallenge(&adapter)
	}
}

// BenchmarkOrchestratorListChallenges measures the cost of
// listing all registered challenge IDs from the orchestrator.
func BenchmarkOrchestratorListChallenges(b *testing.B) {
	o := NewOrchestrator("http://localhost:7061")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ids := o.ListChallenges()
		if len(ids) != 22 {
			b.Fatalf(
				"expected 22 challenges, got %d",
				len(ids),
			)
		}
	}
}

// BenchmarkOrchestratorChallengeCount measures the cost of
// retrieving the total count of registered challenges.
func BenchmarkOrchestratorChallengeCount(b *testing.B) {
	o := NewOrchestrator("http://localhost:7061")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := o.ChallengeCount()
		if n != 22 {
			b.Fatalf(
				"expected 22 challenges, got %d", n,
			)
		}
	}
}

// BenchmarkOrchestratorSummary measures the cost of
// generating the summary string from the orchestrator.
func BenchmarkOrchestratorSummary(b *testing.B) {
	o := NewOrchestrator("http://localhost:7061")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := o.Summary()
		if len(s) == 0 {
			b.Fatal("summary must not be empty")
		}
	}
}

// BenchmarkOrchestratorChallenges measures the cost of
// retrieving the full challenge list from the orchestrator.
func BenchmarkOrchestratorChallenges(b *testing.B) {
	o := NewOrchestrator("http://localhost:7061")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs := o.Challenges()
		if len(cs) != 22 {
			b.Fatalf(
				"expected 22 challenges, got %d",
				len(cs),
			)
		}
	}
}
