# Sprint 5: PV/PVC/StatefulSet Parser Implementation

**Issue**: #39 (BACKLOG-4)  
**Branch**: `feature/sprint5-pvc-parser`  
**Status**: In Progress

---

## Goal
Implement parsers for Kubernetes storage resources:
- PersistentVolumes (PV)
- PersistentVolumeClaims (PVC)
- StatefulSets

## Bundle Files to Parse

```
support-bundle/
└── rke2/
    └── kubectl/
        ├── pv          # kubectl get pv
        ├── pvc         # kubectl get pvc
        └── statefulsets  # kubectl get statefulsets
```

## Implementation Plan

### 1. Data Structures
Add to `internal/rancher/`:

```go
// PersistentVolume represents a PV resource
type PersistentVolume struct {
    Name          string
    Status        string // Available, Bound, Released, Failed
    StorageClass  string
    Capacity      string
    Claim         string // namespace/claim-name
    Age           string
}

// PersistentVolumeClaim represents a PVC resource
type PersistentVolumeClaim struct {
    Name          string
    Namespace     string
    Status        string // Pending, Bound, Lost
    StorageClass  string
    Capacity      string
    Volume        string // Bound PV name
    Age           string
}

// StatefulSet represents a StatefulSet resource
type StatefulSet struct {
    Name          string
    Namespace     string
    Replicas      string // ready/total
    StorageClass  string
    Age           string
}
```

### 2. Parser Functions
Add to `internal/bundle/kubectl.go`:

```go
func ParsePVs(extractPath string) ([]rancher.PersistentVolume, error)
func ParsePVCs(extractPath string) ([]rancher.PersistentVolumeClaim, error)
func ParseStatefulSets(extractPath string) ([]rancher.StatefulSet, error)
```

### 3. TUI Integration
Add to `internal/tui/`:
- Storage view (new view type)
- Table display for PV/PVC/StatefulSets
- Navigation key bindings

### 4. Testing
- Unit tests for parsers
- Integration tests with sample data

---

## Progress

- [ ] Data structures defined
- [ ] PV parser implemented
- [ ] PVC parser implemented
- [ ] StatefulSet parser implemented
- [ ] TUI view added
- [ ] Tests written
- [ ] Documentation updated

---

## Notes

Sample kubectl output format:

```
NAME                                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-7f45c8b2-1a2b-4c3d-8e9f-0a1b2c3d4e5f   Bound    pv-1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d   10Gi       RWO            standard       3d
```

Parsing approach:
1. Skip header line
2. Split by whitespace
3. Map columns to struct fields
4. Handle missing/empty values

