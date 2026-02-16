package rancher

import (
	"testing"
	"time"
)

func TestCluster(t *testing.T) {
	cluster := Cluster{
		ID:        "c-12345",
		Name:      "test-cluster",
		State:     "active",
		Provider:  "rke2",
		Created:   time.Now(),
	}

	if cluster.ID != "c-12345" {
		t.Errorf("Cluster.ID = %v, want c-12345", cluster.ID)
	}

	if cluster.Name != "test-cluster" {
		t.Errorf("Cluster.Name = %v", cluster.Name)
	}

	if cluster.Provider != "rke2" {
		t.Errorf("Cluster.Provider = %v, want rke2", cluster.Provider)
	}
}

func TestNamespace(t *testing.T) {
	ns := Namespace{
		ID:        "n-default",
		Name:      "default",
		ProjectID: "p-default",
		ClusterID: "c-12345",
		State:     "active",
	}

	if ns.Name != "default" {
		t.Errorf("Namespace.Name = %v", ns.Name)
	}

	if ns.ProjectID != "p-default" {
		t.Errorf("Namespace.ProjectID = %v", ns.ProjectID)
	}
}

func TestPod(t *testing.T) {
	pod := Pod{
		ID:            "p-abc123",
		Name:          "nginx-pod",
		NamespaceID:   "default",
		NodeName:      "worker-1",
		State:         "running",
		PodIP:         "10.0.0.1",
		RestartCount:  0,
		Created:       time.Now(),
		KubectlReady:  "1/1",
		KubectlStatus: "Running",
		KubectlEvents: []string{"Started container", "Created container"},
	}

	if pod.Name != "nginx-pod" {
		t.Errorf("Pod.Name = %v", pod.Name)
	}

	if pod.NamespaceID != "default" {
		t.Errorf("Pod.NamespaceID = %v", pod.NamespaceID)
	}

	if pod.KubectlReady != "1/1" {
		t.Errorf("Pod.KubectlReady = %v", pod.KubectlReady)
	}

	if len(pod.KubectlEvents) != 2 {
		t.Errorf("Pod.KubectlEvents length = %d, want 2", len(pod.KubectlEvents))
	}
}

func TestDeployment(t *testing.T) {
	dep := Deployment{
		ID:          "d-xyz789",
		Name:        "nginx-deployment",
		NamespaceID: "default",
		State:       "active",
		Replicas:    3,
	}

	if dep.Name != "nginx-deployment" {
		t.Errorf("Deployment.Name = %v", dep.Name)
	}

	if dep.NamespaceID != "default" {
		t.Errorf("Deployment.NamespaceID = %v", dep.NamespaceID)
	}

	if dep.Replicas != 3 {
		t.Errorf("Deployment.Replicas = %d, want 3", dep.Replicas)
	}
}

func TestService(t *testing.T) {
	svc := Service{
		ID:          "s-svc123",
		Name:        "nginx-service",
		NamespaceID: "default",
		State:       "active",
		Kind:        "ClusterIP",
		ClusterIP:   "10.43.0.100",
	}

	if svc.Name != "nginx-service" {
		t.Errorf("Service.Name = %v", svc.Name)
	}

	if svc.Kind != "ClusterIP" {
		t.Errorf("Service.Kind = %v", svc.Kind)
	}

	if svc.ClusterIP != "10.43.0.100" {
		t.Errorf("Service.ClusterIP = %v", svc.ClusterIP)
	}
}

func TestCRD(t *testing.T) {
	crd := CRD{
		Metadata: ObjectMeta{
			Name: "addons.k3s.cattle.io",
		},
		Spec: CRDSpec{
			Group: "k3s.cattle.io",
			Names: CRDNames{
				Plural:   "addons",
				Singular: "addon",
				Kind:     "Addon",
			},
			Scope: "Namespaced",
			Versions: []CRDVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
		},
	}

	if crd.Spec.Group != "k3s.cattle.io" {
		t.Errorf("CRD.Spec.Group = %v", crd.Spec.Group)
	}

	if crd.Spec.Names.Kind != "Addon" {
		t.Errorf("CRD.Spec.Names.Kind = %v", crd.Spec.Names.Kind)
	}

	if len(crd.Spec.Versions) != 1 {
		t.Errorf("CRD.Spec.Versions length = %d", len(crd.Spec.Versions))
	}
}

