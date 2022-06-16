package terminalcfg

import (
	"testing"
)

func TestSetKubeEnv(t *testing.T) {
	type args struct {
		env string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Failure",
			args: args{
				env: "",
			},
			want: false,
		},
		{
			name: "live Success",
			args: args{
				env: "live",
			},
			want: true,
		},
		{
			name: "manager Success",
			args: args{
				env: "manager",
			},
			want: true,
		},
		{
			name: "test Success",
			args: args{
				env: "cluster-test",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetKubeEnv(tt.args.env); got != tt.want {
				t.Errorf("LiveManagerEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetTFWksp(t *testing.T) {
	type args struct {
		clusterName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Failure",
			args: args{
				clusterName: "",
			},
			want: false,
		},
		{
			name: "live Success",
			args: args{
				clusterName: "live",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTFWksp(tt.args.clusterName)
		})
	}
}
