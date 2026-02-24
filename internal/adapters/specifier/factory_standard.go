//go:build nohelixspecifier

package specifier

// NewOptimalSpecAdapter returns nil when HelixSpecifier is opted out.
// Build with -tags nohelixspecifier to use this fallback.
// The existing SpecKitOrchestrator in internal/services/ will be used.
func NewOptimalSpecAdapter() *SpecAdapter {
	return nil
}

// IsHelixSpecifierEnabled returns false -- opted out.
func IsHelixSpecifierEnabled() bool {
	return false
}

// SpecifierBackendName returns the standard backend name.
func SpecifierBackendName() string {
	return "internal.speckit"
}
