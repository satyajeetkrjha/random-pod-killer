#!/bin/bash

# PDB Test Script
# This script tests PDB functionality with your random pod killer

set -e

echo "🧪 PDB Functionality Testing"
echo "============================"

# Function to check PDB status
check_pdb_status() {
    local pdb_name=$1
    echo "📊 Checking PDB status for $pdb_name..."
    kubectl describe pdb $pdb_name | grep -E "(Min available|Max unavailable|Allowed disruptions)"
}

# Function to test pod killer with specific selector
test_pod_killer() {
    local selector=$1
    local scenario_name=$2
    local expected_behavior=$3
    
    echo ""
    echo "🎯 Testing: $scenario_name"
    echo "Selector: $selector"
    echo "Expected: $expected_behavior"
    echo "----------------------------------------"
    
    # Check if random-pod-killer binary exists
    if [ -f "../../bin/random-pod-killer" ]; then
        echo "Running: ../../bin/random-pod-killer --selector='$selector' --dry-run"
        ../../bin/random-pod-killer --selector="$selector" --dry-run || echo "❌ Test failed or binary not working"
    else
        echo "⚠️  Binary not found at ../../bin/random-pod-killer"
        echo "💡 Build it first with: make build"
        echo "🔧 Manual test command: ./random-pod-killer --selector='$selector' --dry-run"
    fi
    
    echo ""
}

echo ""
echo "🛡️  Current PDB Status:"
echo "======================="
if kubectl get pdb &>/dev/null; then
    check_pdb_status "web-app-pdb"
    echo ""
    check_pdb_status "api-service-pdb"
    echo ""
    check_pdb_status "critical-service-pdb"
    echo ""
    
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
    echo "❌ No PDBs found. Run ./deployall.sh first!"
    exit 1
fi

echo "🧪 Running PDB Tests:"
echo "===================="

# Test each scenario
test_pod_killer "app=web-app" "Web App (minAvailable=3)" "✅ Should allow 2 evictions, block 3rd"
test_pod_killer "app=api-service" "API Service (maxUnavailable=1)" "✅ Should allow 1 eviction, block 2nd"
test_pod_killer "app=worker" "Worker (no PDB)" "✅ Should allow any eviction"
test_pod_killer "app=critical-service" "Critical Service (minAvailable=4)" "❌ Should block all evictions"

echo "🎯 Advanced Test Scenarios:"
echo "==========================="

# Test by tier
test_pod_killer "tier=frontend" "Frontend Tier" "Should respect web-app PDB"
test_pod_killer "tier=backend" "Backend Tier" "Should respect api-service PDB"
test_pod_killer "tier=worker" "Worker Tier" "No PDB restrictions"
test_pod_killer "tier=critical" "Critical Tier" "Should block all evictions"

# Test by PDB type
test_pod_killer "pdb-test=min-available" "Min Available PDB" "Should respect minAvailable constraints"
test_pod_killer "pdb-test=max-unavailable" "Max Unavailable PDB" "Should respect maxUnavailable constraints"
test_pod_killer "pdb-test=no-pdb" "No PDB Protection" "Should allow all evictions"
test_pod_killer "pdb-test=strict-pdb" "Strict PDB" "Should block all evictions"

echo "🔬 Manual Testing Commands:"
echo "==========================="
echo "# Test individual scenarios:"
echo "./random-pod-killer --selector='app=web-app'"
echo "./random-pod-killer --selector='app=api-service'"
echo "./random-pod-killer --selector='app=worker'"
echo "./random-pod-killer --selector='app=critical-service'"
echo ""
echo "# Test with dry-run first:"
echo "./random-pod-killer --selector='app=web-app' --dry-run"
echo ""
echo "# Test multiple evictions in sequence:"
echo "./random-pod-killer --selector='app=web-app'  # 1st eviction"
echo "./random-pod-killer --selector='app=web-app'  # 2nd eviction"
echo "./random-pod-killer --selector='app=web-app'  # 3rd should fail"
echo ""

echo "🔧 Monitoring Commands:"
echo "======================="
echo "# Watch PDB status in real-time:"
echo "watch 'kubectl get pdb'"
echo ""
echo "# Monitor eviction events:"
echo "kubectl get events --field-selector reason=Evicted --watch"
echo ""
echo "# Check current pod status:"
echo "kubectl get pods -l pdb-test -o wide"
echo ""
echo "# Check PDB details:"
echo "kubectl describe pdb web-app-pdb"
echo "kubectl describe pdb api-service-pdb"
echo "kubectl describe pdb critical-service-pdb"
echo ""

echo "📊 Expected Test Results:"
echo "========================"
echo "✅ Web App: Can evict 2 pods (minAvailable=3)"
echo "✅ API Service: Can evict 1 pod (maxUnavailable=1)"
echo "✅ Worker: Can evict any pod (no PDB)"
echo "❌ Critical Service: Cannot evict any pod (minAvailable=4)"
echo ""

echo "💡 Development Tips:"
echo "==================="
echo "1. Start with --dry-run to test PDB detection"
echo "2. Test edge cases (evicting when at PDB limits)"
echo "3. Verify error handling when PDB would be violated"
echo "4. Test with multiple PDBs affecting same pods"
echo "5. Monitor 'kubectl get events' for eviction attempts"
echo ""

echo "🎊 PDB testing environment ready!"
echo "Build your pod killer with PDB support and test against these scenarios."
