// Package unit provides unit tests for HelixAgent core functionality.
package unit

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestBasicJSONParsing verifies JSON parsing functionality used throughout HelixAgent
func TestBasicJSONParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid_json_object",
			input:   `{"key": "value", "number": 42}`,
			wantErr: false,
		},
		{
			name:    "valid_json_array",
			input:   `[1, 2, 3, "test"]`,
			wantErr: false,
		},
		{
			name:    "empty_object",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "invalid_json",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "nested_structure",
			input:   `{"outer": {"inner": {"deep": true}}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			err := json.Unmarshal([]byte(tt.input), &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStringOperations verifies string operations used in HelixAgent
func TestStringOperations(t *testing.T) {
	t.Parallel()

	t.Run("prefix_detection", func(t *testing.T) {
		tests := []struct {
			s      string
			prefix string
			want   bool
		}{
			{"skill_test", "skill_", true},
			{"skill.test", "skill.", true},
			{"helixagent.skill.test", "helixagent.skill.", true},
			{"other", "skill_", false},
		}

		for _, tt := range tests {
			if got := strings.HasPrefix(tt.s, tt.prefix); got != tt.want {
				t.Errorf("HasPrefix(%q, %q) = %v, want %v", tt.s, tt.prefix, got, tt.want)
			}
		}
	})

	t.Run("suffix_detection", func(t *testing.T) {
		tests := []struct {
			s      string
			suffix string
			want   bool
		}{
			{"test.go", ".go", true},
			{"test.json", ".json", true},
			{"test", ".go", false},
		}

		for _, tt := range tests {
			if got := strings.HasSuffix(tt.s, tt.suffix); got != tt.want {
				t.Errorf("HasSuffix(%q, %q) = %v, want %v", tt.s, tt.suffix, got, tt.want)
			}
		}
	})
}

// TestTimeOperations verifies time operations used in HelixAgent
func TestTimeOperations(t *testing.T) {
	t.Parallel()

	t.Run("time_since", func(t *testing.T) {
		start := time.Now()
		time.Sleep(10 * time.Millisecond)
		elapsed := time.Since(start)

		if elapsed < 10*time.Millisecond {
			t.Errorf("time.Since() = %v, want >= 10ms", elapsed)
		}
	})

	t.Run("time_parsing", func(t *testing.T) {
		tests := []struct {
			layout string
			value  string
			valid  bool
		}{
			{time.RFC3339, "2026-01-23T12:00:00Z", true},
			{time.RFC3339, "invalid", false},
		}

		for _, tt := range tests {
			_, err := time.Parse(tt.layout, tt.value)
			if (err == nil) != tt.valid {
				t.Errorf("time.Parse(%q, %q) validity = %v, want %v", tt.layout, tt.value, err == nil, tt.valid)
			}
		}
	})
}

// TestMapOperations verifies map operations used in HelixAgent
func TestMapOperations(t *testing.T) {
	t.Parallel()

	t.Run("map_access", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}

		if v, ok := m["a"]; !ok || v != 1 {
			t.Errorf("map[\"a\"] = %d, %v; want 1, true", v, ok)
		}

		if _, ok := m["c"]; ok {
			t.Error("map[\"c\"] should not exist")
		}
	})

	t.Run("map_deletion", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		delete(m, "a")

		if _, ok := m["a"]; ok {
			t.Error("map[\"a\"] should be deleted")
		}
		if len(m) != 1 {
			t.Errorf("len(m) = %d, want 1", len(m))
		}
	})
}

// TestSliceOperations verifies slice operations used in HelixAgent
func TestSliceOperations(t *testing.T) {
	t.Parallel()

	t.Run("slice_append", func(t *testing.T) {
		s := []int{1, 2}
		s = append(s, 3)

		if len(s) != 3 || s[2] != 3 {
			t.Errorf("append() result = %v, want [1 2 3]", s)
		}
	})

	t.Run("slice_copy", func(t *testing.T) {
		src := []int{1, 2, 3}
		dst := make([]int, len(src))
		copy(dst, src)

		// Modify source
		src[0] = 99

		if dst[0] != 1 {
			t.Error("copy() should create independent slice")
		}
	})
}

// TestInterfaceTypeAssertion verifies type assertions used in HelixAgent
func TestInterfaceTypeAssertion(t *testing.T) {
	t.Parallel()

	t.Run("successful_assertion", func(t *testing.T) {
		var i interface{} = "hello"
		s, ok := i.(string)
		if !ok || s != "hello" {
			t.Errorf("type assertion failed: %v, %v", s, ok)
		}
	})

	t.Run("failed_assertion", func(t *testing.T) {
		var i interface{} = 42
		_, ok := i.(string)
		if ok {
			t.Error("type assertion should fail for int to string")
		}
	})

	t.Run("nil_interface_assertion", func(t *testing.T) {
		var i interface{}
		_, ok := i.([]interface{})
		if ok {
			t.Error("type assertion on nil should fail")
		}
	})
}

// TestErrorHandling verifies error handling patterns used in HelixAgent
func TestErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("nil_error", func(t *testing.T) {
		var err error
		if err != nil {
			t.Error("nil error should be nil")
		}
	})

	t.Run("error_wrapping", func(t *testing.T) {
		baseErr := &testError{msg: "base error"}
		if baseErr.Error() != "base error" {
			t.Errorf("Error() = %q, want %q", baseErr.Error(), "base error")
		}
	})
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
