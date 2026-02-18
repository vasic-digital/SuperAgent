package challenges

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnvFile reads a .env file and returns key-value pairs.
// Lines starting with # are comments. Empty lines are skipped.
// Values may be optionally quoted with single or double quotes.
func LoadEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		// Strip optional quotes.
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		if key != "" {
			env[key] = val
		}
	}

	return env, scanner.Err()
}

// MergeEnvFiles loads multiple .env files and merges them.
// Later files override earlier ones. Files that don't exist
// are silently skipped.
func MergeEnvFiles(paths ...string) map[string]string {
	merged := make(map[string]string)
	for _, p := range paths {
		env, err := LoadEnvFile(p)
		if err != nil {
			continue
		}
		for k, v := range env {
			merged[k] = v
		}
	}
	return merged
}
