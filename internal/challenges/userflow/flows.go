// Package userflow provides HelixAgent-specific user flow
// challenges that exercise all application endpoints using the
// generic Challenges module framework. Each flow simulates a
// real user or QA tester interacting with the system.
package userflow

import (
	"digital.vasic.challenges/pkg/challenge"
	uf "digital.vasic.challenges/pkg/userflow"
)

// --- API Flow Definitions ---

// HealthCheckFlow returns a flow that verifies all public
// health and status endpoints respond correctly.
func HealthCheckFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:           "health_root",
				Method:         "GET",
				Path:           "/health",
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{Type: "not_empty", Target: "body"},
				},
			},
			{
				Name:           "health_enhanced",
				Method:         "GET",
				Path:           "/v1/health",
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{
						Type:   "response_contains",
						Target: "providers",
						Value:  "providers",
					},
				},
			},
			{
				Name:           "monitoring_status",
				Method:         "GET",
				Path:           "/v1/monitoring/status",
				ExpectedStatus: 200,
			},
		},
	}
}

// ProviderDiscoveryFlow returns a flow that lists and inspects
// available LLM providers and models.
func ProviderDiscoveryFlow(token string) uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:           "list_models",
				Method:         "GET",
				Path:           "/v1/models",
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{
						Type:   "response_contains",
						Target: "models",
						Value:  "data",
					},
				},
			},
			{
				Name:           "provider_health",
				Method:         "GET",
				Path:           "/v1/monitoring/provider-health",
				ExpectedStatus: 200,
			},
			{
				Name:           "provider_fallback_chain",
				Method:         "GET",
				Path:           "/v1/monitoring/fallback-chain",
				ExpectedStatus: 200,
			},
			{
				Name:           "monitoring_status",
				Method:         "GET",
				Path:           "/v1/monitoring/status",
				ExpectedStatus: 200,
			},
			{
				Name:           "circuit_breakers",
				Method:         "GET",
				Path:           "/v1/monitoring/circuit-breakers",
				ExpectedStatus: 200,
			},
		},
	}
}

// ChatCompletionFlow returns a flow that exercises the OpenAI-
// compatible chat completion endpoint with a simple prompt.
func ChatCompletionFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:   "chat_completion",
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: `{
					"model": "helixagent-debate",
					"messages": [
						{
							"role": "user",
							"content": "Say hello in one word"
						}
					],
					"max_tokens": 50
				}`,
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{
						Type:   "response_contains",
						Target: "choices",
						Value:  "choices",
					},
				},
			},
		},
	}
}

// StreamingCompletionFlow returns a flow that verifies the
// server-sent events streaming endpoint.
func StreamingCompletionFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:   "streaming_completion",
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: `{
					"model": "helixagent-debate",
					"messages": [
						{
							"role": "user",
							"content": "Count from 1 to 3"
						}
					],
					"stream": true,
					"max_tokens": 50
				}`,
				AcceptedStatuses: []int{200, 201},
				Assertions: []uf.StepAssertion{
					{Type: "not_empty", Target: "body"},
				},
			},
		},
	}
}

// EmbeddingsFlow returns a flow that tests the embeddings
// generation endpoint.
func EmbeddingsFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:   "create_embedding",
				Method: "POST",
				Path:   "/v1/embeddings/generate",
				Body: `{
					"model": "text-embedding-ada-002",
					"input": "Hello world"
				}`,
				AcceptedStatuses: []int{200, 503},
				Assertions: []uf.StepAssertion{
					{Type: "not_empty", Target: "body"},
				},
			},
		},
	}
}

// FormattersFlow returns a flow that exercises the public code
// formatters API endpoint.
func FormattersFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:           "list_formatters",
				Method:         "GET",
				Path:           "/v1/formatters",
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{
						Type:   "response_contains",
						Target: "formatters",
						Value:  "formatters",
					},
				},
			},
			{
				Name:   "format_go_code",
				Method: "POST",
				Path:   "/v1/format",
				Body: `{
					"language": "go",
					"code": "package main\nfunc main(){\nfmt.Println(\"hello\")}"
				}`,
				AcceptedStatuses: []int{200, 503},
			},
		},
	}
}

// DebateFlow returns a flow that creates and monitors a debate
// session.
func DebateFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:   "create_debate",
				Method: "POST",
				Path:   "/v1/debates",
				Body: `{
					"topic": "Implement a simple rate limiter",
					"max_rounds": 2
				}`,
				AcceptedStatuses: []int{200, 201, 503},
				ExtractTo: map[string]string{
					"id": "debate_id",
				},
			},
			{
				Name:             "get_debate_status",
				Method:           "GET",
				Path:             "/v1/debates",
				AcceptedStatuses: []int{200, 503},
			},
		},
	}
}

