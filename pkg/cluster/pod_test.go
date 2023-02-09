package cluster

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCluster_DeleteStuckPods(t *testing.T) {
	type args struct {
		c    *KubeClient
		node *v1.Node
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "DeleteStuckPods",
			args: args{
				c: &mockClient,
				node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := mockCluster
			if err := cluster.DeleteStuckPods(tt.args.c, tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("Cluster.DeleteStuckPods() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getNodePods(t *testing.T) {
	type args struct {
		c *KubeClient
		n *v1.Node
	}
	tests := []struct {
		name     string
		args     args
		wantPods *v1.PodList
		wantErr  bool
	}{
		{
			name: "getNodePods",
			args: args{
				c: &mockClient,
				n: &mockCluster.OldestNode,
			},
			wantPods: &v1.PodList{},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPods, err := getNodePods(tt.args.c, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodePods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPods, tt.wantPods) {
				t.Errorf("getNodePods() = %v,/n/n want %v", gotPods, tt.wantPods)
			}
		})
	}
}

func Test_getPods(t *testing.T) {
	type args struct {
		c *KubeClient
	}
	tests := []struct {
		name    string
		args    args
		want    []v1.Pod
		wantErr bool
	}{
		{
			name: "getPods",
			args: args{
				c: &mockClient,
			},
			want:    []v1.Pod{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPods(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPods() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_FindNode(t *testing.T) {
	type fields struct {
		Name       string
		Nodes      []v1.Node
		Pods       []v1.Pod
		OldestNode v1.Node
		StuckPods  []v1.Pod
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1.Node
		wantErr bool
	}{
		{
			name: "findNode",
			fields: fields{
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
			args: args{
				name: "node1",
			},
			want: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &CloudPlatformCluster{
				Name:       tt.fields.Name,
				Nodes:      tt.fields.Nodes,
				Pods:       tt.fields.Pods,
				OldestNode: tt.fields.OldestNode,
				StuckPods:  tt.fields.StuckPods,
			}
			got, err := cluster.FindNode(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cluster.FindNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cluster.FindNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_GetStuckPods(t *testing.T) {
	type args struct {
		c *KubeClient
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "getPods",
			args: args{
				c: &mockClient,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &CloudPlatformCluster{}
			if err := cluster.GetStuckPods(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Cluster.GetStuckPods() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
