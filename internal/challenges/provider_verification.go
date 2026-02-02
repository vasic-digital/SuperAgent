package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// ProviderVerificationChallenge verifies all configured LLM
// providers can respond to basic requests by checking their
// verification status and minimum score.
type ProviderVerificationChallenge struct {
	challenge.BaseChallenge
	providers []ProviderInfo
	results   map[string]bool
}

// Ensure it satisfies the Challenge interface.
var _ challenge.Challenge = (*ProviderVerificationChallenge)(nil)

// Execute runs the provider verification checks.
func (c *ProviderVerificationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {
	start := time.Now()
	c.results = make(map[string]bool)

	var failed []string
	for _, p := range c.providers {
		// Each provider is considered verified if its score
		// meets the minimum threshold.
		if p.Verified && p.Score >= 5.0 {
			c.results[p.Name] = true
		} else {
			c.results[p.Name] = false
			failed = append(failed, p.Name)
		}
	}

	status := challenge.StatusPassed
	errMsg := ""
	if len(failed) > 0 {
		status = challenge.StatusFailed
		errMsg = fmt.Sprintf(
			"providers failed verification: %v", failed,
		)
	}

	assertions := make(
		[]challenge.AssertionResult, 0, len(c.providers),
	)
	for _, p := range c.providers {
		assertions = append(
			assertions,
			challenge.AssertionResult{
				Type:   "provider_verified",
				Target: p.Name,
				Passed: c.results[p.Name],
				Message: fmt.Sprintf(
					"provider %s verified=%v score=%.1f",
					p.Name, p.Verified, p.Score,
				),
			},
		)
	}

	return c.CreateResult(
		status, start, assertions, nil, nil, errMsg,
	), nil
}
