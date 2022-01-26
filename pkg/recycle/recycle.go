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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/drain"
)

// Options are used to configure recycle sessions.
// These options are normally passed via flags in a command line.
type Options struct {
	ResourceName string
	Debug        bool
	Force        bool
	TimeOut      int
	AwsProfile   string
	AwsRegion    string
	Oldest       bool
	KubecfgPath  string
}

// Recycler is used to store objects used in a
// recycle session.
type Recycler struct {
	Client   *client.Client
	Cluster  *cluster.Cluster
	Snapshot *cluster.Snapshot

	Options       *Options
	nodeToRecycle *v1.Node
}

// Node is called by the cluster recycle-node command
// which passes arguments as a struct.
// It is used to populate the Recycler struct and instigates the
// node recycle process.
func (r *Recycler) Node() (err error) {
	r.setupLogging()

	log.Debug().Msg("Debug enabled")
	log.Debug().Msgf("Kube config file used: %s", r.Options.KubecfgPath)
	log.Debug().Msgf("Using cluster: %s", r.Cluster.Name)
	log.Debug().Msgf("Successfully taken snapshot of cluster: %s", r.Snapshot.Cluster.Name)

	// specify which node to recycle
	err = r.defineResource()
	if err != nil {
		return err
	}

	log.Info().Msgf("Checking cluster: %s is in a valid state to recycle node", r.Cluster.Name)
	err = r.Cluster.HealthCheck()
	if err != nil {
		return err
	}

	// Check for node labels from previous recycle sessions
	for _, node := range r.Cluster.Nodes {
		if node.Labels["node-cordon"] == "true" {
			return fmt.Errorf("node %s is already cordoned, abort", node.Name)
		}

		if node.Labels["node-drain"] == "true" {
			return fmt.Errorf("node %s is already drained, abort", node.Name)
		}
	}

	return r.RecycleNode()
}

// RecycleNode performs the heavy lifting of the node recycle process.
// It will attempt to drain, cordon and delete the node.
func (r *Recycler) RecycleNode() (err error) {
	drainHelper := r.getDrainHelper()

	err = r.cordonNode(drainHelper)
	if apierrors.IsInvalid(err) {
		log.Debug().Msgf("An API error occurred: %s - will continue", err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to cordon node: %s", err)
	}

	err = r.drainNode(drainHelper)
	if apierrors.IsInvalid(err) {
		log.Debug().Msgf("An API error occurred: %s - will continue", err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to drain node: %s", err)
	}

	err = r.terminateNode()
	if err != nil {
		return err
	}

	return r.postRecycleValidation()
}

func (r *Recycler) getDrainHelper() *drain.Helper {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Options.TimeOut)*time.Second)
	defer cancel()

	return &drain.Helper{
		Ctx:                 ctx,
		Client:              r.Client.Clientset,
		Force:               r.Options.Force,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Out:                 log.Logger,
		ErrOut:              log.Logger,
		// We want to proceed even when pods are using emptyDir volumes
		DeleteEmptyDirData: true,
		Timeout:            time.Duration(r.Options.TimeOut) * time.Second,
	}
}

// Drain utilises the kubectl drain command to perform the node recycle process.
func (r *Recycler) drainNode(helper *drain.Helper) error {
	err := r.addLabel("node-drain", "true")
	if err != nil {
		return err
	}

	log.Info().Msgf("Draining node: %s", r.nodeToRecycle.Name)
	return drain.RunNodeDrain(helper, r.nodeToRecycle.Name)
}

func (r *Recycler) cordonNode(helper *drain.Helper) error {
	log.Info().Msgf("Cordoning node: %s", r.nodeToRecycle.Name)
	err := r.addLabel("node-cordon", "true")
	if err != nil {
		return err
	}

	return drain.RunCordonOrUncordon(helper, r.nodeToRecycle, true)
}

func (r *Recycler) terminateNode() error {
	log.Info().Msgf("Deleting node: %s", r.nodeToRecycle.Name)
	err := cluster.DeleteNode(r.Client, r.Options.AwsProfile, r.Options.AwsRegion, r.nodeToRecycle)
	if err != nil {
		return err
	}

	return nil
}

// RemoveLabel is called when the recycle-node command fails
func (r *Recycler) RemoveLabel(key string) error {
	if r.nodeToRecycle.Labels == nil {
		return nil
	}

	delete(r.nodeToRecycle.Labels, key)
	_, err := r.Client.Clientset.CoreV1().Nodes().Patch(context.TODO(), r.nodeToRecycle.Name, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"%s":""}}}`, key)), metav1.PatchOptions{})
	return err
}

func (r *Recycler) addLabel(key, value string) error {
	if r.nodeToRecycle.Labels == nil {
		r.nodeToRecycle.Labels = make(map[string]string)
	}

	r.nodeToRecycle.Labels[key] = value
	_, err := r.Client.Clientset.CoreV1().Nodes().Patch(context.TODO(), r.nodeToRecycle.Name, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"%s":"%s"}}}`, key, value)), metav1.PatchOptions{})
	return err
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
	err := r.Cluster.RefreshStatus(r.Client)
	if err != nil {
		return fmt.Errorf("failed to refresh status: %s", err)
	}

	log.Debug().Msg("Performing new health check")
	err = r.Cluster.HealthCheck()
	if err != nil {
		return fmt.Errorf("node %s is not healthy: %s", r.nodeToRecycle, err)
	}

	log.Debug().Msg("Comparing old snapshot to new cluster information")
	err = r.Cluster.CompareNodes(r.Snapshot)
	if err != nil {
		return fmt.Errorf("node numbers are not in the same state as before: want %v, got %v", len(r.Snapshot.Cluster.Nodes), len(r.Cluster.Nodes))
	}

	return nil
}

// defineResource ensures the Recycler process is populated with the correct node to recycle.
func (r *Recycler) defineResource() (err error) {
	if r.Options.Oldest {
		r.nodeToRecycle = &r.Cluster.OldestNode
		log.Debug().Msgf("Using oldest node: %s", r.nodeToRecycle.Name)
	}

	// Has the user specified a node to recycle?
	if r.Options.ResourceName != "" {
		log.Debug().Msgf("Node name defined as: %s. Gathering node information", r.Options.ResourceName)
		r.nodeToRecycle, err = r.Cluster.FindNode(r.Options.ResourceName)
		if err != nil {
			return
		}
	}

	// Fail if no node is provided
	if r.nodeToRecycle.Name == "" {
		return errors.New("please either choose a node to recycle or use the oldest node")
	}

	return
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