package challenges

import "time"

// CategoryStallThresholds maps challenge category prefixes to
// their stall detection thresholds. Categories with longer
// expected silence periods (e.g., provider verification that
// waits for API responses) have higher thresholds.
var CategoryStallThresholds = map[string]time.Duration{
	"provider":    120 * time.Second,
	"performance": 180 * time.Second,
	"debate":      120 * time.Second,
	"security":    120 * time.Second,
	"mcp":         90 * time.Second,
	"cli":         90 * time.Second,
	"bigdata":     120 * time.Second,
	"memory":      90 * time.Second,
	"default":     60 * time.Second,
}

// StallThresholdForCategory returns the stall threshold for
// the given category. Falls back to "default" if the category
// is not found.
func StallThresholdForCategory(category string) time.Duration {
	if d, ok := CategoryStallThresholds[category]; ok {
		return d
	}
	return CategoryStallThresholds["default"]
}
