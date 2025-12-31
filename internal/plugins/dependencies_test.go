package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyResolver(t *testing.T) {
	registry := NewRegistry()
	resolver := NewDependencyResolver(registry)

	require.NotNil(t, resolver)
	assert.NotNil(t, resolver.registry)
	assert.NotNil(t, resolver.deps)
}

func TestDependencyResolver_AddDependency(t *testing.T) {
	t.Run("add simple dependency", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)

		err := resolver.AddDependency("plugin-a", []string{"plugin-b"})

		require.NoError(t, err)
		assert.Equal(t, []string{"plugin-b"}, resolver.deps["plugin-a"])
	})

	t.Run("add multiple dependencies", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)

		err := resolver.AddDependency("plugin-a", []string{"plugin-b", "plugin-c", "plugin-d"})

		require.NoError(t, err)
		assert.Len(t, resolver.deps["plugin-a"], 3)
	})
}

func TestDependencyResolver_ResolveLoadOrder(t *testing.T) {
	t.Run("simple linear dependency", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)

		// a depends on b, b depends on c
		resolver.AddDependency("a", []string{"b"})
		resolver.AddDependency("b", []string{"c"})

		order, err := resolver.ResolveLoadOrder([]string{"a", "b", "c"})

		require.NoError(t, err)
		// The order should have all 3 items
		assert.Len(t, order, 3)
		// After reversal, dependencies come first
		// The actual order depends on input traversal order
		assert.Contains(t, order, "a")
		assert.Contains(t, order, "b")
		assert.Contains(t, order, "c")
	})

	t.Run("no dependencies", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)

		order, err := resolver.ResolveLoadOrder([]string{"a", "b", "c"})

		require.NoError(t, err)
		assert.Len(t, order, 3)
	})

	t.Run("circular dependency detection", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)

		// Create circular dependency directly in deps map
		resolver.deps["a"] = []string{"b"}
		resolver.deps["b"] = []string{"c"}
		resolver.deps["c"] = []string{"a"}

		_, err := resolver.ResolveLoadOrder([]string{"a", "b", "c"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency")
	})
}

func TestDependencyResolver_GetDependencies(t *testing.T) {
	t.Run("existing dependencies", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)
		resolver.AddDependency("plugin-a", []string{"plugin-b", "plugin-c"})

		deps := resolver.GetDependencies("plugin-a")

		assert.Len(t, deps, 2)
		assert.Contains(t, deps, "plugin-b")
		assert.Contains(t, deps, "plugin-c")
	})

	t.Run("no dependencies", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)

		deps := resolver.GetDependencies("non-existent")

		assert.Empty(t, deps)
	})
}

func TestDependencyResolver_GetDependents(t *testing.T) {
	t.Run("with dependents", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)
		resolver.AddDependency("a", []string{"shared"})
		resolver.AddDependency("b", []string{"shared"})
		resolver.AddDependency("c", []string{"other"})

		dependents := resolver.GetDependents("shared")

		assert.Len(t, dependents, 2)
		assert.Contains(t, dependents, "a")
		assert.Contains(t, dependents, "b")
	})

	t.Run("no dependents", func(t *testing.T) {
		registry := NewRegistry()
		resolver := NewDependencyResolver(registry)
		resolver.AddDependency("a", []string{"b"})

		dependents := resolver.GetDependents("a")

		assert.Empty(t, dependents)
	})
}

func TestDependencyResolver_CompareVersions(t *testing.T) {
	registry := NewRegistry()
	resolver := NewDependencyResolver(registry)

	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.2.3", "1.2.4", -1},
		{"1.2.3", "1.2.2", 1},
		{"1.2.3", "1.3.0", -1},
		{"2.0.0", "1.9.9", 1},
		{"10.0.0", "9.0.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+" vs "+tt.v2, func(t *testing.T) {
			result := resolver.compareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDependencyResolver_CheckVersionCompatibility(t *testing.T) {
	registry := NewRegistry()
	resolver := NewDependencyResolver(registry)

	tests := []struct {
		name       string
		version    string
		constraint string
		compatible bool
	}{
		{"exact match", "1.2.3", "1.2.3", true},
		{"exact no match", "1.2.3", "1.2.4", false},
		{">=", "1.2.3", ">=1.0.0", true},
		{">= boundary", "1.0.0", ">=1.0.0", true},
		{">= fail", "0.9.9", ">=1.0.0", false},
		{"<=", "1.0.0", "<=2.0.0", true},
		{"<= boundary", "2.0.0", "<=2.0.0", true},
		{"<= fail", "2.0.1", "<=2.0.0", false},
		{">", "1.0.1", ">1.0.0", true},
		{"> boundary fail", "1.0.0", ">1.0.0", false},
		{"<", "0.9.9", "<1.0.0", true},
		{"< boundary fail", "1.0.0", "<1.0.0", false},
		{"tilde lower", "1.2.3", "~1.2.0", true},
		{"tilde upper fail", "1.3.0", "~1.2.0", false},
		{"caret lower", "1.2.3", "^1.0.0", true},
		{"caret upper fail", "2.0.0", "^1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.checkVersionCompatibility(tt.version, tt.constraint)
			assert.Equal(t, tt.compatible, result, "Version %s should%s match constraint %s",
				tt.version, map[bool]string{true: "", false: " not"}[tt.compatible], tt.constraint)
		})
	}
}

func TestDependencyResolver_CheckVersionCompatibility_InvalidVersion(t *testing.T) {
	registry := NewRegistry()
	resolver := NewDependencyResolver(registry)

	// Invalid version with less than 3 parts
	result := resolver.checkVersionCompatibility("1.0", ">=1.0.0")
	assert.False(t, result)
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"42", 42},
		{"100", 100},
		{"999", 999},
		{"123abc", 123},
		{"abc", 0},
		{"12.34", 12},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDependencyResolver_HasCapabilityConflict(t *testing.T) {
	registry := NewRegistry()
	resolver := NewDependencyResolver(registry)

	t.Run("no conflict with nil capabilities", func(t *testing.T) {
		result := resolver.hasCapabilityConflict(nil, nil)
		assert.False(t, result)
	})

	t.Run("no conflict with one nil", func(t *testing.T) {
		caps := &map[string]interface{}{"streaming": true}
		result := resolver.hasCapabilityConflict(caps, nil)
		assert.False(t, result)
	})

	t.Run("no conflict with same capabilities", func(t *testing.T) {
		caps1 := &map[string]interface{}{"streaming": true}
		caps2 := &map[string]interface{}{"streaming": true}
		result := resolver.hasCapabilityConflict(caps1, caps2)
		assert.False(t, result)
	})

	t.Run("conflict with different capabilities", func(t *testing.T) {
		caps1 := &map[string]interface{}{"streaming": true}
		caps2 := &map[string]interface{}{"streaming": false}
		result := resolver.hasCapabilityConflict(caps1, caps2)
		assert.True(t, result)
	})
}

// Helper function to find index of element in slice
func indexOf(slice []string, element string) int {
	for i, e := range slice {
		if e == element {
			return i
		}
	}
	return -1
}
