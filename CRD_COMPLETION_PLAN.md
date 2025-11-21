# CRD Explorer Completion Plan

## Current Status
✅ **Completed:**
- CRD listing with proper table alignment
- Description toggle with 'i' key
- Dynamic description updates on row navigation
- CRD schema extraction from OpenAPI v3

## Outstanding Issues

### 1. CRD Instance Viewing (CRITICAL)
**Problem:** While we can list CRDs, pressing Enter on a CRD currently shows instances but we need to verify this works properly for all CRD types, especially Rancher-specific CRDs.

**Rancher System CRDs to Test:**
- `clusters.management.cattle.io` - Cluster definitions
- `machines.cluster.x-k8s.io` - Machine resources
- `plans.upgrade.cattle.io` - Upgrade plans
- `apps.catalog.cattle.io` - App catalog entries
- Fleet CRDs (`clusters.fleet.cattle.io`, `gitrepos.fleet.cattle.io`, etc.)

**Implementation Tasks:**
1. Test CRD instance listing with real Rancher CRDs
2. Verify namespace filtering works for namespaced CRDs
3. Add instance count to CRD list view
4. Handle edge cases (empty instance lists, 404s, etc.)
5. Add visual indicator showing which CRDs have instances

### 2. Missing Descriptions Enhancement
**Problem:** Many CRDs lack OpenAPI schema descriptions, showing "No description available."

**Proposed Solutions:**

#### Option A: AI-Generated Descriptions (Preferred)
- Integrate with LLM to generate descriptions based on CRD metadata
- Cache generated descriptions locally
- Use CRD name, group, and kind as context
- Example: "Machine represents a physical or virtual compute resource in the cluster"

#### Option B: Community Description Database
- Maintain a JSON file mapping CRD names to descriptions
- Start with well-known CRDs (Rancher, K8s ecosystem)
- Allow community contributions
- Fallback when OpenAPI schema is empty

#### Option C: Description Scraping
- Fetch descriptions from official documentation
- Use CRD group URL as hint (e.g., `cattle.io` → Rancher docs)
- Parse README or docs for CRD explanations

**Recommended Approach:**
Hybrid: Try OpenAPI schema → Local description database → AI generation → "No description"

### 3. CRD Instance Actions
**Missing Actions When Viewing Instances:**
- `d` - Describe instance (show full YAML/JSON)
- `e` - Edit instance (open in editor)
- `Del` - Delete instance (with confirmation)
- `y` - Export instance as YAML
- `l` - View logs (for controller CRDs)

### 4. CRD Schema Viewer
**Enhancement:** Show detailed schema in a tree view
- Display field types, required vs optional
- Show validation rules (min/max, regex, enum)
- Provide example YAML snippets
- Navigate schema hierarchy

### 5. CRD Search/Filter
**Missing Functionality:**
- Filter CRDs by group
- Search by name/kind
- Show only CRDs with instances
- Filter by scope (Namespaced vs Cluster)

## Testing Strategy (CRITICAL)
**Always test UI changes by running the application:**
```bash
make build
./bin/r9s
# Navigate and interact with TUI
# Verify visual alignment, functionality, and error handling
```

## Implementation Priority
1. **P0:** Fix/verify CRD instance viewing for Rancher CRDs
2. **P1:** Add instance counts to CRD list
3. **P2:** Implement description enhancement (start with local database)
4. **P3:** Add CRD instance actions (describe, edit, delete)
5. **P4:** Schema viewer and search/filter

## Notes for Next Conversation
- Test the Enter key on various CRDs to see if instances load
- Check specifically for 404 errors with Rancher system CRDs
- The description API already extracts from OpenAPI schema
- Consider using the WARP.md file to document completed features
