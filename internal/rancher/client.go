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
	token      string
	httpClient *http.Client
}

// NewClient creates a new Rancher API client
func NewClient(url, token string, insecure bool) *Client {
	// Ensure URL has /v3 prefix
	if !strings.HasSuffix(url, "/v3") {
		url = strings.TrimSuffix(url, "/") + "/v3"
	}
	
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	
	return &Client{
		baseURL: url,
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

// ListNamespaces returns all namespaces
func (c *Client) ListNamespaces(projectID string) (*NamespaceCollection, error) {
	var result NamespaceCollection
	path := "/clusters/" + extractClusterID(projectID) + "/namespaces"
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// extractClusterID extracts cluster ID from project ID (format: c-xxxxx:p-yyyyy)
func extractClusterID(projectID string) string {
	parts := strings.Split(projectID, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return projectID
}
