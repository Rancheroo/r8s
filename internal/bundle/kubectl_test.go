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

func TestParseConfigMaps(t *testing.T) {
	tmpDir := t.TempDir()
	kubectlDir := filepath.Join(tmpDir, "rke2", "kubectl")
	if err := os.MkdirAll(kubectlDir, 0755); err != nil {
		t.Fatalf("Failed to create kubectl dir: %v", err)
	}

	configmapData := `NAMESPACE     NAME                              DATA   AGE
default       my-config                         3      5d
cattle-system rancher-config                    1      10d
kube-system   kube-proxy                        2      30d
`

	cmPath := filepath.Join(kubectlDir, "configmaps")
	if err := os.WriteFile(cmPath, []byte(configmapData), 0644); err != nil {
		t.Fatalf("Failed to write test configmaps file: %v", err)
	}

	configmaps, err := ParseConfigMaps(tmpDir)
	if err != nil {
		t.Fatalf("ParseConfigMaps failed: %v", err)
	}

	if len(configmaps) != 3 {
		t.Errorf("Expected 3 configmaps, got %d", len(configmaps))
	}

	if configmaps[0].Name != "my-config" || configmaps[0].Namespace != "default" {
		t.Errorf("First configmap mismatch: got %s/%s", configmaps[0].Namespace, configmaps[0].Name)
	}
	if configmaps[0].Data != 3 {
		t.Errorf("Expected first configmap to have 3 data entries, got %d", configmaps[0].Data)
	}
}

func TestParseHelmCharts(t *testing.T) {
	tmpDir := t.TempDir()
	kubectlDir := filepath.Join(tmpDir, "rke2", "kubectl")
	if err := os.MkdirAll(kubectlDir, 0755); err != nil {
		t.Fatalf("Failed to create kubectl dir: %v", err)
	}

	helmchartData := `NAMESPACE     NAME                    CHART                   VERSION    STATUS    REPO                        AGE
cattle-system rancher-monitoring      rancher-monitoring      102.0.0    Deployed  rancher-charts              7d
cattle-system rancher-logging         rancher-logging         102.0.1    Failed    rancher-charts              3d
fleet-system  fleet-agent             fleet-agent             0.7.0      Pending   rancher-charts              1d
`

	hcPath := filepath.Join(kubectlDir, "helmcharts")
	if err := os.WriteFile(hcPath, []byte(helmchartData), 0644); err != nil {
		t.Fatalf("Failed to write test helmcharts file: %v", err)
	}

	helmcharts, err := ParseHelmCharts(tmpDir)
	if err != nil {
		t.Fatalf("ParseHelmCharts failed: %v", err)
	}

	if len(helmcharts) != 3 {
		t.Errorf("Expected 3 helmcharts, got %d", len(helmcharts))
	}

	if helmcharts[0].Name != "rancher-monitoring" {
		t.Errorf("Expected first chart name 'rancher-monitoring', got '%s'", helmcharts[0].Name)
	}
	if helmcharts[0].Chart != "rancher-monitoring" {
		t.Errorf("Expected first chart 'rancher-monitoring', got '%s'", helmcharts[0].Chart)
	}
	if helmcharts[0].Version != "102.0.0" {
		t.Errorf("Expected first chart version '102.0.0', got '%s'", helmcharts[0].Version)
	}
	if helmcharts[0].Status != "Deployed" {
		t.Errorf("Expected first chart status 'Deployed', got '%s'", helmcharts[0].Status)
	}
	if helmcharts[0].Repo != "rancher-charts" {
		t.Errorf("Expected first chart repo 'rancher-charts', got '%s'", helmcharts[0].Repo)
	}

	// Check failed chart
	if helmcharts[1].Status != "Failed" {
		t.Errorf("Expected second chart status 'Failed', got '%s'", helmcharts[1].Status)
	}
}
