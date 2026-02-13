package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePVs_UnboundPVs(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	kubectlDir := filepath.Join(tmpDir, "rke2", "kubectl")
	if err := os.MkdirAll(kubectlDir, 0755); err != nil {
		t.Fatalf("Failed to create kubectl dir: %v", err)
	}

	// Test data: mix of bound and unbound PVs
	pvData := `NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM                           STORAGECLASS   REASON   AGE
pvc-7f45c8b2-1a2b-4c3d-8e9f-0a1b2c3d4e5f   10Gi       RWO            Delete           Bound     default/my-pvc                standard                3d
pv-available-1                             5Gi        RWO            Retain           Available                                 fast                    7d
pvc-abcd1234-5678-90ab-cdef-example11111   20Gi       RWX            Delete           Bound     cattle-system/rancher-data    standard                1d
pv-available-2                             100Gi      RWO            Delete           Available                                 premium                 2d
`

	pvPath := filepath.Join(kubectlDir, "pv")
	if err := os.WriteFile(pvPath, []byte(pvData), 0644); err != nil {
		t.Fatalf("Failed to write test pv file: %v", err)
	}

	// Parse the PVs
	pvs, err := ParsePVs(tmpDir)
	if err != nil {
		t.Fatalf("ParsePVs failed: %v", err)
	}

	// Should have 4 PVs
	if len(pvs) != 4 {
		t.Errorf("Expected 4 PVs, got %d", len(pvs))
	}

	// Check first bound PV
	if pvs[0].Name != "pvc-7f45c8b2-1a2b-4c3d-8e9f-0a1b2c3d4e5f" {
		t.Errorf("Expected first PV name to be 'pvc-7f45c8b2-1a2b-4c3d-8e9f-0a1b2c3d4e5f', got '%s'", pvs[0].Name)
	}
	if pvs[0].Status != "Bound" {
		t.Errorf("Expected first PV status to be 'Bound', got '%s'", pvs[0].Status)
	}
	if pvs[0].Claim != "default/my-pvc" {
		t.Errorf("Expected first PV claim to be 'default/my-pvc', got '%s'", pvs[0].Claim)
	}
	if pvs[0].StorageClass != "standard" {
		t.Errorf("Expected first PV storageclass to be 'standard', got '%s'", pvs[0].StorageClass)
	}

	// Check unbound PV (Available status)
	if pvs[1].Name != "pv-available-1" {
		t.Errorf("Expected second PV name to be 'pv-available-1', got '%s'", pvs[1].Name)
	}
	if pvs[1].Status != "Available" {
		t.Errorf("Expected second PV status to be 'Available', got '%s'", pvs[1].Status)
	}
	if pvs[1].Claim != "" {
		t.Errorf("Expected unbound PV to have empty claim, got '%s'", pvs[1].Claim)
	}
	if pvs[1].StorageClass != "fast" {
		t.Errorf("Expected second PV storageclass to be 'fast', got '%s'", pvs[1].StorageClass)
	}

	// Check second unbound PV
	if pvs[3].Name != "pv-available-2" {
		t.Errorf("Expected fourth PV name to be 'pv-available-2', got '%s'", pvs[3].Name)
	}
	if pvs[3].Status != "Available" {
		t.Errorf("Expected fourth PV status to be 'Available', got '%s'", pvs[3].Status)
	}
	if pvs[3].Claim != "" {
		t.Errorf("Expected unbound PV to have empty claim, got '%s'", pvs[3].Claim)
	}
}

func TestParsePVs_AllUnbound(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	kubectlDir := filepath.Join(tmpDir, "rke2", "kubectl")
	if err := os.MkdirAll(kubectlDir, 0755); err != nil {
		t.Fatalf("Failed to create kubectl dir: %v", err)
	}

	// Test data: all unbound PVs
	pvData := `NAME              CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM   STORAGECLASS   REASON   AGE
pv-available-1    5Gi        RWO            Retain           Available                   fast                    7d
pv-available-2    100Gi      RWO            Delete           Available                   premium                 2d
`

	pvPath := filepath.Join(kubectlDir, "pv")
	if err := os.WriteFile(pvPath, []byte(pvData), 0644); err != nil {
		t.Fatalf("Failed to write test pv file: %v", err)
	}

	// Parse the PVs
	pvs, err := ParsePVs(tmpDir)
	if err != nil {
		t.Fatalf("ParsePVs failed: %v", err)
	}

	// Should have 2 PVs
	if len(pvs) != 2 {
		t.Errorf("Expected 2 PVs, got %d", len(pvs))
	}

	// Both should have empty claims
	for i, pv := range pvs {
		if pv.Claim != "" {
			t.Errorf("Expected PV %d to have empty claim, got '%s'", i, pv.Claim)
		}
		if pv.Status != "Available" {
			t.Errorf("Expected PV %d to have status 'Available', got '%s'", i, pv.Status)
		}
	}
}
