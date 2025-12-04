#!/bin/bash
# test-attention-dashboard.sh
# 
# Generate test alerts for r8s Attention Dashboard validation
# 
# WARNING: This creates INTENTIONALLY BROKEN pods on your cluster!
#          Only run on TEST/DEV clusters, NEVER production!
#
# Usage:
#   ./test-attention-dashboard.sh setup    - Create test pods
#   ./test-attention-dashboard.sh status   - Check test pod status
#   ./test-attention-dashboard.sh cleanup  - Delete test pods
#   ./test-attention-dashboard.sh help     - Show this help

set -e

# Configuration
NAMESPACE="${TEST_NAMESPACE:-default}"
WAIT_TIME=3  # minutes to wait for alerts to trigger

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC}  $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC}  $1"
}

# Check prerequisites
check_prerequisites() {
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl not found. Please install kubectl first."
        exit 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster."
        print_info "Make sure kubectl is configured and you have access to a cluster."
        exit 1
    fi

    print_success "kubectl configured and connected to cluster"
}

# Confirm running on test cluster
confirm_test_cluster() {
    print_warning "This script will create INTENTIONALLY BROKEN pods!"
    echo -e "${YELLOW}Current context: $(kubectl config current-context)${NC}"
    echo -e "${YELLOW}Namespace: ${NAMESPACE}${NC}"
    echo ""
    read -p "Is this a TEST/DEV cluster? (yes/no): " confirm
    
    if [[ "$confirm" != "yes" ]]; then
        print_error "Aborted. Only run on test clusters!"
        exit 1
    fi
    
    print_success "Confirmed test cluster"
}

# Setup function - creates all test pods
setup_alerts() {
    print_header "Creating Test Pods for Attention Dashboard"
    
    check_prerequisites
    confirm_test_cluster
    
    echo -e "${BLUE}Creating 5 test pods to trigger different alerts...${NC}\n"
    
    # 1. ImagePullBackOff Alert
    print_info "Creating: test-imagepull (ImagePullBackOff)"
    kubectl run test-imagepull \
        --image=nonexistent/invalid:tag \
        --namespace=${NAMESPACE} \
        --labels="test=r8s-attention" 2>/dev/null || print_warning "Pod may already exist"
    print_success "test-imagepull created (will trigger in ~30 seconds)"
    
    # 2. CrashLoopBackOff Alert
    print_info "Creating: test-crash (CrashLoopBackOff)"
    kubectl run test-crash \
        --image=busybox \
        --namespace=${NAMESPACE} \
        --labels="test=r8s-attention" \
        --command -- /bin/sh -c "exit 1" 2>/dev/null || print_warning "Pod may already exist"
    print_success "test-crash created (will trigger in ~2-3 minutes)"
    
    # 3. OOMKilled Alert
    print_info "Creating: test-oom (OOMKilled)"
    cat <<EOF | kubectl apply -f - 2>/dev/null || print_warning "Pod may already exist"
apiVersion: v1
kind: Pod
metadata:
  name: test-oom
  namespace: ${NAMESPACE}
  labels:
    test: r8s-attention
spec:
  containers:
  - name: memory-hog
    image: polinux/stress
    resources:
      limits:
        memory: "50Mi"
      requests:
        memory: "50Mi"
    command: ["stress"]
    args: ["--vm", "1", "--vm-bytes", "100M", "--vm-hang", "1"]
  restartPolicy: OnFailure
EOF
    print_success "test-oom created (will trigger in ~30 seconds)"
    
    # 4. High Restart Count Alert
    print_info "Creating: test-restarts (High Restart Count)"
    kubectl run test-restarts \
        --image=busybox \
        --namespace=${NAMESPACE} \
        --labels="test=r8s-attention" \
        --command -- /bin/sh -c "sleep 5; exit 1" 2>/dev/null || print_warning "Pod may already exist"
    print_success "test-restarts created (will accumulate restarts in ~5-10 minutes)"
    
    # 5. Multi-Container Not Ready Alert (CRITICAL - tests the bug fix!)
    print_info "Creating: test-notready (Multi-Container Not Ready)"
    cat <<EOF | kubectl apply -f - 2>/dev/null || print_warning "Pod may already exist"
apiVersion: v1
kind: Pod
metadata:
  name: test-notready
  namespace: ${NAMESPACE}
  labels:
    test: r8s-attention
spec:
  containers:
  - name: ok-container
    image: nginx:alpine
    ports:
    - containerPort: 80
  - name: failing-container
    image: busybox
    command: ["/bin/sh", "-c", "exit 1"]
  restartPolicy: Never
EOF
    print_success "test-notready created (will show 1/2 ready)"
    
    print_header "Setup Complete!"
    
    echo -e "${GREEN}All test pods created successfully!${NC}\n"
    
    print_info "Wait times for alerts to trigger:"
    echo "  • ImagePullBackOff: ~30 seconds"
    echo "  • OOMKilled: ~30 seconds"
    echo "  • Multi-Container Not Ready: ~30 seconds"
    echo "  • CrashLoopBackOff: ~2-3 minutes"
    echo "  • High Restarts: ~5-10 minutes (for ≥3 restarts)"
    
    echo ""
    print_info "Next steps:"
    echo "  1. Wait ${WAIT_TIME} minutes for alerts to develop"
    echo "  2. Check status: ./test-attention-dashboard.sh status"
    echo "  3. Run r8s to see alerts: ./r8s"
    echo "  4. Cleanup when done: ./test-attention-dashboard.sh cleanup"
    
    echo ""
    print_warning "Starting ${WAIT_TIME}-minute countdown..."
    echo -e "${YELLOW}You can Ctrl+C to skip the wait and check manually.${NC}"
    
    for i in $(seq ${WAIT_TIME} -1 1); do
        echo -ne "\rTime remaining: ${i} minute(s)..."
        sleep 60
    done
    echo -e "\n"
    
    print_success "Wait complete! Check pod status now."
    status_alerts
}

