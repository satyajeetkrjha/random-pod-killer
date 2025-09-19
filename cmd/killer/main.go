package main

import (
	"context"
	"flag"
	"log"
	"time"

	K8s "github.com/satyajeetkrjha/random-pod-killer/pkg/k8s"
	killer "github.com/satyajeetkrjha/random-pod-killer/pkg/killer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Parse command line flags for namespace and selector
	namespace := flag.String("namespace", "default", "The namespace to watch for pods")
	selector := flag.String("selector", "", "Label selector to filter pods")
	log.Printf("Watching namespace: %s", *namespace)
	log.Printf("Using label selector: %s", *selector)

	flag.Parse()

	// Create Kubernetes client and verify connection
	clientset, err := K8s.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	log.Printf("Kubernetes client created successfully: %+v", clientset)

	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("Failed to get server version: %v", err)
	}
	log.Printf("Connected to Kubernetes cluster, version: %s", version.String())

	// Set timeout for API operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// List and display all pods in the specified namespace
	pods, err := clientset.CoreV1().Pods(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list pods in namespace %s: %v", *namespace, err)
	}
	log.Printf("Found %d pods in namespace %s", len(pods.Items), *namespace)

	// Print pod details
	for _, pod := range pods.Items {
		log.Printf("Pod Name: %s, Status: %s", pod.Name, pod.Status.Phase)
	}

	log.Println("Pod listing completed successfully")

	// List eligible pods based on the provided selector
	eligiblePods, err := killer.ListEligiblePods(clientset, *namespace, *selector, ctx)
	if err != nil {
		log.Printf("Failed to list eligible pods: %v", err)
	}
	log.Printf(" Total Eligible pods: %+v", len(eligiblePods))

	// If there are eligible pods, select one at random and delete it
	if len(eligiblePods) > 0 {
		podToDelete := killer.SelectRandomPod(eligiblePods)
		log.Printf("Selected pod %s for deletion", podToDelete.Name)

		// Kill the selected pod
		err = killer.KillPod(clientset, podToDelete, ctx)
		if err != nil {
			log.Fatalf("Failed to kill pod: %v", err)
		}

		// Wait a moment for the deletion to take effect
		time.Sleep(2 * time.Second)

		// List pods again to show the updated count
		log.Println("Listing pods after deletion...")
		podsAfter, err := clientset.CoreV1().Pods(*namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("Failed to list pods after deletion: %v", err)
		} else {
			log.Printf("Found %d pods in namespace %s after deletion", len(podsAfter.Items), *namespace)

			// Check if the deleted pod still exists
			deletedPodExists := false
			var newPods []string

			for _, pod := range podsAfter.Items {
				if pod.Name == podToDelete.Name {
					log.Printf("Original pod %s still exists with status: %s", pod.Name, pod.Status.Phase)
					deletedPodExists = true
				} else {
					// Check if this is a new pod (not in original list)
					isNew := true
					for _, originalPod := range pods.Items {
						if originalPod.Name == pod.Name {
							isNew = false
							break
						}
					}
					if isNew {
						newPods = append(newPods, pod.Name)
					}
				}
			}

			if !deletedPodExists {
				log.Printf("✓ Pod %s was successfully deleted", podToDelete.Name)
			}

			if len(newPods) > 0 {
				log.Printf("🔄 New pods created by Kubernetes: %v", newPods)
				log.Println("ℹ️  Pod count remains the same because Kubernetes recreated the deleted pod to maintain desired replica count")
			}
		}
	} else {
		log.Println("No eligible pods found to delete")
	}

}
