package graphql

import (
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/graphql/types"
)

func TestInitSchema(t *testing.T) {
	err := InitSchema()
	require.NoError(t, err)
	assert.NotNil(t, Schema)
}

func TestSchema_HasQueryType(t *testing.T) {
	err := InitSchema()
	require.NoError(t, err)

	queryType := Schema.QueryType()
	assert.NotNil(t, queryType)
	assert.Equal(t, "Query", queryType.Name())
}

func TestSchema_HasMutationType(t *testing.T) {
	err := InitSchema()
	require.NoError(t, err)

	mutationType := Schema.MutationType()
	assert.NotNil(t, mutationType)
	assert.Equal(t, "Mutation", mutationType.Name())
}

func TestQueryType_HasProviderFields(t *testing.T) {
	fields := QueryType.Fields()

	assert.Contains(t, fields, "providers")
	assert.Contains(t, fields, "provider")
}

func TestQueryType_HasDebateFields(t *testing.T) {
	fields := QueryType.Fields()

	assert.Contains(t, fields, "debates")
	assert.Contains(t, fields, "debate")
}

func TestQueryType_HasTaskFields(t *testing.T) {
	fields := QueryType.Fields()

	assert.Contains(t, fields, "tasks")
	assert.Contains(t, fields, "task")
}

func TestQueryType_HasVerificationFields(t *testing.T) {
	fields := QueryType.Fields()

	assert.Contains(t, fields, "verificationResults")
	assert.Contains(t, fields, "providerScores")
}

func TestMutationType_HasDebateMutations(t *testing.T) {
	fields := MutationType.Fields()

	assert.Contains(t, fields, "createDebate")
	assert.Contains(t, fields, "submitDebateResponse")
}

func TestMutationType_HasTaskMutations(t *testing.T) {
	fields := MutationType.Fields()

	assert.Contains(t, fields, "createTask")
	assert.Contains(t, fields, "cancelTask")
}

func TestMutationType_HasProviderMutations(t *testing.T) {
	fields := MutationType.Fields()

	assert.Contains(t, fields, "refreshProvider")
}

func TestExecuteQuery_Providers(t *testing.T) {
	err := InitSchema()
	require.NoError(t, err)

	query := `{ providers { id name status } }`
	result := ExecuteQuery(query, nil)

	assert.Empty(t, result.Errors)
	assert.NotNil(t, result.Data)
}

func TestExecuteQuery_VerificationResults(t *testing.T) {
	err := InitSchema()
	require.NoError(t, err)

	query := `{ verificationResults { total_providers verified_providers } }`
	result := ExecuteQuery(query, nil)

	assert.Empty(t, result.Errors)
	assert.NotNil(t, result.Data)
}

func TestExecuteQuery_ProviderScores(t *testing.T) {
	err := InitSchema()
	require.NoError(t, err)

	query := `{ providerScores { provider_id overall_score } }`
	result := ExecuteQuery(query, nil)

	assert.Empty(t, result.Errors)
	assert.NotNil(t, result.Data)
}

func TestExecuteQuery_WithVariables(t *testing.T) {
	err := InitSchema()
	require.NoError(t, err)

	query := `query GetProvider($id: ID!) { provider(id: $id) { id name } }`
	variables := map[string]interface{}{
		"id": "test-provider-123",
	}
	result := ExecuteQuery(query, variables)

	// Result may be nil since resolver returns nil, but no errors
	assert.Empty(t, result.Errors)
}

func TestExecuteQuery_InvalidQuery(t *testing.T) {
	err := InitSchema()
	require.NoError(t, err)

	query := `{ nonExistentField }`
	result := ExecuteQuery(query, nil)

	assert.NotEmpty(t, result.Errors)
}

func TestResolveProviders(t *testing.T) {
	params := graphql.ResolveParams{}
	result, err := ResolveProviders(params)

	assert.NoError(t, err)
	assert.IsType(t, []types.Provider{}, result)
}

func TestResolveProvider(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id": "test-id",
		},
	}
	result, err := ResolveProvider(params)

	assert.NoError(t, err)
	assert.Nil(t, result) // Placeholder returns nil
}

func TestResolveDebates(t *testing.T) {
	params := graphql.ResolveParams{}
	result, err := ResolveDebates(params)

	assert.NoError(t, err)
	assert.IsType(t, []types.Debate{}, result)
}

func TestResolveDebate(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id": "test-id",
		},
	}
	result, err := ResolveDebate(params)

	assert.NoError(t, err)
	assert.Nil(t, result) // Placeholder returns nil
}

func TestResolveTasks(t *testing.T) {
	params := graphql.ResolveParams{}
	result, err := ResolveTasks(params)

	assert.NoError(t, err)
	assert.IsType(t, []types.Task{}, result)
}

func TestResolveTask(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id": "test-id",
		},
	}
	result, err := ResolveTask(params)

	assert.NoError(t, err)
	assert.Nil(t, result) // Placeholder returns nil
}

