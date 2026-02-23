// Package cmd provides tests for the get pods command
package cmd

import (
	"testing"

	"github.com/Rancheroo/r8s/internal/bundle"
)

func TestParseDescribeArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantKind   string
		wantBundle string
		wantName   string
	}{
		{
			name:       "3 args - kind bundle name",
			args:       []string{"pod", "./bundle/", "nginx"},
			wantKind:   "pod",
			wantBundle: "./bundle/",
			wantName:   "nginx",
		},
		{
			name:       "2 args - kind/name bundle",
			args:       []string{"pod/nginx", "./bundle/"},
			wantKind:   "pod",
			wantBundle: "./bundle/",
			wantName:   "nginx",
		},
		{
			name:       "2 args - bundle name (auto-detect)",
			args:       []string{"./bundle/", "nginx"},
			wantKind:   "",  // Auto-detect returns empty kind
			wantBundle: "./bundle/",
			wantName:   "nginx",
		},
		{
			name:       "pods alias normalized",
			args:       []string{"pods", "./bundle/", "nginx"},
			wantKind:   "pod",
			wantBundle: "./bundle/",
			wantName:   "nginx",
		},
		{
			name:       "nodes alias normalized",
			args:       []string{"nodes", "./bundle/", "node-1"},
			wantKind:   "node",
			wantBundle: "./bundle/",
			wantName:   "node-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKind, gotBundle, gotName := parseDescribeArgs(tt.args)
			if gotKind != tt.wantKind {
				t.Errorf("parseDescribeArgs() kind = %v, want %v", gotKind, tt.wantKind)
			}
			if gotBundle != tt.wantBundle {
				t.Errorf("parseDescribeArgs() bundle = %v, want %v", gotBundle, tt.wantBundle)
			}
			if gotName != tt.wantName {
				t.Errorf("parseDescribeArgs() name = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestNormalizeKind(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"pod", "pod"},
		{"pods", "pod"},
		{"po", "pod"},
		{"node", "node"},
		{"nodes", "node"},
		{"no", "node"},
		{"deployment", "deployment"},
		{"deployments", "deployment"},
		{"deploy", "deployment"},
		{"service", "service"},
		{"services", "service"},
		{"svc", "service"},
		{"configmap", "configmap"},
		{"configmaps", "configmap"},
		{"cm", "configmap"},
		{"event", "event"},
		{"events", "event"},
		{"ev", "event"},
		{"daemonset", "daemonset"},
		{"daemonsets", "daemonset"},
		{"ds", "daemonset"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeKind(tt.input)
			if got != tt.want {
				t.Errorf("normalizeKind(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeResource(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"po", "pods"},
		{"pod", "pods"},
		{"pods", "pods"},
		{"no", "nodes"},
		{"node", "nodes"},
		{"nodes", "nodes"},
		{"ns", "namespaces"},
		{"deploy", "deployments"},
		{"svc", "services"},
		{"ev", "events"},
		{"cm", "configmaps"},
		{"ds", "daemonsets"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeResource(tt.input)
			if got != tt.want {
				t.Errorf("normalizeResource(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidResource(t *testing.T) {
	tests := []struct {
		resource string
		valid    bool
	}{
		{"pods", true},
		{"pod", true},
		{"po", true},
		{"nodes", true},
		{"node", true},
		{"no", true},
		{"namespaces", true},
		{"namespace", true},
		{"ns", true},
		{"deployments", true},
		{"deployment", true},
		{"deploy", true},
		{"services", true},
		{"service", true},
		{"svc", true},
		{"events", true},
		{"event", true},
		{"ev", true},
		{"configmaps", true},
		{"configmap", true},
		{"cm", true},
		{"daemonsets", true},
		{"daemonset", true},
		{"ds", true},
		{"invalid", false},
		{"pods ", false}, // trailing space
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			got := isValidResource(tt.resource)
			if got != tt.valid {
				t.Errorf("isValidResource(%q) = %v, want %v", tt.resource, got, tt.valid)
			}
		})
	}
}

func TestDetermineRoles(t *testing.T) {
	tests := []struct {
		name             string
		nodeName         string
		manifestNodeName string
		want             string
	}{
		{
			name:             "control plane node",
			nodeName:         "server-1",
			manifestNodeName: "server-1",
			want:             "control-plane", // The actual implementation returns just control-plane for server-1
		},
		{
			name:             "worker node",
			nodeName:         "worker-1",
			manifestNodeName: "server-1",
			want:             "worker",
		},
		{
			name:             "etcd node",
			nodeName:         "etcd-1",
			manifestNodeName: "server-1",
			want:             "etcd",
		},
		{
			name:             "agent node",
			nodeName:         "agent-1",
			manifestNodeName: "server-1",
			want:             "worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineRoles(tt.nodeName, tt.manifestNodeName)
			if got != tt.want {
				t.Errorf("determineRoles(%q, %q) = %q, want %q", tt.nodeName, tt.manifestNodeName, got, tt.want)
			}
		})
	}
}

func TestFormatRoles(t *testing.T) {
	tests := []struct {
		name string
		nc   bundle.NodeConditions
		want string
	}{
		{
			name: "control plane only",
			nc: bundle.NodeConditions{
				Name:           "node-1",
				IsControlPlane: true,
			},
			want: "control-plane",
		},
		{
			name: "control plane and etcd",
			nc: bundle.NodeConditions{
				Name:           "node-1",
				IsControlPlane: true,
				IsEtcd:         true,
			},
			want: "control-plane,etcd",
		},
		{
			name: "worker only",
			nc: bundle.NodeConditions{
				Name:     "node-1",
				IsWorker: true,
			},
			want: "worker",
		},
		{
			name: "none",
			nc: bundle.NodeConditions{
				Name: "node-1",
			},
			want: "<none>",
		},
		{
			name: "fallback to roles field",
			nc: bundle.NodeConditions{
				Name:  "node-1",
				Roles: "master,worker",
			},
			want: "master,worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRoles(tt.nc)
			if got != tt.want {
				t.Errorf("formatRoles() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"this is a long message", 10, "this is..."},
		{"exactly ten", 11, "exactly ten"},
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := truncateString(tt.s, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}
