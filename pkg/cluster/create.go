package cluster

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

// ApplyVpc when executed will Apply terraform code to create a Cloud Platform VPC and ensure it is up and running.
// It will return an error if the VPC is not up and running or the terraform commands fail.
func (c *Cluster) ApplyVpc(tf *terraform.TerraformCLIConfig, creds *client.AwsCredentials, dir string) error {
	tf.WorkingDir = dir

	output, err := terraformApply(tf)
	if err != nil {
		return err
	}

	vpcID := output["vpc_id"]
	if vpcID.Value == nil {
		return errors.New("vpc_id not found in terraform output")
	}

	// Trim the vpcId to remove quotes
	vpc := strings.Trim(string(vpcID.Value), "\"")
	return checkVpc(tf.Workspace, vpc, creds.Ec2)
}

// ApplyEks will apply the terraform code to create a Cloud Platform EKS cluster and ensure it is up and running.
// It will return an error if the EKS cluster is not up and running or the terraform commands fail.
func (c *Cluster) ApplyEks(tf *terraform.TerraformCLIConfig, creds *client.AwsCredentials, dir string) error {
	tf.WorkingDir = dir

	_, err := terraformApply(tf)
	if err != nil {
		return err
	}

	if err := checkCluster(tf.Workspace, creds.Eks); err != nil {
		return err
	}

	c.HealthStatus = "Good"

	return nil
}

// ApplyComponents will apply the Cloud Platform specific components on top of a running cluster. At this point your
// cluster should be up and running and you should be able to connect to it.

// Unfortunaltey the AWS SDK does not provide a nice method to grab the kube-config for a cluster so we use a raw aws command.
func (c *Cluster) ApplyComponents(tf *terraform.TerraformCLIConfig, awsCreds *client.AwsCredentials, dir, kubeconf string) error {
	// There is a requirement for the aws binary to exist at this point.
	clientset, err := authToCluster(tf.Workspace, awsCreds.Eks, kubeconf)
	if err != nil {
		return fmt.Errorf("failed to auth to cluster: %w", err)
	}

	tf.WorkingDir = dir

	_, err = terraformApply(tf)
	if err != nil {
		return err
	}

	if err := applyTacticalPspFix(clientset); err != nil {
		return err
	}

	kube, err := client.NewKubeClient(kubeconf)
	if err != nil {
		return err
	}

	if err := c.GetStuckPods(kube); err != nil {
		return err
	}

	nodes, err := GetAllNodes(kube)
	if err != nil {
		return err
	}

	c.Nodes = nodes

	return nil
}

func terraformApply(tf *terraform.TerraformCLIConfig) (map[string]tfexec.OutputMeta, error) {
	terraform, error := terraform.NewTerraformCLI(tf)
	if error != nil {
		return nil, error
	}

	// Start fresh and remove any local state.
	if err := deleteLocalState(tf.WorkingDir, ".terraform", ".terraform.lock.hcl"); err != nil {
		return nil, fmt.Errorf("failed to delete temp directory: %w", err)
	}

	err := terraform.Init(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to init terraform: %w", err)
	}

	if err = terraform.Apply(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to apply terraform: %w", err)
	}

	return terraform.Output(context.Background())
}

func authToCluster(name string, eksSvc eksiface.EKSAPI, path string) (*kubernetes.Clientset, error) {
	input := &eks.DescribeClusterInput{
		Name: aws.String(name),
	}
	result, err := eksSvc.DescribeCluster(input)
	if err != nil {
		log.Fatalf("Error calling DescribeCluster: %v", err)
	}

	if err := writeKubeConfig(result.Cluster, path); err != nil {
		return nil, err
	}

	clientset, err := newClientset(result.Cluster)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}
	return clientset, nil
}

