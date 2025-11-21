#!/bin/bash
set -e

# Config
RANCHER_URL="https://rancher.do.4rl.io"
CLUSTER_ID="c-m-5n9lnrfl"
TOKEN="token-96csf:mbcmv6tcsb8j2qx9rlpxz9dcr4l5724lt8nsqq7s9xsd2j95g5qgmp"
KUBECONFIG_PATH="/tmp/r9s_default_kubeconfig"

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

echo "Creating basic web app in 'default' namespace (Default Project)..."

# Basic Deployment in default namespace
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: basic-web
  namespace: default
  labels:
    app: basic-web
spec:
  replicas: 1
  selector:
    matchLabels:
      app: basic-web
  template:
    metadata:
      labels:
        app: basic-web
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
EOF

# Basic Service in default namespace
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: basic-web-svc
  namespace: default
spec:
  selector:
    app: basic-web
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: ClusterIP
EOF

echo "Done! Basic resources created in default namespace."
rm $KUBECONFIG_PATH
