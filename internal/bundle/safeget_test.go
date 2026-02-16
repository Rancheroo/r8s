package bundle

import (
	"testing"
)

// --- SafeString tests ---

func TestSafeString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"empty string", "", ""},
		{"int", 42, "42"},
		{"int64", int64(42), "42"},
		{"float64", 3.14, "3.14"},
		{"bool", true, "true"},
		{"map", map[string]string{"a": "b"}, "map[a:b]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeString(tt.input)
			if got != tt.expected {
				t.Errorf("SafeString(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// --- SafeStringDefault tests ---

func TestSafeStringDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		defaultVal   string
		expected     string
	}{
		{"nil uses default", nil, "default", "default"},
		{"string value", "hello", "default", "hello"},
		{"int falls back to default", 42, "default", "default"},
		{"empty string is valid", "", "default", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeStringDefault(tt.input, tt.defaultVal)
			if got != tt.expected {
				t.Errorf("SafeStringDefault(%v, %q) = %q, want %q", tt.input, tt.defaultVal, got, tt.expected)
			}
		})
	}
}

// --- SafeInt tests ---

func TestSafeInt(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"nil", nil, 0},
		{"int", 42, 42},
		{"int64", int64(42), 42},
		{"float64", 3.14, 3},
		{"float64 truncates", 3.99, 3},
		{"string number", "42", 42},
		{"string invalid", "abc", 0},
		{"string with prefix", "123abc", 123},
		{"string negative", "-5", -5},
		{"empty string", "", 0},
		{"bool", true, 0},
		{"map", map[string]string{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeInt(tt.input)
			if got != tt.expected {
				t.Errorf("SafeInt(%v) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// --- SafeMap tests ---

func TestSafeMap(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		hasValue bool
	}{
		{"nil", nil, false},
		{"valid map", map[string]interface{}{"a": 1}, true},
		{"empty map", map[string]interface{}{}, true},
		{"wrong type", map[int]string{1: "a"}, false},
		{"string", "hello", false},
		{"int", 42, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeMap(tt.input)
			hasGot := got != nil
			if hasGot != tt.hasValue {
				t.Errorf("SafeMap(%v) returned nil=%v, expected nil=%v", tt.input, !hasGot, !tt.hasValue)
			}
		})
	}
}

// --- SafeMapValue tests ---

func TestSafeMapValue(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{"nil map", nil, "key", ""},
		{"key exists", map[string]interface{}{"key": "value"}, "key", "value"},
		{"key missing", map[string]interface{}{"other": "value"}, "key", ""},
		{"empty map", map[string]interface{}{}, "key", ""},
		{"int value", map[string]interface{}{"count": 42}, "count", "42"},
		{"nil value", map[string]interface{}{"key": nil}, "key", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeMapValue(tt.m, tt.key)
			if got != tt.expected {
				t.Errorf("SafeMapValue(%v, %q) = %q, want %q", tt.m, tt.key, got, tt.expected)
			}
		})
	}
}

// --- SafeNestedString tests ---

func TestSafeNestedString(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		keys     []string
		expected string
	}{
		{"nil map", nil, []string{"a"}, ""},
		{"no keys", map[string]interface{}{}, []string{}, ""},
		{"single key", map[string]interface{}{"name": "test"}, []string{"name"}, "test"},
		{"nested", map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "pod-1",
			},
		}, []string{"metadata", "name"}, "pod-1"},
		{"deeply nested", map[string]interface{}{
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"nodeName": "worker-1",
					},
				},
			},
		}, []string{"spec", "template", "spec", "nodeName"}, "worker-1"},
		{"missing intermediate", map[string]interface{}{
			"spec": "not-a-map",
		}, []string{"spec", "name"}, ""},
		{"missing final key", map[string]interface{}{
			"metadata": map[string]interface{}{},
		}, []string{"metadata", "name"}, ""},
		{"intermediate nil", map[string]interface{}{
			"spec": nil,
		}, []string{"spec", "name"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeNestedString(tt.m, tt.keys...)
			if got != tt.expected {
				t.Errorf("SafeNestedString(%v, %v) = %q, want %q", tt.m, tt.keys, got, tt.expected)
			}
		})
	}
}

// --- SafeNestedMap tests ---

func TestSafeNestedMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		keys     []string
		hasValue bool
	}{
		{"nil map", nil, []string{"a"}, false},
		{"no keys", map[string]interface{}{}, []string{}, false},
		{"single key", map[string]interface{}{"data": map[string]interface{}{}}, []string{"data"}, true},
		{"nested", map[string]interface{}{
			"spec": map[string]interface{}{
				"template": map[string]interface{}{},
			},
		}, []string{"spec", "template"}, true},
		{"missing intermediate", map[string]interface{}{
			"spec": "not-a-map",
		}, []string{"spec", "template"}, false},
		{"missing final key", map[string]interface{}{
			"spec": map[string]interface{}{},
		}, []string{"spec", "template"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeNestedMap(tt.m, tt.keys...)
			hasGot := got != nil
			if hasGot != tt.hasValue {
				t.Errorf("SafeNestedMap(%v, %v) returned nil=%v, expected nil=%v", tt.m, tt.keys, !hasGot, !tt.hasValue)
			}
		})
	}
}

// --- SafeSlice tests ---

func TestSafeSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		hasValue bool
	}{
		{"nil", nil, false},
		{"valid slice", []interface{}{1, 2, 3}, true},
		{"empty slice", []interface{}{}, true},
		{"string slice", []string{"a", "b"}, false},
		{"int slice", []int{1, 2}, false},
		{"string", "hello", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeSlice(tt.input)
			hasGot := got != nil
			if hasGot != tt.hasValue {
				t.Errorf("SafeSlice(%v) returned nil=%v, expected nil=%v", tt.input, !hasGot, !tt.hasValue)
			}
		})
	}
}

// --- RecoverToError tests ---

func TestRecoverToError(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		fn := func() error {
			return nil
		}
		err := RecoverToError(fn, "test")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("with panic", func(t *testing.T) {
		fn := func() error {
			panic("something went wrong")
		}
		err := RecoverToError(fn, "test context")
		if err == nil {
			t.Error("expected error from panic recovery, got nil")
		}
		if err != nil && err.Error() != "test context: recovered from panic: something went wrong" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("panic with non-string", func(t *testing.T) {
		fn := func() error {
			panic(42)
		}
		err := RecoverToError(fn, "numeric panic")
		if err == nil {
			t.Error("expected error from panic recovery, got nil")
		}
	})
}
