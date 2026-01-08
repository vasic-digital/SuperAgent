package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"dev.helix.agent/internal/llm"
)

// DebateTeamPosition represents a position in the AI debate team
type DebateTeamPosition int

const (
	PositionAnalyst   DebateTeamPosition = 1 // Claude Sonnet - Primary analyst
	PositionProposer  DebateTeamPosition = 2 // Claude Opus - Primary proposer
	PositionCritic    DebateTeamPosition = 3 // LLMsVerifier scored - Primary critic
	PositionSynthesis DebateTeamPosition = 4 // LLMsVerifier scored - Synthesis expert
	PositionMediator  DebateTeamPosition = 5 // LLMsVerifier scored - Mediator/consensus
)

// DebateRole represents the role a team member plays
type DebateRole string

const (
	RoleAnalyst   DebateRole = "analyst"
	RoleProposer  DebateRole = "proposer"
	RoleCritic    DebateRole = "critic"
	RoleSynthesis DebateRole = "synthesis"
	RoleMediator  DebateRole = "mediator"
)

// ClaudeModels defines the available Claude models for debate positions
var ClaudeModels = struct {
	Sonnet        string // Position 1 - Analyst
	Opus          string // Position 2 - Proposer
	Haiku         string // Fallback for positions 3, 4, 5
	SonnetLatest  string // Latest Sonnet version
	OpusLatest    string // Latest Opus version
}{
	Sonnet:        "claude-3-sonnet-20240229",
	Opus:          "claude-3-opus-20240229",
	Haiku:         "claude-3-haiku-20240307",
	SonnetLatest:  "claude-3-5-sonnet-20241022",
	OpusLatest:    "claude-3-opus-20240229", // Opus latest is still the same
}

// QwenModels defines the available Qwen models for fallback positions
var QwenModels = struct {
	Turbo   string // Fast, efficient model
	Plus    string // Balanced model
	Max     string // Most capable model
	Coder   string // Code-focused model
	Long    string // Long context model
}{
	Turbo:   "qwen-turbo",
	Plus:    "qwen-plus",
	Max:     "qwen-max",
	Coder:   "qwen-coder-turbo",
	Long:    "qwen-long",
}

// DebateTeamMember represents a member of the AI debate team
type DebateTeamMember struct {
	Position     DebateTeamPosition `json:"position"`
	Role         DebateRole         `json:"role"`
	ProviderName string             `json:"provider_name"`
	ModelName    string             `json:"model_name"`
	Provider     llm.LLMProvider    `json:"-"`
	Fallback     *DebateTeamMember  `json:"fallback,omitempty"`
	Score        float64            `json:"score"`
	IsActive     bool               `json:"is_active"`
}

// DebateTeamConfig manages the AI debate team configuration
type DebateTeamConfig struct {
	mu               sync.RWMutex
	members          map[DebateTeamPosition]*DebateTeamMember
	providerRegistry *ProviderRegistry
	discovery        *ProviderDiscovery
	logger           *logrus.Logger
}

// NewDebateTeamConfig creates a new debate team configuration
func NewDebateTeamConfig(
	providerRegistry *ProviderRegistry,
	discovery *ProviderDiscovery,
	logger *logrus.Logger,
) *DebateTeamConfig {
	config := &DebateTeamConfig{
		members:          make(map[DebateTeamPosition]*DebateTeamMember),
		providerRegistry: providerRegistry,
		discovery:        discovery,
		logger:           logger,
	}
	return config
}

// InitializeTeam sets up the debate team with Claude, Qwen, and LLMsVerifier-scored providers
// All providers are verified before being added to the team
func (dtc *DebateTeamConfig) InitializeTeam(ctx context.Context) error {
	dtc.mu.Lock()
	defer dtc.mu.Unlock()

	dtc.logger.Info("Initializing AI Debate Team with provider verification...")

	// Run provider verification first to ensure we have valid, working providers
	dtc.verifyProviders(ctx)

	// Position 1: Claude Sonnet as Analyst
	if err := dtc.assignClaudePosition(ctx, PositionAnalyst, RoleAnalyst, ClaudeModels.SonnetLatest); err != nil {
		dtc.logger.WithError(err).Warn("Failed to assign Claude Sonnet to Position 1, will use fallback")
	}

	// Position 2: Claude Opus as Proposer
	if err := dtc.assignClaudePosition(ctx, PositionProposer, RoleProposer, ClaudeModels.Opus); err != nil {
		dtc.logger.WithError(err).Warn("Failed to assign Claude Opus to Position 2, will use fallback")
	}

	// Positions 3-5: LLMsVerifier scored providers (only verified ones)
	if err := dtc.assignVerifiedPositions(ctx); err != nil {
		dtc.logger.WithError(err).Warn("Failed to assign some verified positions, will use fallbacks")
	}

	// Assign fallbacks using Claude Haiku for positions 3-5 and Qwen for all positions
	dtc.assignFallbacks(ctx)

	// Log final team composition
	dtc.logTeamComposition()

	dtc.logger.WithField("team_size", len(dtc.members)).Info("AI Debate Team initialized")
	return nil
}

