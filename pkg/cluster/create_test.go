package cluster

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

type mockEKSClient struct {
	eksiface.EKSAPI
}

type mockEC2Client struct {
	ec2iface.EC2API
}

func (m *mockEC2Client) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	if *input.Filters[0].Values[0] != "test" {
		return nil, errors.New("bad vpc")
	}

	return &ec2.DescribeVpcsOutput{
		Vpcs: []*ec2.Vpc{
			{
				VpcId: aws.String("test"),
			},
		},
	}, nil
}

func TestWriteKubeConfig(t *testing.T) {
	cluster := &eks.Cluster{
		Name:     aws.String("test"),
		Endpoint: aws.String("https://test"),
		CertificateAuthority: &eks.Certificate{
			Data: aws.String("test"),
		},
	}

	toks := token.Token{
		Token: "test",
	}

	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		t.Errorf("error decoding certificate: %v", err)
	}

	err = writeKubeConfig(cluster, "config.yaml", "aws", toks, ca)
	if err != nil {
		t.Errorf("WriteKubeConfig() error = %v", err)
	}

	defer func() {
		err := os.Remove("config.yaml")
		if err != nil {
			t.Errorf("error removing config file: %v", err)
		}
	}()

	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		t.Errorf("WriteKubeConfig() error = %v", "config.yaml not found")
	}

	// Expect the test to fail from here.
	err = writeKubeConfig(cluster, "", "aws", toks, []byte{})
	if err == nil {
		t.Errorf("WriteKubeConfig() error = %v", "expected error")
	}

	err = writeKubeConfig(cluster, "badPath", "aws", toks, nil)
	if err == nil {
		t.Errorf("WriteKubeConfig() error = %v", "expected error")
	}

	cluster.Name = nil
	err = writeKubeConfig(cluster, "badPath", "aws", toks, nil)
	if err == nil {
		t.Errorf("WriteKubeConfig() error = %v", "expected error")
	}
}

func TestNewClientset(t *testing.T) {
	cluster := &eks.Cluster{
		Name:     aws.String("test"),
		Endpoint: aws.String("https://test"),
		CertificateAuthority: &eks.Certificate{
			Data: aws.String("test"),
		},
	}

	_, err := newClientset(cluster, "aws", "test")
	if err == nil {
		t.Errorf("newClientset() error = %v", err)
	}
}

func (m *mockEKSClient) DescribeCluster(input *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error) {
	if *input.Name != "test" {
		return nil, errors.New("cluster not found")
	}

	return &eks.DescribeClusterOutput{
		Cluster: &eks.Cluster{
			Name:   aws.String("test"),
			Status: aws.String("ACTIVE"),
		},
	}, nil
}

func TestCheckVpc(t *testing.T) {
	mockSvc := &mockEC2Client{}

	testOpt := "test"

	// Good path
	err := checkVpc(testOpt, "test", mockSvc)
	if err != nil {
		t.Errorf("checkVpc() error = %v", err)
	}

	// Bad path
	err = checkVpc(testOpt, "obviouslyWrong", mockSvc)
	if err == nil {
		t.Errorf("checkVpc() error = %v", err)
	}

	testOpt = "incorrectWorkspace"
	err = checkVpc(testOpt, "test", mockSvc)
	if err == nil {
		t.Error("we expected an error here checkVpc() error")
	}
}

func TestGetVpc(t *testing.T) {
	mockSvc := &mockEC2Client{}

	out, err := getVpc("test", mockSvc)
	if err != nil {
		t.Errorf("checkVpc() error = %v", err)
	}

	if *out.Vpcs[0].VpcId != "test" {
		t.Errorf("checkVpc() error = %v", "vpc not found")
	}

	_, err = getVpc("bad", mockSvc)
	if err == nil {
		t.Errorf("was expecting an error. checkVpc() error = %v", "expected error")
	}

}

func TestCheckCluster(t *testing.T) {
	mockSvc := &mockEKSClient{}

	// Good path
	err := checkCluster("test", mockSvc)
	if err != nil {
		t.Errorf("checkCluster() error = %v", err)
	}

	// Bad path
	err = checkCluster("bad", mockSvc)
	if err == nil {
		t.Errorf("checkCluster() error = %v", "expected error")
	}
}

func TestGetCluster(t *testing.T) {
	mockSvc := &mockEKSClient{}

	// Good path
	mockCluster, err := getCluster("test", mockSvc)
	if err != nil {
		t.Errorf("getCluster() error = %v", err)
	}

	if *mockCluster.Name != "test" {
		t.Errorf("getCluster() error = %v", "cluster name not set")
	}

	// Bad path
	_, err = getCluster("bad", mockSvc)
	if err == nil {
		t.Errorf("was expecting an error here. getCluster() error = %v", "expected error")
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
