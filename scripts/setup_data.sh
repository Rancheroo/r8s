#!/bin/bash
set -e

# Config
RANCHER_URL="https://rancher.do.4rl.io"
CLUSTER_ID="c-m-5n9lnrfl"
TOKEN="token-96csf:mbcmv6tcsb8j2qx9rlpxz9dcr4l5724lt8nsqq7s9xsd2j95g5qgmp"
KUBECONFIG_PATH="/tmp/r9s_dev_kubeconfig"

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

echo "Creating simulation data..."

# 1. Namespace: demo-frontend
echo "Setting up demo-frontend..."
kubectl create namespace demo-frontend --dry-run=client -o yaml | kubectl apply -f -

# Frontend Deployment
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend-app
  namespace: demo-frontend
  labels:
    app: frontend
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
EOF

# Frontend Service
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: frontend-svc
  namespace: demo-frontend
spec:
  selector:
    app: frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: ClusterIP
EOF

# 2. Namespace: demo-backend
echo "Setting up demo-backend..."
kubectl create namespace demo-backend --dry-run=client -o yaml | kubectl apply -f -

# Backend Deployment
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-api
  namespace: demo-backend
  labels:
    app: backend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: api
        image: nginx:alpine
        ports:
        - containerPort: 8080
EOF

# Backend Service (NodePort)
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: backend-svc
  namespace: demo-backend
spec:
  selector:
    app: backend
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
  type: NodePort
EOF

# 3. Namespace: demo-worker
echo "Setting up demo-worker..."
kubectl create namespace demo-worker --dry-run=client -o yaml | kubectl apply -f -

# Worker Deployment (No Service)
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: data-processor
  namespace: demo-worker
  labels:
    app: worker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
    spec:
      containers:
      - name: worker
        image: busybox
        command: ["/bin/sh", "-c", "while true; do echo processing; sleep 10; done"]
EOF

echo "Done! Resources created."
rm $KUBECONFIG_PATH
