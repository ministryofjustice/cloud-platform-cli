package recycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/kubectl/pkg/drain"
)

// Options are used to configure the node recycle command
type Options struct {
	NodeName   string
	Debug      bool
	Force      bool
	TimeOut    int
	AwsProfile string
	AwsRegion  string
	Oldest     bool
	Kubecfg    string
}

// Recycler represents a single node recycle session
type Recycler struct {
	client   *client.Client
	cluster  *cluster.Cluster
	snapshot *cluster.Snapshot

	options       *Options
	nodeToRecycle *v1.Node
}

// New will construct an empty Recycler struct
// for use in a node or pod recycle
func New() *Recycler {
	return &Recycler{}
}

// Node is called by the cluster recycle-node command
// which passes arguments as a struct.
// It is used to populate the Recycler struct and instigates the
// node recycle process.
func (r *Recycler) Node(opt *Options) (err error) {
	// Pretty print log output
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: false,
	})
	// Set default log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if opt.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	r.options = opt
	clientset, err := client.GetClientset(r.options.Kubecfg)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Generating client using kubeconfig: %s", opt.Kubecfg)
	r.client = client.New()
	r.client.Clientset = clientset

	log.Debug().Msg("Getting cluster information from kubeconfig")
	r.cluster, err = cluster.New(r.client)
	if err != nil {
		return err
	}

	log.Debug().Msg("Generating cluster snapshot")
	r.snapshot = r.cluster.NewSnapshot()

	return r.RecycleNode()
}

// RecycleNode performs the heavy lifting of the node recycle process.
// It will attempt to drain, cordon and delete the node.
func (r *Recycler) RecycleNode() (err error) {
	// Does the user want to recycle the oldest node?
	if r.options.Oldest {
		r.nodeToRecycle = &r.cluster.OldestNode
		log.Debug().Msgf("Using oldest node: %s", r.nodeToRecycle.Name)
	}

	// Has the user specified a node to recycle?
	if r.options.NodeName != "" {
		log.Debug().Msgf("Node name defined as: %s. Gathering node information", r.options.NodeName)
		r.nodeToRecycle, err = r.cluster.FindNode(r.options.NodeName)
		if err != nil {
			return
		}
	}

	// Fail if no node is provided
	if r.nodeToRecycle.Name == "" {
		return errors.New("please either choose a node to recycle or use the oldest node")
	}

	log.Info().Msgf("Checking cluster: %s is in a valid state to recycle node", r.cluster.Name)
	err = r.cluster.HealthCheck()
	if err != nil {
		return err
	}

	return r.drainAndCordon()
}

// drainAndCordon utilises the kubectl drain and cordon commands to perform the node recycle process.
func (r *Recycler) drainAndCordon() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.options.TimeOut)*time.Second)
	defer cancel()

	helper := &drain.Helper{
		Ctx:                 ctx,
		Client:              r.client.Clientset,
		Force:               r.options.Force,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Out:                 log.Logger,
		ErrOut:              log.Logger,
		// We want to proceed even when pods are using emptyDir volumes
		DeleteEmptyDirData: true,
		Timeout:            time.Duration(r.options.TimeOut) * time.Second,
	}

	log.Info().Msgf("Cordoning node: %s", r.nodeToRecycle.Name)
	err = drain.RunCordonOrUncordon(helper, r.nodeToRecycle, true)
	if apierrors.IsInvalid(err) {
		log.Debug().Msgf("An API error occurred: %s - will continue", err)
		return nil
	}
	if err != nil {
		return err
	}

	log.Info().Msgf("Draining node: %s", r.nodeToRecycle.Name)
	err = drain.RunNodeDrain(helper, r.nodeToRecycle.Name)
	if apierrors.IsInvalid(err) {
		log.Debug().Msgf("An API error occurred: %s - will continue", err)
		return nil
	}
	if err != nil {
		return err
	}

	log.Info().Msgf("Deleting node: %s", r.nodeToRecycle.Name)
	err = r.cluster.DeleteNode(r.client, r.options.AwsProfile, r.options.AwsRegion, r.nodeToRecycle)
	if err != nil {
		return err
	}
	return r.postRecycleValidation()
}

// postRecycleValidation performs a final check on the cluster to ensure it is in a valid state.
func (r *Recycler) postRecycleValidation() (err error) {
	log.Info().Msg("Validating cluster health")
	for i := 0; i < 5; i++ {
		err := r.validate()
		if err == nil {
			break
		}
		log.Info().Msg("Cluster validation failed, retrying in 1 minutes")
		time.Sleep(time.Minute)
	}

	log.Info().Msgf("Finished recycling node %s", r.nodeToRecycle.Name)
	return nil
}

// validate is called by postRecycleValidation to perform the actual checks on the cluster.
// It ensures the cluster is healthy and the cluster has the same amount of nodes as when the
// node recycle process was started.
func (r *Recycler) validate() error {
	log.Debug().Msg("Refreshing cluster information")
	err := r.cluster.RefreshStatus(r.client)
	if err != nil {
		return fmt.Errorf("failed to refresh status: %s", err)
	}

	log.Debug().Msg("Performing new healther check")
	err = r.cluster.HealthCheck()
	if err != nil {
		return fmt.Errorf("node %s is not healthy: %s", r.nodeToRecycle, err)
	}

	log.Debug().Msg("Comparing old snapshot to new cluster information")
	err = r.cluster.CompareNodes(r.snapshot)
	if err != nil {
		return fmt.Errorf("node numbers are not in the same state as before: want %v, got %v", len(r.snapshot.Cluster.Nodes), len(r.cluster.Nodes))
	}

	return nil
}
