package client

import (
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes"
)

func TestGetClientset(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name    string
		args    args
		want    kubernetes.Interface
		wantErr bool
	}{
		{
			name: "FailToGenerateConfig",
			args: args{
				p: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetClientset(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClientset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClientset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *Client
	}{
		{
			name: "grabClient",
			want: &Client{
				Clientset: &kubernetes.Clientset{},
			},
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
