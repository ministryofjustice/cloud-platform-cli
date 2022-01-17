package cluster

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func deleteStuckPods(client *kubernetes.Clientset, n v1.Node) error {
	stuckStates := stuckStates()

	// Get a collection of all pods on the node
	log.Debug().Msg("Checking if there are any stuck pods on the node")
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
				log.Debug().Msg("Delete stuck pod: " + pod.Name)
				err := client.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					return fmt.Errorf("error deleting stuck pod: %s", err)
				}
			}
		}
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
