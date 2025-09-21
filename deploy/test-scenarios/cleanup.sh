#!/bin/bash

# PDB Test Scenarios Cleanup Script
# This script removes all test scenarios and resources

set -e

echo "🧹 Cleaning up PDB Test Scenarios..."
echo "===================================="

echo ""
echo "🗑️  Removing PodDisruptionBudgets..."
kubectl delete pdb web-app-pdb --ignore-not-found=true
kubectl delete pdb api-service-pdb --ignore-not-found=true
kubectl delete pdb critical-service-pdb --ignore-not-found=true

echo ""
echo "🗑️  Removing Deployments..."
kubectl delete deployment web-app --ignore-not-found=true
kubectl delete deployment api-service --ignore-not-found=true
kubectl delete deployment worker --ignore-not-found=true
kubectl delete deployment critical-service --ignore-not-found=true

echo ""
echo "⏳ Waiting for pods to terminate..."
kubectl wait --for=delete pods -l pdb-test --timeout=120s || true

echo ""
echo "🔍 Verifying cleanup..."
echo "Remaining deployments with pdb-test label:"
kubectl get deployments -l pdb-test 2>/dev/null || echo "✅ No deployments found"

echo ""
echo "Remaining pods with pdb-test label:"
kubectl get pods -l pdb-test 2>/dev/null || echo "✅ No pods found"

echo ""
echo "Remaining PDBs:"
kubectl get pdb 2>/dev/null || echo "✅ No PDBs found"

echo ""
echo "🎉 Cleanup completed successfully!"
echo ""
echo "💡 To redeploy test scenarios, run:"
echo "   ./apply-all.sh"
