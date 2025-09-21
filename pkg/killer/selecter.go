package killer

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// isDaemonSetPod checks if a pod is managed by a DaemonSet
func isDaemonSetPod(pod *v1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

// isSystemNamespace checks if a namespace is a system namespace that should be protected
func isSystemNamespace(namespace string) bool {
	systemNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"kubernetes-dashboard",
	}

	for _, sysNs := range systemNamespaces {
		if namespace == sysNs {
			return true
		}
	}
	return false
}

// isProtectedPod checks if a pod should be protected from eviction (DaemonSet or system namespace)
func isProtectedPod(pod *v1.Pod) (bool, string) {
	// Check if it's a DaemonSet pod
	if isDaemonSetPod(pod) {
		return true, "DaemonSet pod"
	}

	// Check if it's in a system namespace
	if isSystemNamespace(pod.Namespace) {
		return true, "System namespace pod"
	}

	return false, ""
}

// ListEligiblePods returns pods matching the label selector in the specified namespace
func ListEligiblePods(clientset *kubernetes.Clientset, namespace string, labelSelector string, ctx context.Context) ([]v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	var runningPods []v1.Pod
	var protectedCount int

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			// Check if pod is protected (DaemonSet or system namespace)
			if isProtected, reason := isProtectedPod(&pod); isProtected {
				log.Printf("🚫 Skipping protected pod %s: %s", pod.Name, reason)
				protectedCount++
				continue
			}
			runningPods = append(runningPods, pod)
		}
	}

	log.Printf("Found %d eligible pods in namespace %s with selector %s", len(runningPods), namespace, labelSelector)
	if protectedCount > 0 {
		log.Printf("🛡️  DaemonSet/System protection: Filtered out %d protected pods", protectedCount)
	}

	return runningPods, nil
}

// ListPDBSafePods returns pods that can be safely evicted without violating PDBs
func ListPDBSafePods(clientset *kubernetes.Clientset, namespace string, labelSelector string, ctx context.Context) ([]v1.Pod, error) {
	// First get all eligible running pods
	eligiblePods, err := ListEligiblePods(clientset, namespace, labelSelector, ctx)
	if err != nil {
		return nil, err
	}

	if len(eligiblePods) == 0 {
		log.Printf("No eligible pods found")
		return eligiblePods, nil
	}

	// Get all PDBs in the namespace
	pdbs, err := ListPDBs(clientset, namespace, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list PDBs - cannot proceed without PDB checks: %v", err)
	}

	if len(pdbs) == 0 {
		log.Printf("No PDBs found in namespace %s - all eligible pods can be evicted", namespace)
		return eligiblePods, nil
	}

	// Get all pods in namespace for PDB calculations
	// We need all pods (not just eligible ones) because PDBs may protect pods
	// that don't match our selector, and we need accurate counts for PDB math
	allPodsResult, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all pods for PDB calculations: %v", err)
	}
	allPods := allPodsResult.Items

	// Filter pods that can be safely evicted
	var safePods []v1.Pod
	for _, pod := range eligiblePods {
		canEvict, reason := CanEvictPod(&pod, pdbs, allPods)
		if canEvict {
			safePods = append(safePods, pod)
		} else {
			log.Printf("🛡️  Pod %s cannot be evicted: %s", pod.Name, reason)
		}
	}

	log.Printf("PDB-safe pods: %d out of %d eligible pods can be safely evicted", len(safePods), len(eligiblePods))
	return safePods, nil
}

func SelectRandomPod(pods []v1.Pod) *v1.Pod {
	if len(pods) == 0 {
		return nil
	}
	// Select a random pod from the list using math and random

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(pods))

	// Select the pod at the random index
	selectedPod := pods[randomIndex]

	return &selectedPod
}
