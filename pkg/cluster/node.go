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
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/drain"
)

type RecycleNodeOpt struct {
	// Node is the name of the node to drain
	Node v1.Node
	// Age of the node to drain
	Age metav1.Time
	// Force drain and ignore customer uptime requests
	Force bool
	// DryRun specifies that no changes will be made to the cluster
	DryRun bool
	// Timout is the time to wait for pods to be drained
	TimeOut int
	// Oldest specifies that the oldest node should be drained
	Oldest bool
	// KubeConfigPath is the path to the kubeconfig file
	KubeConfigPath string
	// c is the kubernetes client
	c *kubernetes.Clientset
}

func RecycleNode(opt *RecycleNodeOpt) error {
	// auth to cluster
	if opt.KubeConfigPath == "" {
		opt.KubeConfigPath = getKubeConfigPath()
	}
	config, err := clientcmd.BuildConfigFromFlags("", opt.KubeConfigPath)
	if err != nil {
		return fmt.Errorf("error building config: %s", err)
	}

	// create a new kubernetes client interface
	opt.c, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error building kubernetes client: %s", err)
	}

	// ensure all nodes are in a ready state
	err = workerNodesRunning(opt.c)
	if err != nil {
		return fmt.Errorf("failed to ensure all nodes are running: %s", err)
	}

	// if oldest flag true, check the oldest node is the one we want to drain
	if opt.Oldest {
		err = opt.getOldestNode(opt.c)
		if err != nil {
			return fmt.Errorf("error getting oldest node: %s", err)
		}
	}

	// Catch empty node name
	if opt.Node.Name == "" {
		return errors.New("node name is required")
	}

	return RecycleNodeByName(opt)
}

// check for the existance of the node
// ensure we have the correct number of nodes in the cluster
// define stuck states
// ensure the label cloud-platform-recycle-nodes exists on the node
// cordon the node
// delete any stuck pods
// drain the nodes

func (o *RecycleNodeOpt) getOldestNode(client *kubernetes.Clientset) error {
	nodes := client.CoreV1().Nodes()

	// get the oldest node
	list, err := nodes.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var oldestNodeAge metav1.Time = list.Items[0].CreationTimestamp
	for node := range list.Items {
		nodeAge := list.Items[node].CreationTimestamp

		if nodeAge.Before(&oldestNodeAge) {
			oldestNodeAge = nodeAge

			o.Node = list.Items[node]
			o.Age = nodeAge
		}
	}

	return nil
}

func deleteStuckPods(client *kubernetes.Clientset, n *v1.Node) error {
	stuckStates := stuckStates()

	// Get a collection of all pods on the node
	pods, err := client.CoreV1().Pods(n.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + n.Name,
	})
	if err != nil {
		return err
	}

	// If there are no stuck pods then return
	if len(pods.Items) <= 0 {
		return nil
	}

	// if there are any pods on the node that are stuck, delete them
	for _, pod := range pods.Items {
		for _, state := range stuckStates {
			if pod.Status.Phase == state {
				fmt.Println("deleting stuck pod: ", pod.Name)
				err := client.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					return fmt.Errorf("error deleting stuck pod: %s", err)
				}
			}
		}
	}
	return nil
}

func RecycleNodeByName(opt *RecycleNodeOpt) error {
	// check the node exists
	nodes := opt.c.CoreV1().Nodes()

	node, err := nodes.Get(context.Background(), opt.Node.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// delete any stuck pods
	err = deleteStuckPods(opt.c, node)
	if err != nil {
		return err
	}

	err = drainNode(*opt, node)
	if err != nil {
		return err
	}

	err = deleteNode(opt.c, node)
	if err != nil {
		return err
	}

	return nil
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
	fmt.Println("terminating ec2 instance: ", instanceId)
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
func drainNode(opt RecycleNodeOpt, node *v1.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opt.TimeOut)*time.Second)
	defer cancel()

	if node == nil {
		return errors.New("node is nil and is not set")
	}

	helper := &drain.Helper{
		Ctx:                 ctx,
		Client:              opt.c,
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

// workerNodesRunning takes a client-go argument and checks if all nodes reported by
// `kubectl get nodes` report in a "Ready" state. If a node reports anything
// other than "Ready" the function will error. If all nodes
// report "Ready" then it'll return a nil.
//
// This acts as a validation to ensure we can start recycling nodes.
// You wouldn't want to start recycling on an unhealthy cluster.
func workerNodesRunning(client *kubernetes.Clientset) error {
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
