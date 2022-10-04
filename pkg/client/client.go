package client

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeClient is a wrapper around the kubernetes client interface
type KubeClient struct {
	Clientset kubernetes.Interface
}

// AwsCredentials represents the AWS credentials used to connect to an AWS account.
type AwsCredentials struct {
	Session *session.Session
	Profile string
	Eks     *eks.EKS
	Ec2     *ec2.EC2
	Region  string
}

// NewKubeClient will construct a Client struct to interact with a kubernetes cluster
func NewKubeClient(p string) (*KubeClient, error) {
	clientset, err := GetClientset(p)
	if err != nil {
		return nil, err
	}

	return &KubeClient{
		Clientset: clientset,
	}, nil
}

// NewAwsCredentials constructs and populates a new AwsCredentials object
func NewAwsCreds(region string) (*AwsCredentials, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}

	eks := eks.New(sess)

	ec2 := ec2.New(sess)

	return &AwsCredentials{
		Session: sess,
		Region:  region,
		Eks:     eks,
		Ec2:     ec2,
	}, nil
}

// GetClientset takes the path to a kubeconfig file and returns a clientset
func GetClientset(p string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", p)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
