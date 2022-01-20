package cluster

import (
	"reflect"
	"testing"
	"time"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_GetAllNodes(t *testing.T) {
	type args struct {
		c *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    []v1.Node
		wantErr bool
	}{
		{
			name: "getAllNodes",
			args: args{
				c: c,
			},
			want:    []v1.Node{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAllNodes(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAllNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCluster_areNodesReady(t *testing.T) {
	type fields struct {
		Name       string
		Nodes      []v1.Node
		Node       v1.Node
		Pods       []v1.Pod
		OldestNode v1.Node
		StuckPods  []v1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "nodesReady",
			fields: fields{
				Name:       "test",
				Nodes:      []v1.Node{},
				Node:       m.OldestNode,
				Pods:       []v1.Pod{},
				OldestNode: m.OldestNode,
				StuckPods:  []v1.Pod{},
			},
			wantErr: false,
		},
		{
			name: "nodesAreNotReady",
			fields: fields{
				Name: "test",
				Nodes: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: v1.NodeStatus{
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeDiskPressure,
									Status: v1.ConditionTrue,
								},
							},
						},
					},
				},
				Pods:       []v1.Pod{},
				Node:       m.OldestNode,
				OldestNode: m.OldestNode,
				StuckPods:  []v1.Pod{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				Name:       tt.fields.Name,
				Nodes:      tt.fields.Nodes,
				Pods:       tt.fields.Pods,
				OldestNode: tt.fields.OldestNode,
				StuckPods:  tt.fields.StuckPods,
			}
			if err := c.areNodesReady(); (err != nil) != tt.wantErr {
				t.Errorf("Cluster.areNodesReady() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCluster_CompareNodes(t *testing.T) {
	type args struct {
		snap *Snapshot
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "compareNodes",
			args: args{
				snap: s,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.CompareNodes(tt.args.snap); (err != nil) != tt.wantErr {
				t.Errorf("Cluster.CompareNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCluster_DeleteNode(t *testing.T) {
	type args struct {
		client     *client.Client
		awsProfile string
		awsRegion  string
		node       *v1.Node
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "failDeleteNode",
			args: args{
				client:     c,
				awsProfile: "",
				awsRegion:  "",
				node:       &v1.Node{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.DeleteNode(tt.args.client, tt.args.awsProfile, tt.args.awsRegion, tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("Cluster.DeleteNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_oldestNode(t *testing.T) {
	var (
		timeNow    = time.Now()
		timeMinus  = timeNow.Add(-time.Hour)
		timeOldest = timeNow.Add(-time.Hour * 2)
	)

	type args struct {
		nodes []v1.Node
	}

	tests := []struct {
		name    string
		args    args
		want    v1.Node
		wantErr bool
	}{
		{
			name: "OldestNode",
			args: args{
				nodes: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{
								Time: timeNow,
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{
								Time: timeMinus,
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{
								Time: timeOldest,
							},
						},
					},
				},
			},
			want: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{
						Time: timeOldest,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := oldestNode(tt.args.nodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("oldestNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("oldestNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
