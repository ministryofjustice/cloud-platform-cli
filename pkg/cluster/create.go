package cluster

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/ministryofjustice/cloud-platform-go-library/client"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOptions struct represents the options passed to the Create method.
type CreateOptions struct {
	// Name is the name of the cluster.
	Name string
	// ClusterSuffix is the suffix to append to the cluster name.
	// This will be used to create the cluster ingress, such as "live.service.justice.gov.uk".
	ClusterSuffix string

	// NodeCount is the number of nodes to create in the cluster.
	NodeCount int
	// VpcName is the name of the VPC to create the cluster in.
	// Often clusters will be built in a single VPC.
	VpcName string

	// MaxNameLength is the maximum length of the cluster name.
	// This limit exists due to the length of the name of the ingress.
	MaxNameLength int
	// TimeOut is the maximum time to wait for the cluster to be created.
	TimeOut int
	// Debug is true if the cluster should be created in debug mode.
	Debug bool
	// Fast creates the fastest possible cluster.
	Fast bool

	// Auth0 is the Auth0 domain and secret information.
	Auth0 AuthOpts
}

// TerraformOptions are the options to pass to Terraform plan and apply.
type TerraformOptions struct {
	// Apply allows you to group apply options passed to Terraform.
	ApplyVars []tfexec.ApplyOption
	// Plan allows you to group plan options passed to Terraform.
	// Plan []tfexec.PlanOption
	// Init allows you to group init options passed to Terraform.
	InitVars []tfexec.InitOption
	// PlanPath is a string of the path to the Terraform plan file.
	// This is used to both save the output of plan and to pass to apply.
	// PlanPath string
	// Version is the version of Terraform to use.
	Version string
	// ExecPath is the path to the Terraform executable.
	ExecPath string
	// Workspace is the name of the Terraform workspace to use.
	Workspace string
	// FilePath is the location of the cloud-platform-infrastructure repository.
	// This repository contains all the Terraform used to create the cluster.
	FilePath     string
	DirStructure TerraformDirStructure
}

