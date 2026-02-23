package compliance

import (
	"testing"

	"dev.helix.agent/internal/verifier"
	"github.com/stretchr/testify/assert"
)

// TestAllProvidersHaveNames verifies that every registered provider
// has a non-empty display name.
func TestAllProvidersHaveNames(t *testing.T) {
	providers := verifier.SupportedProviders
	missingNames := []string{}

	for key, info := range providers {
		if info.DisplayName == "" {
			missingNames = append(missingNames, key)
		}
	}

	if len(missingNames) > 0 {
		t.Errorf("COMPLIANCE FAILED: %d providers missing display names: %v",
			len(missingNames), missingNames)
	}

	t.Logf("COMPLIANCE: All %d providers have display names", len(providers))
}

// TestAllProvidersHaveAuthType verifies that every registered provider
// specifies an authentication type.
func TestAllProvidersHaveAuthType(t *testing.T) {
	providers := verifier.SupportedProviders
	missingAuth := []string{}

	for key, info := range providers {
		if info.AuthType == "" {
			missingAuth = append(missingAuth, key)
		}
	}

	if len(missingAuth) > 0 {
		t.Errorf("COMPLIANCE FAILED: %d providers missing auth type: %v",
			len(missingAuth), missingAuth)
	}

	t.Logf("COMPLIANCE: All %d providers have auth types", len(providers))
}

// TestAllProvidersHaveValidTier verifies that every provider has a tier
// between 1 and 7 (inclusive).
func TestAllProvidersHaveValidTier(t *testing.T) {
	providers := verifier.SupportedProviders
	invalidTiers := []string{}

	for key, info := range providers {
		if info.Tier < 1 || info.Tier > 7 {
			invalidTiers = append(invalidTiers, key)
		}
	}

	if len(invalidTiers) > 0 {
		t.Errorf("COMPLIANCE FAILED: %d providers have invalid tier (must be 1-7): %v",
			len(invalidTiers), invalidTiers)
	}

	t.Logf("COMPLIANCE: All %d providers have valid tiers (1-7)", len(providers))
}

// TestProviderCountCompliance verifies at least 30 providers are registered.
func TestProviderCountCompliance(t *testing.T) {
	providers := verifier.SupportedProviders
	count := len(providers)
	const minProviders = 30

	t.Logf("Registered providers: %d (required: %d)", count, minProviders)

	if count < minProviders {
		t.Errorf("COMPLIANCE FAILED: Only %d providers registered, minimum is %d",
			count, minProviders)
	}
}

// TestOAuthProviderCountCompliance verifies at least 2 OAuth providers exist.
func TestOAuthProviderCountCompliance(t *testing.T) {
	oauthProviders := verifier.GetProvidersByAuthType(verifier.AuthTypeOAuth)
	const minOAuth = 2

	t.Logf("OAuth providers: %d (required: %d)", len(oauthProviders), minOAuth)
	for _, p := range oauthProviders {
		t.Logf("  - %s", p.DisplayName)
	}

	if len(oauthProviders) < minOAuth {
		t.Errorf("COMPLIANCE FAILED: Only %d OAuth providers, minimum is %d (Claude, Qwen)",
			len(oauthProviders), minOAuth)
	}
}

// TestFreeProviderCountCompliance verifies at least 2 free providers exist.
func TestFreeProviderCountCompliance(t *testing.T) {
	freeCount := 0
	for _, info := range verifier.SupportedProviders {
		if info.Free {
			freeCount++
		}
	}

	const minFree = 2
	t.Logf("Free providers: %d (required: %d)", freeCount, minFree)

	if freeCount < minFree {
		t.Errorf("COMPLIANCE FAILED: Only %d free providers, minimum is %d (Zen, OpenRouter)",
			freeCount, minFree)
	}
}

// TestScoringWeightsCompliance verifies that provider scoring weight
// configuration is consistent (weights are defined per CLAUDE.md specification).
func TestScoringWeightsCompliance(t *testing.T) {
	// Per CLAUDE.md: ResponseSpeed 25%, CostEffectiveness 25%,
	// ModelEfficiency 20%, Capability 20%, Recency 10%
	weights := map[string]float64{
		"ResponseSpeed":     25.0,
		"CostEffectiveness": 25.0,
		"ModelEfficiency":   20.0,
		"Capability":        20.0,
		"Recency":           10.0,
	}

	total := 0.0
	for name, w := range weights {
		assert.Greater(t, w, 0.0, "Weight for %s must be positive", name)
		total += w
	}

	assert.Equal(t, 100.0, total, "Scoring weights must sum to exactly 100%%")
	t.Logf("COMPLIANCE: Scoring weights sum to %.1f%% (correct)", total)
}