// verifyProviders runs health checks on all discovered providers
func (dtc *DebateTeamConfig) verifyProviders(ctx context.Context) {
	if dtc.discovery == nil {
		dtc.logger.Warn("Provider discovery not available, skipping verification")
		return
	}

	// Trigger verification of all providers
	dtc.discovery.VerifyAllProviders(ctx)

	// Count verified providers from all providers
	allProviders := dtc.discovery.GetAllProviders()
	verifiedCount := 0
	for _, p := range allProviders {
		if p.Verified {
			verifiedCount++
		}
	}

	dtc.logger.WithFields(logrus.Fields{
		"total_discovered": len(allProviders),
		"total_verified":   verifiedCount,
	}).Info("Provider verification completed")

	// Log any failed providers
	for _, p := range allProviders {
		if !p.Verified {
			dtc.logger.WithFields(logrus.Fields{
				"provider": p.Name,
				"status":   p.Status,
				"error":    p.Error,
			}).Debug("Provider not verified, may have invalid credentials or subscription")
		}
	}
}

// logTeamComposition logs the final team composition
func (dtc *DebateTeamConfig) logTeamComposition() {
	dtc.logger.Info("=== AI Debate Team Composition ===")
	for pos := PositionAnalyst; pos <= PositionMediator; pos++ {
		member := dtc.members[pos]
		if member != nil {
			fields := logrus.Fields{
				"position": pos,
				"role":     member.Role,
				"provider": member.ProviderName,
				"model":    member.ModelName,
				"score":    member.Score,
			}
			if member.Fallback != nil {
				fields["fallback"] = member.Fallback.ProviderName
			}
			dtc.logger.WithFields(fields).Info("Team position assigned")
		} else {
			dtc.logger.WithField("position", pos).Warn("Team position unassigned")
		}
	}
}

// assignClaudePosition assigns a Claude model to a specific position
func (dtc *DebateTeamConfig) assignClaudePosition(ctx context.Context, position DebateTeamPosition, role DebateRole, model string) error {
	var provider llm.LLMProvider

	// Try to get Claude provider from registry
	if dtc.providerRegistry != nil {
		p, err := dtc.providerRegistry.GetProvider("claude")
		if err == nil && p != nil {
			provider = p
		}
	}

	// Try to get from discovery if not found in registry
	if provider == nil && dtc.discovery != nil {
		discovered := dtc.discovery.GetProviderByName("claude")
		if discovered != nil && discovered.Provider != nil {
			provider = discovered.Provider
		}
	}

	if provider == nil {
		return fmt.Errorf("Claude provider not available")
	}

	member := &DebateTeamMember{
		Position:     position,
		Role:         role,
		ProviderName: "claude",
		ModelName:    model,
		Provider:     provider,
		Score:        9.5, // Claude has high baseline score
		IsActive:     true,
	}

	dtc.members[position] = member
	dtc.logger.WithFields(logrus.Fields{
		"position": position,
		"role":     role,
		"model":    model,
	}).Info("Assigned Claude model to debate position")

	return nil
}

