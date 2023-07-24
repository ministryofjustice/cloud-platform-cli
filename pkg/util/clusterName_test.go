package util

import "testing"

func TestOptions_IsNameValid(t *testing.T) {
	type fields struct {
		Name          string
		MaxNameLength int
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"GIVEN a cluster name that is longer than the mx length THEN return an error",
			fields{
				"long-cluster-name",
				3,
			},
			true,
		},
		{
			"GIVEN a cluster name that includes 'live' THEN return an error",
			fields{
				"cluster-live-name",
				50,
			},
			true,
		},
		{
			"GIVEN a cluster name that includes 'manager' THEN return an error",
			fields{
				"cluster-manager-name",
				50,
			},
			true,
		},
		{
			"GIVEN a cluster name that is valid THEN return nil",
			fields{
				"cluster-valid-name",
				50,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Options{
				Name:          tt.fields.Name,
				MaxNameLength: tt.fields.MaxNameLength,
			}
			if err := opt.IsNameValid(); (err != nil) != tt.wantErr {
				t.Errorf("Options.IsNameValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
