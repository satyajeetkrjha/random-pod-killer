#!/bin/bash

# PDB Test Scenarios Status Script
# This script shows the current status of all test scenarios

echo "📊 PDB Test Scenarios Status"
echo "============================"

# Function to show PDB details
show_pdb_details() {
    local pdb_name=$1
    local app_name=$2
    
    if kubectl get pdb $pdb_name &>/dev/null; then
        echo "📋 $pdb_name:"
        kubectl get pdb $pdb_name -o custom-columns="NAME:.metadata.name,MIN-AVAILABLE:.spec.minAvailable,MAX-UNAVAILABLE:.spec.maxUnavailable,ALLOWED-DISRUPTIONS:.status.disruptionsAllowed,CURRENT-HEALTHY:.status.currentHealthy,DESIRED-HEALTHY:.status.desiredHealthy" --no-headers
        
        # Show pod count for this app
        local pod_count=$(kubectl get pods -l app=$app_name --no-headers 2>/dev/null | wc -l)
        local running_count=$(kubectl get pods -l app=$app_name --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)
        echo "   Pods: $running_count/$pod_count running"
        echo ""
    else
        echo "❌ $pdb_name: Not found"
        echo ""
    fi
}

echo ""
echo "🚀 Deployments Status:"
echo "======================"
if kubectl get deployments -l pdb-test &>/dev/null; then
    kubectl get deployments -l pdb-test -o custom-columns="NAME:.metadata.name,READY:.status.readyReplicas,UP-TO-DATE:.status.updatedReplicas,AVAILABLE:.status.availableReplicas,DESIRED:.spec.replicas,PDB-TEST:.metadata.labels.pdb-test"
else
    echo "❌ No test deployments found"
fi

echo ""
echo "🏃 Pods Status:"
echo "==============="
if kubectl get pods -l pdb-test &>/dev/null; then
    kubectl get pods -l pdb-test -o custom-columns="NAME:.metadata.name,STATUS:.status.phase,APP:.metadata.labels.app,NODE:.spec.nodeName,AGE:.metadata.creationTimestamp" --sort-by=.metadata.labels.app
else
    echo "❌ No test pods found"
fi

echo ""
echo "🛡️  PodDisruptionBudgets Status:"
echo "================================="
if kubectl get pdb &>/dev/null; then
    show_pdb_details "web-app-pdb" "web-app"
    show_pdb_details "api-service-pdb" "api-service"
    show_pdb_details "critical-service-pdb" "critical-service"
    
    # Worker has no PDB
    local worker_pods=$(kubectl get pods -l app=worker --no-headers 2>/dev/null | wc -l)
    local worker_running=$(kubectl get pods -l app=worker --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)
    if [ $worker_pods -gt 0 ]; then
        echo "📋 worker (no PDB):"
        echo "   No PDB protection - all pods can be evicted"
        echo "   Pods: $worker_running/$worker_pods running"
        echo ""
    fi
else
    echo "❌ No PDBs found"
fi

echo ""
echo "🎯 Test Scenarios Summary:"
echo "=========================="

# Check each scenario
check_scenario() {
    local app_name=$1
    local scenario_name=$2
    local expected_behavior=$3
    
    local pod_count=$(kubectl get pods -l app=$app_name --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)
    
    if [ $pod_count -gt 0 ]; then
        echo "✅ $scenario_name: $pod_count pods running - $expected_behavior"
    else
        echo "❌ $scenario_name: No running pods found"
    fi
}

check_scenario "web-app" "Web App" "Can evict 2 pods (minAvailable=3)"
check_scenario "api-service" "API Service" "Can evict 1 pod (maxUnavailable=1)"
check_scenario "worker" "Worker" "Can evict any pod (no PDB)"
check_scenario "critical-service" "Critical Service" "Cannot evict any pod (minAvailable=4)"

echo ""
echo "🧪 Ready to Test Commands:"
echo "=========================="
echo "# Test web-app (should allow 2 evictions):"
echo "./random-pod-killer --selector='app=web-app' --dry-run"
echo ""
echo "# Test api-service (should allow 1 eviction):"
echo "./random-pod-killer --selector='app=api-service' --dry-run"
echo ""
echo "# Test worker (should allow any eviction):"
echo "./random-pod-killer --selector='app=worker' --dry-run"
echo ""
echo "# Test critical-service (should block all evictions):"
echo "./random-pod-killer --selector='app=critical-service' --dry-run"
echo ""

echo "🔧 Useful Monitoring Commands:"
echo "==============================="
echo "# Watch PDB status in real-time:"
echo "watch 'kubectl get pdb'"
echo ""
echo "# Monitor eviction events:"
echo "kubectl get events --field-selector reason=Evicted --watch"
echo ""
echo "# Check pod distribution across nodes:"
echo "kubectl get pods -l pdb-test -o wide"
echo ""
