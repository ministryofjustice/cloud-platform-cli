package cluster

import (
	"context"
	"errors"
	"fmt"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/drain"
)

// Cluster represents useful values and object in a Kubernetes cluster
type Cluster struct {
	Name       string
	Nodes      []v1.Node
	Node       v1.Node
	Pods       []v1.Pod
	OldestNode v1.Node
	StuckPods  []v1.Pod
}

// Snapshot represents a snapshot of a Kubernetes cluster object
type Snapshot struct {
	Cluster Cluster
}

// New constructs a Cluster object
func New(c *client.Client) (*Cluster, error) {
	name, err := getClusterName(c)
	if err != nil {
		return nil, err
	}

	pods, err := getPods(c)
	if err != nil {
		return nil, err
	}

	nodes, err := getNodes(c)
	if err != nil {
		return nil, err
	}

	oldestNode, err := getOldestNode(c)
	if err != nil {
		return nil, err
	}

	return NewWithValues(name, pods, nodes, oldestNode), nil
}

// NewWithValues constructs a Cluster object with values
func NewWithValues(name string, pods []v1.Pod, nodes []v1.Node, oldest v1.Node) *Cluster {
	return &Cluster{
		Name:       name,
		Nodes:      nodes,
		Pods:       pods,
		OldestNode: oldest,
	}
}

// NewSnapshot constructs a Snapshot cluster object
func (c *Cluster) NewSnapshot() *Snapshot {
	return &Snapshot{
		Cluster: *c,
	}
}

// DeleteNode deletes a all pods on a node that are considered "stuck",
// essentially stuck pods are pods that are in a state that is not
// "Ready" or "Succeeded".
func (cluster *Cluster) DeleteStuckPods(c *client.Client) error {
	states := stuckStates()

	podList, err := getNodePods(c, &cluster.Node)
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

// FindNode takes a node name and returns the node object
func (cluster *Cluster) FindNode(name string) (v1.Node, error) {
	var n v1.Node
	for _, node := range cluster.Nodes {
		if node.Name == name {
			return node, nil
		}
	}

	return n, errors.New("node not found")
}

// getNodePods returns a list of pods on a node
func getNodePods(c *client.Client, n *v1.Node) (pods *v1.PodList, err error) {
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

// getClusterName returns the name of the cluster from a node
func getClusterName(c *client.Client) (string, error) {
	cluster, err := c.Clientset.CoreV1().Nodes().Get(context.TODO(), "", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return cluster.Name, nil
}

// getPods returns a slice of all pods in a cluster
func getPods(c *client.Client) ([]v1.Pod, error) {
	p := make([]v1.Pod, 0)
	pods, err := c.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	p = append(p, pods.Items...)
	return p, nil
}

// getNodes returns a slice of all nodes in a cluster
func getNodes(c *client.Client) ([]v1.Node, error) {
	n := make([]v1.Node, 0)
	nodes, err := c.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	n = append(n, nodes.Items...)
	return n, nil
}

// getOldestNode returns the oldest node in a cluster
func getOldestNode(c *client.Client) (v1.Node, error) {
	nodes, err := getNodes(c)
	if err != nil {
		return v1.Node{}, err
	}

	return oldestNode(nodes)
}

// oldestNode takes a slice of nodes and returns the oldest node
func oldestNode(nodes []v1.Node) (v1.Node, error) {
	oldestNode := nodes[0]
	for _, node := range nodes {
		if node.CreationTimestamp.Before(&oldestNode.CreationTimestamp) {
			oldestNode = node
		}
	}

	return oldestNode, nil
}

// CordonNode takes a node and runs the popular drain package to cordon the node
func (c *Cluster) CordonNode(helper drain.Helper) error {
	if c.Node.Name == "" {
		return errors.New("no node found")
	}

	return drain.RunCordonOrUncordon(&helper, &c.Node, true)
}

// DrainNode takes a node and runs the popular drain package to drain the node
func (c *Cluster) DrainNode(helper drain.Helper) error {
	return drain.RunNodeDrain(&helper, c.Node.Name)
}

// HealthCheck ensures the cluster is in a healthy state
// i.e. all nodes are running and ready
func (c *Cluster) HealthCheck() error {
	err := c.areNodesReady()
	if err != nil {
		return err
	}

	return nil
}

// areNodesReady checks if all nodes are in a ready state
func (c *Cluster) areNodesReady() error {
	for _, node := range c.Nodes {
		if node.Status.Conditions[0].Type != "Ready" && node.Status.Conditions[0].Status != "True" {
			return fmt.Errorf("node %s is not ready", node.Name)
		}
	}

	return nil
}

// CompareNodes confirms if the number of nodes in a snapshot
// is the same as the number of nodes in the cluster.
func (c *Cluster) CompareNodes(snap *Snapshot) error {
	if len(c.Nodes) != len(snap.Cluster.Nodes) {
		return fmt.Errorf("number of nodes are different")
	}

	return nil
}
