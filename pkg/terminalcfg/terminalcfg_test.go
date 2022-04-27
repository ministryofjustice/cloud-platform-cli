package terminalcfg

import "testing"

func TestLiveManagerEnv(t *testing.T) {
	type args struct {
		env string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
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
			name: "Failure",
			args: args{
				env: "default",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LiveManagerEnv(tt.args.env); got != tt.want {
				t.Errorf("LiveManagerEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
