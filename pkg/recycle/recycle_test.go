package recycle

import (
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRecycler_defineResource(t *testing.T) {
	mockCluster := cluster.NewMock()
	mockClient := client.Client{
		Clientset: fake.NewSimpleClientset(),
	}
	mockSnapshot := cluster.Snapshot{
		Cluster: *mockCluster,
	}
	mockOptions := Options{
		ResourceName: "node1",
	}

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
		// {
		// 	name: "define resource error",
		// 	fields: fields{
		// 		client:   &mockClient,
		// 		cluster:  mockCluster,
		// 		snapshot: &mockSnapshot,
		// 		options: &Options{
		// 			ResourceName: "",
		// 		},
		// 		nodeToRecycle: nil,
		// 	},
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Recycler{
				client:        tt.fields.client,
				cluster:       tt.fields.cluster,
				snapshot:      tt.fields.snapshot,
				options:       tt.fields.options,
				nodeToRecycle: tt.fields.nodeToRecycle,
			}
			if err := r.defineResource(); (err != nil) != tt.wantErr {
				t.Errorf("Recycler.defineResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
