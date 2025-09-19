package killer

import (
	"context"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KillPod deletes the specified pod
func KillPod(clientset *kubernetes.Clientset, pod *v1.Pod, ctx context.Context) error {
	log.Printf("Killing pod %s in namespace %s", pod.Name, pod.Namespace)

	err := clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Failed to kill pod %s: %v", pod.Name, err)
		return err
	}

	log.Printf("Successfully killed pod %s", pod.Name)
	return nil
}