// assignVerifiedPositions assigns LLMsVerifier-scored providers to positions 3-5
func (dtc *DebateTeamConfig) assignVerifiedPositions(ctx context.Context) error {
	if dtc.discovery == nil {
		return fmt.Errorf("provider discovery not available")
	}

	// Get best providers from LLMsVerifier scoring
	bestProviders := dtc.discovery.GetBestProviders(5)

	positionRoles := map[DebateTeamPosition]DebateRole{
		PositionCritic:    RoleCritic,
		PositionSynthesis: RoleSynthesis,
		PositionMediator:  RoleMediator,
	}

	providerIdx := 0
	for position := PositionCritic; position <= PositionMediator; position++ {
		// Skip Claude providers (they're already assigned to positions 1-2)
		for providerIdx < len(bestProviders) {
			p := bestProviders[providerIdx]
			if p.Type != "claude" {
				break
			}
			providerIdx++
		}

		if providerIdx >= len(bestProviders) {
			dtc.logger.WithField("position", position).Warn("No more verified providers available")
			continue
		}

		provider := bestProviders[providerIdx]
		providerIdx++

		member := &DebateTeamMember{
			Position:     position,
			Role:         positionRoles[position],
			ProviderName: provider.Name,
			ModelName:    provider.DefaultModel,
			Provider:     provider.Provider,
			Score:        provider.Score,
			IsActive:     true,
		}

		dtc.members[position] = member
		dtc.logger.WithFields(logrus.Fields{
			"position": position,
			"role":     positionRoles[position],
			"provider": provider.Name,
			"model":    provider.DefaultModel,
			"score":    provider.Score,
		}).Info("Assigned verified provider to debate position")
	}

	return nil
}

// assignFallbacks assigns fallback providers to all positions
func (dtc *DebateTeamConfig) assignFallbacks(ctx context.Context) {
	var qwenProvider, claudeProvider llm.LLMProvider

	// Get Qwen provider for fallbacks
	if dtc.providerRegistry != nil {
		p, err := dtc.providerRegistry.GetProvider("qwen")
		if err == nil && p != nil {
			qwenProvider = p
		}
	}
	if qwenProvider == nil && dtc.discovery != nil {
		discovered := dtc.discovery.GetProviderByName("qwen")
		if discovered != nil && discovered.Provider != nil {
			qwenProvider = discovered.Provider
		}
	}

	// Get Claude provider for Haiku fallback
	if dtc.providerRegistry != nil {
		p, err := dtc.providerRegistry.GetProvider("claude")
		if err == nil && p != nil {
			claudeProvider = p
		}
	}
	if claudeProvider == nil && dtc.discovery != nil {
		discovered := dtc.discovery.GetProviderByName("claude")
		if discovered != nil && discovered.Provider != nil {
			claudeProvider = discovered.Provider
		}
	}

	// Qwen models for each fallback position (no duplicates - each position uses different model)
	qwenFallbackModels := map[DebateTeamPosition]string{
		PositionAnalyst:   QwenModels.Max,    // Position 1 fallback
		PositionProposer:  QwenModels.Plus,   // Position 2 fallback
		PositionCritic:    QwenModels.Turbo,  // Position 3 fallback
		PositionSynthesis: QwenModels.Coder,  // Position 4 fallback
		PositionMediator:  QwenModels.Long,   // Position 5 fallback
	}

	for position, member := range dtc.members {
		if member == nil {
			continue
		}

		// First level fallback for positions 3-5: Claude Haiku
		if position >= PositionCritic && claudeProvider != nil {
			haikuFallback := &DebateTeamMember{
				Position:     position,
				Role:         member.Role,
				ProviderName: "claude",
				ModelName:    ClaudeModels.Haiku,
				Provider:     claudeProvider,
				Score:        8.5,
				IsActive:     false,
			}
			member.Fallback = haikuFallback

			// Second level fallback: Qwen
			if qwenProvider != nil {
				qwenFallback := &DebateTeamMember{
					Position:     position,
					Role:         member.Role,
					ProviderName: "qwen",
					ModelName:    qwenFallbackModels[position],
					Provider:     qwenProvider,
					Score:        7.5,
					IsActive:     false,
				}
				haikuFallback.Fallback = qwenFallback
			}
		} else if qwenProvider != nil {
			// For positions 1-2, Qwen is the direct fallback
			qwenFallback := &DebateTeamMember{
				Position:     position,
				Role:         member.Role,
				ProviderName: "qwen",
				ModelName:    qwenFallbackModels[position],
				Provider:     qwenProvider,
				Score:        7.5,
				IsActive:     false,
			}
			member.Fallback = qwenFallback
		}

		dtc.logger.WithFields(logrus.Fields{
			"position":     position,
			"primary":      member.ProviderName,
			"has_fallback": member.Fallback != nil,
		}).Debug("Configured fallback chain for position")
	}
}

