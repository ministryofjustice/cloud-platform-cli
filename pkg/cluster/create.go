package cluster

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	cpclient "github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
)

// CreateOptions struct represents the options passed to the Create method
// by the `cloud-platform create cluster` command.
type CreateOptions struct {
	// Name is the name of the cluster you wish to create/amend.
	Name string
	// ClusterSuffix is the suffix to append to the cluster name.
	// This will be used to create the cluster ingress, such as "live.service.justice.gov.uk".
	ClusterSuffix string

	// NodeCount is the number of nodes to create in the new cluster.
	NodeCount int
	// VpcName is the name of the VPC to create the cluster in.
	// Often clusters will be built in a single VPC.
	VpcName string

	// MaxNameLength is the maximum length of the cluster name.
	// This limit exists due to the length of the name of the ingress.
	MaxNameLength int
	// Debug is true if the cluster should be created in debug mode.
	Debug bool
	// Fast creates the fastest possible cluster.
	Fast bool

	// TfVersion is the version of Terraform to use to create the cluster and components.
	TfVersion string
	// TfDirectories is a list of directories to run Terraform in.
	TfDirectories []string

	// Auth0 the Auth0 client ID and secret to use for the cluster.
	Auth0 AuthOpts
}

// TerraformOptions are the options to pass to Terraform plan and apply.
type TerraformOptions struct {
	// Apply allows you to group apply options passed to Terraform.
	ApplyVars []tfexec.ApplyOption
	// Init allows you to group init options passed to Terraform.
	InitVars []tfexec.InitOption
	// Version is the version of Terraform to use.
	Version string
	// ExecPath is the path to the Terraform executable.
	ExecPath string
	// Workspace is the name of the Terraform workspace to use.
	Workspace string
	// FilePath is the location of the cloud-platform-infrastructure repository.
	// This repository contains all the Terraform used to create the cluster.
	FilePath     string
	DirStructure DirectoryStructure
}

// DirectoryStructure represents the directory structure of the terraform to install.
// This is fairly unique to the cloud-platform-infrastructure repository, but can be amended.
// The structure is as follows:
// - VpcDir == "vpc"
// - ClusterDir == "EKS cluster"
// - ComponentsDir == "kubernetes components that bootstrap a cloud-platform cluster"
type DirectoryStructure struct {
	// ComponentDir is the directory path in the cloud-platform-infrastructure repository.
	ComponentDir string
	// VpcDir is the directory path in the cloud-platform-infrastructure repository.
	VpcDir string
	// ClusterDir is the directory path in the cloud-platform-infrastructure repository.
	ClusterDir string
	// List is a collection of directory paths in the cloud-platform-infrastructure repository.
	List []string
}

// AuthOpts represents the options for Auth0.
type AuthOpts struct {
	// Domain is the Auth0 domain.
	Domain string
	// ClientID is the Auth0 client ID.
	ClientId string
	// ClientSecret is the Auth0 client secret.
	ClientSecret string
}

// Build is the entrypoint for the `cloud-platform create cluster` command. This method takes
// a cluster type and a set of options, and creates a cluster.
func (cluster *Cluster) Build(options *CreateOptions, credentials *AwsCredentials) error {
	// Set environment
	fmt.Println("Setting environment variables")
	// if err := setEnvironment(options, credentials); err != nil {
	// 	return err
	// }

	// Create Terraform object
	fmt.Println("Creating Terraform object")
	terraform, err := newTerraformOptions(options)
	if err != nil {
		return err
	}

	fmt.Println("Creating cluster")
	if err := cluster.create(terraform, credentials, options.Fast); err != nil {
		return fmt.Errorf("failed to run terraform: %w", err)
	}

	return nil
}

func (cluster *Cluster) create(terraform *TerraformOptions, creds *AwsCredentials, fast bool) error {
	fmt.Println("Creating VPC")
	vpc, err := terraform.applyVpc(creds)
	if err != nil {
		return fmt.Errorf("failed to create VPC: %w", err)
	}

	cluster.VpcId = vpc

	fmt.Println("Creating EKS")
	err = terraform.applyEks(creds, fast)
	if err != nil {
		return fmt.Errorf("failed to create EKS: %w", err)
	}

	err = terraform.applyComponents(creds)
	if err != nil {
		return fmt.Errorf("failed to apply components: %w", err)
	}

	return nil
}

