package environment

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestBumpModule(t *testing.T) {
	f, err := createTestFile()
	if err != nil {
		t.Errorf("Error creating test file: %e", err)
	}
	type args struct {
		moduleName   string
		newVersion   string
		checkVersion string
		file         os.File
	}

	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantSuccess bool
	}{
		{
			name: "Correctly bump a module version",
			args: args{
				moduleName:   "test",
				newVersion:   "0.1.2",
				checkVersion: "0.1.2",
				file:         f,
			},
			wantErr:     false,
			wantSuccess: true,
		},
		{
			name: "Incorrectly bump a module version",
			args: args{
				moduleName:   "test",
				newVersion:   "0.1.2",
				checkVersion: "NOTHING",
				file:         f,
			},
			wantErr:     false,
			wantSuccess: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BumpModule(tt.args.moduleName, tt.args.newVersion); (err != nil) != tt.wantErr {
				t.Errorf("BumpModule() error = %v, wantErr %v", err, tt.wantErr)
			}

			check, err := checkModuleChange(tt.args.checkVersion, tt.args.file.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("BumpModule() error = %v, wantErr %v", err, tt.wantErr)
			}

			if check != tt.wantSuccess {
				t.Errorf("BumpModule() checkSourceChange = %v, want %v", check, tt.wantSuccess)
			}
			defer os.Remove(tt.args.file.Name())
		})
	}
}

// createTestFile creates a test file with the given version.
func createTestFile() (os.File, error) {
	f, err := os.Create("test.tf")
	if err != nil {
		return *f, fmt.Errorf("Error creating file: %e", err)
	}

	defer f.Close()
	f.WriteString("module test { source = \"test=1.0.0\" }")

	return *f, nil
}

const chunkSize = 64000

// checkModuleChange checks if the file has been changed and contains
// the string passed to it.
func checkModuleChange(v, f string) (bool, error) {
	file, err := os.Open(f)
	if err != nil {
		return false, fmt.Errorf("Error reading file: %e", err)
	}
	defer file.Close()

	b := make([]byte, chunkSize)
	_, err = file.Read(b)

	if err != nil {
		return false, fmt.Errorf("Error reading file: %e", err)
	}

	return bytes.Contains(b, []byte(v)), nil
}
