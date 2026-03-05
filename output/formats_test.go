// Package output provides output formatting tests
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidFormats(t *testing.T) {
	formats := ValidFormats()
	expected := []string{"table", "json", "yaml", "wide", "name"}

	if len(formats) != len(expected) {
		t.Errorf("ValidFormats() returned %d formats, expected %d", len(formats), len(expected))
	}

	for i, f := range expected {
		if i >= len(formats) || formats[i] != f {
			t.Errorf("ValidFormats()[%d] = %s, expected %s", i, formats[i], f)
		}
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		format  string
		isValid bool
	}{
		{"table", true},
		{"json", true},
		{"yaml", true},
		{"wide", true},
		{"name", true},
		{"invalid", false},
		{"", false},
		{"TABLE", false}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := IsValid(tt.format)
			if result != tt.isValid {
				t.Errorf("IsValid(%q) = %v, expected %v", tt.format, result, tt.isValid)
			}
		})
	}
}

func TestOutputPods(t *testing.T) {
	pods := []PodRow{
		{
			Namespace: "default",
			Name:      "nginx-pod",
			Ready:     "1/1",
			Status:    "Running",
			Restarts:  0,
			Age:       "7d",
			Node:      "node-1",
			IP:        "10.42.0.5",
		},
		{
			Namespace: "kube-system",
			Name:      "coredns-abc",
			Ready:     "1/1",
			Status:    "Running",
			Restarts:  2,
			Age:       "14d",
			Node:      "node-1",
			IP:        "10.42.0.3",
		},
	}

	tests := []struct {
		name     string
		format   Format
		wantErr  bool
		contains []string
	}{
		{
			name:     "table format",
			format:   FormatTable,
			wantErr:  false,
			contains: []string{"nginx-pod", "coredns-abc", "Running"},
		},
		{
			name:     "json format",
			format:   FormatJSON,
			wantErr:  false,
			contains: []string{"\"name\":", "\"namespace\":", "nginx-pod"},
		},
		{
			name:     "yaml format",
			format:   FormatYAML,
			wantErr:  false,
			contains: []string{"name:", "namespace:", "nginx-pod"},
		},
		{
			name:     "wide format",
			format:   FormatWide,
			wantErr:  false,
			contains: []string{"nginx-pod", "10.42.0.5", "node-1"},
		},
		{
			name:     "name format",
			format:   FormatName,
			wantErr:  false,
			contains: []string{"nginx-pod", "coredns-abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			opts := Options{
				Format:        tt.format,
				ShowNamespace: false,
				AllNamespaces: false,
				NoHeaders:     false,
			}

			err := OutputPods(pods, opts)

			w.Close()
			os.Stdout = oldStdout

			if (err != nil) != tt.wantErr {
				t.Errorf("OutputPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output, _ := io.ReadAll(r)
			outputStr := string(output)

			for _, want := range tt.contains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("OutputPods() output missing %q, got:\n%s", want, outputStr)
				}
			}
		})
	}
}

