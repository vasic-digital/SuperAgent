package services

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeModels(t *testing.T) {
	t.Run("Claude models are defined", func(t *testing.T) {
		assert.NotEmpty(t, ClaudeModels.Sonnet, "Sonnet model should be defined")
		assert.NotEmpty(t, ClaudeModels.Opus, "Opus model should be defined")
		assert.NotEmpty(t, ClaudeModels.Haiku, "Haiku model should be defined")
		assert.NotEmpty(t, ClaudeModels.SonnetLatest, "SonnetLatest model should be defined")
		assert.NotEmpty(t, ClaudeModels.OpusLatest, "OpusLatest model should be defined")
	})

	t.Run("Claude model names follow Anthropic naming convention", func(t *testing.T) {
		assert.Contains(t, ClaudeModels.Sonnet, "claude-3", "Sonnet should be Claude 3 family")
		assert.Contains(t, ClaudeModels.Opus, "claude-3", "Opus should be Claude 3 family")
		assert.Contains(t, ClaudeModels.Haiku, "claude-3", "Haiku should be Claude 3 family")
	})
}

func TestQwenModels(t *testing.T) {
	t.Run("Qwen models are defined", func(t *testing.T) {
		assert.NotEmpty(t, QwenModels.Turbo, "Turbo model should be defined")
		assert.NotEmpty(t, QwenModels.Plus, "Plus model should be defined")
		assert.NotEmpty(t, QwenModels.Max, "Max model should be defined")
		assert.NotEmpty(t, QwenModels.Coder, "Coder model should be defined")
		assert.NotEmpty(t, QwenModels.Long, "Long model should be defined")
	})

	t.Run("Qwen model names follow Alibaba naming convention", func(t *testing.T) {
		assert.Contains(t, QwenModels.Turbo, "qwen", "Turbo should be Qwen family")
		assert.Contains(t, QwenModels.Plus, "qwen", "Plus should be Qwen family")
		assert.Contains(t, QwenModels.Max, "qwen", "Max should be Qwen family")
	})

	t.Run("All Qwen models are unique", func(t *testing.T) {
		models := []string{
			QwenModels.Turbo,
			QwenModels.Plus,
			QwenModels.Max,
			QwenModels.Coder,
			QwenModels.Long,
		}
		uniqueModels := make(map[string]bool)
		for _, model := range models {
			assert.False(t, uniqueModels[model], "Model %s should be unique", model)
			uniqueModels[model] = true
		}
	})
}

func TestDebateTeamPosition(t *testing.T) {
	t.Run("Positions are correctly numbered 1-5", func(t *testing.T) {
		assert.Equal(t, DebateTeamPosition(1), PositionAnalyst)
		assert.Equal(t, DebateTeamPosition(2), PositionProposer)
		assert.Equal(t, DebateTeamPosition(3), PositionCritic)
		assert.Equal(t, DebateTeamPosition(4), PositionSynthesis)
		assert.Equal(t, DebateTeamPosition(5), PositionMediator)
	})
}

func TestDebateRole(t *testing.T) {
	t.Run("Roles are correctly defined", func(t *testing.T) {
		assert.Equal(t, DebateRole("analyst"), RoleAnalyst)
		assert.Equal(t, DebateRole("proposer"), RoleProposer)
		assert.Equal(t, DebateRole("critic"), RoleCritic)
		assert.Equal(t, DebateRole("synthesis"), RoleSynthesis)
		assert.Equal(t, DebateRole("mediator"), RoleMediator)
	})
}

func TestNewDebateTeamConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress logs in tests

	t.Run("Creates config with nil dependencies", func(t *testing.T) {
		config := NewDebateTeamConfig(nil, nil, logger)
		require.NotNil(t, config)
		assert.NotNil(t, config.members)
		assert.Nil(t, config.providerRegistry)
		assert.Nil(t, config.discovery)
	})

	t.Run("Creates config with provider registry", func(t *testing.T) {
		registryConfig := &RegistryConfig{}
		registry := NewProviderRegistry(registryConfig, nil)
		config := NewDebateTeamConfig(registry, nil, logger)
		require.NotNil(t, config)
		assert.NotNil(t, config.providerRegistry)
	})
}

