package tui

import (
	"testing"
	"time"

	"github.com/Rancheroo/r8s/internal/config"
	"github.com/Rancheroo/r8s/internal/rancher"
	tea "github.com/charmbracelet/bubbletea"
)

// TestNewApp tests application initialization
func TestNewApp(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		wantOffline  bool
		wantViewType ViewType
		wantError    bool
	}{
		{
			name: "valid config creates app",
			config: &config.Config{
				CurrentProfile: "test",
				Profiles: []config.Profile{
					{
						Name:        "test",
						URL:         "https://test.rancher.com",
						BearerToken: "test-token",
					},
				},
			},
			wantOffline:  true,          // Bundle-only mode: always starts with demo bundle
			wantViewType: ViewAttention, // Bundle-only: always start with Attention Dashboard
			wantError:    false,
		},
		{
			name: "no profiles creates app with error",
			config: &config.Config{
				CurrentProfile: "missing",
				Profiles:       []config.Profile{},
			},
			wantOffline:  true,          // Bundle-only mode
			wantViewType: ViewAttention, // Bundle-only: Attention Dashboardf
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.config, "") // Empty bundle path for live mode

			if app == nil {
				t.Fatal("NewApp returned nil")
			}

			if tt.wantError {
				if app.error == "" {
					t.Error("Expected error to be set, but it was empty")
				}
			}

			if app.currentView.viewType != tt.wantViewType {
				t.Errorf("Expected view type %v, got %v", tt.wantViewType, app.currentView.viewType)
			}

			// Offline mode is expected when no live Rancher is available
			if app.offlineMode != tt.wantOffline {
				t.Errorf("Expected offlineMode %v, got %v", tt.wantOffline, app.offlineMode)
			}
		})
	}
}

// TestViewNavigation tests navigation between views
func TestViewNavigation(t *testing.T) {
	app := createTestApp(t)

	tests := []struct {
		name          string
		startView     ViewType
		action        string
		expectedView  ViewType
		expectedStack int
	}{
		{
			name:          "start at clusters view",
			startView:     ViewClusters,
			action:        "none",
			expectedView:  ViewClusters,
			expectedStack: 0,
		},
		{
			name:          "navigate from clusters to projects",
			startView:     ViewClusters,
			action:        "enter",
			expectedView:  ViewProjects,
			expectedStack: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.currentView = ViewContext{viewType: tt.startView}
			app.viewStack = []ViewContext{}

			if tt.action == "enter" {
				// Simulate entering a cluster
				app.clusters = []rancher.Cluster{
					{ID: "c-test", Name: "test-cluster", State: "active"},
				}
				app.updateTable()

				// Simulate entering the first cluster
				if app.currentView.viewType == ViewClusters {
					app.viewStack = append(app.viewStack, app.currentView)
					app.currentView = ViewContext{
						viewType:    ViewProjects,
						clusterID:   "c-test",
						clusterName: "test-cluster",
					}
				}
			}

			if app.currentView.viewType != tt.expectedView {
				t.Errorf("Expected view %v, got %v", tt.expectedView, app.currentView.viewType)
			}

			if len(app.viewStack) != tt.expectedStack {
				t.Errorf("Expected stack size %d, got %d", tt.expectedStack, len(app.viewStack))
			}
		})
	}
}

