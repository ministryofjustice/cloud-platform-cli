package cluster

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
)

// Cluster represents useful values and object in a Kubernetes cluster
type Cluster struct {
	// Name is the name of the cluster object
	Name string
	// VpcId is the ID of the VPC the cluster is in. This is usually the same
	// as the cluster name and is rarely used.
	VpcId string
	// Nodes contains a slice of nodes in the cluster. These are Kubernetes objects
	// so they contain a lot of information.
	Nodes []v1.Node
	// OldestNode is the node that was created first in the cluster. It is usually the first
	// node to be restarted when a cluster is updated.
	OldestNode v1.Node
	// NewestNode is the node that was created last in the cluster.
	NewestNode v1.Node

	// Pods contains a slice of pods in the cluster.
	Pods []v1.Pod
	// StuckPods contains a slice of pods that are stuck in the cluster. Stuck in this case refers to
	// pods that are in a non-running state for more than 5 minutes.
	StuckPods []v1.Pod

	// Namespaces contains a slice of namespaces in the cluster.
	Namespaces v1.NamespaceList

	// HealthStatus is used to define the health of a cluster. This is manually set by the caller
	// at stages of the cluster lifecycle.
	HealthStatus string
}

// Snapshot represents a snapshot of a Kubernetes cluster object
type Snapshot struct {
	Cluster Cluster
}

// AwsCredentials represents the AWS credentials used to connect to an AWS account.
type AwsCredentials struct {
	Session *session.Session
	Profile string
	Region  string
}

// NewCluster creates a new Cluster object and populates its
// fields with the values from the Kubernetes cluster in the client passed to it.
func NewCluster(c *client.KubeClient, recycleOldest bool) (*Cluster, error) {
	oldNode := v1.Node{}

	pods, err := getPods(c)
	if err != nil {
		return nil, err
	}

	nodes, err := GetAllNodes(c)
	if err != nil {
		return nil, err
	}

	if recycleOldest {
		n, err := oldestNode(nodes)
		if err != nil {
			return nil, err
		}

		oldNode = n
	}

	newestNode, err := GetNewestNode(c, nodes)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		Name:       nodes[0].Labels["Cluster"],
		Pods:       pods,
		Nodes:      nodes,
		OldestNode: oldNode,
		NewestNode: newestNode,
	}, nil
}

// NewSnapshot constructs a Snapshot cluster object
func (c *Cluster) NewSnapshot() *Snapshot {
	return &Snapshot{
		Cluster: *c,
	}
}

// NewAwsCredentials constructs and populates a new AwsCredentials object
func NewAwsCreds(region string) (*AwsCredentials, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}

	return &AwsCredentials{
		Session: sess,
		Region:  region,
	}, nil
}