// MonitoringFlow returns a flow that checks all monitoring and
// observability endpoints.
func MonitoringFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:           "system_status",
				Method:         "GET",
				Path:           "/v1/monitoring/status",
				ExpectedStatus: 200,
			},
			{
				Name:           "circuit_breakers",
				Method:         "GET",
				Path:           "/v1/monitoring/circuit-breakers",
				ExpectedStatus: 200,
			},
			{
				Name:           "provider_health",
				Method:         "GET",
				Path:           "/v1/monitoring/provider-health",
				ExpectedStatus: 200,
			},
			{
				Name:           "fallback_chain",
				Method:         "GET",
				Path:           "/v1/monitoring/fallback-chain",
				ExpectedStatus: 200,
			},
			{
				Name:           "concurrency_stats",
				Method:         "GET",
				Path:           "/v1/monitoring/concurrency",
				ExpectedStatus: 200,
			},
			{
				Name:           "concurrency_alerts",
				Method:         "GET",
				Path:           "/v1/monitoring/concurrency/alerts",
				ExpectedStatus: 200,
			},
		},
	}
}

// MCPProtocolFlow returns a flow that exercises Model Context
// Protocol endpoints.
func MCPProtocolFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:             "mcp_stats",
				Method:           "GET",
				Path:             "/v1/mcp/stats",
				AcceptedStatuses: []int{200, 404},
			},
			{
				Name:             "mcp_adapters_search",
				Method:           "GET",
				Path:             "/v1/mcp/adapters/search",
				AcceptedStatuses: []int{200, 404},
			},
			{
				Name:             "mcp_capabilities",
				Method:           "GET",
				Path:             "/v1/mcp/capabilities",
				AcceptedStatuses: []int{200, 404},
			},
		},
	}
}

// RAGFlow returns a flow that exercises Retrieval-Augmented
// Generation endpoints.
func RAGFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:             "rag_health",
				Method:           "GET",
				Path:             "/v1/rag/health",
				AcceptedStatuses: []int{200, 404, 503},
			},
			{
				Name:   "rag_search",
				Method: "POST",
				Path:   "/v1/rag/search",
				Body: `{
					"query": "How does HelixAgent work?",
					"top_k": 3
				}`,
				AcceptedStatuses: []int{200, 404, 503},
			},
		},
	}
}

// FeatureFlagsFlow returns a flow that checks feature flag
// endpoints (public, no auth required).
func FeatureFlagsFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:           "list_features",
				Method:         "GET",
				Path:           "/v1/features",
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{Type: "not_empty", Target: "body"},
				},
			},
			{
				Name:           "feature_available",
				Method:         "GET",
				Path:           "/v1/features/available",
				ExpectedStatus: 200,
			},
		},
	}
}

// FullSystemFlow returns a comprehensive flow that exercises
// all major subsystems in sequence, simulating a real user
// session from health check through LLM completion.
func FullSystemFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			// Phase 1: Health checks
			{
				Name:           "health",
				Method:         "GET",
				Path:           "/health",
				ExpectedStatus: 200,
			},
			{
				Name:           "health_enhanced",
				Method:         "GET",
				Path:           "/v1/health",
				ExpectedStatus: 200,
			},
			// Phase 2: Discovery
			{
				Name:           "models",
				Method:         "GET",
				Path:           "/v1/models",
				ExpectedStatus: 200,
			},
			{
				Name:           "features",
				Method:         "GET",
				Path:           "/v1/features",
				ExpectedStatus: 200,
			},
			// Phase 3: Monitoring
			{
				Name:           "status",
				Method:         "GET",
				Path:           "/v1/monitoring/status",
				ExpectedStatus: 200,
			},
			{
				Name:           "providers",
				Method:         "GET",
				Path:           "/v1/monitoring/provider-health",
				ExpectedStatus: 200,
			},
			// Phase 4: Formatting (public)
			{
				Name:           "formatters",
				Method:         "GET",
				Path:           "/v1/formatters",
				ExpectedStatus: 200,
			},
			// Phase 5: LLM Completion
			{
				Name:   "completion",
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: `{
					"model": "helixagent-debate",
					"messages": [{"role":"user","content":"Hello"}],
					"max_tokens": 10
				}`,
				AcceptedStatuses: []int{200, 503},
			},
		},
	}
}

