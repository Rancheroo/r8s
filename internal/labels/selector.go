package labels

import (
	"fmt"
	"strings"
)

// Selector matches labels against requirements.
type Selector struct {
	requirements []requirement
}

type requirement struct {
	key      string
	operator string
	value    string
}

// Parse parses a selector string into a Selector.
// Supports "key=value", "key==value", and "key!=value".
// Multiple requirements are separated by commas (AND logic).
func Parse(selectorStr string) (*Selector, error) {
	s := &Selector{}
	if selectorStr == "" {
		return s, nil
	}

	parts := strings.Split(selectorStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var req requirement
		if strings.Contains(part, "!=") {
			kv := strings.SplitN(part, "!=", 2)
			req = requirement{key: strings.TrimSpace(kv[0]), operator: "!=", value: strings.TrimSpace(kv[1])}
		} else if strings.Contains(part, "==") {
			kv := strings.SplitN(part, "==", 2)
			req = requirement{key: strings.TrimSpace(kv[0]), operator: "==", value: strings.TrimSpace(kv[1])}
		} else if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			req = requirement{key: strings.TrimSpace(kv[0]), operator: "=", value: strings.TrimSpace(kv[1])}
		} else {
			return nil, fmt.Errorf("invalid selector requirement: %s", part)
		}

		if req.key == "" {
			return nil, fmt.Errorf("invalid selector requirement: key cannot be empty")
		}
		s.requirements = append(s.requirements, req)
	}

	return s, nil
}

// Matches returns true if the labels satisfy all requirements in the selector.
func (s *Selector) Matches(labels map[string]string) bool {
	for _, req := range s.requirements {
		val, ok := labels[req.key]
		switch req.operator {
		case "=", "==":
			if !ok || val != req.value {
				return false
			}
		case "!=":
			if ok && val == req.value {
				return false
			}
		}
	}
	return true
}
