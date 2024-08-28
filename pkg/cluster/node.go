package cluster

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FindNode takes a node name and returns the node object
func (cluster *Cluster) FindNode(name string) (*v1.Node, error) {
	var n v1.Node
	for _, node := range cluster.Nodes {
		if node.Name == name {
			return &node, nil
		}
	}

	return &n, errors.New("node not found")
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

func (c *Cluster) areNodesReady() error {
	for _, node := range c.Nodes {
		for _, condition := range node.Status.Conditions {
			// There are many conditions that can be true, but we only care about
			// "Ready" - if it's not true, then there's an issue with the kublet
			if condition.Type == "Ready" && condition.Status != "True" {
				return fmt.Errorf("node %s is not ready", node.Name)
			}
		}
	}

	return nil
}

// CompareNodes confirms if the number of nodes in a snapshot
// is the same as the number of nodes in the cluster.
func (c *Cluster) CompareNodes(snap *Snapshot) (err error) {
	if len(c.Nodes) != len(snap.Cluster.Nodes) {
		return fmt.Errorf("number of nodes are different")
	}

	return nil
}

// ValidateCluster allows callers to validate their cluster
// object.
func ValidateNodeHealth(c *client.KubeClient) bool {
	nodes, err := GetAllNodes(c)
	if err != nil {
		return false
	}

	for _, node := range nodes {
		if node.Status.Conditions[0].Type != "Ready" && node.Status.Conditions[0].Status != "True" {
			return false
		}
	}

	return true
}

// GetAllNodes returns a slice of all nodes in a cluster
func GetAllNodes(c *client.KubeClient) ([]v1.Node, error) {
	n := make([]v1.Node, 0)
	nodes, err := c.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	n = append(n, nodes.Items...)
	return n, nil
}

func getOldestNode(c *client.KubeClient) (v1.Node, error) {
	nodes, err := GetAllNodes(c)
	if err != nil {
		return v1.Node{}, err
	}

	return oldestNode(nodes)
}

// oldestNode takes a slice of nodes and returns the oldest node
func oldestNode(nodes []v1.Node) (v1.Node, error) {
	oldestNode := nodes[0]
	for _, node := range nodes {
		// We don't want to recycle nodes tainted with a monitoring label
		if node.CreationTimestamp.Before(&oldestNode.CreationTimestamp) && node.Spec.Taints == nil {
			oldestNode = node
		}
	}

	// Assert the oldest node is not tainted and fail if it is
	for _, taints := range oldestNode.Spec.Taints {
		if taints.Key == "monitoring-node" {
			return v1.Node{}, errors.New("oldest node is tainted with monitoring-node and can't find a new node")
		}
	}

	return oldestNode, nil
}

// GetNodeByName takes a node name and returns the node object that has the newest creation timestamp
func GetNewestNode(c *client.KubeClient, nodes []v1.Node) (v1.Node, error) {
	newest := nodes[0]
	for _, node := range nodes {
		if node.CreationTimestamp.After(newest.CreationTimestamp.Time) {
			newest = node
		}
	}

	return newest, nil
}

// DeleteNode takes a node and authenticates to both the cluster and the AWS account.
// You must have a valid AWS credentials and an aws profile set up in your ~/.aws/credentials file.
func DeleteNode(client *client.KubeClient, awsCreds AwsCredentials, node *v1.Node) error {
	err := client.Clientset.CoreV1().Nodes().Delete(context.Background(), node.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = waitForNodeDeletion(client, *node, 10, 120)
	if err != nil {
		return err
	}

	err = terminateNode(*node, awsCreds)
	if err != nil {
		return err
	}

	return nil
}

// waitForNodeDeletion waits for a specified number of retries to see if the node still exists.
func waitForNodeDeletion(client *client.KubeClient, node v1.Node, interval, retries int) error {
	for i := 0; i < retries; i++ {
		if _, err := getNode(client, node.Name); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return nil
}

// getNode takes the name of a node and return its object
func getNode(client *client.KubeClient, name string) (v1.Node, error) {
	node, err := client.Clientset.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return v1.Node{}, err
	}
	return *node, nil
}

// terminateNode requires an AWS profile and region, it first identifies
// the node's instance ID and then terminates it.
func terminateNode(node v1.Node, awsCreds AwsCredentials) error {
	instanceId := getEc2InstanceId(node)

	err := terminateInstance(instanceId, awsCreds)
	if err != nil {
		return nil
	}
	return nil
}

func getEc2InstanceId(node v1.Node) string {
	return strings.Split(node.Spec.ProviderID, "/")[4]
}

// terminateInstance creates an AwsClient and terminates the specified instance
func terminateInstance(instanceId string, awsCreds AwsCredentials) error {
	svc := ec2.New(awsCreds.Session)
	_, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceId),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// CheckEc2InstanceTerminated takes an AWS profile and region and checks if the node passed to it
// exists in Ec2. If it does, it returns an error.
func CheckEc2InstanceTerminated(node v1.Node, awsCreds AwsCredentials) error {
	instanceId := getEc2InstanceId(node)
	svc := ec2.New(awsCreds.Session)

	output, err := svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{
			aws.String(instanceId),
		},
	})
	if err != nil {
		return err
	}

	for _, status := range output.InstanceStatuses {
		if *status.InstanceState.Name != "terminated" {
			return errors.New("node is not terminated")
		}
	}
	return nil
}
