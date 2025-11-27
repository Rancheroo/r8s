# Phase 5B: kubectl Resource Parsing Implementation Plan

**Date:** November 27, 2025  
**Goal:** Parse kubectl output from bundles to enable full resource browsing

## kubectl Output Format Analysis

### Format Pattern
All `rke2/kubectl/*` files follow kubectl `get` output format:
- **Header row** with column names (whitespace-separated)
- **Data rows** with values (whitespace-separated, may contain spaces in some fields)
- **Consistent structure** but varying column counts per resource type

### Observed Formats

#### 1. CRDs (Simple)
```
NAME                                                              CREATED AT
addons.k3s.cattle.io                                              2025-11-20T00:45:08Z
alertmanagers.monitoring.coreos.com                               2025-11-26T01:26:48Z
```
- 2 columns: NAME, CREATED AT
- Fixed-width alignment

#### 2. Deployments (Complex)
```
NAMESPACE      NAME                  READY   UP-TO-DATE   AVAILABLE   AGE     CONTAINERS   IMAGES                    SELECTOR
calico-system  calico-kube-ctrl      1/1     1            1           7d3h    ctrl         docker.io/rancher/ctrl    k8s-app=ctrl
```
- 9 columns with namespace prefix
- READY format: "1/1"
- AGE format: "7d3h", "26h", "5d18h"
- Multi-value fields (CONTAINERS, IMAGES, SELECTOR)

#### 3. Services (Multi-column)
```
NAMESPACE      NAME           TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)      AGE    SELECTOR
calico-system  calico-typha   ClusterIP   10.43.29.34    <none>        5473/TCP     7d3h   k8s-app=calico-typha
```
- 8 columns
- PORT(S) can have multiple ports: "9093/TCP,9094/TCP,9094/UDP"
- EXTERNAL-IP can be "<none>"

#### 4. Namespaces (Simple)
```
NAME                          STATUS   AGE
calico-system                 Active   7d3h
cattle-monitoring-system      Active   26h
```
- 3 columns
- Simple structure

---

## Parsing Strategy

### Approach 1: Field-Based Parsing (CHOSEN)
Use Go's `strings.Fields()` with column count validation:

```go
func parseKubectlOutput(content []byte, expectedColumns int) ([]map[string]string, error) {
    lines := strings.Split(string(content), "\n")
    if len(lines) < 2 {
        return nil, fmt.Errorf("no data")
    }
    
    // Parse header
    header := strings.Fields(lines[0])
    if len(header) != expectedColumns {
        return nil, fmt.Errorf("unexpected column count")
    }
    
    // Parse data rows
    var results []map[string]string
    for _, line := range lines[1:] {
        if strings.TrimSpace(line) == "" {
            continue
        }
        
        fields := strings.Fields(line)
        if len(fields) < expectedColumns {
            continue // Skip incomplete rows
        }
        
        row := make(map[string]string)
        for i, colName := range header {
            if i < len(fields) {
                row[colName] = fields[i]
            }
        }
        results = append(results, row)
    }
    
    return results, nil
}
```

**Pros:**
- Simple and fast
- Works for most kubectl output
- Easy to debug

**Cons:**
- May have issues with fields containing spaces
- Need special handling for complex fields

### Approach 2: Regex-Based (Alternative)
Use regex patterns for each resource type.

**Decision:** Start with Approach 1, add regex if needed.

---

## Implementation Plan

### Step 1: Add Parsing Functions to bundle/kubectl.go (NEW FILE)

Create `internal/bundle/kubectl.go`:

```go
package bundle

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/Rancheroo/r8s/internal/rancher"
)

// ParseCRDs parses kubectl get crds output
func ParseCRDs(extractPath string) ([]rancher.CRD, error) {
    path := filepath.Join(extractPath, "rke2/kubectl/crds")
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    lines := strings.Split(string(content), "\n")
    var crds []rancher.CRD
    
    for i, line := range lines {
        if i == 0 || strings.TrimSpace(line) == "" {
            continue // Skip header and empty
        }
        
        fields := strings.Fields(line)
        if len(fields) < 2 {
            continue
        }
        
        name := fields[0]
        createdAt := fields[1]
        
        // Parse CRD name into group/kind
        parts := strings.Split(name, ".")
        if len(parts) < 2 {
            continue
        }
        
        kind := parts[0]
        group := strings.Join(parts[1:], ".")
        
        // Parse timestamp
        created, _ := time.Parse(time.RFC3339, createdAt)
        
        crds = append(crds, rancher.CRD{
            Metadata: rancher.ObjectMeta{
                Name:              name,
                CreationTimestamp: created,
            },
            Spec: rancher.CRDSpec{
                Group: group,
                Names: rancher.CRDNames{
                    Kind:     strings.Title(kind),
                    Plural:   kind,
                    Singular: kind,
                },
                Scope: "Cluster", // Default, may need refinement
                Versions: []rancher.CRDVersion{
                    {Name: "v1", Served: true, Storage: true},
                },
            },
        })
    }
    
    return crds, nil
}

// ParseDeployments parses kubectl get deployments output
func ParseDeployments(extractPath string) ([]rancher.Deployment, error) {
    path := filepath.Join(extractPath, "rke2/kubectl/deployments")
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    lines := strings.Split(string(content), "\n")
    var deployments []rancher.Deployment
    
    for i, line := range lines {
        if i == 0 || strings.TrimSpace(line) == "" {
            continue
        }
        
        fields := strings.Fields(line)
        if len(fields) < 6 {
            continue // Need at least namespace, name, ready, uptodate, available, age
        }
        
        namespace := fields[0]
        name := fields[1]
        ready := fields[2] // Format: "1/1"
        // uptodate := fields[3]
        // available := fields[4]
        // age := fields[5]
        
        // Parse ready field "1/1"
        readyParts := strings.Split(ready, "/")
        var readyReplicas, totalReplicas int
        if len(readyParts) == 2 {
            fmt.Sscanf(readyParts[0], "%d", &readyReplicas)
            fmt.Sscanf(readyParts[1], "%d", &totalReplicas)
        }
        
        deployments = append(deployments, rancher.Deployment{
            Name:              name,
            NamespaceID:       namespace,
            State:             "active",
            Replicas:          totalReplicas,
            ReadyReplicas:     readyReplicas,
            AvailableReplicas: readyReplicas,
            UpToDateReplicas:  readyReplicas,
        })
    }
    
    return deployments, nil
}

// ParseServices parses kubectl get services output
func ParseServices(extractPath string) ([]rancher.Service, error) {
    path := filepath.Join(extractPath, "rke2/kubectl/services")
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    lines := strings.Split(string(content), "\n")
    var services []rancher.Service
    
    for i, line := range lines {
        if i == 0 || strings.TrimSpace(line) == "" {
            continue
        }
        
        fields := strings.Fields(line)
        if len(fields) < 6 {
            continue
        }
        
        namespace := fields[0]
        name := fields[1]
        serviceType := fields[2]
        clusterIP := fields[3]
        // externalIP := fields[4]
        portsStr := fields[5]
        
        // Parse ports: "5473/TCP" or "9093/TCP,9094/TCP,9094/UDP"
        var ports []rancher.ServicePort
        for _, portStr := range strings.Split(portsStr, ",") {
            parts := strings.Split(portStr, "/")
            if len(parts) == 2 {
                var port int
                fmt.Sscanf(parts[0], "%d", &port)
                protocol := parts[1]
                
                ports = append(ports, rancher.ServicePort{
                    Protocol:   protocol,
                    Port:       port,
                    TargetPort: port,
                })
            }
        }
        
        services = append(services, rancher.Service{
            Name:        name,
            NamespaceID: namespace,
            State:       "active",
            ClusterIP:   clusterIP,
            Kind:        serviceType,
            Ports:       ports,
        })
    }
    
    return services, nil
}

// ParseNamespaces parses kubectl get namespaces output
func ParseNamespaces(extractPath string) ([]rancher.Namespace, error) {
    path := filepath.Join(extractPath, "rke2/kubectl/namespaces")
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    lines := strings.Split(string(content), "\n")
    var namespaces []rancher.Namespace
    
    for i, line := range lines {
        if i == 0 || strings.TrimSpace(line) == "" {
            continue
        }
        
        fields := strings.Fields(line)
        if len(fields) < 2 {
            continue
        }
        
        name := fields[0]
        status := fields[1]
        
        namespaces = append(namespaces, rancher.Namespace{
            Name:      name,
            State:     strings.ToLower(status),
            ClusterID: "bundle",
            ProjectID: "bundle-project",
        })
    }
    
    return namespaces, nil
}
```

