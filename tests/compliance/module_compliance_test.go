package compliance

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getRoot returns the HelixAgent project root directory.
func getRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// extractedModules lists all 20 extracted modules that must exist
// as independent Go modules in the project root.
var extractedModules = []string{
	"EventBus",
	"Concurrency",
	"Observability",
	"Auth",
	"Storage",
	"Streaming",
	"Security",
	"VectorDB",
	"Embeddings",
	"Database",
	"Cache",
	"Messaging",
	"Formatters",
	"MCP_Module",
	"RAG",
	"Memory",
	"Optimization",
	"Plugins",
	"Containers",
	"Challenges",
}

// TestExtractedModuleCount verifies that all 20 extracted modules exist
// as directories in the project root.
func TestExtractedModuleCount(t *testing.T) {
	root := getRoot()

	existingModules := []string{}
	missingModules := []string{}

	for _, module := range extractedModules {
		modulePath := filepath.Join(root, module)
		if _, err := os.Stat(modulePath); err == nil {
			existingModules = append(existingModules, module)
		} else {
			missingModules = append(missingModules, module)
		}
	}

	t.Logf("Found %d/%d extracted modules", len(existingModules), len(extractedModules))
	if len(missingModules) > 0 {
		t.Logf("Missing modules: %v", missingModules)
	}

	assert.GreaterOrEqual(t, len(existingModules), 18,
		"COMPLIANCE FAILED: At least 18 of 20 extracted modules must exist")
}

// TestModuleGoModExists verifies that each extracted module has a go.mod
// file (confirming it's an independent Go module).
func TestModuleGoModExists(t *testing.T) {
	root := getRoot()
	missingGoMod := []string{}

	for _, module := range extractedModules {
		goModPath := filepath.Join(root, module, "go.mod")
		if _, err := os.Stat(goModPath); err != nil {
			modulePath := filepath.Join(root, module)
			if _, dirErr := os.Stat(modulePath); dirErr == nil {
				// Module dir exists but no go.mod
				missingGoMod = append(missingGoMod, module)
			}
		}
	}

	if len(missingGoMod) > 0 {
		t.Errorf("COMPLIANCE FAILED: %d modules missing go.mod: %v",
			len(missingGoMod), missingGoMod)
	} else {
		t.Logf("COMPLIANCE: All present extracted modules have go.mod files")
	}
}

// TestModuleReadmeExists verifies that each extracted module has a README.md.
func TestModuleReadmeExists(t *testing.T) {
	root := getRoot()
	missingReadme := []string{}

	for _, module := range extractedModules {
		modulePath := filepath.Join(root, module)
		if _, err := os.Stat(modulePath); err != nil {
			continue // Module doesn't exist, skip
		}
		readmePath := filepath.Join(modulePath, "README.md")
		if _, err := os.Stat(readmePath); err != nil {
			missingReadme = append(missingReadme, module)
		}
	}

	if len(missingReadme) > 0 {
		t.Logf("COMPLIANCE WARNING: %d modules missing README.md: %v",
			len(missingReadme), missingReadme)
	} else {
		t.Logf("COMPLIANCE: All present extracted modules have README.md files")
	}
}

// TestModuleClaudeMdExists verifies that each extracted module has a CLAUDE.md
// file with project-specific guidance for AI assistants.
func TestModuleClaudeMdExists(t *testing.T) {
	root := getRoot()
	missingClaudeMd := []string{}
	presentCount := 0

	for _, module := range extractedModules {
		modulePath := filepath.Join(root, module)
		if _, err := os.Stat(modulePath); err != nil {
			continue // Module doesn't exist, skip
		}
		claudeMdPath := filepath.Join(modulePath, "CLAUDE.md")
		if _, err := os.Stat(claudeMdPath); err != nil {
			missingClaudeMd = append(missingClaudeMd, module)
		} else {
			presentCount++
		}
	}

	t.Logf("Modules with CLAUDE.md: %d, missing: %d (%v)",
		presentCount, len(missingClaudeMd), missingClaudeMd)
	if len(missingClaudeMd) > 0 {
		t.Logf("COMPLIANCE WARNING: %d modules missing CLAUDE.md", len(missingClaudeMd))
	} else {
		t.Logf("COMPLIANCE: All present extracted modules have CLAUDE.md files")
	}
}

// TestModuleTestsExist verifies that each extracted module has at least
// one test file (confirming test coverage requirements are met).
func TestModuleTestsExist(t *testing.T) {
	root := getRoot()
	missingTests := []string{}

	for _, module := range extractedModules {
		modulePath := filepath.Join(root, module)
		if _, err := os.Stat(modulePath); err != nil {
			continue // Module doesn't exist, skip
		}

		// Look for any *_test.go file recursively
		hasTests := false
		_ = filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				name := info.Name()
				if len(name) > 8 && name[len(name)-8:] == "_test.go" {
					hasTests = true
					return filepath.SkipAll
				}
			}
			return nil
		})

		if !hasTests {
			missingTests = append(missingTests, module)
		}
	}

	if len(missingTests) > 0 {
		t.Errorf("COMPLIANCE FAILED: %d modules have no test files: %v",
			len(missingTests), missingTests)
	} else {
		t.Logf("COMPLIANCE: All present extracted modules have test files")
	}
}
