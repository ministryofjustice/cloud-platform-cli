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

	if opt.Node.Name != "" {
		node, err := GetNode(opt.Node.Name, opt.Client)
		if err != nil {
			return err
		}

		opt.Node.Meta = node
	}

	// if oldest flag true, check the oldest node is the one we want to drain
	if opt.Oldest {
		opt.Node.Meta, err = GetOldestNode(opt.Client)
		if err != nil {
			return err
		}

		opt.Node.Name = opt.Node.Meta.Name
		opt.Node.Age = opt.Node.Meta.CreationTimestamp
	}

	if opt.Node.Name == "" && !opt.Oldest {
		return errors.New("no node specified")
	}

	return nil
}

func (cluster Cluster) ValidateCluster(client *kubernetes.Clientset) error {
	nodes, err := RunningNodes(client)
	if err != nil {
		return fmt.Errorf("failed to ensure all nodes are running: %s", err)
	}

	err = compareNumberOfNodes(cluster, nodes)
	if err != nil {
		return err
	}

	return nil
}
