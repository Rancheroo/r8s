package bundle

import (
	"testing"
)

func TestCorrelateWithPods(t *testing.T) {
	tests := []struct {
		name     string
		kills    []DMesgOOMKill
		pods     map[string]int
		expected map[string]bool
	}{
		{
			name:     "no kills no pods",
			kills:    nil,
			pods:     map[string]int{},
			expected: map[string]bool{},
		},
		{
			name: "no pods to match",
			kills: []DMesgOOMKill{
				{VictimName: "nginx", VictimPID: 100},
			},
			pods:     map[string]int{},
			expected: map[string]bool{},
		},
		{
			name:     "no kills with pods present",
			kills:    nil,
			pods:     map[string]int{"nginx-abc123": 1},
			expected: map[string]bool{},
		},
		{
			name: "exact substring match victim in pod name",
			kills: []DMesgOOMKill{
				{VictimName: "nginx", VictimPID: 100},
			},
			pods:     map[string]int{"nginx-deployment-abc123": 1},
			expected: map[string]bool{"nginx-deployment-abc123": true},
		},
		{
			name: "exact substring match pod in victim name",
			kills: []DMesgOOMKill{
				{VictimName: "my-nginx-worker", VictimPID: 100},
			},
			pods:     map[string]int{"nginx": 1},
			expected: map[string]bool{"nginx": true},
		},
		{
			name: "case insensitive match",
			kills: []DMesgOOMKill{
				{VictimName: "NGINX", VictimPID: 100},
			},
			pods:     map[string]int{"nginx-pod-xyz": 1},
			expected: map[string]bool{"nginx-pod-xyz": true},
		},
		{
			name: "short victim name skipped (less than 3 chars)",
			kills: []DMesgOOMKill{
				{VictimName: "sh", VictimPID: 1},
			},
			pods:     map[string]int{"shell-runner-abc": 1},
			expected: map[string]bool{},
		},
		{
			name: "exactly 3 char victim name not skipped",
			kills: []DMesgOOMKill{
				{VictimName: "app", VictimPID: 1},
			},
			pods:     map[string]int{"app-server-xyz": 1},
			expected: map[string]bool{"app-server-xyz": true},
		},
		{
			name: "no match when names differ",
			kills: []DMesgOOMKill{
				{VictimName: "postgres", VictimPID: 200},
			},
			pods:     map[string]int{"redis-cache-abc": 1},
			expected: map[string]bool{},
		},
		{
			name: "multiple kills correlate with multiple pods",
			kills: []DMesgOOMKill{
				{VictimName: "nginx", VictimPID: 100},
				{VictimName: "java", VictimPID: 200},
			},
			pods: map[string]int{
				"nginx-frontend-abc": 1,
				"java-backend-xyz":   2,
				"redis-cache-123":    1,
			},
			expected: map[string]bool{
				"nginx-frontend-abc": true,
				"java-backend-xyz":   true,
			},
		},
		{
			name: "one kill matches multiple pods",
			kills: []DMesgOOMKill{
				{VictimName: "nginx", VictimPID: 100},
			},
			pods: map[string]int{
				"nginx-frontend-abc": 1,
				"nginx-backend-xyz":  2,
			},
			expected: map[string]bool{
				"nginx-frontend-abc": true,
				"nginx-backend-xyz":  true,
			},
		},
		{
			name: "mixed short and valid victim names",
			kills: []DMesgOOMKill{
				{VictimName: "ab", VictimPID: 1},
				{VictimName: "nginx", VictimPID: 2},
			},
			pods: map[string]int{
				"abc-pod":   1,
				"nginx-pod": 1,
			},
			expected: map[string]bool{
				"nginx-pod": true,
			},
		},
		{
			name: "nil podEvents map",
			kills: []DMesgOOMKill{
				{VictimName: "nginx", VictimPID: 100},
			},
			pods:     nil,
			expected: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &DMesgAnalysis{
				OOMKills: tt.kills,
			}

			got := analysis.CorrelateWithPods(tt.pods)

			if len(got) != len(tt.expected) {
				t.Errorf("expected %d correlated pods, got %d: %v", len(tt.expected), len(got), got)
				return
			}
			for k := range tt.expected {
				if !got[k] {
					t.Errorf("expected pod %q to be correlated, but it wasn't", k)
				}
			}
		})
	}
}