func TestOutputNodes(t *testing.T) {
	nodes := []NodeRow{
		{
			Name:             "node-1",
			Status:           "Ready",
			Roles:            "control-plane,etcd",
			Age:              "30d",
			Version:          "v1.32.0",
			InternalIP:       "192.168.1.10",
			OSImage:          "Ubuntu 22.04",
			KernelVersion:    "5.15.0",
			ContainerRuntime: "containerd://2.0.0",
		},
	}

	tests := []struct {
		name     string
		format   Format
		wantErr  bool
		contains []string
	}{
		{
			name:     "table format",
			format:   FormatTable,
			wantErr:  false,
			contains: []string{"node-1", "Ready", "control-plane,etcd"},
		},
		{
			name:     "json format",
			format:   FormatJSON,
			wantErr:  false,
			contains: []string{"\"name\":", "\"status\":", "node-1"},
		},
		{
			name:     "wide format",
			format:   FormatWide,
			wantErr:  false,
			contains: []string{"node-1", "192.168.1.10", "Ubuntu 22.04"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			opts := Options{
				Format:    tt.format,
				NoHeaders: false,
			}

			err := OutputNodes(nodes, opts)

			w.Close()
			os.Stdout = oldStdout

			if (err != nil) != tt.wantErr {
				t.Errorf("OutputNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output, _ := io.ReadAll(r)
			for _, want := range tt.contains {
				if !strings.Contains(string(output), want) {
					t.Errorf("OutputNodes() output missing %q", want)
				}
			}
		})
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name    string
		created time.Time
		want    string
	}{
		{
			name:    "seconds",
			created: time.Now().Add(-30 * time.Second),
			want:    "30s",
		},
		{
			name:    "minutes",
			created: time.Now().Add(-5 * time.Minute),
			want:    "5m",
		},
		{
			name:    "hours",
			created: time.Now().Add(-3 * time.Hour),
			want:    "3h",
		},
		{
			name:    "days",
			created: time.Now().Add(-7 * 24 * time.Hour),
			want:    "7d",
		},
		{
			name:    "years",
			created: time.Now().Add(-400 * 24 * time.Hour),
			want:    "1y",
		},
		{
			name:    "zero time",
			created: time.Time{},
			want:    "<unknown>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAge(tt.created)
			if got != tt.want {
				t.Errorf("FormatAge() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		msg    string
		maxLen int
		want   string
	}{
		{"short message", 50, "short message"},
		{"this is a very long message that needs truncation", 20, "this is a very lo..."},
		{"exact fit!!!", 12, "exact fit!!!"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got := TruncateMessage(tt.msg, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateMessage() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "bundle not loaded",
			err:  os.ErrNotExist,
			want: "Error:", // Generic error message for os.ErrNotExist
		},
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "generic error",
			err:  os.ErrPermission,
			want: "Error:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatError(tt.err)
			if !strings.Contains(got, tt.want) {
				t.Errorf("FormatError() = %q, should contain %q", got, tt.want)
			}
		})
	}
}

func TestOutputPodsJSONStructure(t *testing.T) {
	pods := []PodRow{
		{
			Namespace: "default",
			Name:      "test-pod",
			Ready:     "1/1",
			Status:    "Running",
			Restarts:  0,
			Age:       "1d",
			Node:      "node-1",
			IP:        "10.0.0.1",
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := Options{Format: FormatJSON}
	err := OutputPods(pods, opts)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("OutputPods() error = %v", err)
	}

	var result []PodRow
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 pod, got %d", len(result))
	}

	if result[0].Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got %q", result[0].Name)
	}
}

func TestOutputPodsYAMLStructure(t *testing.T) {
	pods := []PodRow{
		{
			Namespace: "default",
			Name:      "test-pod",
			Ready:     "1/1",
			Status:    "Running",
			Restarts:  0,
			Age:       "1d",
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := Options{Format: FormatYAML}
	err := OutputPods(pods, opts)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("OutputPods() error = %v", err)
	}

	output, _ := io.ReadAll(r)

	// Verify YAML contains expected fields
	outputStr := string(output)
	expectedFields := []string{"namespace:", "name:", "ready:", "status:", "restarts:", "age:"}
	for _, field := range expectedFields {
		if !strings.Contains(outputStr, field) {
			t.Errorf("YAML output missing field %q", field)
		}
	}
}

// BenchmarkOutputPods benchmarks the pod output performance
func BenchmarkOutputPods(b *testing.B) {
	pods := make([]PodRow, 100)
	for i := 0; i < 100; i++ {
		pods[i] = PodRow{
			Namespace: "default",
			Name:      fmt.Sprintf("pod-%d", i),
			Ready:     "1/1",
			Status:    "Running",
			Restarts:  0,
			Age:       "1d",
		}
	}

	opts := Options{Format: FormatJSON, NoHeaders: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Redirect stdout to discard output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		OutputPods(pods, opts)
		w.Close()
		os.Stdout = oldStdout
		io.Copy(io.Discard, r)
	}
}

// BenchmarkOutputNodes benchmarks the node output performance
func BenchmarkOutputNodes(b *testing.B) {
	nodes := make([]NodeRow, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = NodeRow{
			Name:    fmt.Sprintf("node-%d", i),
			Status:  "Ready",
			Roles:   "worker",
			Age:     "30d",
			Version: "v1.32.0",
		}
	}

	opts := Options{Format: FormatJSON, NoHeaders: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Redirect stdout to discard output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		OutputNodes(nodes, opts)
		w.Close()
		os.Stdout = oldStdout
		io.Copy(io.Discard, r)
	}
}
