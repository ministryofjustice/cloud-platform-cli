package cluster

import (
	"reflect"
	"testing"
)

func TestCluster_NewSnapshot(t *testing.T) {
	tests := []struct {
		name string
		want *Snapshot
	}{
		{
			name: "NewSnapshot",
			want: &Snapshot{
				Cluster: *mockCluster,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mockCluster
			got := c.NewSnapshot()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cluster.NewSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAwsCreds(t *testing.T) {
	type args struct {
		region string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "NewAwsCreds",
			args: args{
				region: "us-east-1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAwsCreds(tt.args.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAwsCreds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
