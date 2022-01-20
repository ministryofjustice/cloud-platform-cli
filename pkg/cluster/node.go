package cluster

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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

// getAllNodes returns a slice of all nodes in a cluster
func getAllNodes(c *client.Client) ([]v1.Node, error) {
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
	nodes, err := getAllNodes(c)
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
		if node.Status.Conditions[0].Type != "Ready" && node.Status.Conditions[0].Status == "True" {
			return fmt.Errorf("node %s is not ready", node.Name)
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

func (c *Cluster) DeleteNode(client *client.Client, awsProfile, awsRegion string, node *v1.Node) error {
	err := client.Clientset.CoreV1().Nodes().Delete(context.Background(), node.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = waitForNodeDeletion(client, *node, 10, 120)
	if err != nil {
		return err
	}

	err = terminateNode(awsProfile, awsRegion, *node)
	if err != nil {
		return err
	}

	return nil
}

// waitForNodeDeletion waits for a specified number of retries to see if the node still exists.
func waitForNodeDeletion(client *client.Client, node v1.Node, interval, retries int) error {
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
func getNode(client *client.Client, name string) (v1.Node, error) {
	node, err := client.Clientset.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return v1.Node{}, err
	}
	return *node, nil
}

// terminateNode requires an AWS profile and region, it first identifies
// the node's instance ID and then terminates it.
func terminateNode(awsProfile, awsRegion string, node v1.Node) error {
	instanceId := getEc2InstanceId(node)

	err := terminateInstance(instanceId, awsProfile, awsRegion)
	if err != nil {
		return nil
	}
	return nil
}

func getEc2InstanceId(node v1.Node) string {
	return strings.Split(node.Spec.ProviderID, "/")[4]
}

// terminateInstance creates an AwsClient and terminates the specified instance
func terminateInstance(instanceId, awsProfile, awsRegion string) error {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: awsProfile,
		Config: aws.Config{
			Region: aws.String(awsRegion),
		},
	})
	if err != nil {
		return err
	}

	svc := ec2.New(sess)
	_, err = svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceId),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// ValidateCluster allows callers to validate their cluster
// object.
func ValidateNodeHealth(c *client.Client) bool {
	nodes, err := getAllNodes(c)
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

// getClusterName returns the name of the cluster from a node
func getClusterName(nodes []v1.Node) string {
	return nodes[0].Labels["Cluster"]
}
