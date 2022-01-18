package cluster

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// validateRecycleOptions parses the values of the object opt. It
// returns an error if any of the values are invalid.
func (opt *RecycleNodeOpt) validateRecycleOptions() (err error) {
	// Use the default kubeconfig if no kubeconfig is provided
	log.Debug().Msg("Getting kubeconfig from the user environment")
	if opt.KubeConfigPath == "" {
		opt.KubeConfigPath = getKubeConfigPath()
	}
	config, err := clientcmd.BuildConfigFromFlags("", opt.KubeConfigPath)
	if err != nil {
		return fmt.Errorf("error building config: %s", err)
	}

	opt.Client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error building kubernetes client: %s", err)
	}

	// If the node name has been provided, ensure the node exists.
	if opt.Node.Name != "" {
		log.Debug().Msg("Node name passed by the user, checking if node exists")
		node, err := GetNode(opt.Node.Name, opt.Client)
		if err != nil {
			return fmt.Errorf("error getting node: %s", err)
		}

		// Set the node object to the one provided by the user
		opt.Node.Meta = node
	}

	// if oldest flag true, use the oldest node in the cluster
	if opt.Oldest {
		log.Debug().Msg("Oldest flag passed by the user, attempting to get the oldest node")
		opt.Node.Meta, err = GetOldestNode(opt.Client)
		if err != nil {
			return fmt.Errorf("error getting oldest node: %s", err)
		}

		opt.Node.Name = opt.Node.Meta.Name
		opt.Node.Age = opt.Node.Meta.CreationTimestamp
	}

	// If no name or oldest options are requestsed, fail.
	if opt.Node.Name == "" && !opt.Oldest {
		return errors.New("no node specified")
	}

	return nil
}

// ValidateCluster validates the cluster object.
// It returns an error if any of the values are invalid and a bool value for debugging purposes.
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
