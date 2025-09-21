package killer

import (
	"context"
	"log"

	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EvictPod evicts the specified pod
func EvictPod(clientset *kubernetes.Clientset, pod *v1.Pod, ctx context.Context) error {
	log.Printf("Evicting pod %s in namespace %s", pod.Name, pod.Namespace)

	// Create an eviction object
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
	}

	// Attempt to evict the pod
	err := clientset.PolicyV1().Evictions(pod.Namespace).Evict(ctx, eviction)
	if err != nil {
		log.Printf("Failed to evict pod %s: %v", pod.Name, err)
		return err
	}

	log.Printf("Successfully evicted pod %s", pod.Name)
	return nil
}