func newClientset(cluster *eks.Cluster) (*kubernetes.Clientset, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: aws.StringValue(cluster.Name),
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}

	config := &rest.Config{
		Host:        aws.StringValue(cluster.Endpoint),
		BearerToken: tok.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func writeKubeConfig(cluster *eks.Cluster, path string) error {
	kubePath := path
	fmt.Printf("Writing kube config to %s", kubePath)
	fmt.Println(*cluster)
	clusters := make(map[string]*api.Cluster)
	clusters[*cluster.Name] = &api.Cluster{
		Server:                   *cluster.Endpoint,
		CertificateAuthorityData: []byte(*cluster.CertificateAuthority.Data),
	}

	contexts := make(map[string]*api.Context)
	contexts[*cluster.Arn] = &api.Context{
		Cluster:  *cluster.Arn,
		AuthInfo: *cluster.Arn,
	}

	clientConfig := api.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: *cluster.Arn,
	}

	// prevConfigbytes, err := os.ReadFile(kubePath)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(prevConfigbytes)
	// err = os.MkdirAll(filepath.Dir(kubePath), os.ModePerm)
	// if err != nil {
	// 	return err
	// }

	// err = os.WriteFile(kubePath, prevConfigbytes, 0644)
	// if err != nil {
	// 	return err
	// }
	fmt.Println("Wrote kube config to ", kubePath)
	return clientcmd.WriteToFile(clientConfig, kubePath)
}

// CheckVpc asserts that the vpc is up and running. It tests the vpc state and id.
func checkVpc(name, vpcId string, svc ec2iface.EC2API) error {
	vpc, err := getVpc(name, svc)
	if err != nil {
		return fmt.Errorf("error describing vpc: %v", err)
	}

	if len(vpc.Vpcs) == 0 {
		return fmt.Errorf("no vpc found")
	}

	if vpc.Vpcs[0].VpcId != nil && *vpc.Vpcs[0].VpcId != vpcId {
		return fmt.Errorf("vpc id mismatch: %s != %s", *vpc.Vpcs[0].VpcId, vpcId)
	}

	if vpc.Vpcs[0].State != nil && *vpc.Vpcs[0].State != "available" {
		return fmt.Errorf("vpc not available: %s", *vpc.Vpcs[0].State)
	}

	return nil
}

func getVpc(name string, svc ec2iface.EC2API) (*ec2.DescribeVpcsOutput, error) {
	return svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Cluster"),
				Values: []*string{aws.String(name)},
			},
		},
	})
}

// applyTacticalPspFix deletes the current eks.privileged psp in the cluster.
// This allows the cluster to be created with a different psp. All pods are recycled
// so the new psp will be applied.
func applyTacticalPspFix(clientset kubernetes.Interface) error {
	// Delete the eks.privileged psp
	err := clientset.PolicyV1beta1().PodSecurityPolicies().Delete(context.TODO(), "eks.privileged", metav1.DeleteOptions{})
	// if the psp doesn't exist, we don't need to do anything
	if kubeErr.IsNotFound(err) {
		fmt.Println("No eks.privileged psp found, skipping")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to delete eks.privileged psp: %w", err)
	}

	// Get all pods in the cluster
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	// Delete all pods in the cluster
	for _, pod := range pods.Items {
		err = clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete pod: %w", err)
		}
	}

	return nil
}

// checkCluster checks the cluster is created and exists.
func checkCluster(name string, eks eksiface.EKSAPI) error {
	cluster, err := getCluster(name, eks)
	if err != nil {
		return err
	}

	if *cluster.Status != "ACTIVE" {
		return fmt.Errorf("cluster is not active")
	}

	return nil
}

func getCluster(name string, svc eksiface.EKSAPI) (*eks.Cluster, error) {
	cluster, err := svc.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	return cluster.Cluster, nil
}

func deleteLocalState(dir string, paths ...string) error {
	for _, path := range paths {
		p := strings.Join([]string{dir, path}, "/")
		if _, err := os.Stat(p); err == nil {
			err = os.RemoveAll(p)
			if err != nil {
				return fmt.Errorf("failed to delete local state: %w", err)
			}
		}
	}

	return nil
}