// TestBreadcrumbGeneration tests breadcrumb string generation
func TestBreadcrumbGeneration(t *testing.T) {
	tests := []struct {
		name         string
		context      ViewContext
		wantContains string
	}{
		{
			name:         "clusters view",
			context:      ViewContext{viewType: ViewClusters},
			wantContains: "[LIVE] r8s - Clusters",
		},
		{
			name: "projects view",
			context: ViewContext{
				viewType:    ViewProjects,
				clusterName: "test-cluster",
			},
			wantContains: "[LIVE] Cluster: test-cluster > Projects",
		},
		{
			name: "namespaces view",
			context: ViewContext{
				viewType:    ViewNamespaces,
				clusterName: "test-cluster",
				projectName: "test-project",
			},
			wantContains: "[LIVE] Cluster: test-cluster > Project: test-project > Namespaces",
		},
		{
			name: "pods view",
			context: ViewContext{
				viewType:      ViewPods,
				clusterName:   "test-cluster",
				projectName:   "test-project",
				namespaceName: "default",
			},
			wantContains: "[LIVE] Cluster: test-cluster > Project: test-project > Namespace: default > Pods",
		},
		{
			name: "deployments view",
			context: ViewContext{
				viewType:      ViewDeployments,
				clusterName:   "test-cluster",
				projectName:   "test-project",
				namespaceName: "default",
			},
			wantContains: "[LIVE] Cluster: test-cluster > Project: test-project > Namespace: default > Deployments",
		},
		{
			name: "services view",
			context: ViewContext{
				viewType:      ViewServices,
				clusterName:   "test-cluster",
				projectName:   "test-project",
				namespaceName: "default",
			},
			wantContains: "[LIVE] Cluster: test-cluster > Project: test-project > Namespace: default > Services",
		},
		{
			name: "CRDs view",
			context: ViewContext{
				viewType:    ViewCRDs,
				clusterName: "test-cluster",
			},
			wantContains: "[LIVE] Cluster: test-cluster > CRDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := createTestApp(t)
			app.currentView = tt.context

			breadcrumb := app.getBreadcrumb()

			if breadcrumb != tt.wantContains {
				t.Errorf("Expected breadcrumb to be %q, got %q", tt.wantContains, breadcrumb)
			}
		})
	}
}

// TestMockDataGeneration tests mock data generation for offline mode
func TestMockDataGeneration(t *testing.T) {
	app := createTestApp(t)
	app.offlineMode = true

	tests := []struct {
		name     string
		generate func() int
		wantMin  int
	}{
		{
			name: "mock clusters",
			generate: func() int {
				clusters := app.getMockClusters()
				return len(clusters)
			},
			wantMin: 2,
		},
		{
			name: "mock pods",
			generate: func() int {
				pods := app.getMockPods("default")
				return len(pods)
			},
			wantMin: 5,
		},
		{
			name: "mock deployments",
			generate: func() int {
				deployments := app.getMockDeployments("default")
				return len(deployments)
			},
			wantMin: 3,
		},
		{
			name: "mock services",
			generate: func() int {
				services := app.getMockServices("default")
				return len(services)
			},
			wantMin: 3,
		},
		{
			name: "mock CRDs",
			generate: func() int {
				crds := app.getMockCRDs()
				return len(crds)
			},
			wantMin: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tt.generate()
			if count < tt.wantMin {
				t.Errorf("Expected at least %d items, got %d", tt.wantMin, count)
			}
		})
	}
}

// TestPodNodeNameExtraction tests pod node name extraction with fallbacks
func TestPodNodeNameExtraction(t *testing.T) {
	app := createTestApp(t)

	tests := []struct {
		name     string
		pod      rancher.Pod
		wantName string
	}{
		{
			name: "NodeName field populated",
			pod: rancher.Pod{
				Name:     "test-pod",
				NodeName: "node-1",
			},
			wantName: "node-1",
		},
		{
			name: "NodeID fallback",
			pod: rancher.Pod{
				Name:   "test-pod",
				NodeID: "node-2",
			},
			wantName: "node-2",
		},
		{
			name: "Node fallback",
			pod: rancher.Pod{
				Name: "test-pod",
				Node: "node-3",
			},
			wantName: "node-3",
		},
		{
			name: "Hostname fallback",
			pod: rancher.Pod{
				Name:     "test-pod",
				Hostname: "node-4",
			},
			wantName: "node-4",
		},
		{
			name: "no node info",
			pod: rancher.Pod{
				Name: "test-pod",
			},
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeName := app.getPodNodeName(tt.pod)
			if nodeName != tt.wantName {
				t.Errorf("Expected node name %q, got %q", tt.wantName, nodeName)
			}
		})
	}
}