### Step 2: Update Bundle Structure

Add fields to `Bundle` struct in `internal/bundle/types.go`:

```go
type Bundle struct {
    Path        string
    ExtractPath string
    Manifest    ManifestInfo
    Pods        []PodInfo
    LogFiles    []LogFileInfo
    
    // NEW: kubectl resources
    CRDs        []rancher.CRD
    Deployments []rancher.Deployment
    Services    []rancher.Service
    Namespaces  []rancher.Namespace
    
    Loaded      bool
    Size        int64
}
```

### Step 3: Update bundle.Load()

In `internal/bundle/bundle.go`:

```go
func Load(opts ImportOptions) (*Bundle, error) {
    // ... existing code ...
    
    // Inventory pods
    pods, err := InventoryPods(extractPath)
    // ... existing error handling ...
    
    // Inventory log files
    logFiles, err := InventoryLogFiles(extractPath)
    // ... existing error handling ...
    
    // NEW: Parse kubectl resources
    crds, _ := ParseCRDs(extractPath)           // Ignore errors for optional data
    deployments, _ := ParseDeployments(extractPath)
    services, _ := ParseServices(extractPath)
    namespaces, _ := ParseNamespaces(extractPath)
    
    bundle := &Bundle{
        Path:        opts.Path,
        ExtractPath: extractPath,
        Manifest:    manifest,
        Pods:        pods,
        LogFiles:    logFiles,
        CRDs:        crds,          // NEW
        Deployments: deployments,    // NEW
        Services:    services,       // NEW
        Namespaces:  namespaces,     // NEW
        Loaded:      true,
        Size:        bundleSize,
    }
    
    return bundle, nil
}
```

### Step 4: Extend DataSource Interface

In `internal/tui/datasource.go`:

```go
type DataSource interface {
    // Existing
    GetPods(projectID, namespace string) ([]rancher.Pod, error)
    GetLogs(clusterID, namespace, pod, container string, previous bool) ([]string, error)
    GetContainers(namespace, pod string) ([]string, error)
    IsOffline() bool
    GetMode() string
    
    // NEW
    GetCRDs(clusterID string) ([]rancher.CRD, error)
    GetDeployments(projectID, namespace string) ([]rancher.Deployment, error)
    GetServices(projectID, namespace string) ([]rancher.Service, error)
    GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error)
}
```

### Step 5: Implement BundleDataSource Methods

```go
func (ds *BundleDataSource) GetCRDs(clusterID string) ([]rancher.CRD, error) {
    return ds.bundle.CRDs, nil
}

func (ds *BundleDataSource) GetDeployments(projectID, namespace string) ([]rancher.Deployment, error) {
    var filtered []rancher.Deployment
    for _, deployment := range ds.bundle.Deployments {
        if namespace == "" || deployment.NamespaceID == namespace {
            filtered = append(filtered, deployment)
        }
    }
    return filtered, nil
}

func (ds *BundleDataSource) GetServices(projectID, namespace string) ([]rancher.Service, error) {
    var filtered []rancher.Service
    for _, service := range ds.bundle.Services {
        if namespace == "" || service.NamespaceID == namespace {
            filtered = append(filtered, service)
        }
    }
    return filtered, nil
}

func (ds *BundleDataSource) GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error) {
    return ds.bundle.Namespaces, nil
}
```

### Step 6: Implement LiveDataSource Methods

