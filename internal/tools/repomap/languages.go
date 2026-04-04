// Package repomap provides language-specific symbol extraction
package repomap

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// extractGoSymbols extracts Go symbols using regex
func (r *RepoMap) extractGoSymbols(path string) ([]Symbol, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var symbols []Symbol
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex patterns for Go
	funcPattern := regexp.MustCompile(`^func\s+(?:\([^)]+\)\s+)?(\w+)`)
	typePattern := regexp.MustCompile(`^type\s+(\w+)`)
	constPattern := regexp.MustCompile(`^const\s+(\w+)`)
	varPattern := regexp.MustCompile(`^var\s+(\w+)`)
	interfacePattern := regexp.MustCompile(`^type\s+(\w+)\s+interface`)
	structPattern := regexp.MustCompile(`^type\s+(\w+)\s+struct`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Function
		if matches := funcPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:       name,
				Type:       "function",
				File:       path,
				Line:       lineNum,
				Language:   "Go",
				Signature:  strings.TrimSpace(line),
				IsExported: isExported(name),
			})
		}

		// Interface
		if matches := interfacePattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:       name,
				Type:       "interface",
				File:       path,
				Line:       lineNum,
				Language:   "Go",
				IsExported: isExported(name),
			})
			continue
		}

		// Struct
		if matches := structPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:       name,
				Type:       "struct",
				File:       path,
				Line:       lineNum,
				Language:   "Go",
				IsExported: isExported(name),
			})
			continue
		}

		// Type (other)
		if matches := typePattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:       name,
				Type:       "type",
				File:       path,
				Line:       lineNum,
				Language:   "Go",
				IsExported: isExported(name),
			})
		}

		// Const
		if matches := constPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:       name,
				Type:       "const",
				File:       path,
				Line:       lineNum,
				Language:   "Go",
				IsExported: isExported(name),
			})
		}

		// Var
		if matches := varPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:       name,
				Type:       "var",
				File:       path,
				Line:       lineNum,
				Language:   "Go",
				IsExported: isExported(name),
			})
		}
	}

	return symbols, scanner.Err()
}

// extractPythonSymbols extracts Python symbols
func (r *RepoMap) extractPythonSymbols(path string) ([]Symbol, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var symbols []Symbol
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex patterns for Python
	funcPattern := regexp.MustCompile(`^def\s+(\w+)`)
	classPattern := regexp.MustCompile(`^class\s+(\w+)`)
	constPattern := regexp.MustCompile(`^([A-Z][A-Z_0-9]*)\s*=`)

	inClass := false
	classIndent := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Track class context
		if inClass && indent <= classIndent {
			inClass = false
		}

		// Function
		if matches := funcPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symType := "function"
			if inClass {
				symType = "method"
			}
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     symType,
				File:     path,
				Line:     lineNum,
				Language: "Python",
			})
		}

		// Class
		if matches := classPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			inClass = true
			classIndent = indent
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     "class",
				File:     path,
				Line:     lineNum,
				Language: "Python",
			})
		}

		// Constant (ALL_CAPS)
		if matches := constPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     "const",
				File:     path,
				Line:     lineNum,
				Language: "Python",
			})
		}
	}

	return symbols, scanner.Err()
}

// extractJSSymbols extracts JavaScript/TypeScript symbols
func (r *RepoMap) extractJSSymbols(path string) ([]Symbol, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var symbols []Symbol
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex patterns for JS/TS
	funcPattern := regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?function\s+(\w+)`)
	arrowPattern := regexp.MustCompile(`^(?:export\s+)?const\s+(\w+)\s*=\s*(?:async\s*)?\(`)
	classPattern := regexp.MustCompile(`^(?:export\s+)?class\s+(\w+)`)
	constPattern := regexp.MustCompile(`^(?:export\s+)?const\s+(\w+)\s*=`)
	interfacePattern := regexp.MustCompile(`^(?:export\s+)?interface\s+(\w+)`)
	typePattern := regexp.MustCompile(`^(?:export\s+)?type\s+(\w+)`)

	inClass := false
	classIndent := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Track class context
		if inClass && indent <= classIndent {
			inClass = false
		}

		// Skip comments and empty lines for method detection
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") ||
			strings.HasPrefix(trimmed, "*") || trimmed == "" {
			continue
		}

		// Function declaration
		if matches := funcPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     "function",
				File:     path,
				Line:     lineNum,
				Language: "JavaScript",
			})
			continue
		}

		// Arrow function
		if matches := arrowPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     "function",
				File:     path,
				Line:     lineNum,
				Language: "JavaScript",
			})
			continue
		}

		// Class
		if matches := classPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			inClass = true
			classIndent = indent
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     "class",
				File:     path,
				Line:     lineNum,
				Language: "JavaScript",
			})
			continue
		}

		// Interface (TypeScript)
		if matches := interfacePattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     "interface",
				File:     path,
				Line:     lineNum,
				Language: "TypeScript",
			})
			continue
		}

		// Type alias (TypeScript)
		if matches := typePattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     "type",
				File:     path,
				Line:     lineNum,
				Language: "TypeScript",
			})
			continue
		}

		// Const
		if matches := constPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			symbols = append(symbols, Symbol{
				Name:     name,
				Type:     "const",
				File:     path,
				Line:     lineNum,
				Language: "JavaScript",
			})
			continue
		}

		// Method (inside class) - simple detection
		if inClass && indent > classIndent {
			trimmedLine := strings.TrimSpace(line)
			// Match method pattern: methodName(...) {
			if idx := strings.Index(trimmedLine, "("); idx > 0 {
				name := strings.TrimSpace(trimmedLine[:idx])
				// Skip keywords and special methods
				if name == "constructor" || name == "if" || name == "while" || name == "for" || name == "switch" {
					continue
				}
				// Check it looks like a method name
				if !strings.Contains(name, " ") && !strings.Contains(name, "=") {
					symbols = append(symbols, Symbol{
						Name:     name,
						Type:     "method",
						File:     path,
						Line:     lineNum,
						Language: "JavaScript",
					})
				}
			}
		}
	}

	return symbols, scanner.Err()
}

// isExported checks if a name is exported (starts with uppercase)
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}