func (terraform *TerraformOptions) applyComponents(creds *AwsCredentials) error {
	// There is a requirement for the aws binary to exist at this point.
	if err := authToCluster(terraform.Workspace); err != nil {
		return err
	}

	_, err := terraform.apply(terraform.DirStructure.ComponentDir, creds, false)
	if err != nil {
		return err
	}

	// This makes a rather large assumption that the users kubeconfig is in the default location.
	// TODO: Have this passed as an argument.
	var path string
	if home := homedir.HomeDir(); home != "" {
		path = filepath.Join(home, ".kube", "config")
	}

	kube, err := cpclient.GetClientset(path)
	if err != nil {
		return err
	}

	if err := applyTacticalPspFix(kube); err != nil {
		return err
	}

	return nil
}
func (terraform *TerraformOptions) applyEks(creds *AwsCredentials, fast bool) error {
	fmt.Println("Applying Terraform against EKS")
	_, err := terraform.apply(terraform.DirStructure.ClusterDir, creds, fast)
	if err != nil {
		return err
	}

	if err := checkCluster(terraform.Workspace, creds.Eks); err != nil {
		return err
	}

	return nil
}

func (terraform *TerraformOptions) applyVpc(creds *AwsCredentials) (string, error) {
	output, err := terraform.apply(terraform.DirStructure.VpcDir, creds, false)
	if err != nil {
		return "", err
	}

	vpcID := output["vpc_id"]
	if vpcID.Value == nil {
		return "", errors.New("vpc_id not found in terraform output")
	}

	fmt.Println("Starting to check vpc")
	// Trim the vpcId to remove quotes
	vpc := strings.Trim(string(vpcID.Value), "\"")
	return vpc, terraform.checkVpc(vpc, creds.Session)
}

func newTerraformOptions(options *CreateOptions) (*TerraformOptions, error) {
	if options.TfVersion == "" {
		return nil, errors.New("terraform version is required")
	}
	tf := &TerraformOptions{
		Version:   options.TfVersion,
		Workspace: options.Name,
	}
	if options.TfDirectories == nil {
		tf.DefaultDirSetup()
	} else {
		for _, dir := range options.TfDirectories {
			tf.DirStructure.List = append(tf.DirStructure.List, dir)
		}
	}

	if err := tf.CreateTerraformObj(); err != nil {
		return nil, err
	}

	return tf, nil
}

// DefaultDirSetup sets the default directory structure for the cloud-platform-infrastructure repository.
func (terraform *TerraformOptions) DefaultDirSetup() {
	const baseDir = "./terraform/aws-accounts/cloud-platform-aws/"
	var (
		vpcDir        = baseDir + "vpc/"
		clusterDir    = vpcDir + "eks/"
		componentsDir = clusterDir + "components/"
	)
	directories := []string{
		vpcDir,
		clusterDir,
		componentsDir,
	}

	structure := DirectoryStructure{
		VpcDir:       vpcDir,
		ClusterDir:   clusterDir,
		ComponentDir: componentsDir,
		List:         directories,
	}
	terraform.DirStructure = structure
}

// createTerraformObj creates a Terraform object using the version passed as a string.
func (terraform *TerraformOptions) CreateTerraformObj() error {
	i := install.NewInstaller()
	v := version.Must(version.NewVersion(terraform.Version))

	execPath, err := i.Ensure(context.TODO(), []src.Source{
		&fs.ExactVersion{
			Product: product.Terraform,
			Version: v,
		},
		&releases.ExactVersion{
			Product: product.Terraform,
			Version: v,
		},
	})
	if err != nil {
		return err
	}

	defer i.Remove(context.TODO())

	terraform.ExecPath = execPath
	terraform.ApplyVars = []tfexec.ApplyOption{
		tfexec.Parallelism(1),
	}

	return nil
}

func (terraform *TerraformOptions) initAndApply(tf *tfexec.Terraform, creds *AwsCredentials, fast bool) error {
	fmt.Printf("Init terraform in directory %s\n", tf.WorkingDir())
	err := terraform.initialise(tf)
	if err != nil {
		return fmt.Errorf("failed to init terraform: %w", err)
	}

	fmt.Printf("Apply terraform in directory %s\n", tf.WorkingDir())
	err = terraform.executeApply(tf, fast)
	if err != nil {
		return fmt.Errorf("failed to apply: %w", err)
	}

	return nil
}

func (terraform *TerraformOptions) output(tf *tfexec.Terraform) (map[string]tfexec.OutputMeta, error) {
	// We don't want terraform to print out the output here as the package doesn't respect the secret flag.
	tf.SetStdout(nil)
	tf.SetStderr(nil)
	output, err := tf.Output(context.TODO())
	if err != nil {
		if strings.Contains(err.Error(), "plugin") || strings.Contains(err.Error(), "init") {
			fmt.Println("Init again, due to failure")
			err = tf.Init(context.TODO(), terraform.InitVars...)
			if err != nil {
				return nil, fmt.Errorf("failed to init: %w", err)
			}
			output, err = tf.Output(context.TODO())
			if err != nil {
				return nil, fmt.Errorf("failed to create output: %w", err)
			}
			return nil, fmt.Errorf("failed to show terraform output: %w", err)
		}
	}
	return output, nil
}

