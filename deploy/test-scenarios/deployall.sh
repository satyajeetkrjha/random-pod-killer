#!/bin/bash

# PDB Test Scenarios Deployment Script
# This script deploys all test scenarios for PDB-aware pod killer testing

set -e

echo "🚀 Deploying PDB Test Scenarios..."
echo "=================================="

# Function to wait for deployment to be ready
wait_for_deployment() {
    local deployment_name=$1
    echo "⏳ Waiting for deployment $deployment_name to be ready..."
    kubectl wait --for=condition=available --timeout=500s deployment/$deployment_name
    echo "✅ Deployment $deployment_name is ready"
}


echo ""
echo "1️⃣  Deploying Web App (5 replicas, minAvailable=3)..."
kubectl apply -f web-app-deployment.yaml
kubectl apply -f web-app-pdb.yaml
wait_for_deployment web-app

echo ""
echo "2️⃣  Deploying API Service (3 replicas, maxUnavailable=1)..."
kubectl apply -f api-service-deployment.yaml
kubectl apply -f api-service-pdb.yaml
wait_for_deployment api-service

echo ""
echo "3️⃣  Deploying Worker (2 replicas, no PDB)..."
kubectl apply -f worker-deployment.yaml
wait_for_deployment worker

echo ""
echo "4️⃣  Deploying Critical Service (4 replicas, minAvailable=4)..."
kubectl apply -f critical-deployment.yaml
kubectl apply -f critical-pdb.yaml
wait_for_deployment critical-service