// TestIsNamespaceResourceView tests namespace resource view detection
func TestIsNamespaceResourceView(t *testing.T) {
	app := createTestApp(t)

	tests := []struct {
		name     string
		viewType ViewType
		want     bool
	}{
		{"pods view", ViewPods, true},
		{"deployments view", ViewDeployments, true},
		{"services view", ViewServices, true},
		{"clusters view", ViewClusters, false},
		{"projects view", ViewProjects, false},
		{"namespaces view", ViewNamespaces, false},
		{"CRDs view", ViewCRDs, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.currentView.viewType = tt.viewType
			result := app.isNamespaceResourceView()
			if result != tt.want {
				t.Errorf("Expected %v, got %v", tt.want, result)
			}
		})
	}
}

// TestTableUpdate tests table rendering for different view types
func TestTableUpdate(t *testing.T) {
	app := createTestApp(t)

	tests := []struct {
		name     string
		setup    func()
		viewType ViewType
	}{
		{
			name: "clusters table",
			setup: func() {
				app.clusters = []rancher.Cluster{
					{
						Name:     "test-cluster",
						Provider: "k3s",
						State:    "active",
						Created:  time.Now().Add(-time.Hour * 24),
					},
				}
			},
			viewType: ViewClusters,
		},
		{
			name: "pods table",
			setup: func() {
				app.pods = []rancher.Pod{
					{
						Name:        "test-pod",
						NamespaceID: "default",
						State:       "Running",
						NodeName:    "node-1",
					},
				}
			},
			viewType: ViewPods,
		},
		{
			name: "deployments table",
			setup: func() {
				app.deployments = []rancher.Deployment{
					{
						Name:              "test-deployment",
						NamespaceID:       "default",
						State:             "active",
						Replicas:          3,
						ReadyReplicas:     3,
						AvailableReplicas: 3,
					},
				}
			},
			viewType: ViewDeployments,
		},
		{
			name: "services table",
			setup: func() {
				app.services = []rancher.Service{
					{
						Name:        "test-service",
						NamespaceID: "default",
						State:       "active",
						ClusterIP:   "10.43.0.1",
						Kind:        "ClusterIP",
					},
				}
			},
			viewType: ViewServices,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			app.currentView.viewType = tt.viewType
			app.width = 100
			app.height = 30

			// Should not panic
			app.updateTable()

			// Table was successfully updated (no panic means success)
		})
	}
}

// TestMessageTypes tests that message types are properly defined
func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name    string
		message tea.Msg
	}{
		{
			name: "clustersMsg",
			message: clustersMsg{
				clusters: []rancher.Cluster{},
			},
		},
		{
			name: "projectsMsg",
			message: projectsMsg{
				projects:        []rancher.Project{},
				namespaceCounts: map[string]int{},
			},
		},
		{
			name: "namespacesMsg",
			message: namespacesMsg{
				namespaces: []rancher.Namespace{},
			},
		},
		{
			name: "podsMsg",
			message: podsMsg{
				pods: []rancher.Pod{},
			},
		},
		{
			name: "deploymentsMsg",
			message: deploymentsMsg{
				deployments: []rancher.Deployment{},
			},
		},
		{
			name: "servicesMsg",
			message: servicesMsg{
				services: []rancher.Service{},
			},
		},
		{
			name: "errMsg",
			message: errMsg{
				error: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.message == nil {
				t.Error("Message should not be nil")
			}
		})
	}
}

// Helper function to create a test app
func createTestApp(t *testing.T) *App {
	t.Helper()

	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: []config.Profile{
			{
				Name:        "test",
				URL:         "https://test.rancher.com",
				BearerToken: "test-token",
			},
		},
	}

	app := NewApp(cfg, "") // Empty bundle path for live mode tests
	if app == nil {
		t.Fatal("Failed to create test app")
	}

	return app
}
