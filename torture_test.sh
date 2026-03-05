#!/bin/bash

BUNDLE="/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-03-03_03_49_04/"
OUTPUT_FILE="torture_test_results.txt"
R8S_BIN="./bin/r8s"

echo "Starting torture test..." > "$OUTPUT_FILE"
echo "Bundle: $BUNDLE" >> "$OUTPUT_FILE"
echo "Date: $(date)" >> "$OUTPUT_FILE"
echo "----------------------------------------" >> "$OUTPUT_FILE"

run_query() {
    local query="$1"
    echo ">> Query: $query" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    $R8S_BIN ask "$BUNDLE" "$query" >> "$OUTPUT_FILE" 2>&1
    echo "----------------------------------------" >> "$OUTPUT_FILE"
}

# 1. Pod Issues
run_query "why is my pod crashing?"
run_query "which pods are crashlooping?"
run_query "show me OOMkilled pods"
run_query "list pods with memory issues"
run_query "why are pods pending?"
run_query "show image pull errors"
run_query "which containers are restarting?"
run_query "pods in crashloopbackoff"
run_query "find failing pods"
run_query "what pods are broken?"

# 2. Node Issues
run_query "which nodes are not ready?"
run_query "show node pressure"
run_query "is any node out of disk?"
run_query "memory pressure on nodes"
run_query "why is the node failing?"
run_query "list unhealthy nodes"
run_query "check node status"
run_query "are there dead nodes?"
run_query "pid pressure detected?"
run_query "node conditions"

# 3. Control Plane
run_query "is etcd healthy?"
run_query "show etcd latency"
run_query "leader election failures"
run_query "api server errors"
run_query "scheduler issues"
run_query "controller manager log errors"
run_query "etcd corruption"
run_query "control plane status"
run_query "why is leader election failing?"
run_query "database space exceeded?"

# 4. Networking
run_query "dns resolution failed"
run_query "cni plugin errors"
run_query "network connectivity timeout"
run_query "service lookup failed"
run_query "calico errors"
run_query "flannel issues"
run_query "can't resolve host"
run_query "i/o timeout"
run_query "network policy drops"
run_query "ingress controller errors"

# 5. Storage
run_query "pvc not bound"
run_query "persistent volume errors"
run_query "storage pressure"
run_query "failed to mount volume"
run_query "volume attachment failed"
run_query "csi driver errors"
run_query "longhorn issues"
run_query "is storage full?"
run_query "disk space warning"
run_query "provisioning failed"

# 6. Specific Names (from known bundle data)
run_query "why is r8s-test-crash-segfault crashing?"
run_query "what is wrong with r8s-test-crash-exit1?"
run_query "check pod oomkill-v2"
run_query "issues with leader-election-failure"
run_query "status of r8s-test-worker-processor"

# 7. Certificates
run_query "are certificates expired?"
run_query "certificate warnings"
run_query "x509 errors"
run_query "tls handshake failed"
run_query "cert manager issues"

# 8. Comparison/Time
run_query "what happened recently?"
run_query "show errors from last hour"
run_query "what changed today?"
run_query "timeline of failures"

# 9. Meta/Broad
run_query "summarize all issues"
run_query "what is the root cause?"
run_query "show me everything"
run_query "analyze this bundle"
run_query "give me a report"
run_query "health check"
run_query "overview"
run_query "top issues"
run_query "critical errors only"
run_query "just warnings"

# 10. Nonsense/Edge Cases
run_query "hello r8s"
run_query "make me a sandwich"
run_query "sudo rm -rf /"
run_query "..."
run_query "12345"
run_query "pod pod pod"
run_query "why why why"
run_query "is it sunny?"
run_query "kubernetes"
run_query ""

echo "Torture test complete. Check $OUTPUT_FILE"
