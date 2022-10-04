package recycle

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	mockCluster = cluster.NewMock()
	mockClient  = client.KubeClient{
		Clientset: fake.NewSimpleClientset(),
	}
	mockSnapshot = cluster.Snapshot{
		Cluster: *mockCluster,
	}
	mockOptions = Options{
		ResourceName: "node1",
		Force:        true,
		TimeOut:      360,
	}

	mockRecycler = Recycler{
		Client:   &mockClient,
		Cluster:  mockCluster,
		Snapshot: &mockSnapshot,
		Options:  &mockOptions,
	}
)

func TestRecycler_useNode(t *testing.T) {
	type fields struct {
		client        *client.KubeClient
		cluster       *cluster.Cluster
		snapshot      *cluster.Snapshot
		options       *Options
		nodeToRecycle *v1.Node
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "define resource",
			fields: fields{
				client:        &mockClient,
				cluster:       mockCluster,
				snapshot:      &mockSnapshot,
				options:       &mockOptions,
				nodeToRecycle: &v1.Node{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Recycler{
				Client:        tt.fields.client,
				Cluster:       tt.fields.cluster,
				Snapshot:      tt.fields.snapshot,
				Options:       tt.fields.options,
				nodeToRecycle: tt.fields.nodeToRecycle,
			}
			if err := r.useNode(); (err != nil) != tt.wantErr {
				t.Errorf("Recycler.defineResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecycler_getDrainHelper(t *testing.T) {
	assert := assert.New(t)

	helper := mockRecycler.getDrainHelper()

	assert.NotNil(t, helper)
}

func TestRecycler_cordonNode(t *testing.T) {
	node, err := mockClient.Clientset.CoreV1().Nodes().Create(
		context.Background(),
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
		},
		metav1.CreateOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	helper := mockRecycler.getDrainHelper()

	mockRecycler.nodeToRecycle = node

	err = mockRecycler.cordonNode(helper)
	if err != nil {
		t.Errorf("Recycler.cordonNode() error = %v", err)
	}

	assert.Equal(t, node.Labels["node-cordon"], "true")
	assert.Equal(t, mockRecycler.nodeToRecycle.Spec.Unschedulable, node.Spec.Unschedulable)
}

func TestRecycler_drainNode(t *testing.T) {
	node, err := mockClient.Clientset.CoreV1().Nodes().Create(
		context.Background(),
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "PleaseDrainMe",
				ResourceVersion: "1",
				Labels:          map[string]string{"node-cordon": "true"},
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	helper := mockRecycler.getDrainHelper()

	mockRecycler.nodeToRecycle = node

	err = mockRecycler.drainNode(helper)
	if err != nil {
		t.Errorf("Recycler.drainNode() error = %v", err)
	}

	podsOnNode, err := mockClient.Clientset.CoreV1().Pods("").List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: "node-cordon=true",
		},
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(t, node.Labels["node-drain"], "true")
	assert.Nil(t, podsOnNode.Items)
}

func TestRecycler_drainNodeWithPDB(t *testing.T) {
	replicas := int32(1)

	node, err := mockClient.Clientset.CoreV1().Nodes().Create(
		context.Background(),
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "PleaseDrainMe-02",
				ResourceVersion: "1",
				Labels:          map[string]string{"node-cordon": "true"},
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = mockClient.Clientset.PolicyV1().PodDisruptionBudgets("default").Create(
		context.Background(),
		&policyv1.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: "policy/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-pdb",
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/name": "my-app",
					},
				},
				MinAvailable: &intstr.IntOrString{IntVal: 1},
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, _ = mockClient.Clientset.CoreV1().ReplicationControllers("default").Create(
		context.Background(),
		&v1.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "rc",
				Namespace:         "default",
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
			Spec: v1.ReplicationControllerSpec{
				Replicas: &replicas,
			},
		},
		metav1.CreateOptions{})

	_, _ = mockClient.Clientset.CoreV1().Pods("default").Create(
		context.Background(),
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "bar",
				Namespace:         "default",
				CreationTimestamp: metav1.Time{Time: time.Now()},
				Labels: map[string]string{
					"app.kubernetes.io/name": "my-app",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "ReplicationController",
						Name:       "rc",
					},
				},
			},
			Spec: v1.PodSpec{
				NodeName: "PleaseDrainMe-02",
			},
		},
		metav1.CreateOptions{})

	helper := mockRecycler.getDrainHelper()

	mockRecycler.nodeToRecycle = node

	err = mockRecycler.drainNode(helper)
	if err != nil {
		t.Errorf("Recycler.drainNode() error = %v", err)
	}

	podsOnNode, err := mockClient.Clientset.CoreV1().Pods("default").List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: "node-cordon=true",
		},
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(t, node.Labels["node-drain"], "true")
	assert.Nil(t, podsOnNode.Items)
}

func TestRecycler_addLabel(t *testing.T) {
	node, err := mockClient.Clientset.CoreV1().Nodes().Create(
		context.Background(),
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "PleaseLabelMe",
				Labels: map[string]string{"node-cordon": "true"},
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockRecycler.nodeToRecycle = node

	err = mockRecycler.addLabel("test-key", "test-value")
	if err != nil {
		t.Errorf("Recycler.addLabel() = %v", err)
	}

	// Check that the label was added
	assert.Equal(t, node.Labels["test-key"], "test-value")
}

func TestRecycler_checkLabels(t *testing.T) {
	node, err := mockClient.Clientset.CoreV1().Nodes().Create(
		context.Background(),
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "CheckLabel",
				Labels: map[string]string{"node-cordon": "true"},
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockRecycler.nodeToRecycle = node

	mockRecycler.Cluster.Nodes = append(mockRecycler.Cluster.Nodes, *node)

	err = mockRecycler.checkLabels()
	assert.NotEqual(t, err, nil)
}
