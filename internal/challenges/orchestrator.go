package challenges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/monitor"
	"digital.vasic.challenges/pkg/registry"
	"digital.vasic.challenges/pkg/runner"
)

// OrchestratorConfig holds settings for the challenge
// orchestrator.
type OrchestratorConfig struct {
	// Parallel enables concurrent execution.
	Parallel bool

	// MaxConcurrency is the max number of concurrent
	// challenges when Parallel is true.
	MaxConcurrency int

	// StopOnFailure stops execution after the first failure.
	StopOnFailure bool

	// Verbose enables detailed logging.
	Verbose bool

	// StallThreshold is the default stall threshold for stuck
	// detection. Zero uses CategoryStallThresholds.
	StallThreshold time.Duration

	// ResultsDir is the base directory for results.
	ResultsDir string

	// Filter limits execution to specific challenge IDs.
	Filter []string

	// Category limits execution to a specific category.
	Category string

	// SkipInfra skips infrastructure health checks.
	SkipInfra bool

	// ScriptsDir is the path to challenge scripts.
	ScriptsDir string

	// ProjectRoot is the HelixAgent project root directory.
	ProjectRoot string

	// Timeout is the per-challenge timeout.
	Timeout time.Duration
}

// OrchestratorResult holds the outcome of an orchestrator run.
type OrchestratorResult struct {
	Results  []*challenge.Result
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	TimedOut int
	Stuck    int
	Errors   int
	Duration time.Duration
}

// ChallengeInfo describes a registered challenge.
type ChallengeInfo struct {
	ID          string
	Name        string
	Description string
	Category    string
}

// Orchestrator coordinates challenge registration, execution,
// and reporting. It replaces the bash run_all_challenges.sh
// logic with a Go-native implementation.
type Orchestrator struct {
	registry  registry.Registry
	runner    *runner.DefaultRunner
	collector *monitor.EventCollector
	reporter  *Reporter
	config    OrchestratorConfig
	envVars   map[string]string
}

// NewOrchestrator creates an Orchestrator with the given
// configuration.
func NewOrchestrator(cfg OrchestratorConfig) *Orchestrator {
	if cfg.ResultsDir == "" {
		cfg.ResultsDir = "challenge-results"
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 2
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Minute
	}

	reg := registry.NewRegistry()
	collector := monitor.NewEventCollector()

	stallThreshold := cfg.StallThreshold
	if stallThreshold == 0 {
		stallThreshold = StallThresholdForCategory("default")
	}

	r := runner.NewRunner(
		runner.WithRegistry(reg),
		runner.WithResultsDir(cfg.ResultsDir),
		runner.WithTimeout(cfg.Timeout),
		runner.WithStaleThreshold(stallThreshold),
	)

	reporter := NewReporter(cfg.ResultsDir)

	// Load environment variables from project root.
	var envVars map[string]string
	if cfg.ProjectRoot != "" {
		envVars = MergeEnvFiles(
			filepath.Join(cfg.ProjectRoot, ".env"),
			filepath.Join(cfg.ProjectRoot,
				"challenges", ".env"),
		)
	}

	return &Orchestrator{
		registry:  reg,
		runner:    r,
		collector: collector,
		reporter:  reporter,
		config:    cfg,
		envVars:   envVars,
	}
}

// RegisterAll discovers and registers all challenge scripts.
func (o *Orchestrator) RegisterAll() error {
	scriptsDir := o.config.ScriptsDir
	if scriptsDir == "" && o.config.ProjectRoot != "" {
		scriptsDir = filepath.Join(
			o.config.ProjectRoot, "challenges", "scripts",
		)
	}
	if scriptsDir == "" {
		return fmt.Errorf("scripts directory not configured")
	}

	return RegisterShellChallengesEnhanced(
		o.registry, scriptsDir, o.config.ProjectRoot,
	)
}