// GetTeamMember returns the team member at the specified position
func (dtc *DebateTeamConfig) GetTeamMember(position DebateTeamPosition) *DebateTeamMember {
	dtc.mu.RLock()
	defer dtc.mu.RUnlock()
	return dtc.members[position]
}

// GetActiveMembers returns all active team members
func (dtc *DebateTeamConfig) GetActiveMembers() []*DebateTeamMember {
	dtc.mu.RLock()
	defer dtc.mu.RUnlock()

	members := make([]*DebateTeamMember, 0, len(dtc.members))
	for _, member := range dtc.members {
		if member != nil && member.IsActive {
			members = append(members, member)
		}
	}
	return members
}

// GetTeamSummary returns a summary of the debate team configuration
func (dtc *DebateTeamConfig) GetTeamSummary() map[string]interface{} {
	dtc.mu.RLock()
	defer dtc.mu.RUnlock()

	positions := make([]map[string]interface{}, 0, 5)

	for pos := PositionAnalyst; pos <= PositionMediator; pos++ {
		member := dtc.members[pos]
		if member == nil {
			positions = append(positions, map[string]interface{}{
				"position": pos,
				"status":   "unassigned",
			})
			continue
		}

		posInfo := map[string]interface{}{
			"position":      pos,
			"role":          member.Role,
			"provider":      member.ProviderName,
			"model":         member.ModelName,
			"score":         member.Score,
			"is_active":     member.IsActive,
			"has_fallback":  member.Fallback != nil,
		}

		if member.Fallback != nil {
			fallbacks := []map[string]interface{}{}
			fb := member.Fallback
			for fb != nil {
				fallbacks = append(fallbacks, map[string]interface{}{
					"provider": fb.ProviderName,
					"model":    fb.ModelName,
					"score":    fb.Score,
				})
				fb = fb.Fallback
			}
			posInfo["fallback_chain"] = fallbacks
		}

		positions = append(positions, posInfo)
	}

	return map[string]interface{}{
		"team_name":        "HelixAgent AI Debate Team",
		"total_positions":  5,
		"active_positions": len(dtc.GetActiveMembers()),
		"positions":        positions,
		"claude_models": map[string]string{
			"sonnet": ClaudeModels.SonnetLatest,
			"opus":   ClaudeModels.Opus,
			"haiku":  ClaudeModels.Haiku,
		},
		"qwen_models": map[string]string{
			"turbo": QwenModels.Turbo,
			"plus":  QwenModels.Plus,
			"max":   QwenModels.Max,
			"coder": QwenModels.Coder,
			"long":  QwenModels.Long,
		},
	}
}

// ActivateFallback activates the fallback for a position when the primary fails
func (dtc *DebateTeamConfig) ActivateFallback(position DebateTeamPosition) (*DebateTeamMember, error) {
	dtc.mu.Lock()
	defer dtc.mu.Unlock()

	member := dtc.members[position]
	if member == nil {
		return nil, fmt.Errorf("no member at position %d", position)
	}

	if member.Fallback == nil {
		return nil, fmt.Errorf("no fallback available for position %d", position)
	}

	// Deactivate current member
	member.IsActive = false

	// Activate fallback
	fallback := member.Fallback
	fallback.IsActive = true
	dtc.members[position] = fallback

	dtc.logger.WithFields(logrus.Fields{
		"position":     position,
		"old_provider": member.ProviderName,
		"new_provider": fallback.ProviderName,
		"new_model":    fallback.ModelName,
	}).Info("Activated fallback for debate position")

	return fallback, nil
}

// GetProviderForPosition returns the appropriate provider for a debate position
func (dtc *DebateTeamConfig) GetProviderForPosition(position DebateTeamPosition) (llm.LLMProvider, string, error) {
	member := dtc.GetTeamMember(position)
	if member == nil {
		return nil, "", fmt.Errorf("no member assigned to position %d", position)
	}

	if member.Provider == nil {
		// Try to activate fallback
		fallback, err := dtc.ActivateFallback(position)
		if err != nil {
			return nil, "", fmt.Errorf("provider unavailable and no fallback: %w", err)
		}
		return fallback.Provider, fallback.ModelName, nil
	}

	return member.Provider, member.ModelName, nil
}
