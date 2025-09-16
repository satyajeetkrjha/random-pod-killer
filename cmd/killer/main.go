package main

import (
	"context"
	"flag"
	"log"
	"time"

	K8s "github.com/satyajeetkrjha/random-pod-killer/pkg/k8s"
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

}
