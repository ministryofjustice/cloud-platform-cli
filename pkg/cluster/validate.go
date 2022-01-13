package cluster

import (
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (opt *RecycleNodeOpt) validateRecycleOptions() error {
	// auth to cluster
	if opt.KubeConfigPath == "" {
		opt.KubeConfigPath = getKubeConfigPath()
	}
	config, err := clientcmd.BuildConfigFromFlags("", opt.KubeConfigPath)
	if err != nil {
		return fmt.Errorf("error building config: %s", err)
	}

	// create a new kubernetes client interface
	opt.Client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error building kubernetes client: %s", err)
	}

	if opt.Name != "" {
		node, err := GetNode(opt.Name, opt.Client)
		if err != nil {
			return err
		}

		opt.Node = node
	}

	// if oldest flag true, check the oldest node is the one we want to drain
	if opt.Oldest {
		opt.Node, err = GetOldestNode(opt.Client)
		if err != nil {
			return err
		}

		opt.Name = opt.Node.Name
		opt.Age = opt.Node.CreationTimestamp
	}

	if opt.Name == "" && !opt.Oldest {
		return errors.New("no node specified")
	}

	return nil
}

func validateCluster(client *kubernetes.Clientset) error {
	err := CheckNodesRunning(client)
	if err != nil {
		return fmt.Errorf("failed to ensure all nodes are running: %s", err)
	}

	return nil
}
