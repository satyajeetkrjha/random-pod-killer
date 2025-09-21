package killer

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// PDBInfo holds information about a PodDisruptionBudget and its current status
type PDBInfo struct {
	PDB                *policyv1.PodDisruptionBudget
	CurrentHealthy     int32
	DesiredHealthy     int32
	AllowedDisruptions int32
}

// ListPDBs returns all PodDisruptionBudgets in the specified namespace
func ListPDBs(clientset *kubernetes.Clientset, namespace string, ctx context.Context) ([]policyv1.PodDisruptionBudget, error) {
	pdbList, err := clientset.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list PodDisruptionBudgets: %v", err)
	}

	log.Printf("Found %d PodDisruptionBudgets in namespace %s", len(pdbList.Items), namespace)
	for _, pdb := range pdbList.Items {
		log.Printf("  - PDB: %s, MinAvailable: %v, MaxUnavailable: %v",
			pdb.Name, pdb.Spec.MinAvailable, pdb.Spec.MaxUnavailable)
	}

	return pdbList.Items, nil
}

// GetPDBsForPod returns all PDBs that apply to the given pod
func GetPDBsForPod(pod *v1.Pod, pdbs []policyv1.PodDisruptionBudget) []policyv1.PodDisruptionBudget {
	var applicablePDBs []policyv1.PodDisruptionBudget

	for _, pdb := range pdbs {
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			log.Printf("Warning: Invalid selector in PDB %s: %v", pdb.Name, err)
			continue
		}

		if selector.Matches(labels.Set(pod.Labels)) {
			applicablePDBs = append(applicablePDBs, pdb)
			log.Printf("Pod %s is protected by PDB %s", pod.Name, pdb.Name)
		}
	}

	return applicablePDBs
}

// CalculatePDBStatus calculates the current status of a PDB given the current pods
func CalculatePDBStatus(pdb *policyv1.PodDisruptionBudget, allPods []v1.Pod) (*PDBInfo, error) {
	selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("invalid selector in PDB %s: %v", pdb.Name, err)
	}

	// Count healthy pods (Running and Ready)
	var healthyPods int32
	var totalMatchingPods int32

	for _, pod := range allPods {
		if selector.Matches(labels.Set(pod.Labels)) {
			totalMatchingPods++
			if pod.Status.Phase == v1.PodRunning {
				// Check if pod is ready
				ready := true
				for _, condition := range pod.Status.Conditions {
					if condition.Type == v1.PodReady && condition.Status != v1.ConditionTrue {
						ready = false
						break
					}
				}
				if ready {
					healthyPods++
				}
			}
		}
	}

	var desiredHealthy int32
	var allowedDisruptions int32

	if pdb.Spec.MinAvailable != nil {
		if pdb.Spec.MinAvailable.Type == intstr.Int {
			desiredHealthy = pdb.Spec.MinAvailable.IntVal
		} else {
			// Handle percentage (simplified - in production you'd want more robust percentage handling)
			desiredHealthy = totalMatchingPods // Simplified: assume 100% for percentage case
		}
		allowedDisruptions = healthyPods - desiredHealthy
	} else if pdb.Spec.MaxUnavailable != nil {
		if pdb.Spec.MaxUnavailable.Type == intstr.Int {
			allowedDisruptions = pdb.Spec.MaxUnavailable.IntVal
		} else {
			// Handle percentage (simplified)
			allowedDisruptions = 1 // Simplified: assume 1 for percentage case
		}
		desiredHealthy = healthyPods - allowedDisruptions
	}

	// Ensure allowed disruptions is not negative
	if allowedDisruptions < 0 {
		allowedDisruptions = 0
	}

	return &PDBInfo{
		PDB:                pdb,
		CurrentHealthy:     healthyPods,
		DesiredHealthy:     desiredHealthy,
		AllowedDisruptions: allowedDisruptions,
	}, nil
}

// CanEvictPod checks if a pod can be safely evicted without violating any PDBs
func CanEvictPod(pod *v1.Pod, pdbs []policyv1.PodDisruptionBudget, allPods []v1.Pod) (bool, string) {
	// First check if this is a protected pod (DaemonSet or system namespace)
	// This is a safety double-check in case protected pods somehow made it this far
	if isProtected, reason := isProtectedPod(pod); isProtected {
		return false, fmt.Sprintf("Protected pod: %s", reason)
	}

	applicablePDBs := GetPDBsForPod(pod, pdbs)

	if len(applicablePDBs) == 0 {
		log.Printf("✅ Pod %s has no PDB protection - eviction allowed", pod.Name)
		return true, "No PDB protection"
	}

	for _, pdb := range applicablePDBs {
		pdbInfo, err := CalculatePDBStatus(&pdb, allPods)
		if err != nil {
			log.Printf("❌ Error calculating PDB status for %s: %v", pdb.Name, err)
			return false, fmt.Sprintf("Error calculating PDB status: %v", err)
		}

		log.Printf("📊 PDB %s status: CurrentHealthy=%d, DesiredHealthy=%d, AllowedDisruptions=%d",
			pdb.Name, pdbInfo.CurrentHealthy, pdbInfo.DesiredHealthy, pdbInfo.AllowedDisruptions)

		if pdbInfo.AllowedDisruptions <= 0 {
			reason := fmt.Sprintf("PDB %s would be violated (CurrentHealthy=%d, DesiredHealthy=%d, AllowedDisruptions=%d)",
				pdb.Name, pdbInfo.CurrentHealthy, pdbInfo.DesiredHealthy, pdbInfo.AllowedDisruptions)
			log.Printf("❌ Pod %s eviction blocked: %s", pod.Name, reason)
			return false, reason
		}
	}

	log.Printf("✅ Pod %s can be safely evicted (all PDBs allow disruption)", pod.Name)
	return true, "All PDBs allow disruption"
}
