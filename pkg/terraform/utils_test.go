package terraform

import (
	"bytes"
	"testing"
)

func Test_Redacted(t *testing.T) {
	type args struct {
		output string
		redact bool
	}
	tests := []struct {
		name   string
		args   args
		expect string
	}{
		{
			name: "Redacted Password Content",
			args: args{
				output: "password: 1234567890",
				redact: true,
			},
			expect: "REDACTED\n",
		},
		{
			name: "Redacted Sercet Content",
			args: args{
				output: "secret: 1234567890",
				redact: true,
			},
			expect: "REDACTED\n",
		},
		{
			name: "Redacted Token Content",
			args: args{
				output: "token: 1234567890",
				redact: true,
			},
			expect: "REDACTED\n",
		},
		{
			name: "Redacted Key Content",
			args: args{
				output: "key: 1234567890",
				redact: true,
			},
			expect: "REDACTED\n",
		},
		{
			name: "Redacted Webhook Content",
			args: args{
				output: "https://hooks.slack.com",
				redact: true,
			},
			expect: "REDACTED\n",
		},
		{
			name: "Dont Redacted Webhook Content",
			args: args{
				output: "https://hooks.slack.com",
				redact: false,
			},
			expect: "https://hooks.slack.com\n",
		},
		{
			name: "Unredacted Content",
			args: args{
				output: "This test should not be redacted",
				redact: true,
			},
			expect: "This test should not be redacted\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			Redacted(w, tt.args.output, tt.args.redact)
			if got := w.String(); got != tt.expect {
				t.Errorf("redacted() = %v, want %v", got, tt.expect)
			}
		})
	}
}
