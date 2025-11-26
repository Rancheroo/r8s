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
	Limit    int    `json:"limit,omitempty"`
	Total    int    `json:"total,omitempty"`
	First    string `json:"first,omitempty"`
	Previous string `json:"previous,omitempty"`
	Next     string `json:"next,omitempty"`
	Last     string `json:"last,omitempty"`
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
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	ClusterID   string            `json:"clusterId"`
	DisplayName string            `json:"displayName"`
	Description string            `json:"description,omitempty"`
	State       string            `json:"state"`
	Created     time.Time         `json:"created"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Links       map[string]string `json:"links"`
	Actions     map[string]string `json:"actions"`
}

// --- Kubernetes API Types (via Proxy) ---

// CRDList represents a list of CustomResourceDefinitions (K8s API)
type CRDList struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Items      []CRD  `json:"items"`
}

// CRD represents a CustomResourceDefinition
type CRD struct {
	Metadata ObjectMeta `json:"metadata"`
	Spec     CRDSpec    `json:"spec"`
}

// ObjectMeta represents standard K8s metadata
type ObjectMeta struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	ResourceVersion   string            `json:"resourceVersion,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
}

// CRDSpec represents the spec of a CRD
type CRDSpec struct {
	Group    string       `json:"group"`
	Names    CRDNames     `json:"names"`
	Scope    string       `json:"scope"`
	Versions []CRDVersion `json:"versions"`
}

// CRDNames represents the names of a CRD
type CRDNames struct {
	Plural     string   `json:"plural"`
	Singular   string   `json:"singular"`
	ShortNames []string `json:"shortNames,omitempty"`
	Kind       string   `json:"kind"`
	ListKind   string   `json:"listKind"`
	Categories []string `json:"categories,omitempty"`
}

// CRDVersion represents a version of a CRD
type CRDVersion struct {
	Name    string     `json:"name"`
	Served  bool       `json:"served"`
	Storage bool       `json:"storage"`
	Schema  *CRDSchema `json:"schema,omitempty"`
}

// CRDSchema represents the schema of a CRD version
type CRDSchema struct {
	OpenAPIV3Schema *OpenAPIV3Schema `json:"openAPIV3Schema,omitempty"`
}

// OpenAPIV3Schema represents the OpenAPI v3 schema
type OpenAPIV3Schema struct {
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// UnstructuredList represents a generic K8s list response
type UnstructuredList struct {
	ApiVersion string                   `json:"apiVersion"`
	Kind       string                   `json:"kind"`
	Metadata   map[string]interface{}   `json:"metadata"`
	Items      []map[string]interface{} `json:"items"`
}

// NamespaceCollection represents a collection of namespaces
type NamespaceCollection struct {
	Collection
	Data []Namespace `json:"data"`
}

// Namespace represents a Kubernetes namespace in Rancher
type Namespace struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	ClusterID   string            `json:"clusterId"`
	ProjectID   string            `json:"projectId"`
	State       string            `json:"state"`
	Created     time.Time         `json:"created"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Links       map[string]string `json:"links"`
	Actions     map[string]string `json:"actions"`
}

// PodCollection represents a collection of pods
type PodCollection struct {
	Collection
	Data []Pod `json:"data"`
}

// Pod represents a Kubernetes pod
type Pod struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	NamespaceID  string            `json:"namespaceId"`
	NodeName     string            `json:"nodeName"` // Try this first
	NodeID       string            `json:"nodeId"`   // Fallback 1
	Node         string            `json:"node"`     // Fallback 2
	Hostname     string            `json:"hostname"` // Fallback 3
	State        string            `json:"state"`
	PodIP        string            `json:"podIP"`
	RestartCount int               `json:"restartCount"`
	Created      time.Time         `json:"created"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Links        map[string]string `json:"links"`
	Actions      map[string]string `json:"actions"`
}

// DeploymentCollection represents a collection of deployments
type DeploymentCollection struct {
	Collection
	Data []Deployment `json:"data"`
}

// Deployment represents a Kubernetes deployment
type Deployment struct {
	ID                string            `json:"id"`
	Type              string            `json:"type"`
	Name              string            `json:"name"`
	NamespaceID       string            `json:"namespaceId"`
	State             string            `json:"state"`
	Replicas          int               `json:"replicas"`
	AvailableReplicas int               `json:"availableReplicas"`
	ReadyReplicas     int               `json:"readyReplicas"`
	UpToDateReplicas  int               `json:"updatedReplicas"`
	Created           time.Time         `json:"created"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Links             map[string]string `json:"links"`
	Actions           map[string]string `json:"actions"`
}

// ServiceCollection represents a collection of services
type ServiceCollection struct {
	Collection
	Data []Service `json:"data"`
}

// ServicePort represents a service port
type ServicePort struct {
	Name       string      `json:"name"`
	Protocol   string      `json:"protocol"`
	Port       int         `json:"port"`
	TargetPort interface{} `json:"targetPort"` // Can be int or string
	NodePort   int         `json:"nodePort,omitempty"`
}

// Service represents a Kubernetes service
type Service struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	NamespaceID string            `json:"namespaceId"`
	State       string            `json:"state"`
	ClusterIP   string            `json:"clusterIp"`
	Kind        string            `json:"kind"` // Service type (ClusterIP, NodePort, etc.)
	Ports       []ServicePort     `json:"ports,omitempty"`
	Created     time.Time         `json:"created"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Links       map[string]string `json:"links"`
	Actions     map[string]string `json:"actions"`
}