# Status function - check test pod status
status_alerts() {
    print_header "Test Pod Status"
    
    check_prerequisites
    
    echo -e "${BLUE}Checking test pods in namespace: ${NAMESPACE}${NC}\n"
    
    if ! kubectl get pods -n ${NAMESPACE} -l test=r8s-attention &> /dev/null; then
        print_warning "No test pods found. Run './test-attention-dashboard.sh setup' first."
        exit 0
    fi
    
    # Show pod status
    echo -e "${CYAN}Pod Status:${NC}"
    kubectl get pods -n ${NAMESPACE} -l test=r8s-attention -o wide
    
    echo ""
    print_info "Expected alerts in r8s dashboard:"
    echo ""
    
    # Check each pod's status and show expected alert
    check_pod_alert "test-imagepull" "🚫 test-imagepull | ImagePullBackOff | ${NAMESPACE}"
    check_pod_alert "test-crash" "💀 test-crash | CrashLoopBackOff | ${NAMESPACE}"
    check_pod_alert "test-oom" "💀 test-oom | OOMKilled | ${NAMESPACE}"
    check_pod_alert "test-restarts" "🔥 test-restarts | X restarts | ${NAMESPACE} (X ≥ 3)"
    check_pod_alert "test-notready" "⚠️  test-notready | Not ready (1/2) | ${NAMESPACE}"
    
    echo ""
    print_info "To view alerts in r8s:"
    echo "  ./r8s"
    echo ""
    print_info "To cleanup test pods:"
    echo "  ./test-attention-dashboard.sh cleanup"
}

