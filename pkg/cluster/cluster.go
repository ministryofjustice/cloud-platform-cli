package cluster

import (
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
)

// Cluster represents useful values and object in a Kubernetes cluster
type Cluster struct {
	Name       string
	Nodes      []v1.Node
	Pods       []v1.Pod
	OldestNode v1.Node
	StuckPods  []v1.Pod
}

// Snapshot represents a snapshot of a Kubernetes cluster object
type Snapshot struct {
	Cluster Cluster
}

// New constructs a new Cluster object
func New(c *client.Client) (*Cluster, error) {
	pods, err := getPods(c)
	if err != nil {
		return nil, err
	}

	nodes, err := getAllNodes(c)
	if err != nil {
		return nil, err
	}

	oldestNode, err := getOldestNode(c)
	if err != nil {
		return nil, err
	}

	return NewWithValues(nodes[0].Labels["Cluster"], pods, nodes, oldestNode), nil
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

// RefreshStatus performs a value overwrite of the cluster status.
// This is useful for when the cluster is being updated.
func (c *Cluster) RefreshStatus(client *client.Client) (err error) {
	c.Nodes, err = getAllNodes(client)
	if err != nil {
		return err
	}

	c.OldestNode, err = getOldestNode(client)
	if err != nil {
		return err
	}
	return nil
}