// AuthenticationFlow returns a flow that verifies JWT token
// acquisition, authenticated requests, token refresh, and
// rejection of invalid credentials.
func AuthenticationFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:   "login_valid",
				Method: "POST",
				Path:   "/v1/auth/login",
				Body: `{
					"username": "admin",
					"password": "admin"
				}`,
				AcceptedStatuses: []int{200, 501},
				ExtractTo: map[string]string{
					"token": "auth_token",
				},
			},
			{
				Name:             "models_with_auth",
				Method:           "GET",
				Path:             "/v1/models",
				AcceptedStatuses: []int{200},
				Assertions: []uf.StepAssertion{
					{
						Type:   "not_empty",
						Target: "body",
					},
				},
			},
			{
				Name:   "token_refresh",
				Method: "POST",
				Path:   "/v1/auth/refresh",
				Body: `{
					"token": "{{auth_token}}"
				}`,
				AcceptedStatuses: []int{
					200, 401, 404, 501,
				},
			},
			{
				Name:   "models_no_auth",
				Method: "GET",
				Path:   "/v1/models",
				Headers: map[string]string{
					"Authorization": "",
				},
				AcceptedStatuses: []int{
					200, 401,
				},
			},
			{
				Name:   "login_bad_credentials",
				Method: "POST",
				Path:   "/v1/auth/login",
				Body: `{
					"username": "invalid",
					"password": "wrong"
				}`,
				AcceptedStatuses: []int{
					401, 403, 404, 501,
				},
			},
		},
	}
}

// ErrorHandlingFlow returns a flow that validates proper
// error responses for invalid requests, bad payloads, and
// nonexistent endpoints.
func ErrorHandlingFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:   "nonexistent_model",
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: `{
					"model": "no-such-model-xyz",
					"messages": [
						{
							"role": "user",
							"content": "test"
						}
					]
				}`,
				AcceptedStatuses: []int{
					400, 404, 422, 500, 503,
				},
				Assertions: []uf.StepAssertion{
					{
						Type:   "not_empty",
						Target: "body",
					},
				},
			},
			{
				Name:   "invalid_json",
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body:   `{not valid json!!!`,
				AcceptedStatuses: []int{
					400, 422,
				},
			},
			{
				Name:   "empty_messages",
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: `{
					"model": "helixagent-debate",
					"messages": []
				}`,
				AcceptedStatuses: []int{
					400, 422, 500,
				},
			},
			{
				Name:   "nonexistent_endpoint",
				Method: "GET",
				Path:   "/v1/nonexistent-endpoint",
				AcceptedStatuses: []int{
					404, 405,
				},
			},
			{
				Name:   "empty_embedding_input",
				Method: "POST",
				Path:   "/v1/embeddings/generate",
				Body: `{
					"model": "text-embedding-ada-002",
					"input": ""
				}`,
				AcceptedStatuses: []int{
					400, 422, 500, 503,
				},
			},
		},
	}
}

// ConcurrentUsersFlow returns a flow that exercises multiple
// endpoints to verify the system remains stable under
// concurrent access patterns.
func ConcurrentUsersFlow() uf.APIFlow {
	return uf.APIFlow{
		Steps: []uf.APIStep{
			{
				Name:           "baseline_health",
				Method:         "GET",
				Path:           "/health",
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{
						Type:   "not_empty",
						Target: "body",
					},
				},
			},
			{
				Name:           "concurrent_models",
				Method:         "GET",
				Path:           "/v1/models",
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{
						Type:   "not_empty",
						Target: "body",
					},
				},
			},
			{
				Name:           "concurrent_status",
				Method:         "GET",
				Path:           "/v1/monitoring/status",
				ExpectedStatus: 200,
			},
			{
				Name:   "concurrent_completion",
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: `{
					"model": "helixagent-debate",
					"messages": [
						{
							"role": "user",
							"content": "Ping"
						}
					],
					"max_tokens": 10
				}`,
				AcceptedStatuses: []int{200, 503},
			},
			{
				Name:           "post_load_health",
				Method:         "GET",
				Path:           "/v1/health",
				ExpectedStatus: 200,
				Assertions: []uf.StepAssertion{
					{
						Type:   "not_empty",
						Target: "body",
					},
				},
			},
		},
	}
}

// --- Challenge Constructors ---

// NewHealthCheckChallenge creates a challenge verifying all
// health endpoints.
func NewHealthCheckChallenge(
	adapter uf.APIAdapter,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-health-check",
		"HelixAgent Health Check",
		"Verify all health and liveness endpoints",
		nil,
		adapter,
		HealthCheckFlow(),
	)
}

