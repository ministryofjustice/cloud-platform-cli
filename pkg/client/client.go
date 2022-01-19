package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Client is a wrapper around the kubernetes client
type Client struct {
	Clientset kubernetes.Interface
}

// New is the constructor for the Client
func New() *Client {
	return &Client{
		Clientset: &kubernetes.Clientset{},
	}
}

// GetClientset takes the path to a kubeconfig file and returns a clientset
func GetClientset(p string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", p)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