func TestResolveVerificationResults(t *testing.T) {
	params := graphql.ResolveParams{}
	result, err := ResolveVerificationResults(params)

	assert.NoError(t, err)
	assert.IsType(t, &types.VerificationResults{}, result)
}

func TestResolveProviderScores(t *testing.T) {
	params := graphql.ResolveParams{}
	result, err := ResolveProviderScores(params)

	assert.NoError(t, err)
	assert.IsType(t, []types.ProviderScore{}, result)
}

func TestResolveCreateDebate(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"input": map[string]interface{}{
				"topic": "Test topic",
			},
		},
	}
	result, err := ResolveCreateDebate(params)

	assert.NoError(t, err)
	assert.Nil(t, result) // Placeholder returns nil
}

func TestResolveSubmitDebateResponse(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"input": map[string]interface{}{
				"debate_id":      "debate-123",
				"participant_id": "participant-456",
				"content":        "Test response",
			},
		},
	}
	result, err := ResolveSubmitDebateResponse(params)

	assert.NoError(t, err)
	assert.Nil(t, result) // Placeholder returns nil
}

func TestResolveCreateTask(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"input": map[string]interface{}{
				"type":     "background",
				"priority": 5,
			},
		},
	}
	result, err := ResolveCreateTask(params)

	assert.NoError(t, err)
	assert.Nil(t, result) // Placeholder returns nil
}

func TestResolveCancelTask(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id": "task-123",
		},
	}
	result, err := ResolveCancelTask(params)

	assert.NoError(t, err)
	assert.Nil(t, result) // Placeholder returns nil
}

func TestResolveRefreshProvider(t *testing.T) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"id": "provider-123",
		},
	}
	result, err := ResolveRefreshProvider(params)

	assert.NoError(t, err)
	assert.Nil(t, result) // Placeholder returns nil
}

func TestProviderType_Fields(t *testing.T) {
	fields := providerType.Fields()

	expectedFields := []string{
		"id", "name", "type", "status", "score",
		"models", "health_status", "capabilities",
		"created_at", "updated_at",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "Provider type should have field: %s", field)
	}
}

func TestModelType_Fields(t *testing.T) {
	fields := modelType.Fields()

	expectedFields := []string{
		"id", "name", "provider_id", "version",
		"context_window", "max_tokens", "supports_tools",
		"supports_vision", "supports_streaming", "score", "rank",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "Model type should have field: %s", field)
	}
}

func TestDebateType_Fields(t *testing.T) {
	fields := debateType.Fields()

	expectedFields := []string{
		"id", "topic", "status", "participants",
		"rounds", "conclusion", "confidence",
		"created_at", "updated_at", "completed_at",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "Debate type should have field: %s", field)
	}
}

func TestTaskType_Fields(t *testing.T) {
	fields := taskType.Fields()

	expectedFields := []string{
		"id", "type", "status", "priority",
		"progress", "result", "error",
		"created_at", "started_at", "completed_at",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "Task type should have field: %s", field)
	}
}

func TestHealthStatusType_Fields(t *testing.T) {
	fields := healthStatusType.Fields()

	expectedFields := []string{
		"status", "latency_ms", "last_check", "error_message",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "HealthStatus type should have field: %s", field)
	}
}

func TestCapabilitiesType_Fields(t *testing.T) {
	fields := capabilitiesType.Fields()

	expectedFields := []string{
		"chat", "completions", "embeddings", "vision",
		"tool_use", "streaming", "function_calling",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "Capabilities type should have field: %s", field)
	}
}

func TestVerificationResultsType_Fields(t *testing.T) {
	fields := verificationResultsType.Fields()

	expectedFields := []string{
		"total_providers", "verified_providers",
		"total_models", "verified_models",
		"overall_score", "last_verified",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "VerificationResults type should have field: %s", field)
	}
}

func TestProviderScoreType_Fields(t *testing.T) {
	fields := providerScoreType.Fields()

	expectedFields := []string{
		"provider_id", "provider_name", "overall_score",
		"response_speed", "model_efficiency",
		"cost_effectiveness", "capability", "recency",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "ProviderScore type should have field: %s", field)
	}
}

func TestDebateRoundType_Fields(t *testing.T) {
	fields := debateRoundType.Fields()

	expectedFields := []string{
		"id", "debate_id", "round_number",
		"responses", "summary", "created_at",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "DebateRound type should have field: %s", field)
	}
}

func TestParticipantType_Fields(t *testing.T) {
	fields := participantType.Fields()

	expectedFields := []string{
		"id", "provider_id", "model_id",
		"position", "role", "score",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "Participant type should have field: %s", field)
	}
}

func TestResponseType_Fields(t *testing.T) {
	fields := responseType.Fields()

	expectedFields := []string{
		"participant_id", "content", "confidence",
		"token_count", "latency_ms", "created_at",
	}

	for _, field := range expectedFields {
		assert.Contains(t, fields, field, "Response type should have field: %s", field)
	}
}
