package cluster

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
)

// RecycleNodeOpt provides options for the recycle-node command
type RecycleNodeOpt struct {
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
	// Client is the kubernetes client used to authenticate
	Client *kubernetes.Clientset
	// Cluster is the cluster object the recycle-node command is being run against
	Cluster Cluster
	// Node is the node object to be recycled
	Node Node
	// Debug enables debug logging
	Debug bool
}

func (opt *RecycleNodeOpt) RecycleNode() error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if opt.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if opt.DryRun {
		log.Info().Msg("dry-run mode enabled")
	}

	log.Debug().Msg("Attempting to validating recycle options: " + opt.Cluster.Name)
	err := opt.validateRecycleOptions()
	if err != nil {
		return fmt.Errorf("unable to validate recycle options: %v", err)
	}

	log.Debug().Msg("Taking a snapshot of the cluster")
	err = opt.Cluster.Snapshot(opt.Client)
	if err != nil {
		return err
	}

	log.Debug().Msg("Attempting to validating the cluster " + opt.Cluster.Name)
	working, err := opt.Cluster.ValidateCluster(opt.Client)
	if err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}
	if working {
		log.Debug().Msg("Cluster is valid")
	}

	return opt.Recycle()
}

func (opt *RecycleNodeOpt) Recycle() error {
	// validate node exists
	log.Debug().Msg("Attempting to validating node: " + opt.Node.Name)
	_, err := GetNode(opt.Node.Name, opt.Client)
	if err != nil {
		return fmt.Errorf("unable to recycle node as it cannot be found: %v", err)
	}
	log.Debug().Msg("Finished validating node: " + opt.Node.Name)

	// delete any pods that might be considered "stuck" i.e. Not "Ready"
	// TODO(jasonBirchall): move to pod package
	log.Debug().Msg("Attempting to delete pods on node: " + opt.Node.Name)
	err = deleteStuckPods(opt.Client, opt.Node.Meta)
	if err != nil {
		return err
	}
	log.Debug().Msg("Finished deleting stuck pods on node: " + opt.Node.Name)

	// Drain the node of all pods
	log.Debug().Msg("Attempting to drain node: " + opt.Node.Name)
	err = drainNode(opt.Client, &opt.Node.Meta, opt.TimeOut)
	if err != nil {
		return err
	}
	log.Debug().Msg("Finished draining node: " + opt.Node.Name)

	log.Debug().Msg("Attempting to delete node: " + opt.Node.Name)
	err = deleteNode(opt.Client, &opt.Node.Meta)
	if err != nil {
		return err
	}
	log.Debug().Msg("Finished deleting node: " + opt.Node.Name)

	// Attempt to validate the cluster for 4 minutes
	for i := 0; i < 4; i++ {
		_, err = opt.Cluster.ValidateCluster(opt.Client)
		if err == nil {
			break
		}
		log.Debug().Msg("Cluster validation failed, retrying in 1 minute")
		time.Sleep(time.Minute)
	}

	return nil
}
