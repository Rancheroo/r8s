package rancher

import "time"

// Sort represents Rancher API sort options
type Sort struct {
	Order   string            `json:"order,omitempty"`
	Reverse string            `json:"reverse,omitempty"`
	Links   map[string]string `json:"links,omitempty"`
}

// Collection represents a Rancher API collection response
type Collection struct {
	Type         string                   `json:"type"`
	ResourceType string                   `json:"resourceType"`
	Links        map[string]string        `json:"links"`
	CreateTypes  map[string]string        `json:"createTypes"`
	Actions      map[string]string        `json:"actions"`
	Pagination   *Pagination              `json:"pagination,omitempty"`
	Sort         *Sort                    `json:"sort,omitempty"`
	Filters      map[string][]interface{} `json:"filters,omitempty"`
}

// Pagination represents pagination information
type Pagination struct {
	Limit   int    `json:"limit,omitempty"`
	Total   int    `json:"total,omitempty"`
	First   string `json:"first,omitempty"`
	Previous string `json:"previous,omitempty"`
	Next    string `json:"next,omitempty"`
	Last    string `json:"last,omitempty"`
}

// ClusterCollection represents a collection of clusters
type ClusterCollection struct {
	Collection
	Data []Cluster `json:"data"`
}

// ClusterVersion represents Kubernetes version info
type ClusterVersion struct {
	GitVersion string `json:"gitVersion,omitempty"`
	Major      string `json:"major,omitempty"`
	Minor      string `json:"minor,omitempty"`
}

// Cluster represents a Rancher cluster
type Cluster struct {
	ID                   string            `json:"id"`
	Type                 string            `json:"type"`
	Name                 string            `json:"name"`
	State                string            `json:"state"`
	Transitioning        string            `json:"transitioning"`
	TransitioningMessage string            `json:"transitioningMessage"`
	Version              *ClusterVersion   `json:"version,omitempty"`
	Provider             string            `json:"provider"`
	Created              time.Time         `json:"created"`
	Labels               map[string]string `json:"labels,omitempty"`
	Annotations          map[string]string `json:"annotations,omitempty"`
	Links                map[string]string `json:"links"`
	Actions              map[string]string `json:"actions"`
}

// ProjectCollection represents a collection of projects
type ProjectCollection struct {
	Collection
	Data []Project `json:"data"`
}

// Project represents a Rancher project
type Project struct {
	ID                  string            `json:"id"`
	Type                string            `json:"type"`
	Name                string            `json:"name"`
	ClusterID           string            `json:"clusterId"`
	DisplayName         string            `json:"displayName"`
	Description         string            `json:"description,omitempty"`
	State               string            `json:"state"`
	Created             time.Time         `json:"created"`
	Labels              map[string]string `json:"labels,omitempty"`
	Annotations         map[string]string `json:"annotations,omitempty"`
	Links               map[string]string `json:"links"`
	Actions             map[string]string `json:"actions"`
}

// NamespaceCollection represents a collection of namespaces
type NamespaceCollection struct {
	Collection
	Data []Namespace `json:"data"`
}

// Namespace represents a Kubernetes namespace in Rancher
type Namespace struct {
	ID                  string            `json:"id"`
	Type                string            `json:"type"`
	Name                string            `json:"name"`
	ClusterID           string            `json:"clusterId"`
	ProjectID           string            `json:"projectId"`
	State               string            `json:"state"`
	Created             time.Time         `json:"created"`
	Labels              map[string]string `json:"labels,omitempty"`
	Annotations         map[string]string `json:"annotations,omitempty"`
	Links               map[string]string `json:"links"`
	Actions             map[string]string `json:"actions"`
}

// PodCollection represents a collection of pods
type PodCollection struct {
	Collection
	Data []Pod `json:"data"`
}

// Pod represents a Kubernetes pod
type Pod struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	NamespaceID string            `json:"namespaceId"`
	NodeName    string            `json:"nodeName"`
	State       string            `json:"state"`
	PodIP       string            `json:"podIP"`
	RestartCount int              `json:"restartCount"`
	Created     time.Time         `json:"created"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Links       map[string]string `json:"links"`
	Actions     map[string]string `json:"actions"`
}
