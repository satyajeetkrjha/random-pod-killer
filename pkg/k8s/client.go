package k8s

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient() (*kubernetes.Clientset, error) {
	// Try in-cluster configuration first (when running inside a pod)
	cfg, err := rest.InClusterConfig()
	if err == nil {
		cfg.UserAgent = "random-pod-killer"
		return kubernetes.NewForConfig(cfg)
	}

	// Fall back to local kubeconfig file for development
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.ExpandEnv("$HOME/.kube/config")
	}

	cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	cfg.UserAgent = "random-pod-killer"

	return kubernetes.NewForConfig(cfg)

}
