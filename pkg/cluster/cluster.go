package cluster

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/go-git/go-git/v5" // with go modules enabled (GO111MODULE=on or outside GOPATH)

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"

	v1 "k8s.io/api/core/v1"
)

// Cluster represents useful values and object in a Kubernetes cluster
type Cluster struct {
	Name       string
	VpcId      string
	Nodes      []v1.Node
	OldestNode v1.Node
	NewestNode v1.Node

	Pods      []v1.Pod
	StuckPods []v1.Pod

	Namespaces v1.NamespaceList
}

// Snapshot represents a snapshot of a Kubernetes cluster object
type Snapshot struct {
	Cluster Cluster
}

// AwsCredentials represents the AWS credentials used to connect to an AWS account.
type AwsCredentials struct {
	Session *session.Session
	Profile string
	Region  string
}

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
	// AwsCredentials contains the AWS credentials to use when creating the cluster.
	// AwsCredentials AwsCredentials

	// TerraformOptions are the options to pass to Terraform plan and apply.
	// TerraformOptions TerraformOptions
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
	// FilePath is the location of the cloud-platform-infrastructure reporisitory.
	// This reporisitory contains all the Terraform used to create the cluster.
	FilePath string
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

// NewCluster creates a new Cluster object and populates its
// fields with the values from the Kubernetes cluster in the client passed to it.
func NewCluster(c *client.Client) (*Cluster, error) {
	pods, err := getPods(c)
	if err != nil {
		return nil, err
	}

	nodes, err := GetAllNodes(c)
	if err != nil {
		return nil, err
	}

	oldestNode, err := getOldestNode(c)
	if err != nil {
		return nil, err
	}

	newestNode, err := GetNewestNode(c, nodes)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		Name:       nodes[0].Labels["Cluster"],
		Pods:       pods,
		Nodes:      nodes,
		OldestNode: oldestNode,
		NewestNode: newestNode,
	}, nil
}

// NewSnapshot constructs a Snapshot cluster object
func (c *Cluster) NewSnapshot() *Snapshot {
	return &Snapshot{
		Cluster: *c,
	}
}

// NewAwsCredentials constructs and populates a new AwsCredentials object
func NewAwsCreds(region string) (*AwsCredentials, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}

	return &AwsCredentials{
		Session: sess,
		Region:  region,
	}, nil
}

// RefreshStatus performs a value overwrite of the cluster status.
// This is useful for when the cluster is being updated.
func (c *Cluster) RefreshStatus(client *client.Client) (err error) {
	c.Nodes, err = GetAllNodes(client)
	if err != nil {
		return err
	}

	c.OldestNode, err = getOldestNode(client)
	if err != nil {
		return err
	}
	return nil
}

// Create creates a new Kubernetes cluster using the options passed to it.
func (cluster *Cluster) Create(opts *CreateOptions, terraform *TerraformOptions, awsCred *AwsCredentials) error {
	// Clone the repository to a directory.
	fmt.Println("Cloning repository")
	_, err := git.PlainClone("./tmp", false, &git.CloneOptions{
		URL:      "https://" + terraform.FilePath,
		Progress: os.Stdout,
	})
	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			fmt.Println("Repository already exists")
		} else {
			return err
		}
	}
	defer os.RemoveAll("./tmp")

	// Checkout the correct branch.
	// wt, err := repo.Worktree()
	// if err != nil {
	// 	return fmt.Errorf("Failed to get worktree: %s", err)
	// }

	// fmt.Println("Checking out branch")
	// branch := "create-cluster"
	// err = wt.Checkout(&git.CheckoutOptions{
	// 	Branch: plumbing.NewBranchReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
	// 	Force:  true,
	// })
	// if err != nil {
	// 	return err
	// }

	fmt.Println("Creating cluster")
	err = terraform.Run(awsCred, opts.Fast)
	if err != nil {
		return err
	}

	// TODO: Build the cluster object and perform general cluster readiness checks.
	// TODO: Display a nice table of the cluster status.

	return nil
}

