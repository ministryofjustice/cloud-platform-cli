package cluster

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RecycleNodeOpt provides options for the recycle-node command
type RecycleNodeOpt struct {
	// Node is the node object to be recycled
	Node v1.Node
	// Name is passed by a user argument
	Name string
	// Age of the node to drain
	Age metav1.Time
	// Force drain and ignore customer uptime requests
	Force bool
	// DryRun specifies that no changes will be made to the cluster
	DryRun bool
	// Timout is the time to wait for pods to be drained
	TimeOut int
	// Oldest specifies that the oldest node should be drained
	Oldest bool
	// KubeConfigPath is the path to the kubeconfig file
	KubeConfigPath string
	// c is the kubernetes client
	Client *kubernetes.Clientset
}

func (opt *RecycleNodeOpt) RecycleNode() error {
	fmt.Println("Recycling node: validating options")
	// validate options
	err := opt.validateRecycleOptions()
	if err != nil {
		return err
	}
	fmt.Println(opt)

	// validate cluster
	fmt.Println("Recycling node: validating cluster")
	err = validateCluster(opt.Client)
	if err != nil {
		return err
	}
	fmt.Println(opt)

	return opt.Recycle()
}

func (opt *RecycleNodeOpt) Recycle() error {
	// validate node exists
	_, err := GetNode(opt.Name, opt.Client)
	if err != nil {
		return fmt.Errorf("unable to recycle node as it cannot be found: %v", err)
	}

	// delete any pods that might be considered "stuck" i.e. Not "Ready"
	err = deleteStuckPods(opt.Client, opt.Node)
	if err != nil {
		return err
	}

	// Drain the node of all pods
	err = drainNode(opt.Client, &opt.Node, opt.TimeOut)
	if err != nil {
		return err
	}

	err = deleteNode(opt.Client, &opt.Node)
	if err != nil {
		return err
	}

	return nil
}
