package client

import (
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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

func TestNewKubeClient(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name    string
		args    args
		want    *KubeClient
		wantErr bool
	}{
		{
			name: "Pass empty path",
			args: args{
				p: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKubeClient(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubeClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKubeClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
