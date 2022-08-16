package environment

import (
	"bytes"
	"os"
	"testing"
)

func TestBumpModule(t *testing.T) {
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
				file:         createTestFile(t),
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
				file:         createTestFile(),
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

			if checkModuleChange(tt.args.checkVersion, tt.args.file.Name()) != tt.wantSuccess {
				t.Errorf("BumpModule() checkSourceChange = %v, want %v", checkModuleChange(tt.args.checkVersion, tt.args.file.Name()), tt.wantSuccess)
			}
			defer os.Remove(tt.args.file.Name())
		})
	}
}

// createTestFile creates a test file with the given version.
func createTestFile(t *testing.T) os.File {
	f, err := os.Create("test.tf")
	if err != nil {
		t.Errorf("Error creating file: %s", err)
	}

	defer f.Close()
	f.WriteString("module test { source = \"test=1.0.0\" }")

	return *f
}

const chunkSize = 64000

// checkModuleChange checks if the file has been changed and contains
// the string passed to it.
func checkModuleChange(t *testing.T, v, f string) bool {
	file, err := os.Open(f)
	if err != nil {
		t.Errorf("Error reading file: %s", err)
	}
	defer file.Close()

	b := make([]byte, chunkSize)
	_, err = file.Read(b)

	if err != nil {
		t.Errorf("Error reading file: %s", err)
	}

	return bytes.Contains(b, []byte(v))
}
