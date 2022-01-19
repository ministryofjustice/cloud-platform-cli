package cluster

import (
	"reflect"
	"testing"
	"time"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	c = &client.Client{
		Clientset: fake.NewSimpleClientset(),
	}

	m = NewMock()
	s = m.NewSnapshot()
)

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

func TestCluster_DeleteStuckPods(t *testing.T) {
	type args struct {
		c *client.Client
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "DeleteStuckPods",
			args: args{
				c: c,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := m
			if err := cluster.DeleteStuckPods(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Cluster.DeleteStuckPods() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCluster_FindNode(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    v1.Node
		wantErr bool
	}{
		{
			name: "FindNode",
			args: args{
				name: "node1",
			},
			want: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := m
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

func TestCluster_NewSnapshot(t *testing.T) {
	tests := []struct {
		name string
		want *Snapshot
	}{
		{
			name: "NewSnapshot",
			want: &Snapshot{
				Cluster: *m,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := m
			got := c.NewSnapshot()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cluster.NewSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNodePods(t *testing.T) {
	type args struct {
		c *client.Client
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
				c: c,
				n: &m.OldestNode,
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

func Test_getClusterName(t *testing.T) {
	type args struct {
		c *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "failGetClusterName",
			args: args{
				c: c,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getClusterName(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("getClusterName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getClusterName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPods(t *testing.T) {
	type args struct {
		c *client.Client
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
				c: c,
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

func Test_getNodes(t *testing.T) {
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
			name: "getNodes",
			args: args{
				c: c,
			},
			want:    []v1.Node{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNodes(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNodes() = %v, want %v", got, tt.want)
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
									Type:   v1.NodeMemoryPressure,
									Status: v1.ConditionFalse,
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
				Node:       tt.fields.Node,
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
			c := m
			if err := c.CompareNodes(tt.args.snap); (err != nil) != tt.wantErr {
				t.Errorf("Cluster.CompareNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
