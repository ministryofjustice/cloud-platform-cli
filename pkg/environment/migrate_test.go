package environment

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestGrepFile(t *testing.T) {
	hasBusinessUnit := grepFile("fixtures/foobar-namespace.yml", []byte("cloud-platform.justice.gov.uk/business-unit"))
	if hasBusinessUnit == 0 {
		t.Errorf("Business Unit annotation exist inside fixures file, grepFile() returned %v - expected: 1", hasBusinessUnit)
	}

	hasRandomAnnotation := grepFile("fixtures/foobar-namespace.yml", []byte("whatever"))
	if hasRandomAnnotation != 0 {
		t.Errorf("whatever annotation DOES NOT exist inside fixures file, grepFile() returned %v - expected: 0", hasRandomAnnotation)
	}
}

func Test_changeElasticSearch(t *testing.T) {
	fileContent := []byte(
		"module testmodule {\n" +
			"  foo = a.x + a.y * b.c\n" +
			"  bar = max(a.z, b.c)\n" +
			"}",
	)

	file, err := ioutil.TempFile("./", "testFile")
	if err != nil {
		log.Printf("Error creating test file: %e", err)
	}

	defer os.Remove(file.Name())

	if _, err = file.Write(fileContent); err != nil {
		log.Printf("Error writing to test file: %e", err)
	}

	if err := file.Close(); err != nil {
		log.Printf("Error closing test file: %e", err)
	}

	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successfully write lines to file",
			args: args{
				file: file.Name(),
			},
			wantErr: false,
		},
		{
			name: "Incorrect filename",
			args: args{
				file: "obviouslyFakeName",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := changeElasticSearch(tt.args.file); (err != nil) != tt.wantErr {
				t.Errorf("changeElasticSearch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
