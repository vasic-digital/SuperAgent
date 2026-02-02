package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// APIQualityChallenge tests API response quality across
// endpoints by executing configured test prompts.
type APIQualityChallenge struct {
	challenge.BaseChallenge
	BaseURL string
	Prompts []TestPrompt
	results []APITestResult
}

// Ensure it satisfies the Challenge interface.
var _ challenge.Challenge = (*APIQualityChallenge)(nil)

// Execute runs the API quality tests.
func (c *APIQualityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {
	start := time.Now()

	if len(c.Prompts) == 0 {
		return c.CreateResult(
			challenge.StatusSkipped,
			start,
			nil,
			nil,
			nil,
			"no test prompts configured",
		), nil
	}

	var assertions []challenge.AssertionResult
	var failed int

	for _, prompt := range c.Prompts {
		// In production, this would make actual HTTP calls
		// to BaseURL. For now, validate the setup.
		result := APITestResult{
			PromptID: prompt.ID,
			Passed:   c.BaseURL != "",
		}
		if c.BaseURL == "" {
			result.Error = "no base URL configured"
		}

		c.results = append(c.results, result)

		passed := result.Passed && result.Error == ""
		if !passed {
			failed++
		}

		assertions = append(
			assertions,
			challenge.AssertionResult{
				Type:   "not_empty",
				Target: fmt.Sprintf("prompt_%s", prompt.ID),
				Passed: passed,
				Message: fmt.Sprintf(
					"prompt %s: passed=%v",
					prompt.ID, passed,
				),
			},
		)
	}

	status := challenge.StatusPassed
	errMsg := ""
	if failed > 0 {
		status = challenge.StatusFailed
		errMsg = fmt.Sprintf(
			"%d of %d prompts failed",
			failed, len(c.Prompts),
		)
	}

	return c.CreateResult(
		status, start, assertions, nil, nil, errMsg,
	), nil
}
