//go:build performance
// +build performance

package performance

import (
	"context"
	"fmt"
	"testing"

	"dev.helix.agent/internal/agents"
	"dev.helix.agent/internal/skills"
)

// =============================================================================
// Skill benchmark helpers
// =============================================================================

// newBenchSkillService creates a SkillService pre-populated with a realistic
// set of skills covering multiple categories.  No disk I/O is performed.
func newBenchSkillService() *skills.Service {
	config := skills.DefaultSkillConfig()
	config.EnableSemanticMatching = false // no LLM in benchmarks
	config.MinConfidence = 0.3
	svc := skills.NewService(config)
	svc.Start()

	categories := []string{"development", "devops", "testing", "security", "database"}
	triggers := [][]string{
		{"run tests", "go test", "test coverage", "unit test", "integration test"},
		{"docker build", "deploy container", "kubernetes", "k8s", "helm install"},
		{"write tests", "test automation", "e2e test", "pytest", "jest"},
		{"security scan", "vulnerability check", "snyk", "gosec", "owasp"},
		{"run migration", "database schema", "sql query", "pg dump", "redis cache"},
	}

	// Register 48 skills (one per CLI agent slot) spread across categories
	for i := 0; i < 48; i++ {
		cat := categories[i%len(categories)]
		trig := triggers[i%len(triggers)]
		svc.RegisterSkill(&skills.Skill{
			Name:           fmt.Sprintf("skill-%02d-%s", i, cat),
			Description:    fmt.Sprintf("Skill %d for %s workflows", i, cat),
			Category:       cat,
			TriggerPhrases: trig,
			Instructions:   fmt.Sprintf("Perform %s operation %d.", cat, i),
		})
	}

	return svc
}

// =============================================================================
// Skill service benchmarks
// =============================================================================

// BenchmarkSkillService_FindSkills measures the latency of matching a query
// against a registry of 48 skills using keyword matching only.
func BenchmarkSkillService_FindSkills(b *testing.B) {
	svc := newBenchSkillService()
	defer svc.Shutdown() //nolint:errcheck

	ctx := context.Background()
	queries := []string{
		"run unit tests",
		"docker build image",
		"security vulnerability scan",
		"database migration",
		"write pytest tests",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = svc.FindSkills(ctx, queries[i%len(queries)])
	}
}

// BenchmarkSkillService_FindSkills_Parallel measures parallel skill-lookup
// throughput.
func BenchmarkSkillService_FindSkills_Parallel(b *testing.B) {
	svc := newBenchSkillService()
	defer svc.Shutdown() //nolint:errcheck

	ctx := context.Background()
	queries := []string{
		"run unit tests",
		"deploy kubernetes",
		"gosec scan",
		"redis cache",
		"e2e test automation",
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = svc.FindSkills(ctx, queries[i%len(queries)])
			i++
		}
	})
}

// BenchmarkSkillService_FindBestSkill measures the latency of returning a
// single best-matching skill.
func BenchmarkSkillService_FindBestSkill(b *testing.B) {
	svc := newBenchSkillService()
	defer svc.Shutdown() //nolint:errcheck

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = svc.FindBestSkill(ctx, "run go test coverage")
	}
}

// BenchmarkSkillService_GetSkill measures direct name-based skill lookup.
func BenchmarkSkillService_GetSkill(b *testing.B) {
	svc := newBenchSkillService()
	defer svc.Shutdown() //nolint:errcheck

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("skill-%02d-development", i%10)
		_, _ = svc.GetSkill(name)
	}
}

// BenchmarkSkillService_GetAllSkills measures the cost of enumerating all
// registered skills.
func BenchmarkSkillService_GetAllSkills(b *testing.B) {
	svc := newBenchSkillService()
	defer svc.Shutdown() //nolint:errcheck

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = svc.GetAllSkills()
	}
}

// BenchmarkSkillService_GetByCategory measures category-filtered lookup.
func BenchmarkSkillService_GetByCategory(b *testing.B) {
	svc := newBenchSkillService()
	defer svc.Shutdown() //nolint:errcheck

	categories := []string{"development", "devops", "testing", "security", "database"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = svc.GetSkillsByCategory(categories[i%len(categories)])
	}
}

// BenchmarkSkillService_SearchSkills measures free-text search across skill
// names and descriptions.
func BenchmarkSkillService_SearchSkills(b *testing.B) {
	svc := newBenchSkillService()
	defer svc.Shutdown() //nolint:errcheck

	terms := []string{"docker", "test", "security", "database", "deployment"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = svc.SearchSkills(terms[i%len(terms)])
	}
}

// BenchmarkSkillService_RegisterSkill measures the cost of registering a new
// skill into a running service.
func BenchmarkSkillService_RegisterSkill(b *testing.B) {
	config := skills.DefaultSkillConfig()
	config.EnableSemanticMatching = false
	svc := skills.NewService(config)
	svc.Start()
	defer svc.Shutdown() //nolint:errcheck

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		svc.RegisterSkill(&skills.Skill{
			Name:           fmt.Sprintf("bench-skill-%d", i),
			Description:    "Benchmark registration skill",
			Category:       "bench",
			TriggerPhrases: []string{fmt.Sprintf("trigger-%d", i)},
		})
	}
}

// =============================================================================
// CLI agent registry benchmarks
// =============================================================================

// BenchmarkCLIAgentRegistry_Lookup measures map lookup time for the static CLI
// agent registry which backs all 48 agent config generators.
func BenchmarkCLIAgentRegistry_Lookup(b *testing.B) {
	registry := agents.CLIAgentRegistry
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = registry[keys[i%len(keys)]]
	}
}

// BenchmarkCLIAgentRegistry_Enumerate measures the cost of iterating over all
// 48 agents in the registry (e.g. when building config lists).
func BenchmarkCLIAgentRegistry_Enumerate(b *testing.B) {
	registry := agents.CLIAgentRegistry

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, agent := range registry {
			_ = agent.Name
		}
	}
}

// BenchmarkSkillTypes_ParseAllowedTools measures AllowedTools parsing which
// runs on every skill load from disk.
func BenchmarkSkillTypes_ParseAllowedTools(b *testing.B) {
	inputs := []string{
		"",
		"Read",
		"Read, Write, Edit",
		"Read, Write, Edit, Bash, Glob, Grep, Git",
		"Bash(cmd:*), Glob(pattern:*.go), Read, Write, Edit, Test, Lint",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = skills.ParseAllowedTools(inputs[i%len(inputs)])
	}
}
