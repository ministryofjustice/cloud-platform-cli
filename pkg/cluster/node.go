package cluster

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/drain"
)

// get the node name
func GetNode(name string, client *kubernetes.Clientset) (v1.Node, error) {
	var node *v1.Node
	node, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return *node, fmt.Errorf("error getting node: %s", err)
	}

	return *node, nil
}

// check for the existance of the node
// ensure we have the correct number of nodes in the cluster
// define stuck states
// ensure the label cloud-platform-recycle-nodes exists on the node
// cordon the node
// delete any stuck pods
// drain the nodes

func GetOldestNode(client *kubernetes.Clientset) (v1.Node, error) {
	var oldestNode v1.Node
	nodes := client.CoreV1().Nodes()

	// get the oldest node
	list, err := nodes.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return oldestNode, err
	}

	// Starting node
	oldestNode = list.Items[0]
	for _, n := range list.Items {
		if n.CreationTimestamp.Before(&oldestNode.CreationTimestamp) {
			oldestNode = n
		}
	}

	return oldestNode, nil
}

func deleteNode(client *kubernetes.Clientset, node *v1.Node) error {
	// Delete the node from the cluster
	err := client.CoreV1().Nodes().Delete(context.Background(), node.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// Wait for the node to be deleted
	err = waitForNodeDeletion(client, node.Name, 10, 120)
	if err != nil {
		return err
	}

	err = terminateNode(client, node)
	if err != nil {
		return err
	}

	return nil
}

// terminateNode will terminate the Ec2 instance associated with the node
func terminateNode(client *kubernetes.Clientset, node *v1.Node) error {
	// Get the node's ec2 instance id
	ec2InstanceId, err := getEc2InstanceId(node)
	if err != nil {
		return err
	}

	// Terminate the ec2 instance
	err = terminateEc2Instance(ec2InstanceId)
	if err != nil {
		return err
	}

	return nil
}

// terminateEc2Instance will terminate the ec2 instance
func terminateEc2Instance(instanceId string) error {
	// Create a new ec2 client
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: "default",
		Config: aws.Config{
			Region: aws.String("eu-west-2"),
		},
	})
	if err != nil {
		return err
	}

	svc := ec2.New(sess)

	// Terminate the ec2 instance
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

// getEc2InstanceId will return the ec2 instance id associated with the node
func getEc2InstanceId(node *v1.Node) (string, error) {
	// Get the node's ec2 instance id
	s := node.Spec.ProviderID
	ec2InstanceId := strings.Split(s, "/")[4]

	if ec2InstanceId == "" {
		return "", errors.New("ec2 instance id not found")
	}

	return ec2InstanceId, nil
}

// waitForNodeDeletion waits for the node to be deleted
func waitForNodeDeletion(client *kubernetes.Clientset, name string, interval, tries int) error {
	// Wait for the node to be deleted
	for i := 0; i < tries; i++ {
		_, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return nil
}

// drainNode takes a node and performs a cordon and drain.
// If it succeeds, it returns nil.
func drainNode(client *kubernetes.Clientset, node *v1.Node, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	if node == nil {
		return errors.New("node is nil and is not set")
	}

	helper := &drain.Helper{
		Ctx:                 ctx,
		Client:              client,
		Force:               true,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Out:                 os.Stdout,
		ErrOut:              os.Stdout,
		// We want to proceed even when pods are using emptyDir volumes
		DeleteEmptyDirData: true,
		Timeout:            time.Duration(120) * time.Second,
	}

	if err := drain.RunCordonOrUncordon(helper, node, true); err != nil {
		if apierrors.IsInvalid(err) {
			return nil
		}
		return fmt.Errorf("error cordoning node: %v", err)
	}

	if err := drain.RunNodeDrain(helper, node.Name); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error draining node: %v", err)
	}

	return nil
}

// CheckNodesRunning takes a client-go argument and checks if all nodes reported by
// `kubectl get nodes` report in a "Ready" state. If a node reports anything
// other than "Ready" the function will error. If all nodes
// report "Ready" then it'll return a nil.
//
// This acts as a validation to ensure we can start recycling nodes.
// You wouldn't want to start recycling on an unhealthy cluster.
func CheckNodesRunning(client *kubernetes.Clientset) error {
	nodes := client.CoreV1().Nodes()

	list, err := nodes.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var workingNodes int
	for _, node := range list.Items {
		if node.Status.Conditions[0].Type == "Ready" && node.Status.Conditions[0].Status != "True" {
			return fmt.Errorf("node %s is not ready", node.Name)
		}
		workingNodes++
	}

	// if all nodes report a "Ready" state then approve
	if workingNodes != len(list.Items) {
		return fmt.Errorf("all nodes check failed to report a ready state. Please ensure all nodes are Ready")
	}

	return nil
}

func getKubeConfigPath() string {
	// Set the filepath of the kubeconfig file. This assumes
	// the user has either the envname set or stores their config file
	// in the default location.
	configFile := os.Getenv("KUBECONFIG")
	if configFile == "" {
		configFile = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	return configFile
}
