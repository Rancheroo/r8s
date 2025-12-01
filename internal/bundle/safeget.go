// Package bundle provides safe getter utilities for defensive data access.
// Real kubectl bundles violate every schema - never trust interface{} conversions.
package bundle

import (
	"fmt"
)

// SafeString safely converts an interface{} to string, returning empty string if nil or wrong type.
// This prevents panic on nil interface conversions that plague bundle parsing.
func SafeString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	// Fallback: try to convert numeric types to string
	return fmt.Sprintf("%v", v)
}

// SafeStringDefault safely converts an interface{} to string with a custom default value.
func SafeStringDefault(v interface{}, defaultVal string) string {
	if v == nil {
		return defaultVal
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultVal
}

// SafeInt safely converts an interface{} to int, returning 0 if nil or wrong type.
func SafeInt(v interface{}) int {
	if v == nil {
		return 0
	}

	// Try various numeric types
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		// Try to parse string to int
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	default:
		return 0
	}
}

// SafeMap safely converts an interface{} to map[string]interface{}, returning nil if wrong type.
func SafeMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return nil
}

// SafeMapValue safely retrieves a value from a map[string]interface{} as string.
// Returns empty string if map is nil, key doesn't exist, or value is wrong type.
func SafeMapValue(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	return SafeString(m[key])
}

// SafeNestedString retrieves a nested string value from a map using a path of keys.
// Example: SafeNestedString(data, "metadata", "name") accesses data["metadata"]["name"]
// Returns empty string if any level is nil or wrong type.
func SafeNestedString(m map[string]interface{}, keys ...string) string {
	if m == nil || len(keys) == 0 {
		return ""
	}

	current := m
	for i, key := range keys {
		val, exists := current[key]
		if !exists {
			return ""
		}

		// Last key - return as string
		if i == len(keys)-1 {
			return SafeString(val)
		}

		// Not last key - must be a map to continue
		current = SafeMap(val)
		if current == nil {
			return ""
		}
	}

	return ""
}

// SafeNestedMap retrieves a nested map from a map using a path of keys.
// Returns nil if any level is nil or wrong type.
func SafeNestedMap(m map[string]interface{}, keys ...string) map[string]interface{} {
	if m == nil || len(keys) == 0 {
		return nil
	}

	current := m
	for _, key := range keys {
		val, exists := current[key]
		if !exists {
			return nil
		}

		current = SafeMap(val)
		if current == nil {
			return nil
		}
	}

	return current
}

// SafeSlice safely converts an interface{} to []interface{}, returning nil if wrong type.
func SafeSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

// RecoverToError wraps a function with panic recovery, converting panics to errors.
// Use this around any bundle parsing code that might panic on malformed data.
func RecoverToError(fn func() error, context string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s: recovered from panic: %v", context, r)
		}
	}()

	return fn()
}
