package recycle

import (
	"context"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	mockCluster = cluster.NewMock()
	mockClient  = client.Client{
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
		Client:        &mockClient,
		Cluster:       mockCluster,
		Snapshot:      &mockSnapshot,
		Options:       &mockOptions,
		nodeToRecycle: &v1.Node{},
	}
)

func TestRecycler_defineResource(t *testing.T) {
	type fields struct {
		client        *client.Client
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
			if err := r.defineResource(); (err != nil) != tt.wantErr {
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
				ClusterName:     "cluster1",
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
				Name:   "PleaseLabelMe",
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
