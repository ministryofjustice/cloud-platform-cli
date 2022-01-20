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

// Node takes an options argument specified by the user and
// starts the node recycle process.
func Node(options Options) error {
	// Pretty print log output
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: false,
	})
	// Set default log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if options.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	clientset, err := client.GetClientset(options.Kubecfg)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Generating client using kubeconfig: %s", options.Kubecfg)
	client := client.New()
	client.Clientset = clientset

	log.Debug().Msg("Getting cluster information")
	cluster, err := cluster.New(client)
	if err != nil {
		return err
	}

	log.Info().Msg("Starting node recycler process")
	return Recycle(options, cluster, client)
}

// Recycle validates the cluster, takes a snapshot for later use and triggers the
// drain and cordon process.
func Recycle(options Options, c *cluster.Cluster, client *client.Client) (err error) {
	log.Debug().Msg("Generating cluster snapshot")
	snapshot := c.NewSnapshot()

	if options.Oldest {
		log.Debug().Msgf("Using oldest node: %s", c.Node.Name)
		c.Node = c.OldestNode
	}

	if options.NodeName != "" {
		log.Debug().Msgf("Node name defined as: %s. Gathering node information", options.NodeName)
		c.Node, err = c.FindNode(options.NodeName)
		if err != nil {
			return err
		}
	}

	if c.Node.Name == "" {
		return errors.New("no node found")
	}

	log.Info().Msgf("Checking cluster: %s is in a valid state to recycle node", c.Name)
	err = c.HealthCheck()
	if err != nil {
		return err
	}

	return drainAndCordon(options, c, snapshot, client)
}

// drainAndCordon performs the required node cordon and drain actions, it deletes the node and instance and then validates the cluster.
func drainAndCordon(options Options, c *cluster.Cluster, snapshot *cluster.Snapshot, client *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(options.TimeOut)*time.Second)
	defer cancel()

	helper := &drain.Helper{
		Ctx:                 ctx,
		Client:              client.Clientset,
		Force:               options.Force,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Out:                 log.Logger,
		ErrOut:              log.Logger,
		// We want to proceed even when pods are using emptyDir volumes
		DeleteEmptyDirData: true,
		Timeout:            time.Duration(options.TimeOut) * time.Second,
	}

	log.Info().Msgf("Cordoning node: %s", c.Node.Name)
	err := c.CordonNode(*helper)
	if apierrors.IsInvalid(err) {
		log.Debug().Msgf("An API error occurred: %s - will continue", err)
		return nil
	}
	if err != nil {
		return err
	}

	log.Info().Msgf("Draining node: %s", c.Node.Name)
	err = c.DrainNode(*helper)
	if apierrors.IsInvalid(err) {
		log.Debug().Msgf("An API error occurred: %s - will continue", err)
		return nil
	}
	if err != nil {
		return err
	}

	log.Info().Msgf("Deleting node: %s", c.Node.Name)
	err = c.DeleteNode(client, options.AwsProfile, options.AwsRegion)
	if err != nil {
		return err
	}

	return postDrainValidation(c, client, *snapshot)
}

// postDrainValidation validates the cluster after the drain and cordon process.
func postDrainValidation(c *cluster.Cluster, client *client.Client, snapshot cluster.Snapshot) error {
	log.Info().Msg("Validating cluster health")
	for i := 0; i < 5; i++ {
		err := validateSuccess(client, c, &snapshot)
		if err == nil {
			break
		}
		log.Info().Msg("Cluster validation failed, retrying in 1 minutes")
		time.Sleep(time.Minute)
	}

	log.Info().Msgf("Finished recycling node %s", c.Node.Name)
	return nil
}

// validateSuccess returns an error if the cluster is not in a valid state.
func validateSuccess(client *client.Client, c *cluster.Cluster, snapshot *cluster.Snapshot) error {
	log.Debug().Msg("Refreshing cluster information")
	err := c.RefreshStatus(client)
	if err != nil {
		return fmt.Errorf("failed to refresh status: %s", err)
	}

	log.Debug().Msg("Performing new healther check")
	err = c.HealthCheck()
	if err != nil {
		return fmt.Errorf("node %s is not healthy: %s", c.Node.Name, err)
	}

	log.Debug().Msg("Comparing old snapshot to new cluster information")
	err = c.CompareNodes(snapshot)
	if err != nil {
		return fmt.Errorf("node numbers are not in the same state as before: want %v, got %v", len(snapshot.Cluster.Nodes), len(c.Nodes))
	}

	return nil
}