# Check individual pod and show expected alert
check_pod_alert() {
    local pod_name=$1
    local expected_alert=$2
    
    if kubectl get pod ${pod_name} -n ${NAMESPACE} &> /dev/null; then
        local status=$(kubectl get pod ${pod_name} -n ${NAMESPACE} -o jsonpath='{.status.phase}')
        local ready=$(kubectl get pod ${pod_name} -n ${NAMESPACE} -o jsonpath='{.status.containerStatuses[0].ready}')
        local restarts=$(kubectl get pod ${pod_name} -n ${NAMESPACE} -o jsonpath='{.status.containerStatuses[0].restartCount}')
        
        echo -e "  ${GREEN}✓${NC} ${expected_alert}"
        echo -e "    ${BLUE}Current: Status=${status}, Ready=${ready}, Restarts=${restarts}${NC}"
    else
        echo -e "  ${YELLOW}⚠${NC}  ${expected_alert}"
        echo -e "    ${YELLOW}Pod not found - may have been cleaned up${NC}"
    fi
}

# Cleanup function - remove all test pods
cleanup_alerts() {
    print_header "Cleanup Test Pods"
    
    check_prerequisites
    
    echo -e "${BLUE}Deleting test pods from namespace: ${NAMESPACE}${NC}\n"
    
    if ! kubectl get pods -n ${NAMESPACE} -l test=r8s-attention &> /dev/null; then
        print_warning "No test pods found to cleanup."
        exit 0
    fi
    
    # Show what will be deleted
    echo -e "${CYAN}Pods to delete:${NC}"
    kubectl get pods -n ${NAMESPACE} -l test=r8s-attention
    echo ""
    
    read -p "Delete these pods? (yes/no): " confirm
    
    if [[ "$confirm" != "yes" ]]; then
        print_warning "Cleanup cancelled."
        exit 0
    fi
    
    # Delete by label
    kubectl delete pods -n ${NAMESPACE} -l test=r8s-attention --wait=true
    
    print_success "All test pods deleted!"
    
    echo ""
    print_info "Your cluster is clean. Test pods removed:"
    echo "  • test-imagepull"
    echo "  • test-crash"
    echo "  • test-oom"
    echo "  • test-restarts"
    echo "  • test-notready"
}

# Help function
show_help() {
    cat <<EOF
${CYAN}r8s Attention Dashboard Test Script${NC}

${YELLOW}WARNING:${NC} This creates INTENTIONALLY BROKEN pods!
         Only run on TEST/DEV clusters, NEVER production!

${CYAN}Usage:${NC}
  ./test-attention-dashboard.sh setup    - Create test pods
  ./test-attention-dashboard.sh status   - Check test pod status
  ./test-attention-dashboard.sh cleanup  - Delete test pods
  ./test-attention-dashboard.sh help     - Show this help

${CYAN}What This Script Does:${NC}

${YELLOW}Creates 5 test pods to trigger alerts:${NC}
  🚫 test-imagepull    - ImagePullBackOff (invalid image)
  💀 test-crash        - CrashLoopBackOff (exits with error)
  💀 test-oom          - OOMKilled (exceeds memory limit)
  🔥 test-restarts     - High restart count (≥3 restarts)
  ⚠️  test-notready     - Not ready (1/2 containers ready)

${CYAN}Testing Workflow:${NC}
  1. Run: ./test-attention-dashboard.sh setup
  2. Wait 3 minutes (script auto-waits)
  3. Check: ./test-attention-dashboard.sh status
  4. View alerts: ./r8s
  5. Cleanup: ./test-attention-dashboard.sh cleanup

${CYAN}Environment Variables:${NC}
  TEST_NAMESPACE - Namespace to use (default: default)
  
${CYAN}Examples:${NC}
  # Use custom namespace
  TEST_NAMESPACE=test ./test-attention-dashboard.sh setup
  
  # Quick status check
  ./test-attention-dashboard.sh status
  
  # Remove all test pods
  ./test-attention-dashboard.sh cleanup

${CYAN}Critical Test:${NC}
  The test-notready pod verifies the bug fix:
  ✅ CORRECT: "1/2" triggers "Not ready" alert
  ❌ BUG (fixed): "2/2" should NOT trigger alert
  
${YELLOW}Remember:${NC} Always cleanup after testing!

EOF
}

# Main script logic
case "${1:-help}" in
    setup)
        setup_alerts
        ;;
    status)
        status_alerts
        ;;
    cleanup)
        cleanup_alerts
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
