#!/bin/bash
set -e

# Base bundles (Real data)
BASE_SRC_1="/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57"
BASE_SRC_2="/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-03-03_03_49_04"

# Destinations
DEMO_DIR="/tmp"
BUNDLE1="${DEMO_DIR}/r8s-demo-critical"
BUNDLE2="${DEMO_DIR}/r8s-demo-deepdive"

echo "🎨 Setting up Demo Bundles in ${DEMO_DIR}..." >&2

# ---------------------------------------------------------
# Check Base Bundles
# ---------------------------------------------------------
if [ ! -d "$BASE_SRC_1" ] || [ ! -d "$BASE_SRC_2" ]; then
    echo "❌ Base bundles not found in ~/Downloads/logBundles/" >&2
    echo "Please ensure the real log bundles are present." >&2
    exit 1
fi

# ---------------------------------------------------------
# Setup Bundle 1: "Critical Failure" (Base: 2026-02-03)
# ---------------------------------------------------------
echo "  - Creating ${BUNDLE1} (Rich Data + Critical Errors)..." >&2
rm -rf "$BUNDLE1"
cp -r "$BASE_SRC_1" "$BUNDLE1"

# Inject 1.1: Etcd Quorum Loss Log
mkdir -p "${BUNDLE1}/rke2/podlogs"
cat <<LOGEOF > "${BUNDLE1}/rke2/podlogs/kube-system-etcd-r8s-cp-wlp7h-lhvgq.log"
2026-03-01T10:00:00.000Z [INFO] etcdserver: starting to check integrity...
2026-03-01T10:00:05.000Z [WARN] etcdserver: request timed out, waiting for WAL write
2026-03-01T10:00:10.000Z [CRITICAL] etcdserver: lost majority (quorum loss), stopping members
2026-03-01T10:00:11.000Z [ERROR] raft: failed to send message: output buffer full
LOGEOF

# Inject 1.2: CNI Failure Log
cat <<LOGEOF > "${BUNDLE1}/rke2/podlogs/calico-system-calico-node-x9f2k.log"
2026-03-01T10:00:00.000Z [INFO] Felix is running
2026-03-01T10:01:00.000Z [ERROR] felix/int_dataplane.go 1023: Failed to list ip sets error=cni plugin error: failed to connect to CNI
2026-03-01T10:01:05.000Z [FATAL] CNI network not ready: NetworkUnreachable
LOGEOF

# Inject 1.3: Expired Certificate Log (using real path structure if possible, but creating new file is safer)
cat <<LOGEOF > "${BUNDLE1}/rke2/podlogs/kube-system-rke2-server-r8s-cp-wlp7h-lhvgq.log"
I0301 10:00:00.123456       1 server.go:123] Starting kube-apiserver...
E0301 10:05:00.123456       1 authenticator.go:123] x509: certificate has expired or is not yet valid: current time 2026-03-01T10:05:00Z is after 2026-02-28T10:00:00Z
E0301 10:05:01.123456       1 secure_serving.go:123] Serving cert is expired: /var/lib/rancher/rke2/server/tls/serving-kube-apiserver.crt
LOGEOF

# Inject 1.4: Update kubectl/pods to show these failures
# We append to the existing file so we keep the real pods too!
PODS_FILE="${BUNDLE1}/rke2/kubectl/pods"
if [ -f "$PODS_FILE" ]; then
    echo "" >> "$PODS_FILE"
    echo "kube-system     etcd-r8s-cp-wlp7h-lhvgq             0/1     Error              5          10d   10.0.0.1      r8s-cp-wlp7h-lhvgq" >> "$PODS_FILE"
    echo "calico-system   calico-node-x9f2k                   0/1     CrashLoopBackOff   12         10d   10.0.0.2      r8s-wk-jnhwv-4xqzn" >> "$PODS_FILE"
    echo "default         app-backend-5d9b7f8c9d-abcde        0/1     OOMKilled          2          4h    10.42.1.5     r8s-wk-jnhwv-4xqzn" >> "$PODS_FILE"
fi

# ---------------------------------------------------------
# Setup Bundle 2: "Deep Dive" (Base: 2026-03-03)
# ---------------------------------------------------------
echo "  - Creating ${BUNDLE2} (Rich Data + CrashLoop)..." >&2
rm -rf "$BUNDLE2"
cp -r "$BASE_SRC_2" "$BUNDLE2"

# Inject 2.1: The CrashLoop Pod Log
mkdir -p "${BUNDLE2}/rke2/podlogs"
cat <<LOGEOF > "${BUNDLE2}/rke2/podlogs/cattle-system-rancher-webhook-5d9b7f8c9d-abcde.log"
2026-03-02T14:00:00.000Z [INFO] Starting rancher-webhook v0.3.5
2026-03-02T14:00:01.000Z [INFO] Validating configuration...
2026-03-02T14:00:02.000Z [INFO] Connect to kubernetes... OK
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x1234567]

goroutine 1 [running]:
main.main()
        /go/src/github.com/rancher/webhook/main.go:123 +0x456
LOGEOF

# Inject 2.2: The Pod Describe
mkdir -p "${BUNDLE2}/rke2/kubectl/poddescribe"
cat <<LOGEOF > "${BUNDLE2}/rke2/kubectl/poddescribe/cattle-system-rancher-webhook-5d9b7f8c9d-abcde"
Name:         rancher-webhook-5d9b7f8c9d-abcde
Namespace:    cattle-system
Priority:     0
Node:         r8s-wk-jnhwv-j5k8h/10.0.0.3
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
  Normal   Scheduled  15m                   default-scheduler  Successfully assigned cattle-system/rancher-webhook-5d9b7f8c9d-abcde to r8s-wk-jnhwv-j5k8h
  Normal   Pulled     15m                   kubelet            Container image "rancher/webhook:v0.3.5" already present on machine
  Normal   Created    15m                   kubelet            Created container rancher-webhook
  Normal   Started    15m                   kubelet            Started container rancher-webhook
  Warning  BackOff    2m (x5 over 4m)       kubelet            Back-off restarting failed container
LOGEOF

# Inject 2.3: Update kubectl/pods
PODS_FILE_2="${BUNDLE2}/rke2/kubectl/pods"
if [ -f "$PODS_FILE_2" ]; then
    echo "" >> "$PODS_FILE_2"
    echo "cattle-system   rancher-webhook-5d9b7f8c9d-abcde  0/1     CrashLoopBackOff   5 (2m ago)        15m   10.42.2.10    r8s-wk-jnhwv-j5k8h" >> "$PODS_FILE_2"
fi

# Inject 2.4: Ensure nodes exist (REMOVED - using real nodes)
# We use existing r8s-wk-jnhwv-j5k8h node

echo "✅ Rich Demo Bundles Ready in /tmp!" >&2
echo "   - $BUNDLE1 (Critical)" >&2
echo "   - $BUNDLE2 (Deep Dive)" >&2
echo "" >&2
echo "To configure your environment, run:" >&2
echo "eval \$($0 env)" >&2

if [ "$1" == "env" ]; then
    # Output solely the export commands for eval
    echo "export r8sbundle1=\"$BUNDLE1\""
    echo "export r8sbundle2=\"$BUNDLE2\""
    exit 0
fi

echo "" >&2
echo "Or manually run:" >&2
echo "export r8sbundle1=\"$BUNDLE1\"" >&2
echo "export r8sbundle2=\"$BUNDLE2\"" >&2
