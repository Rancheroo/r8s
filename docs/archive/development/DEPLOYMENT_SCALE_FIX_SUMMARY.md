# Deployment Scale Fix - Quick Reference

## The Problem
Deployments view shows error:
```
Error: failed to fetch deployments: failed to decode response: 
json: cannot unmarshal number into Go struct field Deployment.data.scale 
of type rancher.DeploymentScale
```

## Root Cause
Rancher API returns `scale` in two formats:
- **Format A:** `"scale": 1` (number) ← Current issue
- **Format B:** `"scale": {"scale": 1, "ready": 1, "total": 1}` (object)

Our code only handles Format B.

## The Fix

### Step 1: Update types.go
Add custom unmarshaler to handle both formats:

```go
type Deployment struct {
    // ... existing fields ...
    ScaleRaw json.RawMessage `json:"scale,omitempty"`
    Scale *DeploymentScale `json:"-"`
    // ... rest of fields ...
}

func (d *Deployment) UnmarshalJSON(data []byte) error {
    type Alias Deployment
    aux := &struct{ *Alias }{Alias: (*Alias)(d)}
    
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    
    if len(d.ScaleRaw) > 0 {
        var scaleObj DeploymentScale
        if err := json.Unmarshal(d.ScaleRaw, &scaleObj); err == nil {
            d.Scale = &scaleObj  // Object format
        } else {
            var scaleNum int
            if err := json.Unmarshal(d.ScaleRaw, &scaleNum); err == nil {
                d.Scale = &DeploymentScale{
                    Scale: scaleNum,
                    Ready: scaleNum,
                    Total: scaleNum,
                }
            }
        }
    }
    return nil
}
```

### Step 2: Update app.go display logic
Use fallback for replica counts:

```go
var desired, ready, uptodate, available int

if deployment.Scale != nil {
    desired = deployment.Scale.Scale
    ready = deployment.Scale.Ready
    uptodate = deployment.Scale.Total
    available = ready
} else {
    desired = deployment.Replicas
    ready = deployment.ReadyReplicas
    uptodate = deployment.UpToDateReplicas
    if uptodate == 0 {
        uptodate = deployment.UpdatedReplicas
    }
    available = deployment.AvailableReplicas
}

rows = append(rows, table.NewRow(table.RowData{
    "ready":     fmt.Sprintf("%d/%d", ready, desired),
    "uptodate":  fmt.Sprintf("%d", uptodate),
    "available": fmt.Sprintf("%d", available),
}))
```

## Files to Change
1. `internal/rancher/types.go` - Add UnmarshalJSON method
2. `internal/tui/app.go` - Update display logic (lines ~664-673)

## Test After Fix
- [ ] Deployments view loads without error
- [ ] Shows "1/1" instead of "0/0"
- [ ] Works in all namespaces
- [ ] No regression in other views

## Full Details
See: `CLINE_FIX_SPECIFICATION.md`
