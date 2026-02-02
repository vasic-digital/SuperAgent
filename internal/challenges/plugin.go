package challenges

import (
	"fmt"

	"digital.vasic.challenges/pkg/assertion"
	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/plugin"
	"digital.vasic.challenges/pkg/registry"
)

// HelixPlugin registers HelixAgent-specific challenges with the
// generic challenge framework. It implements the plugin.Plugin
// interface so it can be loaded by the plugin registry.
type HelixPlugin struct {
	providers []ProviderInfo
	reg       registry.Registry
	engine    assertion.Engine
}

// NewHelixPlugin creates a new HelixAgent challenge plugin.
func NewHelixPlugin(providers []ProviderInfo) *HelixPlugin {
	return &HelixPlugin{providers: providers}
}

// Ensure HelixPlugin satisfies the plugin.Plugin interface.
var _ plugin.Plugin = (*HelixPlugin)(nil)

func (p *HelixPlugin) Name() string    { return "helix-agent" }
func (p *HelixPlugin) Version() string { return "1.0.0" }

// Init initializes the plugin with the given context. It expects
// "registry" and "assertion_engine" keys in the config map.
func (p *HelixPlugin) Init(ctx *plugin.PluginContext) error {
	if ctx == nil {
		return fmt.Errorf("plugin context must not be nil")
	}

	// Extract registry from context config.
	if regVal, ok := ctx.Config["registry"]; ok {
		if r, ok := regVal.(registry.Registry); ok {
			p.reg = r
		}
	}

	// Extract assertion engine from context config.
	if engVal, ok := ctx.Config["assertion_engine"]; ok {
		if e, ok := engVal.(assertion.Engine); ok {
			p.engine = e
		}
	}

	if p.reg != nil {
		if err := p.registerChallenges(p.reg); err != nil {
			return fmt.Errorf("register challenges: %w", err)
		}
	}

	if p.engine != nil {
		if err := p.registerAssertions(p.engine); err != nil {
			return fmt.Errorf("register assertions: %w", err)
		}
	}

	return nil
}

// registerChallenges registers HelixAgent-specific challenge
// implementations with the given registry.
func (p *HelixPlugin) registerChallenges(
	reg registry.Registry,
) error {
	// Register provider verification challenge.
	if err := reg.Register(&ProviderVerificationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"helix-provider-verification",
			"Provider Verification",
			"Verifies all configured LLM providers can "+
				"accept and respond to requests",
			"validation",
			nil,
		),
		providers: p.providers,
	}); err != nil {
		return err
	}

	// Register debate formation challenge.
	if err := reg.Register(&DebateFormationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"helix-debate-formation",
			"Debate Group Formation",
			"Validates debate group formation from "+
				"verified providers",
			"validation",
			[]challenge.ID{"helix-provider-verification"},
		),
		providers: p.providers,
	}); err != nil {
		return err
	}

	// Register API quality challenge.
	if err := reg.Register(&APIQualityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"helix-api-quality",
			"API Quality",
			"Tests API response quality and latency",
			"validation",
			[]challenge.ID{"helix-provider-verification"},
		),
	}); err != nil {
		return err
	}

	return nil
}

// registerAssertions registers HelixAgent-specific assertion
// evaluators with the given engine.
func (p *HelixPlugin) registerAssertions(
	engine assertion.Engine,
) error {
	// Register provider_verified evaluator.
	if err := engine.Register(
		"provider_verified",
		func(
			def assertion.Definition,
			value any,
		) (bool, string) {
			info, ok := value.(*ProviderInfo)
			if !ok {
				return false, "expected *ProviderInfo"
			}
			if !info.Verified {
				return false,
					"provider " + info.Name +
						" is not verified"
			}
			return true,
				"provider " + info.Name + " is verified"
		},
	); err != nil {
		return err
	}

	// Register min_provider_score evaluator.
	if err := engine.Register(
		"min_provider_score",
		func(
			def assertion.Definition,
			value any,
		) (bool, string) {
			info, ok := value.(*ProviderInfo)
			if !ok {
				return false, "expected *ProviderInfo"
			}
			if info.Score < 5.0 {
				return false, fmt.Sprintf(
					"provider %s score %.1f below "+
						"minimum 5.0",
					info.Name, info.Score,
				)
			}
			return true, fmt.Sprintf(
				"provider %s score %.1f meets minimum",
				info.Name, info.Score,
			)
		},
	); err != nil {
		return err
	}

	return nil
}
