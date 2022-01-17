package cluster

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Cluster struct {
	Name  string
	Nodes []v1.Node
	Pods  []v1.Pod
}

func getClusterName(client *kubernetes.Clientset) string {
	cluster, err := client.CoreV1().Nodes().Get(context.TODO(), "", metav1.GetOptions{})
	if err != nil {
		fmt.Println("error getting cluster name: ", err)
		return ""
	}
	fmt.Println(cluster.Name)

	return cluster.Name
}
