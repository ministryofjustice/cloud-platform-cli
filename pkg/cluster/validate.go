package cluster

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (opt *RecycleNodeOpt) validateRecycleOptions() error {
	// auth to cluster
	log.Debug().Msg("Getting kubeconfig from the user environment")
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
		log.Debug().Msg("Node name passed by the user, checking if node exists")
		node, err := GetNode(opt.Node.Name, opt.Client)
		if err != nil {
			return fmt.Errorf("error getting node: %s", err)
		}

		opt.Node.Meta = node
	}

	// if oldest flag true, check the oldest node is the one we want to drain
	if opt.Oldest {
		log.Debug().Msg("Oldest flag passed by the user, attempting to get the oldest node")
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

func (cluster Cluster) ValidateCluster(client *kubernetes.Clientset) (bool, error) {
	log.Debug().Msg("Checking if all nodes in the cluster are ready")
	nodes, err := RunningNodes(client)
	if err != nil {
		return false, fmt.Errorf("failed to ensure all nodes are running: %s", err)
	}

	log.Debug().Msg("Comparing the number of nodes currently with the snapshot")
	err = compareNumberOfNodes(cluster, nodes)
	if err != nil {
		return false, fmt.Errorf("failed to compare all nodes in the cluster: %s", err)
	}

	return true, nil
}
