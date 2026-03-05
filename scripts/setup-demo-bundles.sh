#!/bin/bash
set -e

# Setup Demo Bundles for r8s v1.3.2 Presentation

DEMO_DIR="$(pwd)"
BUNDLE1="${DEMO_DIR}/support-bundle-2026-03-01"
BUNDLE2="${DEMO_DIR}/support-bundle-2026-03-02"

echo "🎨 Creating Demo Bundles..."

# ==========================================
# Bundle 1: "Everything is Broken" (Critical)
# ==========================================
echo "  - Generating ${BUNDLE1} (Critical Failures)..."
mkdir -p "${BUNDLE1}/rke2/kubectl/poddescribe"
mkdir -p "${BUNDLE1}/rke2/podlogs"

# 1.1 Etcd Quorum Loss Log
cat <<EOF > "${BUNDLE1}/rke2/podlogs/kube-system-etcd-ip-10-0-0-1.log"
2026-03-01T10:00:00.000Z [INFO] etcdserver: starting to check integrity...
2026-03-01T10:00:05.000Z [WARN] etcdserver: request timed out, waiting for WAL write
2026-03-01T10:00:10.000Z [CRITICAL] etcdserver: lost majority (quorum loss), stopping members
2026-03-01T10:00:11.000Z [ERROR] raft: failed to send message: output buffer full
EOF

# 1.2 CNI Failure Log
cat <<EOF > "${BUNDLE1}/rke2/podlogs/calico-system-calico-node-x9f2k.log"
2026-03-01T10:00:00.000Z [INFO] Felix is running
2026-03-01T10:01:00.000Z [ERROR] felix/int_dataplane.go 1023: Failed to list ip sets error=cni plugin error: failed to connect to CNI
2026-03-01T10:01:05.000Z [FATAL] CNI network not ready: NetworkUnreachable
EOF

# 1.3 Expired Certificate Log
cat <<EOF > "${BUNDLE1}/rke2/podlogs/kube-system-rke2-server-master-1.log"
I0301 10:00:00.123456       1 server.go:123] Starting kube-apiserver...
E0301 10:05:00.123456       1 authenticator.go:123] x509: certificate has expired or is not yet valid: current time 2026-03-01T10:05:00Z is after 2026-02-28T10:00:00Z
E0301 10:05:01.123456       1 secure_serving.go:123] Serving cert is expired: /var/lib/rancher/rke2/server/tls/serving-kube-apiserver.crt
EOF

# 1.4 Pods List (OOMKilled)
cat <<EOF > "${BUNDLE1}/rke2/kubectl/pods"
NAMESPACE       NAME                                READY   STATUS             RESTARTS   AGE   IP            NODE
kube-system     etcd-ip-10-0-0-1                    0/1     Error              5          10d   10.0.0.1      master-1
calico-system   calico-node-x9f2k                   0/1     CrashLoopBackOff   12         10d   10.0.0.2      worker-1
default         app-backend-5d9b7f8c9d-abcde        0/1     OOMKilled          2          4h    10.42.1.5     worker-1
kube-system     rke2-server-master-1                1/1     Running            0          10d   10.0.0.1      master-1
EOF

# 1.5 Nodes List (NotReady)
cat <<EOF > "${BUNDLE1}/rke2/kubectl/nodes"
NAME       STATUS     ROLES                       AGE     VERSION
master-1   Ready      control-plane,etcd,master   10d     v1.28.4+rke2r1
worker-1   NotReady   worker                      10d     v1.28.4+rke2r1
EOF

# 1.6 Metadata
cat <<EOF > "${BUNDLE1}/metadata.json"
{"bundle_type": "rke2", "collected_at": "2026-03-01T12:00:00Z"}
EOF

# ==========================================
# Bundle 2: "Deep Dive" (CrashLoop Investigation)
# ==========================================
echo "  - Generating ${BUNDLE2} (Deep Dive)..."
mkdir -p "${BUNDLE2}/rke2/kubectl/poddescribe"
mkdir -p "${BUNDLE2}/rke2/podlogs"

# 2.1 Pods List
cat <<EOF > "${BUNDLE2}/rke2/kubectl/pods"
NAMESPACE       NAME                              READY   STATUS             RESTARTS          AGE   IP            NODE
cattle-system   rancher-webhook-5d9b7f8c9d-abcde  0/1     CrashLoopBackOff   5 (2m ago)        15m   10.42.2.10    worker-2
cattle-system   rancher-6d7b8c9d0e-12345          3/3     Running            0                 5d    10.42.2.11    worker-2
kube-system     coredns-7c8d9e0f1a-54321          1/1     Running            0                 5d    10.42.2.12    worker-2
EOF

# 2.2 Pod Logs (The Panic)
cat <<EOF > "${BUNDLE2}/rke2/podlogs/cattle-system-rancher-webhook-5d9b7f8c9d-abcde.log"
2026-03-02T14:00:00.000Z [INFO] Starting rancher-webhook v0.3.5
2026-03-02T14:00:01.000Z [INFO] Validating configuration...
2026-03-02T14:00:02.000Z [INFO] Connect to kubernetes... OK
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x1234567]

goroutine 1 [running]:
main.main()
        /go/src/github.com/rancher/webhook/main.go:123 +0x456
EOF

# 2.3 Pod Describe
cat <<EOF > "${BUNDLE2}/rke2/kubectl/poddescribe/cattle-system-rancher-webhook-5d9b7f8c9d-abcde"
Name:         rancher-webhook-5d9b7f8c9d-abcde
Namespace:    cattle-system
Priority:     0
Node:         worker-2/10.0.0.3
Start Time:   Mon, 02 Mar 2026 14:00:00 +0000
Labels:       app=rancher-webhook
              pod-template-hash=5d9b7f8c9d
Status:       Running
IP:           10.42.2.10
Containers:
  rancher-webhook:
    Container ID:   containerd://1234567890abcdef
    Image:          rancher/webhook:v0.3.5
    Image ID:       sha256:abcdef1234567890
    Port:           443/TCP
    Host Port:      0/TCP
    State:          Waiting
      Reason:       CrashLoopBackOff
    Last State:     Terminated
      Reason:       Error
      Exit Code:    2
      Started:      Mon, 02 Mar 2026 14:00:00 +0000
      Finished:     Mon, 02 Mar 2026 14:00:02 +0000
    Ready:          False
    Restart Count:  5
Events:
  Type     Reason     Age                   From               Message
  ----     ------     ----                  ----               -------
  Normal   Scheduled  15m                   default-scheduler  Successfully assigned cattle-system/rancher-webhook-5d9b7f8c9d-abcde to worker-2
  Normal   Pulled     15m                   kubelet            Container image "rancher/webhook:v0.3.5" already present on machine
  Normal   Created    15m                   kubelet            Created container rancher-webhook
  Normal   Started    15m                   kubelet            Started container rancher-webhook
  Warning  BackOff    2m (x5 over 4m)       kubelet            Back-off restarting failed container
EOF

# 2.4 Metadata
cat <<EOF > "${BUNDLE2}/metadata.json"
{"bundle_type": "rke2", "collected_at": "2026-03-02T14:15:00Z"}
EOF

echo "✅ Demo Bundles Ready!"
echo "   - $BUNDLE1"
echo "   - $BUNDLE2"
echo ""
echo "Run source usage:"
echo "export r8sbundle1=\"$BUNDLE1\""
echo "export r8sbundle2=\"$BUNDLE2\""
