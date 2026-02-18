package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"dev.helix.agent/internal/challenges"
)

// handleListChallenges discovers and lists all available
// challenges.
func handleListChallenges(appCfg *AppConfig) error {
	orch := challenges.NewOrchestrator(
		challenges.OrchestratorConfig{
			ProjectRoot: getProjectRoot(),
		},
	)
	if err := orch.RegisterAll(); err != nil {
		return fmt.Errorf("register challenges: %w", err)
	}

	list := orch.List()
	if len(list) == 0 {
		fmt.Println("No challenges found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tNAME\tCATEGORY\n")
	fmt.Fprintf(w, "--\t----\t--------\n")
	for _, c := range list {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			c.ID, c.Name, c.Category)
	}
	w.Flush()

	fmt.Printf("\nTotal: %d challenges\n", len(list))
	return nil
}

// handleRunChallenges executes challenges according to the
// provided flags.
func handleRunChallenges(appCfg *AppConfig) error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	cfg := challenges.OrchestratorConfig{
		ProjectRoot:    getProjectRoot(),
		ResultsDir:     "challenge-results",
		Parallel:       appCfg.ChallengeParallel,
		MaxConcurrency: 2,
		Verbose:        appCfg.ChallengeVerbose,
		StallThreshold: appCfg.ChallengeStallThreshold,
		Timeout:        10 * 60 * 1e9, // 10 minutes
	}

	// Parse the run target.
	target := appCfg.RunChallenges
	if target != "all" {
		// Check if it's a category or a specific challenge ID.
		knownCategories := []string{
			"provider", "security", "debate", "cli", "mcp",
			"bigdata", "memory", "performance", "grpc",
			"release", "speckit", "subscription",
			"verification", "fallback", "semantic",
			"integration", "shell",
		}
		isCategory := false
		for _, cat := range knownCategories {
			if target == cat {
				isCategory = true
				break
			}
		}
		if isCategory {
			cfg.Category = target
		} else {
			cfg.Filter = []string{target}
		}
	}

	orch := challenges.NewOrchestrator(cfg)
	if err := orch.RegisterAll(); err != nil {
		return fmt.Errorf("register challenges: %w", err)
	}

	appCfg.Logger.Infof(
		"Running challenges (target=%s, parallel=%v, "+
			"stall_threshold=%v)",
		target, cfg.Parallel, cfg.StallThreshold,
	)

	var result *challenges.OrchestratorResult
	var err error

	if len(cfg.Filter) == 1 {
		result, err = orch.RunSingle(ctx, cfg.Filter[0])
	} else {
		result, err = orch.Run(ctx)
	}
	if err != nil {
		return fmt.Errorf("run challenges: %w", err)
	}

	// Print summary.
	printChallengeSummary(result)

	if result.Failed > 0 || result.Errors > 0 ||
		result.Stuck > 0 {
		return fmt.Errorf(
			"challenges: %d failed, %d errors, %d stuck",
			result.Failed, result.Errors, result.Stuck,
		)
	}

	return nil
}

// printChallengeSummary outputs a formatted summary of results.
func printChallengeSummary(result *challenges.OrchestratorResult) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("CHALLENGE RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total:    %d\n", result.Total)
	fmt.Printf("Passed:   %d\n", result.Passed)
	fmt.Printf("Failed:   %d\n", result.Failed)
	fmt.Printf("Skipped:  %d\n", result.Skipped)
	fmt.Printf("TimedOut: %d\n", result.TimedOut)
	fmt.Printf("Stuck:    %d\n", result.Stuck)
	fmt.Printf("Errors:   %d\n", result.Errors)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Println(strings.Repeat("=", 60))

	// Print per-challenge details if verbose.
	if result.Results != nil {
		for _, r := range result.Results {
			icon := "?"
			switch r.Status {
			case "passed":
				icon = "OK"
			case "failed":
				icon = "FAIL"
			case "stuck":
				icon = "STUCK"
			case "timed_out":
				icon = "TIMEOUT"
			case "skipped":
				icon = "SKIP"
			case "error":
				icon = "ERROR"
			}
			fmt.Printf("  [%-7s] %s (%v)\n",
				icon, r.ChallengeID, r.Duration)
			if r.Error != "" {
				fmt.Printf("            %s\n", r.Error)
			}
		}
	}
	fmt.Println()
}

// getProjectRoot returns the HelixAgent project root directory.
func getProjectRoot() string {
	// Try to detect from working directory.
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
