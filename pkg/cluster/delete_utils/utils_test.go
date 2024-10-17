package deleteutils

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	clienttesting "k8s.io/client-go/testing"
)

type mockEKSClient struct {
	eksiface.EKSAPI
	mockBehaviour string
}

type mockEc2Client struct {
	ec2iface.EC2API
	mockBehaviour string
}

func (m *mockEKSClient) ListClusters(input *eks.ListClustersInput) (*eks.ListClustersOutput, error) {
	if m.mockBehaviour == "error" {
		return nil, errors.New("big eks client err")
	}
	clusters := []*string{aws.String("invalid-test")}
	response := &eks.ListClustersOutput{
		Clusters: clusters,
	}

	return response, nil
}

func (m *mockEc2Client) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	if m.mockBehaviour == "error" {
		return nil, errors.New("big ec2 vpc client err")
	}
	vpcs := []*ec2.Vpc{{VpcId: aws.String("invalid-test")}}
	response := &ec2.DescribeVpcsOutput{
		Vpcs: vpcs,
	}

	return response, nil
}

func Test_AbortIfUserNamespacesExist(t *testing.T) {
	type args struct {
		namespaces       []v1.Namespace
		systemNamespaces []string
	}

	userNSArgs := args{
		namespaces: []v1.Namespace{
			{ObjectMeta: metav1.ObjectMeta{Name: "error-user-ns"}},
		},
		systemNamespaces: []string{"default", "kube-system"},
	}

	pureSystemArgs := args{
		namespaces: []v1.Namespace{{
			ObjectMeta: metav1.ObjectMeta{Name: "kube-system"},
		}},
		systemNamespaces: []string{"default", "kube-system"},
	}

	smokeTestNSArgs := args{
		namespaces:       []v1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "smoketest-example"}}, {ObjectMeta: metav1.ObjectMeta{Name: "smoketest-example2"}}},
		systemNamespaces: []string{"default", "kube-system"},
	}

	starterPackNSArgs := args{
		namespaces:       []v1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "starter-pack-example"}}, {ObjectMeta: metav1.ObjectMeta{Name: "starter-pack-example2"}}},
		systemNamespaces: []string{"default", "kube-system"},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "GIVEN a list of namespaces AND it contains user created (non-system) namespaces THEN return an error", args: userNSArgs, wantErr: true},
		{name: "GIVEN a list of namespaces AND it DOES NOT contain any user created namespaces THEN return nil", args: pureSystemArgs, wantErr: false},
		{name: "GIVEN a list of namespaces AND it contains smoketest namespaces THEN return nil", args: smokeTestNSArgs, wantErr: false},
		{name: "GIVEN a list of namespaces AND it contains starter-pack namespaces THEN return nil", args: starterPackNSArgs, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AbortIfUserNamespacesExist(tt.args.namespaces, tt.args.systemNamespaces); (err != nil) != tt.wantErr {
				t.Errorf("abortIfUserNamespacesExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_contains(t *testing.T) {
	type args struct {
		s   []string
		str string
	}

	validStrArr := []string{"testVal"}
	validArgs := args{
		s:   validStrArr,
		str: "testVal",
	}

	invalidArgs := args{
		s:   validStrArr,
		str: "incorrect",
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "GIVEN a string and that string is also in the string array THEN return true", args: validArgs, want: true},
		{name: "GIVEN a string and that string is NOT in the string array THEN return false", args: invalidArgs, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := util.Contains(tt.args.s, tt.args.str); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetNamespaces(t *testing.T) {
	type args struct {
		clientset kubernetes.Interface
	}

	mockClient := client.KubeClient{
		Clientset: fake.NewSimpleClientset(),
	}

	mockErrClient := client.KubeClient{
		Clientset: fake.NewSimpleClientset(),
	}

	mockErrClient.Clientset.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("list", "namespaces", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1.NamespaceList{}, errors.New("Error listing namespaces")
	})

	standardNamespaces := []v1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}}}

	tests := []struct {
		name               string
		args               args
		want               []v1.Namespace
		existingNamespaces []v1.Namespace
		wantErr            bool
	}{
		{name: "GIVEN a k8s client AND namespaces THEN return a list of namespaces", args: args{clientset: mockClient.Clientset}, want: standardNamespaces, existingNamespaces: standardNamespaces, wantErr: false},
		{name: "GIVEN a k8s client that errors THEN return the error", args: args{clientset: mockErrClient.Clientset}, want: nil, existingNamespaces: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, ns := range tt.want {
				initNamespaces, err := tt.args.clientset.CoreV1().Namespaces().Create(context.Background(), &ns, metav1.CreateOptions{})

				t.Log("create namespaces...", initNamespaces)

				if err != nil {
					t.Errorf("getUserNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			got, err := GetNamespaces(tt.args.clientset)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUserNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_CheckClusterIsDestroyed(t *testing.T) {
	type args struct {
		clusterName   string
		eksAPI        eksiface.EKSAPI
		mockBehaviour string
	}

	mockEKSClient := &mockEKSClient{}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "GIVEN an eks client, list all the namespaces THEN return successfully if the cluster is NOT listed", args: args{clusterName: "valid-test", eksAPI: mockEKSClient, mockBehaviour: "normal"}, wantErr: false},
		{name: "GIVEN an eks client, list all the namespaces THEN return an error if the cluster is listed", args: args{clusterName: "invalid-test", eksAPI: mockEKSClient, mockBehaviour: "normal"}, wantErr: true},
		{name: "GIVEN an eks client THEN error WHEN the client errors", args: args{clusterName: "test", eksAPI: mockEKSClient, mockBehaviour: "error"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEKSClient.mockBehaviour = tt.args.mockBehaviour
			if err := CheckClusterIsDestroyed(tt.args.clusterName, tt.args.eksAPI); (err != nil) != tt.wantErr {
				t.Errorf("CheckClusterIsDestroyed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_CheckVpcIsDestroyed(t *testing.T) {
	type args struct {
		vpcID         string
		ec2API        ec2iface.EC2API
		mockBehaviour string
	}

	mockEC2Client := &mockEc2Client{}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "GIVEN an ec2 client, describe all the vpcs THEN return successfully if the vpc is NOT listed", args: args{vpcID: "valid-test", ec2API: mockEC2Client, mockBehaviour: "normal"}, wantErr: false},
		{name: "GIVEN an ec2 client, decsribe all the vpcs THEN return an error if the vpc is listed", args: args{vpcID: "invalid-test", ec2API: mockEC2Client, mockBehaviour: "normal"}, wantErr: true},
		{name: "GIVEN an ec2 client THEN error WHEN the client errors", args: args{vpcID: "test", ec2API: mockEC2Client, mockBehaviour: "error"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEC2Client.mockBehaviour = tt.args.mockBehaviour
			if err := CheckVpcIsDestroyed(tt.args.vpcID, tt.args.ec2API); (err != nil) != tt.wantErr {
				t.Errorf("CheckVpcIsDestroyed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
