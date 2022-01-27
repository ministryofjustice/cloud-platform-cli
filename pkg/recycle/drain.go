package recycle

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/drain"
)

func (r *Recycler) getDrainHelper() *drain.Helper {
	return &drain.Helper{
		Ctx:                 context.TODO(),
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

func (r *Recycler) cordonNode(helper *drain.Helper) error {
	log.Debug().Msgf("Adding 'node-cordon' label to: %s", r.nodeToRecycle.Name)
	err := r.addLabel("node-cordon", "true")
	if err != nil {
		return err
	}

	return drain.RunCordonOrUncordon(helper, r.nodeToRecycle, true)
}

func (r *Recycler) drainNode(helper *drain.Helper) error {
	log.Debug().Msgf("Adding 'node-drain' label to: %s", r.nodeToRecycle.Name)
	err := r.addLabel("node-drain", "true")
	if err != nil {
		return err
	}

	return drain.RunNodeDrain(helper, r.nodeToRecycle.Name)
}

func (r *Recycler) terminateNode() error {
	err := cluster.DeleteNode(r.Client, r.Options.AwsProfile, r.Options.AwsRegion, r.nodeToRecycle)
	if err != nil {
		return err
	}

	return nil
}

func (r *Recycler) checkLabels() error {
	for _, node := range r.Cluster.Nodes {
		if node.Labels["node-cordon"] == "true" {
			return fmt.Errorf("node %s is already cordoned, abort", node.Name)
		}

		if node.Labels["node-drain"] == "true" {
			return fmt.Errorf("node %s is already drained, abort", node.Name)
		}
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

// defineResource ensures the Recycler process is populated with the correct node to recycle.
func (r *Recycler) useNode() (err error) {
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