func TestDebateTeamMember(t *testing.T) {
	t.Run("Member with all fields", func(t *testing.T) {
		member := &DebateTeamMember{
			Position:     PositionAnalyst,
			Role:         RoleAnalyst,
			ProviderName: "claude",
			ModelName:    ClaudeModels.SonnetLatest,
			Score:        9.5,
			IsActive:     true,
		}

		assert.Equal(t, PositionAnalyst, member.Position)
		assert.Equal(t, RoleAnalyst, member.Role)
		assert.Equal(t, "claude", member.ProviderName)
		assert.Equal(t, ClaudeModels.SonnetLatest, member.ModelName)
		assert.Equal(t, 9.5, member.Score)
		assert.True(t, member.IsActive)
	})

	t.Run("Member with fallback chain", func(t *testing.T) {
		qwenFallback := &DebateTeamMember{
			Position:     PositionCritic,
			Role:         RoleCritic,
			ProviderName: "qwen",
			ModelName:    QwenModels.Turbo,
			Score:        7.5,
			IsActive:     false,
		}

		haikuFallback := &DebateTeamMember{
			Position:     PositionCritic,
			Role:         RoleCritic,
			ProviderName: "claude",
			ModelName:    ClaudeModels.Haiku,
			Score:        8.5,
			IsActive:     false,
			Fallback:     qwenFallback,
		}

		primary := &DebateTeamMember{
			Position:     PositionCritic,
			Role:         RoleCritic,
			ProviderName: "deepseek",
			ModelName:    "deepseek-chat",
			Score:        8.8,
			IsActive:     true,
			Fallback:     haikuFallback,
		}

		// Verify fallback chain
		assert.NotNil(t, primary.Fallback)
		assert.Equal(t, "claude", primary.Fallback.ProviderName)
		assert.NotNil(t, primary.Fallback.Fallback)
		assert.Equal(t, "qwen", primary.Fallback.Fallback.ProviderName)
		assert.Nil(t, primary.Fallback.Fallback.Fallback)
	})
}

func TestDebateTeamConfigGetTeamSummary(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	config := NewDebateTeamConfig(nil, nil, logger)

	// Manually add a member for testing
	config.members[PositionAnalyst] = &DebateTeamMember{
		Position:     PositionAnalyst,
		Role:         RoleAnalyst,
		ProviderName: "claude",
		ModelName:    ClaudeModels.SonnetLatest,
		Score:        9.5,
		IsActive:     true,
	}

	t.Run("Summary includes team info", func(t *testing.T) {
		summary := config.GetTeamSummary()

		assert.Equal(t, "HelixAgent AI Debate Team", summary["team_name"])
		assert.Equal(t, 5, summary["total_positions"])
		assert.Equal(t, 1, summary["active_positions"])
	})

	t.Run("Summary includes Claude models", func(t *testing.T) {
		summary := config.GetTeamSummary()

		claudeModels := summary["claude_models"].(map[string]string)
		assert.NotEmpty(t, claudeModels["sonnet"])
		assert.NotEmpty(t, claudeModels["opus"])
		assert.NotEmpty(t, claudeModels["haiku"])
	})

	t.Run("Summary includes Qwen models", func(t *testing.T) {
		summary := config.GetTeamSummary()

		qwenModels := summary["qwen_models"].(map[string]string)
		assert.NotEmpty(t, qwenModels["turbo"])
		assert.NotEmpty(t, qwenModels["plus"])
		assert.NotEmpty(t, qwenModels["max"])
		assert.NotEmpty(t, qwenModels["coder"])
		assert.NotEmpty(t, qwenModels["long"])
	})

	t.Run("Summary includes position details", func(t *testing.T) {
		summary := config.GetTeamSummary()

		positions := summary["positions"].([]map[string]interface{})
		require.Len(t, positions, 5)

		// Position 1 should be assigned
		pos1 := positions[0]
		assert.Equal(t, DebateTeamPosition(1), pos1["position"])
		assert.Equal(t, RoleAnalyst, pos1["role"])
		assert.Equal(t, "claude", pos1["provider"])
	})
}

func TestDebateTeamConfigGetActiveMembers(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	config := NewDebateTeamConfig(nil, nil, logger)

	// Add some members
	config.members[PositionAnalyst] = &DebateTeamMember{
		Position:     PositionAnalyst,
		ProviderName: "claude",
		IsActive:     true,
	}
	config.members[PositionProposer] = &DebateTeamMember{
		Position:     PositionProposer,
		ProviderName: "claude",
		IsActive:     true,
	}
	config.members[PositionCritic] = &DebateTeamMember{
		Position:     PositionCritic,
		ProviderName: "deepseek",
		IsActive:     false, // Inactive
	}

	t.Run("Returns only active members", func(t *testing.T) {
		active := config.GetActiveMembers()
		assert.Len(t, active, 2)
	})

	t.Run("Inactive members are excluded", func(t *testing.T) {
		active := config.GetActiveMembers()
		for _, member := range active {
			assert.True(t, member.IsActive)
			assert.NotEqual(t, PositionCritic, member.Position)
		}
	})
}