// NewProviderDiscoveryChallenge creates a challenge that
// discovers and inspects LLM providers.
func NewProviderDiscoveryChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-provider-discovery",
		"HelixAgent Provider Discovery",
		"Discover available LLM providers, models, and health",
		deps,
		adapter,
		ProviderDiscoveryFlow(""),
	)
}

// NewChatCompletionChallenge creates a challenge that sends a
// chat completion request.
func NewChatCompletionChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-chat-completion",
		"HelixAgent Chat Completion",
		"Send a chat completion request via OpenAI-compatible API",
		deps,
		adapter,
		ChatCompletionFlow(),
	)
}

// NewStreamingCompletionChallenge creates a challenge that
// tests SSE streaming.
func NewStreamingCompletionChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-streaming-completion",
		"HelixAgent Streaming Completion",
		"Test server-sent events streaming endpoint",
		deps,
		adapter,
		StreamingCompletionFlow(),
	)
}

// NewEmbeddingsChallenge creates a challenge for embeddings.
func NewEmbeddingsChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-embeddings",
		"HelixAgent Embeddings",
		"Generate text embeddings via embeddings API",
		deps,
		adapter,
		EmbeddingsFlow(),
	)
}

// NewFormattersChallenge creates a challenge for code
// formatters.
func NewFormattersChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-formatters",
		"HelixAgent Code Formatters",
		"List and invoke code formatters",
		deps,
		adapter,
		FormattersFlow(),
	)
}

// NewDebateChallenge creates a challenge for the debate system.
func NewDebateChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-debate",
		"HelixAgent AI Debate",
		"Create and monitor multi-agent debate sessions",
		deps,
		adapter,
		DebateFlow(),
	)
}

// NewMonitoringChallenge creates a challenge for the monitoring
// subsystem.
func NewMonitoringChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-monitoring",
		"HelixAgent Monitoring",
		"Verify all monitoring and observability endpoints",
		deps,
		adapter,
		MonitoringFlow(),
	)
}

// NewMCPChallenge creates a challenge for MCP protocol.
func NewMCPChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-mcp-protocol",
		"HelixAgent MCP Protocol",
		"Exercise Model Context Protocol endpoints",
		deps,
		adapter,
		MCPProtocolFlow(),
	)
}

// NewRAGChallenge creates a challenge for RAG.
func NewRAGChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-rag",
		"HelixAgent RAG Pipeline",
		"Test Retrieval-Augmented Generation endpoints",
		deps,
		adapter,
		RAGFlow(),
	)
}

// NewFeatureFlagsChallenge creates a challenge for feature
// flags.
func NewFeatureFlagsChallenge(
	adapter uf.APIAdapter,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-feature-flags",
		"HelixAgent Feature Flags",
		"Check feature flag endpoints",
		nil,
		adapter,
		FeatureFlagsFlow(),
	)
}

// NewAuthenticationChallenge creates a challenge verifying
// JWT token acquisition, authenticated requests, token
// refresh, and invalid credential rejection.
func NewAuthenticationChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-authentication",
		"HelixAgent Authentication",
		"Verify JWT auth: login, token refresh, "+
			"gated endpoints, invalid credentials",
		deps,
		adapter,
		AuthenticationFlow(),
	)
}

// NewErrorHandlingChallenge creates a challenge that validates
// proper error responses for invalid requests, bad payloads,
// missing fields, and nonexistent endpoints.
func NewErrorHandlingChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-error-handling",
		"HelixAgent Error Handling",
		"Validate error responses: bad model, "+
			"invalid JSON, empty fields, 404",
		deps,
		adapter,
		ErrorHandlingFlow(),
	)
}

// NewConcurrentUsersChallenge creates a challenge that
// exercises multiple endpoints to verify system stability
// under concurrent access patterns.
func NewConcurrentUsersChallenge(
	adapter uf.APIAdapter,
	deps []challenge.ID,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-concurrent-users",
		"HelixAgent Concurrent Users",
		"Verify system stability under parallel "+
			"requests to multiple endpoints",
		deps,
		adapter,
		ConcurrentUsersFlow(),
	)
}

// NewFullSystemChallenge creates a comprehensive end-to-end
// challenge that exercises all major subsystems.
func NewFullSystemChallenge(
	adapter uf.APIAdapter,
) *uf.APIFlowChallenge {
	return uf.NewAPIFlowChallenge(
		"helix-full-system",
		"HelixAgent Full System Flow",
		"End-to-end flow: health "+
			"-> discovery -> monitoring "+
			"-> formatting -> completion",
		nil,
		adapter,
		FullSystemFlow(),
	)
}
