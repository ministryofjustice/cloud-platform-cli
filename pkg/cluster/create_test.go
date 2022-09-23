package cluster

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
)

type mockEKSClient struct {
	eksiface.EKSAPI
}

func (m *mockEKSClient) getName(name string) (*eks.Cluster, error) {
	// mock response/functionality
	return &eks.Cluster{
		Name: &name,
		// Status: aws.String("ACTIVE"),
	}, nil
}

// func TestCheckCluster(t *testing.T) {
// 	mockSvc := &mockEKSClient{}
// 	mockSvc.getName("test")

// 	err := checkCluster("test", mockSvc)
// 	if err != nil {
// 		t.Errorf("checkCluster() error = %v", err)
// 	}
// }

func TestGetCluster(t *testing.T) {
	// Setup Test
	mockSvc := &mockEKSClient{}

	mockSvc.getName("test")

	_, err := getCluster("test", mockSvc)
	if err != nil {
		t.Errorf("getCluster() error = %v", err)
	}

	// if *c.Name != "test" {
	// 	t.Errorf("getCluster() error = %v", "cluster name not set")
	// }

	// Verify myFunc's functionality
}
func Test_setEnvironment(t *testing.T) {
	auth := &AuthOpts{
		ClientId:     "test",
		ClientSecret: "test",
		Domain:       "test",
	}

	options := &CreateOptions{
		Name:  "test",
		Auth0: *auth,
	}

	fakeCredentials := &AwsCredentials{
		Profile: "test",
		Region:  "test",
	}

	if err := setEnvironment(options, fakeCredentials); err != nil {
		t.Errorf("setEnvironment() error = %v", err)
	}

	defer removeTestEnvVars(t)

	if os.Getenv("AUTH0_CLIENT_ID") != "test" {
		t.Errorf("setEnvironment() error = %v", "AUTH0_CLIENT_ID not set")
	}

	if os.Getenv("AUTH0_CLIENT_SECRET") != "test" {
		t.Errorf("setEnvironment() error = %v", "AUTH0_CLIENT_SECRET not set")
	}

	if os.Getenv("AUTH0_DOMAIN") != "test" {
		t.Errorf("setEnvironment() error = %v", "AUTH0_DOMAIN not set")
	}

	if os.Getenv("AWS_PROFILE") != "test" {
		t.Errorf("setEnvironment() error = %v", "AWS_PROFILE not set")
	}

	if os.Getenv("AWS_REGION") != "test" {
		t.Errorf("setEnvironment() error = %v", "AWS_REGION not set")
	}
}

func removeTestEnvVars(t *testing.T) {
	t.Helper()
	os.Unsetenv("AUTH0_CLIENT_ID")
	os.Unsetenv("AUTH0_CLIENT_SECRET")
	os.Unsetenv("AUTH0_DOMAIN")

	os.Unsetenv("AWS_PROFILE")
	os.Unsetenv("AWS_REGION")
}

func Test_deleteLocalState(t *testing.T) {
	parentDir := "testParent"
	file := "testFile"
	siblingDir := "testDir"

	os.RemoveAll(parentDir)
	err := os.Mkdir(parentDir, 0755)
	if err != nil {
		t.Errorf("deleteLocalState() error = %v", err)
	}
	defer os.RemoveAll(parentDir)

	// create file in temp directory
	_, err = os.CreateTemp(parentDir, file)
	if err != nil {
		t.Errorf("deleteLocalState() error = %v", err)
	}

	// create directory in temp directory
	_, err = os.MkdirTemp(parentDir, siblingDir)
	if err != nil {
		t.Errorf("deleteLocalState() error = %v", err)
	}

	if err := deleteLocalState(parentDir, file, siblingDir); err != nil {
		t.Errorf("deleteLocalState() error = %v", err)
	}

	if _, err = os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("deleteLocalState() error = %v", "file not deleted")
	}

	if _, err := os.Stat(siblingDir); !os.IsNotExist(err) {
		t.Errorf("deleteLocalState() error = %v", "directory not deleted")
	}
}

func TestTerraformOptions_DefaultDirSetup(t *testing.T) {
	testTfOptions := &TerraformOptions{}

	testTfOptions.DefaultDirSetup()

	if testTfOptions.DirStructure.List == nil {
		t.Errorf("DefaultDirSetup() error = %v", "Dir not set")
	}
}

func TestNewTerraformOptions(t *testing.T) {
	testCreateOptions := &CreateOptions{
		Name: "test",
	}
	_, err := newTerraformOptions(testCreateOptions)
	if err == nil {
		t.Errorf("newTerraformOptions() error = %v", "expected error")
	}

	testCreateOptions.TfVersion = "0.14.8"
	testCreateOptions.TfDirectories = []string{"test"}
	testTf, err := newTerraformOptions(testCreateOptions)
	if err != nil {
		t.Errorf("newTerraformOptions() error = %v", err)
	}

	if testTf.Version != "0.14.8" {
		t.Errorf("newTerraformOptions() error = %v", "terraform version not set")
	}

	if testTf.DirStructure.List == nil {
		t.Errorf("newTerraformOptions() error = %v", "Dir not set")
	}
}
