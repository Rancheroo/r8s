package bundle

import (
	"testing"
	"time"
)

func TestParseNamespaces(t *testing.T) {
	content := `NAME              STATUS   AGE
kube-system       Active   5d
default           Active   5d
cattle-system     Active   5d
longhorn-system   Active   3d
`

	bundlePath, cleanup := createKubectlTestBundle(t, map[string]string{"namespaces": content})
	defer cleanup()

	namespaces, err := ParseNamespaces(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(namespaces) != 4 {
		t.Fatalf("Expected 4 namespaces, got: %d", len(namespaces))
	}

	if namespaces[0].Name != "kube-system" {
		t.Errorf("Expected first namespace 'kube-system', got: %s", namespaces[0].Name)
	}

	// State is lowercased by parser
	if namespaces[0].State != "active" {
		t.Errorf("Expected state 'active', got: %s", namespaces[0].State)
	}
}

func TestParseNamespaces_Empty(t *testing.T) {
	bundlePath, cleanup := createKubectlTestBundle(t, map[string]string{"namespaces": "NAME\n"})
	defer cleanup()

	namespaces, err := ParseNamespaces(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(namespaces) != 0 {
		t.Errorf("Expected 0 namespaces, got: %d", len(namespaces))
	}
}

func TestParseDaemonSets(t *testing.T) {
	content := `NAMESPACE         NAME                DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
kube-system       calico-node         3         3         3       3            3           <none>          5d
kube-system       kube-proxy          3         3         3       3            3           <none>          5d
`

	bundlePath, cleanup := createKubectlTestBundle(t, map[string]string{"daemonsets": content})
	defer cleanup()

	daemonsets, err := ParseDaemonSets(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(daemonsets) != 2 {
		t.Fatalf("Expected 2 daemonsets, got: %d", len(daemonsets))
	}

	if daemonsets[0].Name != "calico-node" {
		t.Errorf("Expected name 'calico-node', got: %s", daemonsets[0].Name)
	}

	if daemonsets[0].Namespace != "kube-system" {
		t.Errorf("Expected namespace 'kube-system', got: %s", daemonsets[0].Namespace)
	}
}

func TestParsePVs(t *testing.T) {
	content := `NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS   REASON   AGE
pvc-abc123                                 10Gi       RWO            Delete           Bound    default/my-pvc           longhorn                2d
pvc-def456                                 20Gi       RWO            Delete           Bound    database/postgres-pvc    longhorn                3d
`

	bundlePath, cleanup := createKubectlTestBundle(t, map[string]string{"pv": content})
	defer cleanup()

	pvs, err := ParsePVs(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(pvs) != 2 {
		t.Fatalf("Expected 2 PVs, got: %d", len(pvs))
	}

	if pvs[0].Name != "pvc-abc123" {
		t.Errorf("Expected name 'pvc-abc123', got: %s", pvs[0].Name)
	}

	if pvs[0].Capacity != "10Gi" {
		t.Errorf("Expected capacity '10Gi', got: %s", pvs[0].Capacity)
	}

	if pvs[0].Status != "Bound" {
		t.Errorf("Expected status 'Bound', got: %s", pvs[0].Status)
	}
}

func TestParsePVCs(t *testing.T) {
	content := `NAMESPACE   NAME           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
default     my-pvc         Bound    pvc-abc123                                 10Gi       RWO            longhorn       2d
database    postgres-pvc   Bound    pvc-def456                                 20Gi       RWO            longhorn       3d
`

	bundlePath, cleanup := createKubectlTestBundle(t, map[string]string{"pvc": content})
	defer cleanup()

	pvcs, err := ParsePVCs(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(pvcs) != 2 {
		t.Fatalf("Expected 2 PVCs, got: %d", len(pvcs))
	}

	if pvcs[0].Name != "my-pvc" {
		t.Errorf("Expected name 'my-pvc', got: %s", pvcs[0].Name)
	}

	if pvcs[0].Namespace != "default" {
		t.Errorf("Expected namespace 'default', got: %s", pvcs[0].Namespace)
	}
}

func TestParseStatefulSets(t *testing.T) {
	content := `NAMESPACE   NAME           READY   AGE
database    postgres       1/1     5d
cache       redis          3/3     3d
`

	bundlePath, cleanup := createKubectlTestBundle(t, map[string]string{"statefulsets": content})
	defer cleanup()

	statefulsets, err := ParseStatefulSets(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(statefulsets) != 2 {
		t.Fatalf("Expected 2 statefulsets, got: %d", len(statefulsets))
	}

	if statefulsets[0].Name != "postgres" {
		t.Errorf("Expected name 'postgres', got: %s", statefulsets[0].Name)
	}

	if statefulsets[0].Replicas != "1/1" {
		t.Errorf("Expected replicas '1/1', got: %s", statefulsets[0].Replicas)
	}
}

func TestParseConfigMaps(t *testing.T) {
	content := `NAMESPACE         NAME                                 DATA   AGE
kube-system       calico-config                        4      5d
default           my-config                            2      2d
cattle-system     cattle-agent-config                  1      5d
`

	bundlePath, cleanup := createKubectlTestBundle(t, map[string]string{"configmaps": content})
	defer cleanup()

	configmaps, err := ParseConfigMaps(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(configmaps) != 3 {
		t.Fatalf("Expected 3 configmaps, got: %d", len(configmaps))
	}

	if configmaps[0].Name != "calico-config" {
		t.Errorf("Expected name 'calico-config', got: %s", configmaps[0].Name)
	}

	if configmaps[0].Namespace != "kube-system" {
		t.Errorf("Expected namespace 'kube-system', got: %s", configmaps[0].Namespace)
	}
}

func TestParseHelmCharts(t *testing.T) {
	// Parser expects: NAMESPACE NAME CHART VERSION STATUS AGE
	content := `NAMESPACE         NAME         CHART           VERSION    STATUS     AGE
cattle-system     rancher      rancher-2.8.0   v2.8.0     deployed   5d
longhorn-system   longhorn     longhorn-1.5.3  v1.5.3     deployed   3d
`

	bundlePath, cleanup := createKubectlTestBundle(t, map[string]string{"helmcharts": content})
	defer cleanup()

	charts, err := ParseHelmCharts(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(charts) != 2 {
		t.Fatalf("Expected 2 helm charts, got: %d", len(charts))
	}

	if charts[0].Name != "rancher" {
		t.Errorf("Expected name 'rancher', got: %s", charts[0].Name)
	}

	if charts[0].Namespace != "cattle-system" {
		t.Errorf("Expected namespace 'cattle-system', got: %s", charts[0].Namespace)
	}

	if charts[0].Status != "deployed" {
		t.Errorf("Expected status 'deployed', got: %s", charts[0].Status)
	}
}

func TestParseKubectlAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input       string
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{"5d", 4 * 24 * time.Hour, 6 * 24 * time.Hour},
		{"10m", 9 * time.Minute, 11 * time.Minute},
		{"30s", 29 * time.Second, 31 * time.Second},
		{"2h", 1 * time.Hour, 3 * time.Hour},
	}

	for _, tc := range tests {
		result := parseKubectlAge(tc.input)
		age := now.Sub(result)

		if age < tc.expectedMin || age > tc.expectedMax {
			t.Errorf("parseKubectlAge(%q) age %v outside expected range %v-%v",
				tc.input, age, tc.expectedMin, tc.expectedMax)
		}
	}
}
