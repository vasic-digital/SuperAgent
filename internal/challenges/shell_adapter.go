package challenges

import (
	"os"
	"strings"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/registry"
)

// RegisterShellChallenges discovers shell scripts in the given
// directory and registers them as ShellChallenge instances.
func RegisterShellChallenges(
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

		scriptPath := scriptsDir + "/" + name

		sc := challenge.NewShellChallenge(
			challenge.ID(id),
			formatName(id),
			"Shell challenge: "+name,
			"shell",
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

// formatName converts a dash-separated ID to a human-readable
// name.
func formatName(id string) string {
	parts := strings.Split(id, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
