package cluster

import (
	"context"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteStuckPods deletes a all pods on a node that are considered "stuck",
// essentially stuck pods are pods that are in a state that is not
// "Ready" or "Succeeded".
func (cluster *Cluster) DeleteStuckPods(c *client.KubeClient, node *v1.Node) error {
	states := stuckStates()

	podList, err := getNodePods(c, node)
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		return nil
	}

	for _, pod := range podList.Items {
		for _, state := range states {
			if pod.Status.Phase == state {
				err := c.Clientset.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// getStuckPods returns a slice of all pods in a cluster that are considered "stuck"
func (cluster *Cluster) GetStuckPods(c *client.KubeClient) error {
	p := make([]v1.Pod, 0)
	pods, err := getPods(c)
	if err != nil {
		return err
	}

	states := stuckStates()
	for _, pod := range pods {
		for _, state := range states {
			if pod.Status.Phase == state {
				p = append(p, pod)
			}
		}
	}

	return nil
}

// getNodePods returns a list of pods on a node
func getNodePods(c *client.KubeClient, n *v1.Node) (pods *v1.PodList, err error) {
	pods, err = c.Clientset.CoreV1().Pods(n.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + n.Name,
	})
	if err != nil {
		return
	}
	return
}

// stuckStates returns a list of pod states that are considered "stuck"
func stuckStates() []v1.PodPhase {
	return []v1.PodPhase{
		"Pending",
		"Scheduling",
		"Unschedulable",
		"ImagePullBackOff",
		"CrashLoopBackOff",
		"Unknown",
	}
}

// getPods returns a slice of all pods in a cluster
func getPods(c *client.KubeClient) ([]v1.Pod, error) {
	p := make([]v1.Pod, 0)
	pods, err := c.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	p = append(p, pods.Items...)
	return p, nil
}
