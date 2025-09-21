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

// ListEligiblePods returns pods matching the label selector in the specified namespace
func ListEligiblePods(clientset *kubernetes.Clientset, namespace string, labelSelector string, ctx context.Context) ([]v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	var runningPods []v1.Pod
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			runningPods = append(runningPods, pod)
		}
	}
	log.Printf("Found %d eligible pods in namespace %s with selector %s", len(runningPods), namespace, labelSelector)
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
