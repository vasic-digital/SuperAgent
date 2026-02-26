package challenges

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// Reporter writes challenge results in HelixAgent's directory
// structure with a `latest` symlink.
type Reporter struct {
	baseDir string
}

// NewReporter creates a Reporter that writes to the given
// base directory.
func NewReporter(baseDir string) *Reporter {
	return &Reporter{baseDir: baseDir}
}

// WriteResults writes a summary of all challenge results to
// the base directory with a timestamped subdirectory and a
// `latest` symlink.
func (r *Reporter) WriteResults(
	results []*challenge.Result,
) error {
	now := time.Now()
	dir := filepath.Join(
		r.baseDir,
		now.Format("2006-01-02_150405"),
	)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create results dir: %w", err)
	}

	// Write summary.
	summary := r.buildSummary(results)
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal summary: %w", err)
	}
	summaryPath := filepath.Join(dir, "summary.json")
	if err := os.WriteFile(summaryPath, data, 0644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}

	// Write individual results.
	for _, result := range results {
		id := strings.ReplaceAll(
			string(result.ChallengeID), "/", "_",
		)
		resultPath := filepath.Join(
			dir, id+".json",
		)
		resultData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			continue
		}
		_ = os.WriteFile(resultPath, resultData, 0644) //nolint:errcheck
	}

	// Update latest symlink.
	latestPath := filepath.Join(r.baseDir, "latest")
	_ = os.Remove(latestPath)
	_ = os.Symlink(dir, latestPath) //nolint:errcheck

	return nil
}

// ResultSummary holds aggregate results.
type ResultSummary struct {
	Timestamp string        `json:"timestamp"`
	Total     int           `json:"total"`
	Passed    int           `json:"passed"`
	Failed    int           `json:"failed"`
	Skipped   int           `json:"skipped"`
	TimedOut  int           `json:"timed_out"`
	Stuck     int           `json:"stuck"`
	Errors    int           `json:"errors"`
	Duration  string        `json:"duration"`
	Results   []ResultEntry `json:"results"`
}

// ResultEntry is a summary line for one challenge.
type ResultEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

func (r *Reporter) buildSummary(
	results []*challenge.Result,
) ResultSummary {
	s := ResultSummary{
		Timestamp: time.Now().Format(time.RFC3339),
		Results:   make([]ResultEntry, 0, len(results)),
	}
	var totalDur time.Duration

	for _, res := range results {
		s.Total++
		totalDur += res.Duration
		switch res.Status {
		case challenge.StatusPassed:
			s.Passed++
		case challenge.StatusFailed:
			s.Failed++
		case challenge.StatusSkipped:
			s.Skipped++
		case challenge.StatusTimedOut:
			s.TimedOut++
		case challenge.StatusStuck:
			s.Stuck++
		case challenge.StatusError:
			s.Errors++
		}
		s.Results = append(s.Results, ResultEntry{
			ID:       string(res.ChallengeID),
			Name:     res.ChallengeName,
			Status:   res.Status,
			Duration: res.Duration.String(),
			Error:    res.Error,
		})
	}
	s.Duration = totalDur.String()
	return s
}
