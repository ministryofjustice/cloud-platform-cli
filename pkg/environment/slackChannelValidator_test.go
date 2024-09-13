package environment

import "testing"

func TestSlackChannelValidator_IsValid(t *testing.T) {
	v := &slackChannelValidator{}

	tests := []struct {
		name string
		args string
		want bool
	}{
		{
			name: "valid slack channel",
			args: "valid-channel",
			want: true,
		},
		{
			name: "invalid slack channel with special characters",
			args: "invalid#channel",
			want: false,
		},
		{
			name: "invalid slack channel with uppercase letters",
			args: "InvalidChannel",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.isValid(tt.args); got != tt.want {
				t.Errorf("slackChannelValidator.isValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