// Run executes all registered challenges according to config.
func (o *Orchestrator) Run(
	ctx context.Context,
) (*OrchestratorResult, error) {
	start := time.Now()

	ids := o.filterChallenges()
	if len(ids) == 0 {
		return &OrchestratorResult{}, nil
	}

	cfg := challenge.NewConfig("")
	cfg.ResultsDir = o.config.ResultsDir
	cfg.Verbose = o.config.Verbose
	cfg.Environment = o.envVars

	var results []*challenge.Result
	var err error

	if o.config.Parallel && len(ids) > 1 {
		results, err = o.runner.RunParallel(
			ctx, ids, cfg, o.config.MaxConcurrency,
		)
	} else {
		results, err = o.runner.RunSequence(ctx, ids, cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("run challenges: %w", err)
	}

	result := o.buildResult(results, time.Since(start))

	// Write results report.
	if writeErr := o.reporter.WriteResults(results); writeErr != nil {
		fmt.Fprintf(os.Stderr,
			"warning: failed to write results: %v\n", writeErr)
	}

	return result, nil
}

// RunSingle executes a single challenge by ID.
func (o *Orchestrator) RunSingle(
	ctx context.Context,
	id string,
) (*OrchestratorResult, error) {
	start := time.Now()

	cfg := challenge.NewConfig(challenge.ID(id))
	cfg.ResultsDir = o.config.ResultsDir
	cfg.Verbose = o.config.Verbose
	cfg.Environment = o.envVars

	// Apply category-specific stall threshold.
	c, getErr := o.registry.Get(challenge.ID(id))
	if getErr != nil {
		return nil, fmt.Errorf(
			"challenge %s not found: %w", id, getErr,
		)
	}
	cfg.StaleThreshold = StallThresholdForCategory(c.Category())

	result, err := o.runner.Run(
		ctx, challenge.ID(id), cfg,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"run challenge %s: %w", id, err,
		)
	}

	results := []*challenge.Result{result}
	if writeErr := o.reporter.WriteResults(results); writeErr != nil {
		fmt.Fprintf(os.Stderr,
			"warning: failed to write results: %v\n", writeErr)
	}

	return o.buildResult(results, time.Since(start)), nil
}

// List returns information about all registered challenges.
func (o *Orchestrator) List() []ChallengeInfo {
	all := o.registry.List()
	infos := make([]ChallengeInfo, 0, len(all))
	for _, c := range all {
		infos = append(infos, ChallengeInfo{
			ID:          string(c.ID()),
			Name:        c.Name(),
			Description: c.Description(),
			Category:    c.Category(),
		})
	}
	return infos
}

// filterChallenges returns challenge IDs matching the filter
// and category config.
func (o *Orchestrator) filterChallenges() []challenge.ID {
	all := o.registry.List()
	var ids []challenge.ID

	filterSet := make(map[string]bool)
	for _, f := range o.config.Filter {
		filterSet[f] = true
	}

	for _, c := range all {
		// Apply ID filter.
		if len(filterSet) > 0 &&
			!filterSet[string(c.ID())] {
			continue
		}
		// Apply category filter.
		if o.config.Category != "" &&
			c.Category() != o.config.Category {
			continue
		}
		ids = append(ids, c.ID())
	}
	return ids
}

func (o *Orchestrator) buildResult(
	results []*challenge.Result,
	duration time.Duration,
) *OrchestratorResult {
	r := &OrchestratorResult{
		Results:  results,
		Duration: duration,
	}
	for _, res := range results {
		r.Total++
		switch res.Status {
		case challenge.StatusPassed:
			r.Passed++
		case challenge.StatusFailed:
			r.Failed++
		case challenge.StatusSkipped:
			r.Skipped++
		case challenge.StatusTimedOut:
			r.TimedOut++
		case challenge.StatusStuck:
			r.Stuck++
		case challenge.StatusError:
			r.Errors++
		}
	}
	return r
}

// RegisterShellChallengesEnhanced discovers shell scripts and
// registers them with category detection and per-category
// stall thresholds.
func RegisterShellChallengesEnhanced(
	reg registry.Registry,
	scriptsDir string,
	workDir string,
) error {
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, "_challenge.sh") &&
			!strings.HasSuffix(name, "_test.sh") {
			continue
		}

		id := strings.TrimSuffix(name, ".sh")
		id = strings.ReplaceAll(id, "_", "-")

		category := detectCategory(name)
		scriptPath := filepath.Join(scriptsDir, name)

		sc := challenge.NewShellChallenge(
			challenge.ID(id),
			formatName(id),
			"Shell challenge: "+name,
			category,
			nil,
			scriptPath,
			nil,
			workDir,
		)

		if err := reg.Register(sc); err != nil {
			return err
		}
	}

	return nil
}

// detectCategory extracts a category from the script filename
// prefix. For example, "provider_comprehensive_challenge.sh"
// returns "provider".
func detectCategory(filename string) string {
	prefixes := []string{
		"provider_", "security_", "debate_", "cli_",
		"mcp_", "bigdata_", "memory_", "performance_",
		"grpc_", "release_", "speckit_", "subscription_",
		"verification_", "fallback_", "semantic_",
		"integration_", "full_system_", "constitution_",
		"challenge_module_",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(filename, p) {
			return strings.TrimSuffix(p, "_")
		}
	}
	return "shell"
}
