package cluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/kubectl/pkg/drain"
)

type RecycleNodeOpt struct {
	Node    string      // name of the node to drain
	Age     metav1.Time // age of the node to drain
	Force   bool        // force drain and ignore customer uptime requests
	DryRun  bool        // don't actually drain the node
	TimeOut int         // draining a node usually takes around two minutes. If it takes longer than this, it will be cancelled.
	Oldest  bool        // drain the oldest node
}

func RecycleNode(opt *RecycleNodeOpt) error {
	// auth to cluster
	clientset, err := authenticate.CreateClientFromConfigFile(getKubeConfigPath(), "")
	if err != nil {
		return err
	}

	// ensure all nodes are in a ready state
	err = workerNodesRunning(*clientset)
	if err != nil {
		return err
	}

	// if oldest flag true, check the oldest node is the one we want to drain
	if opt.Oldest {
		err = opt.getOldestNode(*clientset)
		if err != nil {
			return err
		}
	}

	// Catch empty node name
	if opt.Node == "" {
		return fmt.Errorf("node name is required")
	}

	return RecycleNodeByName(*clientset, opt)
}

// check for the existance of the node
// ensure we have the correct number of nodes in the cluster
// define stuck states
// ensure the label cloud-platform-recycle-nodes exists on the node
// cordon the node
// delete any stuck pods
// drain the nodes

func (o *RecycleNodeOpt) getOldestNode(client kubernetes.Clientset) error {
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

			o.Node = list.Items[node].Name
			o.Age = nodeAge
		}
	}

	return nil
}

func deleteStuckPods(client kubernetes.Clientset, n *v1.Node) error {
	stuckStates := stuckStates()

	// get the pods on the node
	pods := client.CoreV1().Pods(n.Namespace)
	list, err := pods.List(context.Background(), metav1.ListOptions{
		LabelSelector: "node=" + n.Name,
	})
	if err != nil {
		return err
	}

	// if there are any pods on the node that are stuck, delete them
	if len(list.Items) > 0 {
		for _, pod := range list.Items {
			for _, state := range stuckStates {
				if pod.Status.Reason == state {
					err := pods.Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func drainNode(client kubernetes.Clientset, node *v1.Node) error {
	helper := &drain.Helper{
		// Client:              client,
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
		return fmt.Errorf("error cordoning node: %v", err)
	}

	if err := drain.RunNodeDrain(helper, node.Name); err != nil {
		return fmt.Errorf("error draining node: %v", err)
	}

	return nil
}

func RecycleNodeByName(client kubernetes.Clientset, opt *RecycleNodeOpt) error {

	// check the node exists
	nodes := client.CoreV1().Nodes()

	node, err := nodes.Get(context.Background(), opt.Node, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// delete any stuck pods
	err = deleteStuckPods(client, node)
	if err != nil {
		return err
	}
	// ensure the node is labeled
	// if !node.Labels["cloud-platform-recycle-nodes"] {
	// 	return fmt.Errorf("node %s is not labeled for recycling", opt.Node)
	// }

	// cordon the node
	err = drainNode(client, node)
	if err != nil {
		return err
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
func workerNodesRunning(client kubernetes.Clientset) error {
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

func stuckStates() []string {
	return []string{
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
