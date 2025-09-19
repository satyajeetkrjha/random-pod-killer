package killer

import (
	"context"
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
