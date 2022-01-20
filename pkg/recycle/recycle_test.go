package recycle

import (
	"reflect"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	clientMock = &client.Client{
		Clientset: fake.NewSimpleClientset(),
	}

	clusterMock  = cluster.NewMock()
	snapshotMock = clusterMock.NewSnapshot()

	recyclerMock = &Recycler{
		client:   clientMock,
		cluster:  clusterMock,
		snapshot: snapshotMock,
		options: &Options{
			NodeName:   "node1",
			Debug:      false,
			Force:      false,
			TimeOut:    0,
			AwsProfile: "",
			AwsRegion:  "",
			Oldest:     true,
		},
		nodeToRecycle: &clusterMock.OldestNode,
	}
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *Recycler
	}{
		{
			name: "New",
			want: &Recycler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
