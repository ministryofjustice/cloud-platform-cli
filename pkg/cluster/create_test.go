package cluster

import (
	"context"
	"os"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

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

func TestApplyTacticalPspFix(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&v1beta1.PodSecurityPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "eks.privileged",
			},
		},
		// Add pods
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "FakePod",
			},
		},
	)

	// Good path
	err := applyTacticalPspFix(fakeClientset)
	if err != nil {
		t.Errorf("applyTacticalPspFix() error = %v", err)
	}

	err = fakeClientset.PolicyV1beta1().PodSecurityPolicies().Delete(context.Background(), "eks.privileged", metav1.DeleteOptions{})
	if err == nil {
		t.Errorf("we wanted to delete the eks.privileged psp. applyTacticalPspFix() error = %v", err)
	}
}
