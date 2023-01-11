package util

import (
	"bytes"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// If we assign a string value to 'repository', we get it back
func TestWorkingRepository(t *testing.T) {
	re := Repository{currentRepository: "foobar"}
	str, err := re.Repository()
	if str != "foobar" {
		t.Errorf("Something went wrong: %s", str)
	}
	if err != nil {
		t.Errorf("Expected an error, got nil")
	}
}

// If we aren't in a git repo, we get an Error
func TestRepoErrorRepository(t *testing.T) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}

	// Change to a directory which isn't a git repo
	if err = os.Chdir("/tmp"); err != nil {
		t.Errorf("Something went wrong: %s", err)
	}

	re := Repository{}
	_, err = re.Repository()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}

	// Change back to the original directory
	if err := os.Chdir(wd); err != nil {
		t.Errorf("Something went wrong: %s", err)
	}
}

// If we don't assign a string value to 'repository', we get whatever the
// current git repository is called
func TestRepoDefaultRepository(t *testing.T) {
	re := Repository{}
	str, _ := re.Repository()
	if str != "cloud-platform-cli" {
		t.Errorf("Expected cloud-platform-cli, got: x%sx", str)
	}
}

// If we assign a string value to 'branch', we get it back
func TestWorkingBranch(t *testing.T) {
	re := Repository{branch: "foobar"}
	str, err := re.GetBranch()
	if str != "foobar" {
		t.Errorf("Something went wrong: %s", str)
	}
	if err != nil {
		t.Errorf("Expected an error, got nil")
	}
}

// If we don't assign a string value to 'branch', we get whatever the
// current git branch is called
func TestRepoDefaultBranch(t *testing.T) {
	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}

	re := Repository{}
	str, _ := re.GetBranch()
	if str != strings.Trim(string(branch), "\n") {
		t.Errorf("Expected master, got: x%sx", str)
	}
}

// If we aren't in a git repo, we get an Error
func TestRepoErrorBranch(t *testing.T) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}

	// Change to a directory which isn't a git repo
	if err = os.Chdir("/tmp"); err != nil {
		t.Errorf("Something went wrong: %s", err)
	}

	re := Repository{}
	_, err = re.GetBranch()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}

	// Change back to the original directory
	if err := os.Chdir(wd); err != nil {
		t.Errorf("Something went wrong: %s", err)
	}
}

func TestRedacted(t *testing.T) {
	type args struct {
		output string
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
			},
			expect: "REDACTED\n",
		},
		{
			name: "Redacted Sercet Content",
			args: args{
				output: "secret: 1234567890",
			},
			expect: "REDACTED\n",
		},
		{
			name: "Redacted Token Content",
			args: args{
				output: "token: 1234567890",
			},
			expect: "REDACTED\n",
		},
		{
			name: "Redacted Key Content",
			args: args{
				output: "key: 1234567890",
			},
			expect: "REDACTED\n",
		},
		{
			name: "Redacted Webhook Content",
			args: args{
				output: "https://hooks.slack.com",
			},
			expect: "REDACTED\n",
		},
		{
			name: "Unredacted Content",
			args: args{
				output: "This test should not be redacted",
			},
			expect: "This test should not be redacted\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			Redacted(&output, tt.args.output)
			if tt.expect != output.String() {
				t.Errorf("got %s but expected %s", output.String(), tt.expect)
			}
		})
	}
}

func TestGetDatePastMinute(t *testing.T) {
	type args struct {
		timestamp string
		minutes   int
	}
	tests := []struct {
		name    string
		args    args
		want    *Date
		wantErr bool
	}{
		{
			name: "same date with 1 minutes",
			args: args{
				timestamp: "2022-12-07 18:12:46 +0000",
				minutes:   1,
			},
			want: &Date{
				First: "2022-12-07T18:13:00",
				Last:  "2022-12-07T18:11:46",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDatePastMinute(tt.args.timestamp, tt.args.minutes)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDatePastMinute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDatePastMinute() = %v, want %v", got, tt.want)
			}
		})
	}
}
