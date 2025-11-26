package rancher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		token    string
		insecure bool
		wantBase string
		wantRoot string
	}{
		{
			name:     "basic URL without /v3",
			url:      "https://rancher.example.com",
			token:    "token-123",
			insecure: false,
			wantBase: "https://rancher.example.com/v3",
			wantRoot: "https://rancher.example.com",
		},
		{
			name:     "URL with /v3 suffix",
			url:      "https://rancher.example.com/v3",
			token:    "token-123",
			insecure: false,
			wantBase: "https://rancher.example.com/v3",
			wantRoot: "https://rancher.example.com",
		},
		{
			name:     "URL with trailing slash",
			url:      "https://rancher.example.com/",
			token:    "token-123",
			insecure: false,
			wantBase: "https://rancher.example.com/v3",
			wantRoot: "https://rancher.example.com",
		},
		{
			name:     "URL with /v3 and trailing slash",
			url:      "https://rancher.example.com/v3/",
			token:    "token-123",
			insecure: false,
			wantBase: "https://rancher.example.com/v3",
			wantRoot: "https://rancher.example.com",
		},
		{
			name:     "insecure mode",
			url:      "https://rancher.example.com",
			token:    "token-123",
			insecure: true,
			wantBase: "https://rancher.example.com/v3",
			wantRoot: "https://rancher.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.url, tt.token, tt.insecure)

			if client.baseURL != tt.wantBase {
				t.Errorf("NewClient() baseURL = %v, want %v", client.baseURL, tt.wantBase)
			}
			if client.rootURL != tt.wantRoot {
				t.Errorf("NewClient() rootURL = %v, want %v", client.rootURL, tt.wantRoot)
			}
			if client.token != tt.token {
				t.Errorf("NewClient() token = %v, want %v", client.token, tt.token)
			}
			if client.httpClient == nil {
				t.Error("NewClient() httpClient is nil")
			}
		})
	}
}

func TestClient_TestConnection(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "successful connection",
			statusCode: http.StatusOK,
			response:   `{"data": []}`,
			wantErr:    false,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   `{"error": "unauthorized"}`,
			wantErr:    true,
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			response:   `{"error": "forbidden"}`,
			wantErr:    true,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   `{"error": "internal server error"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify auth header
				if auth := r.Header.Get("Authorization"); auth == "" {
					t.Error("Authorization header not set")
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false)
			err := client.TestConnection()

			if (err != nil) != tt.wantErr {
				t.Errorf("TestConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ListClusters(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		response     string
		wantErr      bool
		wantCount    int
		verifyHeader bool
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			response: `{
				"type": "collection",
				"data": [
					{"id": "c-1", "name": "cluster1", "state": "active"},
					{"id": "c-2", "name": "cluster2", "state": "active"}
				]
			}`,
			wantErr:      false,
			wantCount:    2,
			verifyHeader: true,
		},
		{
			name:       "empty response",
			statusCode: http.StatusOK,
			response:   `{"type": "collection", "data": []}`,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   `{"error": "unauthorized"}`,
			wantErr:    true,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   `{"error": "server error"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path
				if r.URL.Path != "/v3/clusters" {
					t.Errorf("Request path = %v, want /v3/clusters", r.URL.Path)
				}

				// Verify headers
				if tt.verifyHeader {
					if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
						t.Errorf("Authorization header = %v, want 'Bearer test-token'", auth)
					}
					if ct := r.Header.Get("Content-Type"); ct != "application/json" {
						t.Errorf("Content-Type = %v, want 'application/json'", ct)
					}
					if accept := r.Header.Get("Accept"); accept != "application/json" {
						t.Errorf("Accept = %v, want 'application/json'", accept)
					}
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false)
			clusters, err := client.ListClusters()

			if (err != nil) != tt.wantErr {
				t.Errorf("ListClusters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(clusters.Data) != tt.wantCount {
				t.Errorf("ListClusters() count = %v, want %v", len(clusters.Data), tt.wantCount)
			}
		})
	}
}

