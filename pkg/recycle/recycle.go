package recycle

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// Options are used to configure recycle sessions.
// These options are normally passed via flags in a command line.
type Options struct {
	// Resource is the resource to be recycled.
	// This can be a node, pod, service, deployment, etc.
	ResourceName string
	// Debug enables the debug messages in stdout and stderr.
	Debug bool
	// Force enables the force flag for the kubectl drain command.
	Force bool
	// Timeout is the timeout for the kubectl drain command.
	TimeOut int
	// Oldest suggests using the oldest resource in a list, such as a node.
	Oldest bool
	// KubecfgPath is the path to the kubeconfig file to use for the Kubernetes client.
	KubecfgPath string
	// IgnoreLabel indicates that the node-cordon or node-drain labels should not be checked.
	IgnoreLabel bool
	// just cordon and drain the nodes and don't bring up new ones
	DrainOnly bool

	// AwsProfile is the AWS profile to use for resource termination.
	AwsProfile string
	// AwsRegion is the AWS region to use for resource termination.
	AwsRegion string
}

// Recycler is used to store objects used in a recycle session.
type Recycler struct {
	// Client represents the kubernetes client.
	Client *client.KubeClient
	// Cluster is the cluster object obtained from the current context.
	Cluster *cluster.Cluster
	// Snapshot is the snapshot of the cluster. Used for comparison of cluster state.
	Snapshot *cluster.Snapshot
	// AwsCreds is the AWS credentials to use for resource termination.
	AwsCreds *cluster.AwsCredentials

	// Options are used to configure the recycle session.
	Options *Options
	// nodeToRecycle is the Kubernetes node object to be recycled.
	nodeToRecycle *v1.Node
}

// Node is the entry point for the node recycle process.
// It will attempt to find the node to recycle, check the cluster is ready for a node recycle
// and then check for missing labels. It returns an error if the called RecycleNode method fails.
func (r *Recycler) Node() (err error) {
	r.setupLogging()

	log.Debug().Msg("Debug enabled")
	log.Debug().Msgf("Kube config file used: %s", r.Options.KubecfgPath)
	log.Debug().Msgf("Using cluster: %s", r.Cluster.Name)
	log.Debug().Msgf("Successfully taken snapshot of cluster: %s", r.Snapshot.Cluster.Name)

	// Populate the nodeToRecycle object
	err = r.useNode()
	if err != nil {
		return err
	}

	log.Debug().Msg("Checking AWS credentials work")
	err = r.checkAwsCreds()
	if err != nil {
		return fmt.Errorf("unable to validate AWS credentials: %s", err)
	}

	log.Info().Msgf("Checking cluster: %s is in a valid state to recycle node", r.Cluster.Name)
	err = r.Cluster.HealthCheck()
	if err != nil {
		return err
	}

	// Check for node labels from previous recycle sessions
	log.Debug().Msg("Checking if the node-cordon or node-drain labels exist")
	if !r.Options.IgnoreLabel {
		err = r.checkLabels()
		if err != nil {
			return err
		}
	}

	return r.recycleNode()
}

// recycleNode performs the heavy lifting of the node recycle process.
// It will create a drain helper for further usage, utilise the kubectl
// drain package to first cordon the selected node and then drain all pods.
// It then calls the cluster package and terminates said node from AWS Ec2.
func (r *Recycler) recycleNode() (err error) {
	drainHelper := r.getDrainHelper()

	log.Info().Msgf("Cordoning node: %s", r.nodeToRecycle.Name)
	err = r.cordonNode(drainHelper)
	if err != nil {
		return fmt.Errorf("failed to cordon node: %s", err)
	}

	log.Info().Msgf("Draining node: %s", r.nodeToRecycle.Name)
	err = r.drainNode(drainHelper)
	if err != nil {
		return fmt.Errorf("unable to drain node: %s", err)
	}

	if r.Options.DrainOnly {
		log.Info().Msgf("Running in DrainOnly mode, node: %s is cordoned and drained but not terminated", r.nodeToRecycle.Name)
		return nil
	}

	log.Info().Msgf("Terminate node: %s", r.nodeToRecycle.Name)
	err = r.terminateNode()
	if err != nil {
		return err
	}

	return r.postNodeCheck()
}

func (r *Recycler) postNodeCheck() (err error) {
	log.Info().Msg("Validating cluster health")
	// Grab the node being terminated as this can get muddled after the validation.
	// It generally takes about 4 minutes for the node to recycle, but to be sure we'll wait for 7.
	for i := 0; i < 7; i++ {
		err := r.validate()
		if err == nil {
			log.Info().Msgf("Finished recycling node %s", r.nodeToRecycle.Name)
			log.Info().Msgf("New node created: %s", r.Cluster.NewestNode.Name)
			return nil
		}

		log.Debug().Msg(err.Error())
		log.Warn().Msg("Cluster validation failed, retrying in 1 minutes")
		time.Sleep(time.Minute)
	}

	return errors.New("failed to validate cluster after 7 attempts")
}

// validate is called to perform checks on the cluster.
// It ensures the cluster is healthy and the cluster has the same amount of nodes as when the
// node recycle process was started.
func (r *Recycler) validate() (err error) {
	log.Debug().Msg("Refreshing cluster information")
	r.Cluster.Nodes, err = cluster.GetAllNodes(r.Client)
	if err != nil {
		return fmt.Errorf("failed to refresh status: %s", err)
	}

	log.Debug().Msgf("Checking ec2 instance: %s has terminated", r.nodeToRecycle.Name)
	err = cluster.CheckEc2InstanceTerminated(*r.nodeToRecycle, *r.AwsCreds)
	if err != nil {
		return fmt.Errorf("ec2 instance has not terminated: %s", err)
	}

	log.Debug().Msg("Comparing old snapshot to new cluster information")
	err = r.Cluster.CompareNodes(r.Snapshot)
	if err != nil {
		return fmt.Errorf("node numbers are not in the same state as before: want %v, got %v", len(r.Snapshot.Cluster.Nodes), len(r.Cluster.Nodes))
	}

	log.Debug().Msg("Refreshing cluster information")
	r.Cluster.NewestNode, err = cluster.GetNewestNode(r.Client, r.Cluster.Nodes)
	if err != nil {
		return fmt.Errorf("failed to get new node: %s", err)
	}

	log.Debug().Msg("Performing new health check")
	err = r.Cluster.HealthCheck()
	if err != nil {
		return err
	}

	return nil
}

func (r *Recycler) setupLogging() {
	// Pretty print log output
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: false,
	})
	// Set default log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if r.Options.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func (r *Recycler) checkAwsCreds() error {
	_, err := r.AwsCreds.Session.Config.Credentials.Get()
	return err
}