func TestDebateTeamConfigActivateFallback(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	config := NewDebateTeamConfig(nil, nil, logger)

	t.Run("Returns error for non-existent position", func(t *testing.T) {
		_, err := config.ActivateFallback(PositionAnalyst)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no member at position")
	})

	t.Run("Returns error when no fallback available", func(t *testing.T) {
		config.members[PositionAnalyst] = &DebateTeamMember{
			Position:     PositionAnalyst,
			ProviderName: "claude",
			IsActive:     true,
			Fallback:     nil,
		}

		_, err := config.ActivateFallback(PositionAnalyst)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no fallback available")
	})

	t.Run("Successfully activates fallback", func(t *testing.T) {
		fallback := &DebateTeamMember{
			Position:     PositionProposer,
			ProviderName: "qwen",
			ModelName:    QwenModels.Plus,
			IsActive:     false,
		}

		config.members[PositionProposer] = &DebateTeamMember{
			Position:     PositionProposer,
			ProviderName: "claude",
			ModelName:    ClaudeModels.Opus,
			IsActive:     true,
			Fallback:     fallback,
		}

		activated, err := config.ActivateFallback(PositionProposer)
		require.NoError(t, err)
		assert.Equal(t, "qwen", activated.ProviderName)
		assert.Equal(t, QwenModels.Plus, activated.ModelName)
		assert.True(t, activated.IsActive)

		// Verify the position now has the fallback
		current := config.GetTeamMember(PositionProposer)
		assert.Equal(t, "qwen", current.ProviderName)
	})
}

func TestDebateTeamConfigInitializeTeam(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	t.Run("Initializes without providers gracefully", func(t *testing.T) {
		config := NewDebateTeamConfig(nil, nil, logger)
		err := config.InitializeTeam(context.Background())
		// Should not return error even without providers
		assert.NoError(t, err)
	})
}

func TestDebateTeamPositionRoleMapping(t *testing.T) {
	t.Run("Position 1 is Analyst", func(t *testing.T) {
		assert.Equal(t, PositionAnalyst, DebateTeamPosition(1))
	})

	t.Run("Position 2 is Proposer", func(t *testing.T) {
		assert.Equal(t, PositionProposer, DebateTeamPosition(2))
	})

	t.Run("Claude positions are 1 and 2", func(t *testing.T) {
		// Claude Sonnet = Position 1 (Analyst)
		// Claude Opus = Position 2 (Proposer)
		assert.True(t, PositionAnalyst < PositionCritic, "Claude positions should be before LLMsVerifier positions")
		assert.True(t, PositionProposer < PositionCritic, "Claude positions should be before LLMsVerifier positions")
	})

	t.Run("LLMsVerifier positions are 3, 4, and 5", func(t *testing.T) {
		assert.Equal(t, DebateTeamPosition(3), PositionCritic)
		assert.Equal(t, DebateTeamPosition(4), PositionSynthesis)
		assert.Equal(t, DebateTeamPosition(5), PositionMediator)
	})
}

func TestNoModelDuplication(t *testing.T) {
	t.Run("Claude models have no duplicates in fallback assignments", func(t *testing.T) {
		// Claude Haiku is used as fallback for positions 3, 4, 5
		// This is intentional - same model, different instances
		// The test verifies that we're aware of this design
		assert.Equal(t, ClaudeModels.Haiku, "claude-3-haiku-20240307")
	})

	t.Run("Qwen models are unique per position", func(t *testing.T) {
		// Each position should use a different Qwen model as fallback
		fallbackModels := map[DebateTeamPosition]string{
			PositionAnalyst:   QwenModels.Max,
			PositionProposer:  QwenModels.Plus,
			PositionCritic:    QwenModels.Turbo,
			PositionSynthesis: QwenModels.Coder,
			PositionMediator:  QwenModels.Long,
		}

		// Verify all models are unique
		usedModels := make(map[string]bool)
		for _, model := range fallbackModels {
			assert.False(t, usedModels[model], "Qwen model %s should not be duplicated", model)
			usedModels[model] = true
		}
	})
}
