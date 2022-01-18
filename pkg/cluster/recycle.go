package cluster

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
)

// RecycleNodeOpt provides options for recycling a Kubernetes Node.
type RecycleNodeOpt struct {
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
	// AwsProfile is the aws profile to use for the aws client
	AwsProfile string
}

// RecycleNode is the main entry point for the recycle-node command
func (opt *RecycleNodeOpt) RecycleNode() error {
	// Log options
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: false,
	})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if opt.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Parse options provided by user
	err := opt.validateRecycleOptions()
	if err != nil {
		return fmt.Errorf("unable to validate recycle options: %v", err)
	}

	// Take a snapshot of the cluster before we make changes
	err = opt.Cluster.Snapshot(opt.Client)
	if err != nil {
		return fmt.Errorf("unable to snapshot cluster: %v", err)
	}

	// Ensure the nodes are ready and not in a state of "unschedulable"
	log.Info().Msg("Validating the cluster before recycling node: " + opt.Node.Name)
	working, err := opt.Cluster.ValidateCluster(opt.Client)
	if err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}
	if working {
		log.Debug().Msg("Cluster is valid")
	}

	return opt.Recycle()
}

// Recycle performs the actual node recycling duties.
// It uses the options data to determine the node to be recycled.
// It returns an error if any of the steps fail.
func (opt *RecycleNodeOpt) Recycle() error {
	log.Debug().Msg("Validating node: " + opt.Node.Name)
	_, err := GetNode(opt.Node.Name, opt.Client)
	if err != nil {
		return fmt.Errorf("unable to recycle node as it cannot be found: %v", err)
	}
	log.Debug().Msg("Finished validating node: " + opt.Node.Name)

	// delete any pods that might be considered "stuck" i.e. Not "Ready"
	log.Debug().Msg("Deleting pods on node: " + opt.Node.Name)
	err = deleteStuckPods(opt.Client, opt.Node.Meta)
	if err != nil {
		return err
	}
	log.Debug().Msg("Finished deleting stuck pods on node: " + opt.Node.Name)

	// Drain the node of all pods
	log.Debug().Msg("Draining node: " + opt.Node.Name)
	err = drainNode(opt.Client, &opt.Node.Meta, opt.TimeOut)
	if err != nil {
		return err
	}
	log.Debug().Msg("Finished draining node: " + opt.Node.Name)

	log.Debug().Msg("Deleting node: " + opt.Node.Name)
	err = deleteNode(opt.Client, &opt.Node.Meta, opt.AwsProfile)
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
		log.Info().Msg("Cluster validation failed, retrying in 1 minute")
		time.Sleep(time.Minute)
	}
	log.Info().Msg("Finished recycling node: " + opt.Node.Name)

	return nil
}