func TestClient_ListProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameter
		if clusterID := r.URL.Query().Get("clusterId"); clusterID == "" {
			t.Error("clusterId query parameter not set")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"type": "collection",
			"data": [
				{"id": "p-1", "name": "project1"}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	projects, err := client.ListProjects("c-test")

	if err != nil {
		t.Errorf("ListProjects() error = %v", err)
	}
	if len(projects.Data) != 1 {
		t.Errorf("ListProjects() count = %v, want 1", len(projects.Data))
	}
}

func TestClient_GetPodDetails(t *testing.T) {
	tests := []struct {
		name       string
		clusterID  string
		namespace  string
		podName    string
		statusCode int
		response   string
		wantErr    bool
		wantPath   string
	}{
		{
			name:       "successful pod details",
			clusterID:  "c-test",
			namespace:  "default",
			podName:    "nginx-pod",
			statusCode: http.StatusOK,
			response: `{
				"name": "nginx-pod",
				"namespaceId": "default",
				"state": "running",
				"nodeName": "worker-1"
			}`,
			wantErr:  false,
			wantPath: "/k8s/clusters/c-test/api/v1/namespaces/default/pods/nginx-pod",
		},
		{
			name:       "pod not found",
			clusterID:  "c-test",
			namespace:  "default",
			podName:    "missing-pod",
			statusCode: http.StatusNotFound,
			response:   `{"error": "not found"}`,
			wantErr:    true,
			wantPath:   "/k8s/clusters/c-test/api/v1/namespaces/default/pods/missing-pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify path (without server URL prefix)
				if r.URL.Path != tt.wantPath {
					t.Errorf("Request path = %v, want %v", r.URL.Path, tt.wantPath)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false)
			pod, err := client.GetPodDetails(tt.clusterID, tt.namespace, tt.podName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && pod.Name != tt.podName {
				t.Errorf("GetPodDetails() pod name = %v, want %v", pod.Name, tt.podName)
			}
		})
	}
}

func TestClient_GetDeploymentDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/k8s/clusters/c-test/apis/apps/v1/namespaces/default/deployments/nginx"
		if r.URL.Path != expectedPath {
			t.Errorf("Request path = %v, want %v", r.URL.Path, expectedPath)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"name": "nginx",
			"namespaceId": "default",
			"state": "active",
			"replicas": 3
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	deployment, err := client.GetDeploymentDetails("c-test", "default", "nginx")

	if err != nil {
		t.Errorf("GetDeploymentDetails() error = %v", err)
	}
	if deployment.Name != "nginx" {
		t.Errorf("GetDeploymentDetails() name = %v, want 'nginx'", deployment.Name)
	}
}

func TestClient_GetServiceDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/k8s/clusters/c-test/api/v1/namespaces/default/services/nginx-svc"
		if r.URL.Path != expectedPath {
			t.Errorf("Request path = %v, want %v", r.URL.Path, expectedPath)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"name": "nginx-svc",
			"namespaceId": "default",
			"kind": "ClusterIP",
			"clusterIp": "10.43.0.1"
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	service, err := client.GetServiceDetails("c-test", "default", "nginx-svc")

	if err != nil {
		t.Errorf("GetServiceDetails() error = %v", err)
	}
	if service.Name != "nginx-svc" {
		t.Errorf("GetServiceDetails() name = %v, want 'nginx-svc'", service.Name)
	}
}

func TestClient_ListCRDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/k8s/clusters/c-test/apis/apiextensions.k8s.io/v1/customresourcedefinitions"
		if r.URL.Path != expectedPath {
			t.Errorf("Request path = %v, want %v", r.URL.Path, expectedPath)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind": "CustomResourceDefinitionList",
			"items": [
				{
					"metadata": {"name": "certificates.cert-manager.io"},
					"spec": {
						"group": "cert-manager.io",
						"names": {"kind": "Certificate", "plural": "certificates"}
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	crds, err := client.ListCRDs("c-test")

	if err != nil {
		t.Errorf("ListCRDs() error = %v", err)
	}
	if len(crds.Items) != 1 {
		t.Errorf("ListCRDs() count = %v, want 1", len(crds.Items))
	}
	if crds.Items[0].Spec.Group != "cert-manager.io" {
		t.Errorf("ListCRDs() group = %v, want 'cert-manager.io'", crds.Items[0].Spec.Group)
	}
}

func TestClient_ListCustomResources(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantPath  string
	}{
		{
			name:      "cluster-scoped resources",
			namespace: "",
			wantPath:  "/k8s/clusters/c-test/apis/cert-manager.io/v1/certificates",
		},
		{
			name:      "namespace-scoped resources",
			namespace: "default",
			wantPath:  "/k8s/clusters/c-test/apis/cert-manager.io/v1/namespaces/default/certificates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantPath {
					t.Errorf("Request path = %v, want %v", r.URL.Path, tt.wantPath)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"apiVersion": "cert-manager.io/v1",
					"kind": "CertificateList",
					"items": [
						{"metadata": {"name": "test-cert"}}
					]
				}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false)
			resources, err := client.ListCustomResources("c-test", "cert-manager.io", "v1", "certificates", tt.namespace)

			if err != nil {
				t.Errorf("ListCustomResources() error = %v", err)
			}
			if len(resources.Items) != 1 {
				t.Errorf("ListCustomResources() count = %v, want 1", len(resources.Items))
			}
		})
	}
}

// TestClient_ConcurrentRequests verifies the client is safe for concurrent use
func TestClient_ConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"type": "collection", "data": []}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)

	// Run multiple concurrent requests
	const numRequests = 10
	done := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			_, err := client.ListClusters()
			done <- err
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}
