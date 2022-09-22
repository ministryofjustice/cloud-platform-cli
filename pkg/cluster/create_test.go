package cluster

import (
	"testing"
)

func Test_setEnvironment(t *testing.T) {
}

func Test_deleteLocalState(t *testing.T) {
	type args struct {
		dir   string
		paths []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteLocalState(tt.args.dir, tt.args.paths...); (err != nil) != tt.wantErr {
				t.Errorf("deleteLocalState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