type TerraformDirStructure struct {
	// ComponentDir is the directory path in the cloud-platform-infrastructure repository.
	ComponentDir string
	// VpcDir is the directory path in the cloud-platform-infrastructure repository.
	VpcDir string
	// ClusterDir is the directory path in the cloud-platform-infrastructure repository.
	ClusterDir string
	// Directories is a slice of strings of the directory paths in the cloud-platform-infrastructure repository.
	Directories []string
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

// Create creates a new Kubernetes cluster using the options passed to it.
func (terraform *TerraformOptions) CreateCluster(options *CreateOptions, awsCred *AwsCredentials) error {
	fmt.Println("Creating cluster", options.Name)
	if err := terraform.setup(); err != nil {
		return fmt.Errorf("failed to setup terraform: %w", err)
	}

	if err := terraform.build(awsCred, options.Fast); err != nil {
		return fmt.Errorf("failed to run terraform: %w", err)
	}

	return nil
}

func (terraform *TerraformOptions) setup() error {
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

	structure := TerraformDirStructure{
		VpcDir:       vpcDir,
		ClusterDir:   clusterDir,
		ComponentDir: componentsDir,
		Directories:  directories,
	}

	terraform.DirStructure = structure

	err := terraform.CreateTerraformObj()
	if err != nil {
		return fmt.Errorf("error creating terraform obj: %s", err)
	}
	return nil
}

func (terraform *TerraformOptions) build(creds *AwsCredentials, fast bool) error {
	for _, dir := range terraform.DirStructure.Directories {
		fmt.Println("Applying in dir", dir)
		if dir == terraform.DirStructure.ComponentDir {
			err := authToCluster(terraform.Workspace)
			if err != nil {
				return fmt.Errorf("error authenticating to cluster: %s", err)
			}
		}
		err := terraform.apply(dir, creds, fast)
		if err != nil {
			return fmt.Errorf("error applying terraform in dir: %s %s", dir, err)
		}
	}

	return nil
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

func (terraform *TerraformOptions) InitAndApply(tf *tfexec.Terraform, creds *AwsCredentials, fast bool) error {
	fmt.Printf("Init terraform on directory %s\n", tf.WorkingDir())
	err := terraform.Initialise(tf, creds)
	if err != nil {
		return fmt.Errorf("failed to init terraform: %w", err)
	}

	fmt.Printf("Apply terraform on directory %s\n", tf.WorkingDir())
	err = terraform.ExecuteApply(tf, fast)
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

func (terraform *TerraformOptions) HealthCheck(tf *tfexec.Terraform, creds *AwsCredentials) error {
	// switch case for checking which directory we are in
	fmt.Println(tf.WorkingDir())
	switch {
	case tf.WorkingDir() == "./terraform/aws-accounts/cloud-platform-aws/":
		err := terraform.checkVpc(tf, creds.Session)
		if err != nil {
			return fmt.Errorf("failed to check vpc: %w", err)
		}
	case strings.Contains(tf.WorkingDir(), "eks"):
		err := checkCluster(terraform.Workspace, creds.Session)
		if err != nil {
			return fmt.Errorf("failed to check cluster: %w", err)
		}
	case tf.WorkingDir() == "./terraform/aws-accounts/cloud-platform-aws/":
		fmt.Println("Applying tactical fix to get psp working")
		if err := applyTacticalPspFix(); err != nil {
			return fmt.Errorf("failed to apply tactical psp fix: %w", err)
		}
	}
	return nil
}

func (terraform *TerraformOptions) ApplyAndCheck(tf *tfexec.Terraform, creds *AwsCredentials, fast bool) error {
	err := terraform.InitAndApply(tf, creds, fast)
	if err != nil {
		return fmt.Errorf("an error occurred attempting to init and apply %w", err)
	}

	err = terraform.HealthCheck(tf, creds)
	if err != nil {
		return fmt.Errorf("an error occurred attempting to perform a healthcheck %w", err)
	}

	return nil
}

// intialise performs the `terraform init` function.
func (terraform *TerraformOptions) Initialise(tf *tfexec.Terraform, creds *AwsCredentials) error {
	err := tf.Init(context.TODO())
	if err != nil {
		return err
	}
	return terraform.SetWorkspace(tf, creds)
}

func (terraform *TerraformOptions) SetWorkspace(tf *tfexec.Terraform, creds *AwsCredentials) error {
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

func (terraform *TerraformOptions) ExecuteApply(tf *tfexec.Terraform, fast bool) error {
	// I've found to ensure parallelism you need to execute init once more.
	terraform.InitVars = append(terraform.InitVars, tfexec.Reconfigure(true))
	err := tf.Init(context.TODO(), terraform.InitVars...)
	if err != nil {
		return fmt.Errorf("failed to init: %w", err)
	}

	if strings.Contains(tf.WorkingDir(), "eks") && fast {
		terraform.ApplyVars = append(terraform.ApplyVars, tfexec.Var(fmt.Sprintf("%s=%v", "enable_oidc_associate", false)))
	}

	err = tf.Apply(context.TODO(), terraform.ApplyVars...)
	// handle a case where you need to init again
	if err != nil {
		if errors.As(err, &tfexec.ErrNoInit{}) {
			fmt.Println("Init again, due to failure")
			err = tf.Init(context.TODO(), terraform.InitVars...)
			if err != nil {
				return fmt.Errorf("failed to init: %w", err)
			}
			err = tf.Apply(context.TODO(), terraform.ApplyVars...)
			if err != nil {
				return fmt.Errorf("failed to apply: %w", err)
			}
			return err
		}
	}

	return nil
}

// CheckVpc asserts that the vpc is up and running. It tests the vpc state and id.
func (terraform *TerraformOptions) checkVpc(tf *tfexec.Terraform, sess *session.Session) error {
	output, err := terraform.output(tf)
	if err != nil {
		return fmt.Errorf("failed to get output: %w", err)
	}

	vpcID := output["vpc_id"]
	if vpcID.Value == nil {
		return errors.New("vpc_id not found in terraform output")
	}

	// Trim the vpcId to remove quotes
	trimVpc := strings.Trim(string(vpcID.Value), "\"")
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

	if vpc.Vpcs[0].VpcId != nil && *vpc.Vpcs[0].VpcId != trimVpc {
		return fmt.Errorf("vpc id mismatch: %s != %s", *vpc.Vpcs[0].VpcId, trimVpc)
	}

	if vpc.Vpcs[0].State != nil && *vpc.Vpcs[0].State != "available" {
		return fmt.Errorf("vpc not available: %s", *vpc.Vpcs[0].State)
	}

	return nil
}

func (terraform *TerraformOptions) apply(directory string, creds *AwsCredentials, fast bool) error {
	fmt.Println("Creating terraform object")
	tf, err := tfexec.NewTerraform(directory, terraform.ExecPath)
	if err != nil {
		return fmt.Errorf("failed to create terraform object: %w", err)
	}

	tf.SetStdout(log.Writer())
	tf.SetStderr(log.Writer())

	err = deleteLocalState(directory, ".terraform", ".terraform.lock.hcl")
	if err != nil {
		return fmt.Errorf("failed to delete temp directory: %w", err)
	}

	return terraform.ApplyAndCheck(tf, creds, fast)
}

// applyTacticalPspFix deletes the current eks.privileged psp in the cluster.
// This allows the cluster to be created with a different psp. All pods are recycled
// so the new psp will be applied.
func applyTacticalPspFix() error {
	client, err := client.NewKubeClientWithValues("", "")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Delete the eks.privileged psp
	err = client.Clientset.PolicyV1beta1().PodSecurityPolicies().Delete(context.TODO(), "eks.privileged", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete eks.privileged psp: %w", err)
	}

	// Get all pods in the cluster
	pods, err := client.Clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	// Delete all pods in the cluster
	for _, pod := range pods.Items {
		err = client.Clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete pod: %w", err)
		}
	}

	return nil
}

// checkCluster checks the cluster is created and exists.
func checkCluster(name string, session *session.Session) error {
	// Check the cluster is created and exists.
	cluster, err := getCluster(name, session)
	if err != nil {
		return err
	}

	if *cluster.Status != "ACTIVE" {
		return fmt.Errorf("cluster is not active")
	}

	return nil
}

func getCluster(name string, session *session.Session) (*eks.Cluster, error) {
	svc := eks.New(session)
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
		path := strings.Join([]string{dir, path}, "/")
		if _, err := os.Stat(path); err == nil {
			err = os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("failed to delete local state: %w", err)
			}
		}
	}

	return nil
}
