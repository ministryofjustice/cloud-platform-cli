package cluster

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

// Cluster represents useful values and object in a Kubernetes cluster
type CloudPlatformCluster struct {
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
	Cluster CloudPlatformCluster
}

// KubeClient is a wrapper around the kubernetes interface
type KubeClient struct {
	Clientset kubernetes.Interface
}

// AwsCredentials represents the AWS credentials used to connect to an AWS account.
type AwsCredentials struct {
	Session *session.Session
	Profile string
	Eks     *eks.EKS
	Ec2     *ec2.EC2
	Region  string
}

// NewKubeClient will construct a Client struct to interact with a kubernetes cluster
func NewKubeClient(p string) (*KubeClient, error) {
	clientset, err := GetClientset(p)
	if err != nil {
		return nil, err
	}

	return &KubeClient{
		Clientset: clientset,
	}, nil
}

// GetClientset takes the path to a kubeconfig file and returns a clientset
func GetClientset(p string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", p)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// NewCluster creates a new Cluster object and populates its
// fields with the values from the Kubernetes cluster in the client passed to it.
func NewCluster(c *KubeClient) (*CloudPlatformCluster, error) {
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

	return &CloudPlatformCluster{
		Name:       nodes[0].Labels["Cluster"],
		Pods:       pods,
		Nodes:      nodes,
		OldestNode: oldestNode,
		NewestNode: newestNode,
	}, nil
}

// NewSnapshot constructs a Snapshot cluster object
func (c *CloudPlatformCluster) NewSnapshot() *Snapshot {
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
func (c *CloudPlatformCluster) RefreshStatus(client *KubeClient) (err error) {
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

// AuthToCluster will authenticate to the cluster and return a kubernetes clientset. It will also write the kubeconfig
// and set the current context to the eks cluster passed to it.
func AuthToCluster(name string, eksSvc eksiface.EKSAPI, path string, awsProfile string) (*kubernetes.Clientset, error) {
	input := &eks.DescribeClusterInput{
		Name: aws.String(name),
	}
	result, err := eksSvc.DescribeCluster(input)
	if err != nil {
		log.Fatalf("Error calling DescribeCluster: %v", err)
	}

	clientset, err := newClientset(result.Cluster, path, awsProfile)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}
	return clientset, nil
}

func newClientset(cluster *eks.Cluster, path, awsProfile string) (*kubernetes.Clientset, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: aws.StringValue(cluster.Name),
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}

	config := &rest.Config{
		Host:        aws.StringValue(cluster.Endpoint),
		BearerToken: tok.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}

	if err := writeKubeConfig(cluster, path, awsProfile, tok, ca); err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func writeKubeConfig(cluster *eks.Cluster, path, profile string, tok token.Token, ca []byte) error {
	if path == "" {
		return errors.New("kubeconfig path is empty")
	}

	if ca == nil {
		return errors.New("ca is empty")
	}

	kc := api.Config{
		Clusters: map[string]*api.Cluster{
			*cluster.Name: {
				Server:                   *cluster.Endpoint,
				CertificateAuthorityData: ca,
			},
		},
		Contexts: map[string]*api.Context{
			*cluster.Name: {
				Cluster:  *cluster.Name,
				AuthInfo: *cluster.Name,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			*cluster.Name: {
				Token: tok.Token,
			},
		},
		CurrentContext: *cluster.Name,
	}

	// write kubeconfig to disk
	return clientcmd.WriteToFile(kc, path)
}

// CheckVpc asserts that the vpc is up and running. It tests the vpc state and id.
func checkVpc(name, vpcId string, svc ec2iface.EC2API) error {
	vpc, err := getVpc(name, svc)
	if err != nil {
		return fmt.Errorf("error describing vpc: %v", err)
	}

	if len(vpc.Vpcs) == 0 {
		return fmt.Errorf("no vpc found")
	}

	if vpc.Vpcs[0].VpcId != nil && *vpc.Vpcs[0].VpcId != vpcId {
		return fmt.Errorf("vpc id mismatch: %s != %s", *vpc.Vpcs[0].VpcId, vpcId)
	}

	if vpc.Vpcs[0].State != nil && *vpc.Vpcs[0].State != "available" {
		return fmt.Errorf("vpc not available: %s", *vpc.Vpcs[0].State)
	}

	return nil
}

func getVpc(name string, svc ec2iface.EC2API) (*ec2.DescribeVpcsOutput, error) {
	return svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Cluster"),
				Values: []*string{aws.String(name)},
			},
		},
	})
}

// checkCluster checks the cluster is created and exists.
func checkCluster(name string, eks eksiface.EKSAPI) error {
	cluster, err := getCluster(name, eks)
	if err != nil {
		return err
	}

	if *cluster.Status != "ACTIVE" {
		return fmt.Errorf("cluster is not active")
	}

	return nil
}

func getCluster(name string, svc eksiface.EKSAPI) (*eks.Cluster, error) {
	cluster, err := svc.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	return cluster.Cluster, nil
}
