package challenges

import (
	"context"
	"fmt"
	"sort"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// DebateFormationChallenge verifies that debate groups can be
// correctly formed from verified providers.
type DebateFormationChallenge struct {
	challenge.BaseChallenge
	providers    []ProviderInfo
	MinProviders int
	GroupSize    int
}

// Ensure it satisfies the Challenge interface.
var _ challenge.Challenge = (*DebateFormationChallenge)(nil)

// Execute runs the debate formation validation.
func (c *DebateFormationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {
	start := time.Now()

	minProviders := c.MinProviders
	if minProviders == 0 {
		minProviders = 3
	}
	groupSize := c.GroupSize
	if groupSize == 0 {
		groupSize = 3
	}

	// Filter verified providers with sufficient score.
	var verified []ProviderInfo
	for _, p := range c.providers {
		if p.Verified && p.Score >= 5.0 {
			verified = append(verified, p)
		}
	}

	// Sort by score descending.
	sort.Slice(verified, func(i, j int) bool {
		return verified[i].Score > verified[j].Score
	})

	assertions := []challenge.AssertionResult{
		{
			Type:   "min_count",
			Target: "verified_providers",
			Passed: len(verified) >= minProviders,
			Message: fmt.Sprintf(
				"need %d verified providers, have %d",
				minProviders, len(verified),
			),
		},
	}

	status := challenge.StatusPassed
	errMsg := ""

	if len(verified) < minProviders {
		status = challenge.StatusFailed
		errMsg = fmt.Sprintf(
			"insufficient verified providers: "+
				"need %d, have %d",
			minProviders, len(verified),
		)
	} else {
		// Form a debate group from top providers.
		group := DebateGroup{
			Name:     "primary",
			Strategy: "confidence_weighted",
		}
		for i := 0; i < groupSize && i < len(verified); i++ {
			group.Members = append(
				group.Members,
				DebateGroupMember{
					Provider: verified[i].Name,
					Score:    verified[i].Score,
					Role:     "debater",
				},
			)
		}

		assertions = append(
			assertions,
			challenge.AssertionResult{
				Type:   "min_count",
				Target: "debate_group_members",
				Passed: len(group.Members) >= groupSize,
				Message: fmt.Sprintf(
					"debate group has %d members "+
						"(need %d)",
					len(group.Members), groupSize,
				),
			},
		)
	}

	return c.CreateResult(
		status, start, assertions, nil, nil, errMsg,
	), nil
}
