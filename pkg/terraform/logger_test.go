package terraform

import "testing"

func TestCmdOutput_redacted(t *testing.T) {
	type fields struct {
		Stdout string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Redacted Password Content",
			fields: fields{
				Stdout: "password: 1234567890",
			},
		},
		{
			name: "Redacted Sercet Content",
			fields: fields{
				Stdout: "secret: 1234567890",
			},
		},
		{
			name: "Redacted Token Content",
			fields: fields{
				Stdout: "token: 1234567890",
			},
		},
		{
			name: "Redacted Key Content",
			fields: fields{
				Stdout: "key: 1234567890",
			},
		},
		{
			name: "Redacted Webhook Content",
			fields: fields{
				Stdout: "https://hooks.slack.com",
			},
		},
		{
			name: "Unredacted Content",
			fields: fields{
				Stdout: "This test should not be redacted",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &CmdOutput{
				Stdout: tt.fields.Stdout,
			}
			o.redacted()
		})
	}
}
