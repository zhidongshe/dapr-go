# OMS Kubernetes Deployment Guide

## Prerequisites

1. Kubernetes cluster (v1.24+)
2. kubectl configured
3. Dapr installed on cluster
4. Ingress controller (nginx recommended)
5. Docker images built and available

## Install Dapr (if not installed)

```bash
# Install Dapr CLI
curl -fsSL https://raw.githubusercontent.com/dapr/cli/master/install/install.sh | /bin/bash

# Initialize Dapr on Kubernetes
dapr init -k

# Verify Dapr installation
kubectl get pods -n dapr-system
```

## Build and Load Docker Images

### Option 1: Using local Docker images (for local K8s like minikube/kind)

```bash
# Build all images
cd /Users/shezhidong/Documents/代码库/dapr_go
docker-compose build

# For minikube
minikube image load dapr-go-order-service:latest
minikube image load dapr-go-payment-service:latest
minikube image load dapr-go-inventory-service:latest
minikube image load dapr-go-api-gateway:latest
minikube image load dapr-go-web:latest

# For kind
kind load docker-image dapr-go-order-service:latest
kind load docker-image dapr-go-payment-service:latest
kind load docker-image dapr-go-inventory-service:latest
kind load docker-image dapr-go-api-gateway:latest
kind load docker-image dapr-go-web:latest
```

### Option 2: Push to Docker Registry

```bash
# Tag images with registry prefix
docker tag dapr-go-order-service:latest your-registry/dapr-go-order-service:v1.0.0
docker tag dapr-go-payment-service:latest your-registry/dapr-go-payment-service:v1.0.0
docker tag dapr-go-inventory-service:latest your-registry/dapr-go-inventory-service:v1.0.0
docker tag dapr-go-api-gateway:latest your-registry/dapr-go-api-gateway:v1.0.0
docker tag dapr-go-web:latest your-registry/dapr-go-web:v1.0.0

# Push to registry
docker push your-registry/dapr-go-order-service:v1.0.0
docker push your-registry/dapr-go-payment-service:v1.0.0
docker push your-registry/dapr-go-inventory-service:v1.0.0
docker push your-registry/dapr-go-api-gateway:v1.0.0
docker push your-registry/dapr-go-web:v1.0.0

# Update image references in deployment files
```

## Deploy to Kubernetes

### Method 1: Using kubectl apply

```bash
cd /Users/shezhidong/Documents/代码库/dapr_go/k8s

# Create namespace
kubectl apply -f 00-namespace.yaml

# Deploy infrastructure
kubectl apply -f base/01-mysql.yaml
kubectl apply -f base/02-redis.yaml

# Wait for MySQL and Redis to be ready
kubectl wait --for=condition=ready pod -l app=mysql -n oms --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis -n oms --timeout=60s

# Deploy Dapr components
kubectl apply -f dapr/

# Deploy microservices
kubectl apply -f base/03-order-service.yaml
kubectl apply -f base/04-payment-service.yaml
kubectl apply -f base/05-inventory-service.yaml
kubectl apply -f base/06-api-gateway.yaml
kubectl apply -f base/07-web.yaml

# Deploy ingress
kubectl apply -f ingress/ingress.yaml
```

### Method 2: Using Kustomize (recommended)

```bash
cd /Users/shezhidong/Documents/代码库/dapr_go/k8s

# Deploy everything
kubectl apply -k .

# Or with kustomize binary
kustomize build . | kubectl apply -f -
```

## Verify Deployment

```bash
# Check all pods
kubectl get pods -n oms

# Check services
kubectl get svc -n oms

# Check ingress
kubectl get ingress -n oms

# Check Dapr components
kubectl get components -n oms

# View logs
kubectl logs -n oms -l app=order-service --tail=100
kubectl logs -n oms -l app=api-gateway --tail=100
```

## Access the Application

### Option 1: Port Forward

```bash
# Forward web service
kubectl port-forward -n oms svc/web 8080:80

# Access at http://localhost:8080
```

### Option 2: Using Ingress

Add to `/etc/hosts`:
```
127.0.0.1 oms.local
```

For minikube:
```bash
minikube tunnel
# Access at http://oms.local
```

For kind:
```bash
kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80
# Access at http://oms.local:8080
```

For cloud K8s, get the ingress IP:
```bash
kubectl get ingress -n oms
```

## Scale Services

```bash
# Scale order service to 3 replicas
kubectl scale deployment order-service -n oms --replicas=3

# Scale payment service
kubectl scale deployment payment-service -n oms --replicas=2
```

## Update Deployment

```bash
# After updating code and building new images
kubectl rollout restart deployment/order-service -n oms
kubectl rollout restart deployment/payment-service -n oms
kubectl rollout restart deployment/inventory-service -n oms
kubectl rollout restart deployment/api-gateway -n oms
kubectl rollout restart deployment/web -n oms
```

## Cleanup

```bash
# Delete all resources
kubectl delete -k .

# Or delete namespace (removes everything)
kubectl delete namespace oms
```

## Troubleshooting

### Pods not starting

```bash
# Check pod events
kubectl describe pod -n oms <pod-name>

# Check logs
kubectl logs -n oms <pod-name>
```

### Dapr sidecar not injecting

```bash
# Check Dapr is installed
kubectl get pods -n dapr-system

# Check pod annotations
kubectl get pod -n oms <pod-name> -o yaml | grep -A 5 annotations
```

### Database connection issues

```bash
# Check MySQL is ready
kubectl exec -n oms mysql-0 -- mysqladmin ping

# Check service DNS
kubectl exec -n oms <pod-name> -- nslookup mysql
```

## Production Considerations

1. **Security**: Change default passwords in Secrets
2. **Storage**: Use proper storage class for PVCs
3. **Resources**: Adjust resource requests/limits based on load
4. **Monitoring**: Set up Prometheus/Grafana for metrics
5. **Logging**: Configure centralized logging (ELK/Loki)
6. **TLS**: Enable HTTPS with cert-manager
7. **Backup**: Set up MySQL and Redis backups
