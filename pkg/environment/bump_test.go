package environment

import (
	"log"
	"os"
	"testing"
)

func createTestFile() error {
	f, err := os.Create("test.tf")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString("module test { source = \"test=0.1.1\" }")

	return nil
}

func TestBumpModule(t *testing.T) {
	err := createTestFile()
	if err != nil {
		log.Fatalln("Failed to create test Terraform file:", err)
	}

	defer os.Remove("test.tf")

	type args struct {
		m string
		v string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Bump module version",
			args: args{
				m: "test",
				v: "0.1.2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BumpModule(tt.args.m, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("BumpModule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