```go
func (ds *LiveDataSource) GetCRDs(clusterID string) ([]rancher.CRD, error) {
    if ds.offlineMode {
        return nil, fmt.Errorf("offline mode")
    }
    
    crdList, err := ds.client.ListCRDs(clusterID)
    if err != nil {
        return nil, err
    }
    return crdList.Items, nil
}

// Similar for GetDeployments, GetServices, GetNamespaces
```

### Step 7: Wire Up TUI Functions

In `internal/tui/app.go`, update fetch functions:

```go
func (a *App) fetchCRDs(clusterID string) tea.Cmd {
    return func() tea.Msg {
        // Try data source first
        if a.dataSource != nil {
            crds, err := a.dataSource.GetCRDs(clusterID)
            if err == nil && len(crds) > 0 {
                return crdsMsg{crds: crds}
            }
        }
        
        // Fallback to mock
        mockCRDs := a.getMockCRDs()
        return crdsMsg{crds: mockCRDs}
    }
}

// Similar for fetchDeployments, fetchServices, fetchNamespaces
```

---

## Testing Strategy

### Unit Tests
```go
func TestParseCRDs(t *testing.T) {
    content := `NAME                           CREATED AT
addons.k3s.cattle.io          2025-11-20T00:45:08Z
alertmanagers.monitoring.io   2025-11-26T01:26:48Z`

    tmpDir := t.TempDir()
    kubectlDir := filepath.Join(tmpDir, "rke2/kubectl")
    os.MkdirAll(kubectlDir, 0755)
    os.WriteFile(filepath.Join(kubectlDir, "crds"), []byte(content), 0644)
    
    crds, err := ParseCRDs(tmpDir)
    assert.NoError(t, err)
    assert.Equal(t, 2, len(crds))
    assert.Equal(t, "addons.k3s.cattle.io", crds[0].Metadata.Name)
}
```

### Integration Test
```bash
# Load real bundle and verify resources
r8s --bundle=example-log-bundle/*.tar.gz

# Navigate to CRDs, verify real data displayed
# Browse deployments, verify real data shown
```

---

## Error Handling

### Missing Files
- If kubectl output file missing, fallback to empty slice
- Bundle still loads successfully

### Parse Errors
- Log warnings but don't fail bundle load
- Individual resource parsing errors shouldn't break entire bundle

### Data Quality
- Some fields may be incomplete
- Use sensible defaults where needed
- Don't crash on unexpected formats

---

## Performance Considerations

### Parsing Cost
- Parse once during bundle.Load()
- Store in memory (small data ~100KB total)
- No runtime parsing overhead

### Memory Usage
- CRDs: ~50 entries × 1KB = 50KB
- Deployments: ~20 entries × 500B = 10KB
- Services: ~30 entries × 500B = 15KB
- Namespaces: ~10 entries × 200B = 2KB
- **Total:** ~77KB additional memory per bundle

Negligible compared to 100MB bundle limit.

---

## Success Criteria

- [ ] ParseCRDs extracts CRD names and timestamps
- [ ] ParseDeployments extracts namespace, name, replicas
- [ ] ParseServices extracts namespace, name, type, ports
- [ ] ParseNamespaces extracts name and status
- [ ] Bundle.Load() populates all resource fields
- [ ] BundleDataSource returns real data
- [ ] TUI displays real CRDs from bundle (not mocks)
- [ ] TUI displays real Deployments from bundle
- [ ] TUI displays real Services from bundle
- [ ] TUI displays real Namespaces from bundle
- [ ] Graceful fallback if kubectl files missing
- [ ] No performance degradation

---

## Timeline

1. **Create kubectl.go** - 20 min
2. **Update types.go** - 5 min
3. **Update bundle.go** - 10 min
4. **Extend DataSource** - 10 min
5. **Implement methods** - 15 min
6. **Wire up TUI** - 15 min
7. **Test** - 15 min

**Total:** ~90 minutes

---

## Next Steps

1. Create `internal/bundle/kubectl.go`
2. Implement parsing functions
3. Update bundle loading
4. Test with real bundle
5. Commit and document

Then the bundle browser will be COMPLETE!
