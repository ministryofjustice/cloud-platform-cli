package recycle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kubectl/pkg/drain"
)

func (r *Recycler) getDrainHelper() *drain.Helper {
	return &drain.Helper{
		Ctx:                 context.TODO(),
		Client:              r.Client.Clientset,
		Force:               r.Options.Force,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Out:                 io.Discard,
		ErrOut:              log.Logger,
		// We want to proceed even when pods are using emptyDir volumes
		DeleteEmptyDirData: true,
		Timeout:            time.Duration(r.Options.TimeOut) * time.Second,
		OnPodDeletedOrEvicted: func(pod *v1.Pod, true bool) {
			log.Debug().Msgf("Evicting pod %s", pod.Name)
		},
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

	err = drain.RunNodeDrain(helper, r.nodeToRecycle.Name)
	if err != nil {
		pendingList, newErrs := helper.GetPodsForDeletion(r.nodeToRecycle.Name)
		if pendingList != nil {
			pods := pendingList.Pods()
			if len(pods) != 0 {
				log.Warn().Msgf("There are pending pods to drain on node %q - Retrying with disable-eviction enabled.",
					r.nodeToRecycle.Name)
				for _, pendingPod := range pods {
					log.Warn().Msgf("%s/%s\n", "pod", pendingPod.Name)
				}
				// retry deleting the pending pods with disable-eviction enabled
				// This wll delete the pods instead of evicting
				helper.DisableEviction = true
				err = drain.RunNodeDrain(helper, r.nodeToRecycle.Name)
				if err != nil {
					log.Warn().Msgf("Error while retrying to delete pending pods")
					return err
				}
			}
		} else {
			// No pending pods to delete. Some other error during draining
			if newErrs != nil {
				log.Warn().Msgf("Error while getting the list of pending pods to delete:\n%s", utilerrors.NewAggregate(newErrs))
			}
			return err
		}
	}
	return nil
}

func (r *Recycler) terminateNode() error {
	err := cluster.DeleteNode(r.Client, *r.AwsCreds, r.nodeToRecycle)
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

// useNode ensures the Recycler process is populated with the correct node to recycle.
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
