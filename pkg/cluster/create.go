package cluster

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
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
// You can make the eks terraform creation faster by passing the faenable_oidc_associateenable_oidc_associatest flag.
func (c *Cluster) ApplyEks(tf *terraform.TerraformCLIConfig, creds *client.AwsCredentials, dir string, fast bool) error {
	tf.WorkingDir = dir
	if fast {
		tf.ApplyVars = append(tf.ApplyVars, tfexec.Var(fmt.Sprintf("%s=%v", "enable_oidc_associate", false)))
	}

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
func (c *Cluster) ApplyComponents(tf *terraform.TerraformCLIConfig, awsCreds *client.AwsCredentials, dir, kubeconf string) error {
	// Reset any previous varibles that might've been set.
	tf.ApplyVars = nil

	// Turn the monitoring options off.
	vars := []string{
		fmt.Sprintf("%s=%s", "pagerduty_config", "dummydummy"),
		fmt.Sprintf("%s=%s", "slack_hook_id", "dummydummy"),
	}
	for _, v := range vars {
		tf.ApplyVars = append(tf.ApplyVars, tfexec.Var(v))
	}

	// Auth to the cluster and write the kubeconfig to disk.
	_, err := AuthToCluster(tf.Workspace, awsCreds.Eks, kubeconf, awsCreds.Profile)
	if err != nil {
		return fmt.Errorf("failed to auth to cluster: %w", err)
	}

	tf.WorkingDir = dir

	_, err = terraformApply(tf)
	if err != nil {
		return err
	}

	kube, err := client.NewKubeClient(kubeconf)
	if err != nil {
		return err
	}

	nodes, err := GetAllNodes(kube)
	if err != nil {
		return err
	}
	c.Nodes = nodes

	if err := c.GetStuckPods(kube); err != nil {
		return err
	}

	return nil
}

func terraformApply(tf *terraform.TerraformCLIConfig) (map[string]tfexec.OutputMeta, error) {
	tfCli, error := terraform.NewTerraformCLI(tf)
	if error != nil {
		return nil, error
	}

	// Start fresh and remove any local state.
	if err := terraform.DeleteLocalState(tf.WorkingDir, ".terraform", ".terraform.lock.hcl"); err != nil {
		fmt.Println("Failed to delete local state, continuing anyway")
	}

	err := tfCli.Init(context.Background(), log.Writer())
	if err != nil {
		return nil, fmt.Errorf("failed to init terraform: %w", err)
	}

	if err = tfCli.Apply(context.Background(), log.Writer()); err != nil {
		return nil, fmt.Errorf("failed to apply terraform: %w", err)
	}

	// We pass a nil writer to the output command as we don't want to print the output to the console.
	return tfCli.Output(context.Background(), nil)
}

// AuthToCluster will authenticate to the cluster and return a kubernetes clientset. It will also write the kubeconfig
// and set the current context to the eks cluster passed to it.
func AuthToCluster(name string, eksSvc eksiface.EKSAPI, path string, awsProfile string) (*kubernetes.Clientset, error) {
	input := &eks.DescribeClusterInput{
		Name: aws.String(name),
	}
	result, err := eksSvc.DescribeCluster(input)
	if err != nil {
		log.Fatalf("Error calling DescribeCluster: %v", err)
	}

	clientset, err := newClientset(result.Cluster, path, awsProfile)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}
	return clientset, nil
}

func newClientset(cluster *eks.Cluster, path, awsProfile string) (*kubernetes.Clientset, error) {
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

	if err := writeKubeConfig(cluster, path, awsProfile, tok, ca); err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func writeKubeConfig(cluster *eks.Cluster, path, profile string, tok token.Token, ca []byte) error {
	if path == "" {
		return errors.New("kubeconfig path is empty")
	}

	if ca == nil {
		return errors.New("ca is empty")
	}

	kc := api.Config{
		Clusters: map[string]*api.Cluster{
			*cluster.Name: {
				Server:                   *cluster.Endpoint,
				CertificateAuthorityData: ca,
			},
		},
		Contexts: map[string]*api.Context{
			*cluster.Name: {
				Cluster:  *cluster.Name,
				AuthInfo: *cluster.Name,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			*cluster.Name: {
				Token: tok.Token,
			},
		},
		CurrentContext: *cluster.Name,
	}

	// write kubeconfig to disk
	return clientcmd.WriteToFile(kc, path)
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
