package rancher

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCluster_UnmarshalJSON(t *testing.T) {
	data := `{
		"id": "c-12345",
		"type": "cluster",
		"name": "test-cluster",
		"state": "active",
		"transitioning": "no",
		"transitioningMessage": "",
		"provider": "rke2",
		"created": "2025-01-15T10:30:00Z",
		"labels": {"env": "test"},
		"annotations": {"description": "test cluster"},
		"version": {"gitVersion": "v1.28.5"}
	}`

	var cluster Cluster
	if err := json.Unmarshal([]byte(data), &cluster); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if cluster.ID != "c-12345" {
		t.Errorf("Expected ID 'c-12345', got: %s", cluster.ID)
	}

	if cluster.Name != "test-cluster" {
		t.Errorf("Expected name 'test-cluster', got: %s", cluster.Name)
	}

	if cluster.State != "active" {
		t.Errorf("Expected state 'active', got: %s", cluster.State)
	}

	if cluster.Provider != "rke2" {
		t.Errorf("Expected provider 'rke2', got: %s", cluster.Provider)
	}

	if cluster.Version == nil || cluster.Version.GitVersion != "v1.28.5" {
		t.Errorf("Expected version v1.28.5")
	}

	if cluster.Labels["env"] != "test" {
		t.Errorf("Expected label env=test")
	}
}

func TestProject_UnmarshalJSON(t *testing.T) {
	data := `{
		"id": "p-abcde",
		"type": "project",
		"name": "test-project",
		"clusterId": "c-12345",
		"displayName": "Test Project",
		"description": "A test project",
		"state": "active",
		"created": "2025-01-15T10:30:00Z",
		"labels": {"team": "dev"}
	}`

	var project Project
	if err := json.Unmarshal([]byte(data), &project); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if project.ID != "p-abcde" {
		t.Errorf("Expected ID 'p-abcde', got: %s", project.ID)
	}

	if project.Name != "test-project" {
		t.Errorf("Expected name 'test-project', got: %s", project.Name)
	}

	if project.ClusterID != "c-12345" {
		t.Errorf("Expected cluster ID 'c-12345', got: %s", project.ClusterID)
	}

	if project.DisplayName != "Test Project" {
		t.Errorf("Expected display name 'Test Project', got: %s", project.DisplayName)
	}

	if project.Labels["team"] != "dev" {
		t.Errorf("Expected label team=dev")
	}
}

func TestCRD_UnmarshalJSON(t *testing.T) {
	data := `{
		"metadata": {
			"name": "addons.k3s.cattle.io"
		},
		"spec": {
			"group": "k3s.cattle.io",
			"versions": [{"name": "v1"}]
		}
	}`

	var crd CRD
	if err := json.Unmarshal([]byte(data), &crd); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if crd.Metadata.Name != "addons.k3s.cattle.io" {
		t.Errorf("Expected name, got: %s", crd.Metadata.Name)
	}

	if crd.Spec.Group != "k3s.cattle.io" {
		t.Errorf("Expected group k3s.cattle.io, got: %s", crd.Spec.Group)
	}
}

func TestCollection_Pagination(t *testing.T) {
	data := `{
		"type": "collection",
		"resourceType": "cluster",
		"data": [],
		"pagination": {
			"limit": 100,
			"total": 50,
			"first": "/v3/clusters?limit=100",
			"next": null,
			"last": "/v3/clusters?marker=abc"
		}
	}`

	var collection Collection
	if err := json.Unmarshal([]byte(data), &collection); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if collection.Pagination == nil {
		t.Fatal("Expected pagination to be parsed")
	}

	if collection.Pagination.Limit != 100 {
		t.Errorf("Expected limit 100, got: %d", collection.Pagination.Limit)
	}

	if collection.Pagination.Total != 50 {
		t.Errorf("Expected total 50, got: %d", collection.Pagination.Total)
	}

	if collection.Pagination.Next != "" {
		t.Error("Expected next to be empty")
	}
}

func TestTimeParsing(t *testing.T) {
	data := `{
		"id": "test",
		"created": "2025-01-15T10:30:00Z"
	}`

	var cluster Cluster
	if err := json.Unmarshal([]byte(data), &cluster); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	expected := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	if !cluster.Created.Equal(expected) {
		t.Errorf("Expected time %v, got: %v", expected, cluster.Created)
	}
}
