#!/bin/bash
set -e

# Config
RANCHER_URL="https://rancher.do.4rl.io"
CLUSTER_ID="c-m-5n9lnrfl"
TOKEN="token-96csf:mbcmv6tcsb8j2qx9rlpxz9dcr4l5724lt8nsqq7s9xsd2j95g5qgmp"
KUBECONFIG_PATH="/tmp/r9s_empty_kubeconfig"

# Create Kubeconfig
cat <<EOF > $KUBECONFIG_PATH
apiVersion: v1
kind: Config
clusters:
- name: r9s-dev
  cluster:
    server: ${RANCHER_URL}/k8s/clusters/${CLUSTER_ID}
    insecure-skip-tls-verify: true
users:
- name: r9s-user
  user:
    token: ${TOKEN}
contexts:
- name: default
  context:
    cluster: r9s-dev
    user: r9s-user
current-context: default
EOF

export KUBECONFIG=$KUBECONFIG_PATH

echo "Creating empty namespace 'empty-ns'..."
kubectl create namespace empty-ns --dry-run=client -o yaml | kubectl apply -f -

echo "Done! Empty namespace created."
rm $KUBECONFIG_PATH