func (terraform *TerraformOptions) Run(creds *AwsCredentials, fast bool) error {
	// Directory paths in the cloud-platform-infrastructure repository.
	const baseDir = "./tmp/terraform/aws-accounts/cloud-platform-aws/"
	var (
		vpcDir        = baseDir + "vpc/"
		clusterDir    = vpcDir + "eks/"
		componentsDir = clusterDir + "components/"
	)

	directories := []string{
		vpcDir,
		clusterDir,
		// componentsDir,
	}

	// TODO: Remove prints
	fmt.Println("Creating directories...", vpcDir, clusterDir, componentsDir)

	// Create Terraform object to use throught method.
	fmt.Println("Creating Terraform object")
	err := terraform.CreateTerraformObj()
	if err != nil {
		return fmt.Errorf("error creating terraform obj: %s", err)
	}

	for _, dir := range directories {
		fmt.Println("Applying in dir", dir)
		err := terraform.Apply(dir, creds, fast)
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

	execPath, err := i.Ensure(context.Background(), []src.Source{
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

	defer i.Remove(context.Background())

	terraform.ExecPath = execPath
	terraform.ApplyVars = []tfexec.ApplyOption{
		tfexec.Parallelism(1),
	}

	return nil
}

func (terraform *TerraformOptions) ApplyAndCheck(tf *tfexec.Terraform, creds *AwsCredentials, fast bool) error {
	fmt.Println("Init terraform")
	err := terraform.Initialise(tf)
	if err != nil {
		return fmt.Errorf("failed to init terraform: %w", err)
	}

	fmt.Println("Apply terraform")
	err = terraform.ExecuteApply(tf, fast)
	if err != nil {
		return fmt.Errorf("failed to apply: %w", err)
	}

	fmt.Println("Show terraform")
	state, err := tf.Show(context.Background())
	if err != nil {
		return fmt.Errorf("failed to show: %w", err)
	}

	fmt.Println("Check terraform")
	// switch case for checking which directory we are in
	switch {
	case strings.Contains(tf.WorkingDir(), "vpc"):
		vpcEndpointId, err := getVpcFromState(state)
		if err != nil {
			return fmt.Errorf("failed to get vpc endpoint id: %w", err)
		}
		err = checkVpc(vpcEndpointId, terraform.Workspace, creds.Session)
		if err != nil {
			return fmt.Errorf("failed to check vpc: %w", err)
		}
		fmt.Println("Check complete")
	case strings.Contains(tf.WorkingDir(), "eks"):
		err := checkCluster(terraform.Workspace, state, creds.Session)
		if err != nil {
			return fmt.Errorf("failed to check cluster: %w", err)
		}
		// TODO: Check cluster health
	case strings.Contains(tf.WorkingDir(), "components"):
		// TODO: Check components health
	}

	return nil
}

// intialise performs the `terraform init` function.
func (terraform *TerraformOptions) Initialise(tf *tfexec.Terraform) error {
	err := tf.Init(context.Background())
	if err != nil {
		return err
	}
	return terraform.SetWorkspace(tf)
}

func (terraform *TerraformOptions) SetWorkspace(tf *tfexec.Terraform) error {
	list, _, err := tf.WorkspaceList(context.Background())

	for _, ws := range list {
		if ws == terraform.Workspace {
			err = tf.WorkspaceSelect(context.Background(), terraform.Workspace)
			if err != nil {
				return err
			}
			return nil
		}
	}

	err = tf.WorkspaceNew(context.Background(), terraform.Workspace)
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
		terraform.ApplyVars = append(terraform.ApplyVars, tfexec.Var(fmt.Sprintf("%s=%v", "auth0_count", false)))
		terraform.ApplyVars = append(terraform.ApplyVars, tfexec.Var(fmt.Sprintf("%s=%v", "aws_eks_identity_provider_config_oidc_associate", false)))
	}

	err = tf.Apply(context.TODO(), terraform.ApplyVars...)
	// handle a case where you need to init again
	if err != nil {
		if strings.Contains(err.Error(), "init") {
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
func checkVpc(vpcId, workspace string, sess *session.Session) error {
	svc := ec2.New(sess)

	vpc, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Cluster"),
				Values: []*string{aws.String(workspace)},
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

func (terraform *TerraformOptions) Apply(directory string, creds *AwsCredentials, fast bool) error {
	fmt.Println("Creating terraform object")
	tf, err := tfexec.NewTerraform(directory, terraform.ExecPath)
	if err != nil {
		return fmt.Errorf("failed to create terraform object: %w", err)
	}

	tf.SetStdout(log.Writer())
	tf.SetStderr(log.Writer())
	fmt.Println("object looks like:", tf)

	// if .terraform.tfstate directory exists, delete it
	err = deleteLocalState(strings.Join([]string{directory, ".terraform"}, "/"))
	if err != nil {
		return fmt.Errorf("failed to delete .terraform.tfstate directory: %w", err)
	}

	return terraform.ApplyAndCheck(tf, creds, fast)
}

// ApplyTacticalPspFix deletes the current eks.privileged psp in the cluster.
// This allows the cluster to be created with a different psp. All pods are recycled
// so the new psp will be applied.
// func (c *Cluster) ApplyTacticalPspFix() error {
// 	client, err := client.NewKubeClientWithValues("", "")
// 	if err != nil {
// 		return fmt.Errorf("failed to create kubernetes client: %w", err)
// 	}

// 	// Delete the eks.privileged psp
// 	err = client.Clientset.PolicyV1beta1().PodSecurityPolicies().Delete(context.Background(), "eks.privileged", metav1.DeleteOptions{})
// 	if err != nil {
// 		return fmt.Errorf("failed to delete eks.privileged psp: %w", err)
// 	}

// 	// Delete all pods in the cluster
// 	err = client.Clientset.CoreV1().Pods("").DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{})
// 	if err != nil {
// 		return fmt.Errorf("failed to recycle pods: %w", err)
// 	}

// 	return nil
// }

// checkCluster checks the cluster is created and exists.
func checkCluster(name string, state *tfjson.State, session *session.Session) error {
	// Check the cluster is created and exists.
	cluster, err := getCluster(name, session)
	if err != nil {
		return err
	}

	if cluster.Status != aws.String("ACTIVE") {
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

func getVpcFromState(state *tfjson.State) (string, error) {
	var vpcEndpointId string
	for k, v := range state.Values.Outputs {
		if k == "vpc_id" {
			vpcEndpointId = v.Value.(string)
		}
	}
	if vpcEndpointId == "" {
		return "", fmt.Errorf("failed to find vpc endpoint id")
	}

	return vpcEndpointId, nil
}

func deleteLocalState(path string) error {
	if _, err := os.Stat(path); err == nil {
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	return nil
}
