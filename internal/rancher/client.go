package rancher

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents a Rancher API client
type Client struct {
	baseURL    string
	rootURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Rancher API client
func NewClient(url, token string, insecure bool) *Client {
	// Normalize URL to root (remove /v3 suffix if present)
	rootURL := strings.TrimSuffix(url, "/")
	if strings.HasSuffix(rootURL, "/v3") {
		rootURL = strings.TrimSuffix(rootURL, "/v3")
	}

	// Base URL for V3 API
	baseURL := rootURL + "/v3"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	return &Client{
		baseURL: baseURL,
		rootURL: rootURL,
		token:   token,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, fmt.Errorf("authentication failed: invalid token or credentials")
	}

	if resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		return nil, fmt.Errorf("access forbidden: insufficient permissions")
	}

	return resp, nil
}

// get performs a GET request
func (c *Client) get(path string, result interface{}) error {
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// TestConnection tests the connection to Rancher
func (c *Client) TestConnection() error {
	var result map[string]interface{}
	if err := c.get("/clusters", &result); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

// ListClusters returns all clusters
func (c *Client) ListClusters() (*ClusterCollection, error) {
	var result ClusterCollection
	if err := c.get("/clusters", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCluster returns a specific cluster by ID
func (c *Client) GetCluster(id string) (*Cluster, error) {
	var result Cluster
	if err := c.get("/clusters/"+id, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListProjects returns all projects for a cluster
func (c *Client) ListProjects(clusterID string) (*ProjectCollection, error) {
	var result ProjectCollection
	path := "/projects"
	if clusterID != "" {
		path += "?clusterId=" + clusterID
	}
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListNamespaces returns all namespaces for a cluster
func (c *Client) ListNamespaces(clusterID string) (*NamespaceCollection, error) {
	var result NamespaceCollection
	path := "/clusters/" + clusterID + "/namespaces"
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListPods returns all pods for a project
func (c *Client) ListPods(projectID string) (*PodCollection, error) {
	var result PodCollection
	// Rancher uses project-scoped pod endpoints
	path := "/projects/" + projectID + "/pods"
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListDeployments returns all deployments for a project
func (c *Client) ListDeployments(projectID string) (*DeploymentCollection, error) {
	var result DeploymentCollection
	path := "/projects/" + projectID + "/deployments"
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListServices returns all services for a project
func (c *Client) ListServices(projectID string) (*ServiceCollection, error) {
	var result ServiceCollection
	path := "/projects/" + projectID + "/services"
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListCRDs returns all CustomResourceDefinitions in the cluster (via K8s proxy)
func (c *Client) ListCRDs(clusterID string) (*CRDList, error) {
	var result CRDList
	// Using K8s proxy endpoint as Rancher v3 API doesn't expose CRDs directly in a simple way
	path := "/k8s/clusters/" + clusterID + "/apis/apiextensions.k8s.io/v1/customresourcedefinitions"

	// Use getViaRoot to bypass /v3 prefix
	if err := c.getViaRoot(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListCustomResources returns all instances of a CRD (via K8s proxy)
// If namespace is empty, lists all resources (for Cluster scope) or across all namespaces (for Namespaced scope)
func (c *Client) ListCustomResources(clusterID, group, version, plural, namespace string) (*UnstructuredList, error) {
	var result UnstructuredList
	var path string

	// Construct K8s API path
	basePath := "/k8s/clusters/" + clusterID + "/apis/" + group + "/" + version

	if namespace != "" {
		path = basePath + "/namespaces/" + namespace + "/" + plural
	} else {
		path = basePath + "/" + plural
	}

	// Use getViaRoot to bypass /v3 prefix
	if err := c.getViaRoot(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// getViaRoot performs a GET request using the root URL (without /v3 prefix)
func (c *Client) getViaRoot(path string, result interface{}) error {
	// Temporarily override baseURL logic for this request
	url := c.rootURL + path

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// extractClusterID extracts cluster ID from project ID (format: c-xxxxx:p-yyyyy)
func extractClusterID(projectID string) string {
	parts := strings.Split(projectID, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return projectID
}

// GetPodDetails returns detailed information about a specific pod (via K8s proxy)
func (c *Client) GetPodDetails(clusterID, namespace, name string) (*Pod, error) {
	var result Pod
	path := "/k8s/clusters/" + clusterID + "/api/v1/namespaces/" + namespace + "/pods/" + name

	if err := c.getViaRoot(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDeploymentDetails returns detailed information about a specific deployment (via K8s proxy)
func (c *Client) GetDeploymentDetails(clusterID, namespace, name string) (*Deployment, error) {
	var result Deployment
	path := "/k8s/clusters/" + clusterID + "/apis/apps/v1/namespaces/" + namespace + "/deployments/" + name

	if err := c.getViaRoot(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetServiceDetails returns detailed information about a specific service (via K8s proxy)
func (c *Client) GetServiceDetails(clusterID, namespace, name string) (*Service, error) {
	var result Service
	path := "/k8s/clusters/" + clusterID + "/api/v1/namespaces/" + namespace + "/services/" + name

	if err := c.getViaRoot(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
