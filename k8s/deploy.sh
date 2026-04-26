#!/bin/bash

set -e

echo "🚀 OMS Kubernetes Deployment Script"
echo "===================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl not found. Please install kubectl.${NC}"
    exit 1
fi

if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Kubernetes cluster not accessible.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Prerequisites met${NC}"

# Navigate to script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Function to wait for pod
wait_for_pod() {
    local app=$1
    local namespace=$2
    echo -e "${YELLOW}Waiting for $app to be ready...${NC}"
    kubectl wait --for=condition=ready pod -l app=$app -n $namespace --timeout=120s || {
        echo -e "${RED}Timeout waiting for $app${NC}"
        return 1
    }
    echo -e "${GREEN}✓ $app is ready${NC}"
}

# Deploy
echo -e "\n${YELLOW}Step 1: Creating namespace...${NC}"
kubectl apply -f 00-namespace.yaml
echo -e "${GREEN}✓ Namespace created${NC}"

echo -e "\n${YELLOW}Step 2: Deploying MySQL...${NC}"
kubectl apply -f base/01-mysql.yaml
wait_for_pod mysql oms
echo -e "${GREEN}✓ MySQL deployed${NC}"

echo -e "\n${YELLOW}Step 3: Deploying Redis...${NC}"
kubectl apply -f base/02-redis.yaml
wait_for_pod redis oms
echo -e "${GREEN}✓ Redis deployed${NC}"

echo -e "\n${YELLOW}Step 4: Deploying Dapr components...${NC}"
kubectl apply -f dapr/
echo -e "${GREEN}✓ Dapr components deployed${NC}"

echo -e "\n${YELLOW}Step 5: Deploying microservices...${NC}"
kubectl apply -f base/03-order-service.yaml
kubectl apply -f base/04-payment-service.yaml
kubectl apply -f base/05-inventory-service.yaml
kubectl apply -f base/06-api-gateway.yaml
kubectl apply -f base/07-web.yaml

echo -e "${YELLOW}Waiting for services to be ready...${NC}"
sleep 10
wait_for_pod order-service oms
wait_for_pod payment-service oms
wait_for_pod inventory-service oms
wait_for_pod api-gateway oms
wait_for_pod web oms
echo -e "${GREEN}✓ All microservices deployed${NC}"

echo -e "\n${YELLOW}Step 6: Deploying Ingress...${NC}"
kubectl apply -f ingress/ingress.yaml
echo -e "${GREEN}✓ Ingress deployed${NC}"

echo -e "\n${GREEN}====================================${NC}"
echo -e "${GREEN}✓ Deployment Complete!${NC}"
echo -e "${GREEN}====================================${NC}"

echo -e "\n${YELLOW}Useful commands:${NC}"
echo "  kubectl get pods -n oms"
echo "  kubectl get svc -n oms"
echo "  kubectl logs -n oms -l app=order-service"
echo ""
echo -e "${YELLOW}To access the application:${NC}"
echo "  kubectl port-forward -n oms svc/web 8080:80"
echo "  Then open http://localhost:8080"
