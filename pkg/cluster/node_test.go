package cluster

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func Test_compareNumberOfNodes(t *testing.T) {
	type args struct {
		cluster Cluster
		nodes   *v1.NodeList
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "SuccessfullyCompareNodes",
			args: args{
				cluster: Cluster{
					Name: "test",
					Nodes: []v1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node2",
							},
						},
					},
				},
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node2",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "FailToCompareNodes",
			args: args{
				cluster: Cluster{
					Name: "test",
					Nodes: []v1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node2",
							},
						},
					},
				},
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node2",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node3",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := compareNumberOfNodes(tt.args.cluster, tt.args.nodes); (err != nil) != tt.wantErr {
				t.Errorf("compareNumberOfNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
