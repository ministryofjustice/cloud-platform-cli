package cluster

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
)

// Cluster represents useful values and object in a Kubernetes cluster
type Cluster struct {
	Name       string
	Nodes      []v1.Node
	Pods       []v1.Pod
	OldestNode v1.Node
	NewestNode v1.Node
	StuckPods  []v1.Pod
	Spec       NewClusterOptions
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

type NewClusterOptions struct {
	Name          string
	ClusterSuffix string

	NodeCount int
	VpcName   string

	GitCryptUnlock bool

	Kubeconfig string

	MaxNameLength int `default:"12"`
	TimeOut       int
	Debug         bool

	Auth0          AuthOpts
	AwsCredentials AwsCredentials
}

type AuthOpts struct {
	Domain       string
	ClientId     string
	ClientSecret string
}

// NewCluster creates a new Cluster object and populates its
// fields with the values from the Kubernetes cluster in the client passed to it.
func NewCluster(c *client.Client) (*Cluster, error) {
	pods, err := getPods(c)
	if err != nil {
		return nil, err
	}

	nodes, err := GetAllNodes(c)
	if err != nil {
		return nil, err
	}

	oldestNode, err := getOldestNode(c)
	if err != nil {
		return nil, err
	}

	newestNode, err := GetNewestNode(c, nodes)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		Name:       nodes[0].Labels["Cluster"],
		Pods:       pods,
		Nodes:      nodes,
		OldestNode: oldestNode,
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

// RefreshStatus performs a value overwrite of the cluster status.
// This is useful for when the cluster is being updated.
func (c *Cluster) RefreshStatus(client *client.Client) (err error) {
	c.Nodes, err = GetAllNodes(client)
	if err != nil {
		return err
	}

	c.OldestNode, err = getOldestNode(client)
	if err != nil {
		return err
	}
	return nil
}

// Create creates a new Kubernetes cluster using the options passed to it.
func Create(opts *NewClusterOptions) error {
	// create vpc
	err := CreateVpc(opts)
	if err != nil {
		return err
	}

	// create kubernetes cluster
	err = CreateCluster(opts)
	if err != nil {
		return err
	}

	// install components into kubernetes cluster
	err = InstallComponents(opts)
	if err != nil {
		return err
	}

	// perform health check on the cluster
	err = HealthCheck(opts)
	if err != nil {
		return err
	}

	return nil
}

// CreateVpc creates a new VPC in AWS.
func CreateVpc(opts *NewClusterOptions) error {
	return nil
}

// CreateCluster creates a new Kubernetes cluster in AWS.
func CreateCluster(opts *NewClusterOptions) error {
	return nil
}

// InstallComponents installs components into the Kubernetes cluster.
func InstallComponents(opts *NewClusterOptions) error {
	return nil
}

// HealthCheck performs a health check on the Kubernetes cluster.
func HealthCheck(opts *NewClusterOptions) error {
	return nil
}
