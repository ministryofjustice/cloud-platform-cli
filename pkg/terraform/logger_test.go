package terraform

import "testing"

func TestCmdOutput_redacted(t *testing.T) {
	type fields struct {
		Stdout   string
		Stderr   string
		ExitCode int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Redacted Content",
			fields: fields{
				Stdout:   "password: 1234567890",
				Stderr:   "",
				ExitCode: 0,
			},
		},
		{
			name: "Unredacted Content",
			fields: fields{
				Stdout:   "This is a test",
				Stderr:   "",
				ExitCode: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &CmdOutput{
				Stdout:   tt.fields.Stdout,
				Stderr:   tt.fields.Stderr,
				ExitCode: tt.fields.ExitCode,
			}
			o.redacted()
		})
	}
}
