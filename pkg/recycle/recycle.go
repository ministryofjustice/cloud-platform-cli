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
	Node           v1.Node
	NodeName       string
	Debug          bool
	Force          bool
	TimeOut        int
	DeleteEmptyDir bool
	TimeToValidate int
	AwsProfile     string
	AwsRegion      string
	Oldest         bool
	Kubecfg        string
}

// RecycleNode takes an options argument specified by the user and
// starts the node recycle process.
func RecycleNode(options Options) error {
	clientset, err := client.GetClientset(options.Kubecfg)
	if err != nil {
		return err
	}

	client := client.New()
	client.Clientset = clientset

	cluster, err := cluster.New(client)
	if err != nil {
		return err
	}

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

	return Recycle(options, cluster, client)
}

// Recycle validates the cluster, takes a snapshot for later use and triggers the
// drain and cordon process.
func Recycle(options Options, c *cluster.Cluster, client *client.Client) (err error) {
	snapshot := c.NewSnapshot()

	if options.Oldest {
		c.Node = c.OldestNode
	}

	if options.NodeName != "" {
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
		return fmt.Errorf("node %s is not healthy: %s", c.Node.Name, err)
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
		Force:               true,
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
		return nil
	}
	if err != nil {
		return err
	}

	log.Info().Msgf("Draining node: %s", c.Node.Name)
	err = c.DrainNode(*helper)
	if apierrors.IsInvalid(err) {
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
	for range time.NewTicker(time.Duration(2) * time.Minute).C {
		err := validateSuccess(client, c, &snapshot)
		if err == nil {
			break
		}
		log.Info().Msg("Cluster validation failed, retrying in 2 minutes")
	}

	log.Info().Msgf("Finished recycling node %s", c.Node.Name)
	return nil
}

// validateSuccess returns an error if the cluster is not in a valid state.
func validateSuccess(client *client.Client, c *cluster.Cluster, snapshot *cluster.Snapshot) error {
	err := c.RefreshStatus(client)
	if err != nil {
		return fmt.Errorf("failed to refresh status: %s", err)
	}

	err = c.HealthCheck()
	if err != nil {
		return fmt.Errorf("node %s is not healthy: %s", c.Node.Name, err)
	}

	err = c.CompareNodes(snapshot)
	if err != nil {
		return fmt.Errorf("node numbers are not in the same state as before: want %v, got %v", len(snapshot.Cluster.Nodes), len(c.Nodes))
	}

	return nil
}