func TestEvent(t *testing.T) {
	event := Event{
		Type:       "Warning",
		Reason:     "FailedMount",
		Message:    "MountVolume.SetUp failed",
		ObjectKind: "pod",
		Namespace:  "default",
		PodName:    "test-pod",
		Count:      5,
		Source:     "kubelet",
	}

	if event.Type != "Warning" {
		t.Errorf("Event.Type = %v", event.Type)
	}

	if event.Reason != "FailedMount" {
		t.Errorf("Event.Reason = %v", event.Reason)
	}

	if event.Count != 5 {
		t.Errorf("Event.Count = %d, want 5", event.Count)
	}
}

func TestProject(t *testing.T) {
	project := Project{
		ID:          "p-default",
		Name:        "default",
		ClusterID:   "c-12345",
		State:       "active",
		Description: "Default project",
	}

	if project.Name != "default" {
		t.Errorf("Project.Name = %v", project.Name)
	}

	if project.ClusterID != "c-12345" {
		t.Errorf("Project.ClusterID = %v", project.ClusterID)
	}
}

func TestObjectMeta(t *testing.T) {
	om := ObjectMeta{
		Name:              "test-pod",
		Namespace:         "default",
		UID:               "uid-12345",
		ResourceVersion:   "12345",
		CreationTimestamp: time.Now(),
		Labels:            map[string]string{"app": "test"},
		Annotations:       map[string]string{"note": "test"},
	}

	if om.Name != "test-pod" {
		t.Errorf("ObjectMeta.Name = %v", om.Name)
	}

	if om.Labels["app"] != "test" {
		t.Errorf("ObjectMeta.Labels[app] = %v", om.Labels["app"])
	}
}

func TestCRDVersion(t *testing.T) {
	version := CRDVersion{
		Name:    "v1",
		Served:  true,
		Storage: true,
	}

	if version.Name != "v1" {
		t.Errorf("CRDVersion.Name = %v", version.Name)
	}

	if !version.Served {
		t.Error("CRDVersion.Served should be true")
	}
}

func TestPersistentVolume(t *testing.T) {
	pv := PersistentVolume{
		Name:         "pv-001",
		Status:       "Bound",
		StorageClass: "standard",
		Capacity:     "10Gi",
		Claim:        "default/pvc-001",
		Age:          "5d",
	}

	if pv.Name != "pv-001" {
		t.Errorf("PersistentVolume.Name = %v", pv.Name)
	}

	if pv.Status != "Bound" {
		t.Errorf("PersistentVolume.Status = %v", pv.Status)
	}
}

func TestPersistentVolumeClaim(t *testing.T) {
	pvc := PersistentVolumeClaim{
		Name:         "pvc-001",
		Namespace:    "default",
		Status:       "Bound",
		StorageClass: "standard",
		Capacity:     "10Gi",
		Volume:       "pv-001",
		Age:          "5d",
	}

	if pvc.Name != "pvc-001" {
		t.Errorf("PersistentVolumeClaim.Name = %v", pvc.Name)
	}

	if pvc.Status != "Bound" {
		t.Errorf("PersistentVolumeClaim.Status = %v", pvc.Status)
	}
}

func TestStatefulSet(t *testing.T) {
	ss := StatefulSet{
		Name:         "web",
		Namespace:    "default",
		Replicas:     "3/3",
		StorageClass: "standard",
		Age:          "7d",
	}

	if ss.Name != "web" {
		t.Errorf("StatefulSet.Name = %v", ss.Name)
	}

	if ss.Replicas != "3/3" {
		t.Errorf("StatefulSet.Replicas = %v", ss.Replicas)
	}
}

func TestServicePort(t *testing.T) {
	port := ServicePort{
		Name:       "http",
		Port:       80,
		TargetPort: 8080,
		NodePort:   30080,
		Protocol:   "TCP",
	}

	if port.Port != 80 {
		t.Errorf("ServicePort.Port = %d", port.Port)
	}

	if port.Name != "http" {
		t.Errorf("ServicePort.Name = %v", port.Name)
	}
}