// intialise performs the `terraform init` function.
func (terraform *TerraformOptions) initialise(tf *tfexec.Terraform) error {
	terraform.InitVars = append(terraform.InitVars, tfexec.Reconfigure(true))
	err := tf.Init(context.TODO(), terraform.InitVars...)
	// Handle no plugin error
	if err != nil {
		if err := tf.Init(context.TODO(), terraform.InitVars...); err != nil {
			return fmt.Errorf("failed to init: %w", err)
		}
	}

	return terraform.setWorkspace(tf)
}

func (terraform *TerraformOptions) setWorkspace(tf *tfexec.Terraform) error {
	list, _, err := tf.WorkspaceList(context.TODO())
	if err != nil {
		return err
	}

	for _, ws := range list {
		if ws == terraform.Workspace {
			err = tf.WorkspaceSelect(context.TODO(), terraform.Workspace)
			if err != nil {
				return err
			}
			return nil
		}
	}

	err = tf.WorkspaceNew(context.TODO(), terraform.Workspace)
	if err != nil {
		return err
	}

	return nil
}

func authToCluster(cluster string) error {
	_, err := exec.Command("aws", "eks", "--region", "eu-west-2", "update-kubeconfig", "--name", cluster).Output()
	if err != nil {
		return err
	}

	return nil
}

func (terraform *TerraformOptions) executeApply(tf *tfexec.Terraform, fast bool) error {
	if strings.Contains(tf.WorkingDir(), "eks") && fast {
		terraform.ApplyVars = append(terraform.ApplyVars, tfexec.Var(fmt.Sprintf("%s=%v", "enable_oidc_associate", false)))
	}

	err := tf.Apply(context.TODO(), terraform.ApplyVars...)
	// handle a case where you need to init again
	if err != nil {
		fmt.Println("Init again, due to failure")
		err = tf.Init(context.TODO(), terraform.InitVars...)
		if err != nil {
			return fmt.Errorf("failed to init: %w", err)
		}
		err = tf.Apply(context.TODO(), terraform.ApplyVars...)
		if err != nil {
			return fmt.Errorf("failed to apply: %w", err)
		}
	}

	return nil
}

// CheckVpc asserts that the vpc is up and running. It tests the vpc state and id.
func (terraform *TerraformOptions) checkVpc(vpcId string, sess *session.Session) error {
	svc := ec2.New(sess)

	vpc, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Cluster"),
				Values: []*string{aws.String(terraform.Workspace)},
			},
		},
	})
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

func (terraform *TerraformOptions) apply(directory string, creds *AwsCredentials, fast bool) (map[string]tfexec.OutputMeta, error) {
	tf, err := tfexec.NewTerraform(directory, terraform.ExecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform object: %w", err)
	}
	fmt.Println(tf)

	// Write the output to the terminal.
	tf.SetStdout(log.Writer())
	tf.SetStderr(log.Writer())

	// Start fresh and remove any local state.
	if err = deleteLocalState(directory, ".terraform", ".terraform.lock.hcl"); err != nil {
		return nil, fmt.Errorf("failed to delete temp directory: %w", err)
	}

	err = terraform.initAndApply(tf, creds, fast)
	if err != nil {
		return nil, fmt.Errorf("an error occurred attempting to init and apply %w", err)
	}

	output, err := terraform.output(tf)
	if err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}

	return output, nil
}

// applyTacticalPspFix deletes the current eks.privileged psp in the cluster.
// This allows the cluster to be created with a different psp. All pods are recycled
// so the new psp will be applied.
func applyTacticalPspFix(clientset kubernetes.Interface) error {
	// Delete the eks.privileged psp
	err := clientset.PolicyV1beta1().PodSecurityPolicies().Delete(context.TODO(), "eks.privileged", metav1.DeleteOptions{})
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

func setEnvironment(options *CreateOptions, cred *AwsCredentials) error {
	// Set environment variables
	if err := os.Setenv("AWS_PROFILE", cred.Profile); err != nil {
		return fmt.Errorf("error setting AWS_PROFILE: %s", err)
	}

	if err := os.Setenv("AWS_REGION", cred.Region); err != nil {
		return fmt.Errorf("error setting AWS_REGION: %s", err)
	}

	if err := os.Setenv("AUTH0_DOMAIN", options.Auth0.Domain); err != nil {
		return fmt.Errorf("error setting AUTH0_DOMAIN: %s", err)
	}

	if err := os.Setenv("AUTH0_CLIENT_ID", options.Auth0.ClientId); err != nil {
		return fmt.Errorf("error setting AUTH0_CLIENT_ID: %s", err)
	}

	if err := os.Setenv("AUTH0_CLIENT_SECRET", options.Auth0.ClientSecret); err != nil {
		return fmt.Errorf("error setting AUTH0_CLIENT_SECRET: %s", err)
	}

	return nil
}
