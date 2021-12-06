package cluster

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type RecycleNodeOpt struct {
	Node    string      // name of the node to drain
	Age     metav1.Time // age of the node to drain
	Force   bool        // force drain and ignore customer uptime requests
	DryRun  bool        // don't actually drain the node
	TimeOut int         // draining a node usually takes around two minutes. If it takes longer than this, it will be cancelled.
	Oldest  bool        // drain the oldest node
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

func RecycleNode(opt *RecycleNodeOpt) error {
	// auth to cluster
	clientset, err := authenticate.CreateClientFromConfigFile(getKubeConfigPath(), "")
	if err != nil {
		return err
	}

	// if oldest flag true, check the oldest node is the one we want to drain
	if opt.Oldest {
		err = opt.RecycleOldestNode(*clientset)
		if err != nil {
			return err
		}
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

func (o *RecycleNodeOpt) RecycleOldestNode(client kubernetes.Clientset) error {
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

func RecycleNodeByName(client kubernetes.Clientset, opt *RecycleNodeOpt) error {
	return nil
}

func getLocalKubeConfig() (*kubernetes.Clientset, error) {
	// Set the filepath of the kubeconfig file. This assumes
	// the user has either the envname set or stores their config file
	// in the default location.
	configFile := os.Getenv("KUBECONFIG")
	if configFile == "" {
		configFile = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	// Build the Kubernetes client using the default context (user set).
	config, err := clientcmd.BuildConfigFromFlags("", configFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
