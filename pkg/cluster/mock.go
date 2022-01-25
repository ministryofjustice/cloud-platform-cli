package cluster

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewMock mimics the behaviour of New() but returns a mock instead of a real
// Cluster.
func NewMock() *Cluster {
	return &Cluster{
		Name: "mock",
		Nodes: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node2",
					CreationTimestamp: metav1.Time{
						Time: time.Now(),
					},
				},
			},
		},
		Pods: []v1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod2",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod3",
					Namespace: "Namespace1",
				},
				Status: v1.PodStatus{
					Phase: v1.PodFailed,
				},
			},
		},
		OldestNode: v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "node1",
				CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Hour)),
				Namespace:         "Namespace1",
			},
		},
	}
}
