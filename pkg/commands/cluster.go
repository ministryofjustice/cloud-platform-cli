package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	cloudPlatform "github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/recycle"
	terraform "github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// Aws credentials are required to communicate with the aws api.kube
// This is a global variable so it can be used in other commands.
var (
	awsSecret, awsAccessKey, awsProfile, awsRegion string
)

// kubePath is the path to the kubeconfig file
var kubePath string

// CreateOptions struct represents the options passed to the Create method
// by the `cloud-platform create cluster` command.
type clusterOptions struct {
	// Name is the name of the cluster you wish to create/amend.
	Name string
	// ClusterSuffix is the suffix to append to the cluster name.
	// This will be used to create the cluster ingress, such as "live.service.justice.gov.uk".
	ClusterSuffix string
	// VpcName is the name of the VPC to create the cluster in.
	// Often clusters will be built in a single VPC.
	VpcName string

	// MaxNameLength is the maximum length of the cluster name.
	// This limit exists due to the length of the name of the ingress.
	MaxNameLength int

	// TfVersion is the version of Terraform to use to create the cluster and components.
	TfVersion string

	// Auth0 the Auth0 client ID and secret to use for the cluster.
	Auth0 authOpts
}

// AuthOpts represents the options for Auth0.
type authOpts struct {
	// Domain is the Auth0 domain.
	Domain string
	// ClientID is the Auth0 client ID.
	ClientId string
	// ClientSecret is the Auth0 client secret.
	ClientSecret string
}

func addCreateClusterCmd(toplevel *cobra.Command) {
	var (
		auth = authOpts{}
		opt  = clusterOptions{
			MaxNameLength: 12,
			Auth0:         auth,
		}
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: `create a new cloud-platform cluster`,
		Long: heredoc.Doc(`

			Running this command will create a new eks cluster in the cloud-platform aws account.
			It will create you a VPC, a cluster and the components the cloud-platform team defines as being required for a cluster to function.

			You must have the following environment variables set, or passed via arguments:
				- a valid AWS profile or access key and secret.
				- a valid auth0 client id and secret.
				- a valid auth0 domain.

			You must also be in the infrastructure repository, and have decrypted the repository before running this command.
`),
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "create-cluster"})
			if err := checkCreateDirectory(); err != nil {
				contextLogger.Fatal(err)
			}
			_ = cloudPlatform.Cluster{
				Name:         opt.Name,
				VpcId:        opt.VpcName,
				HealthStatus: "Creating",
			}

			_ = terraform.TerraformCLIConfig{
				Workspace: opt.Name,
				Version:   opt.TfVersion,
			}

			_, err := getCredentials(awsRegion)
			if err != nil {
				contextLogger.Fatal(err)
			}

			// TODO: Business logic to create a cluster
		},
	}
	if err := opt.addCreateClusterFlags(cmd, &auth); err != nil {
		log.Fatal(err)
	}

	toplevel.AddCommand(cmd)
}

func (opt *clusterOptions) addCreateClusterFlags(cmd *cobra.Command, auth *authOpts) error {
	var (
		date    = time.Now().Format("0201")
		minHour = time.Now().Format("1504")
	)

	cmd.Flags().StringVar(&opt.Auth0.ClientId, "auth0-client-id", os.Getenv("AUTH0_CLIENT_ID"), "[required] auth0 client id to use")
	cmd.Flags().StringVar(&opt.Auth0.ClientSecret, "auth0-client-secret", "", "[required] auth0 client secret to use")
	cmd.Flags().StringVar(&opt.Auth0.Domain, "auth0-domain", os.Getenv("AUTH0_DOMAIN"), "[required] auth0 domain to use")

	cmd.Flags().StringVar(&opt.Name, "name", "", "[optional] name of the cluster")
	cmd.Flags().StringVar(&opt.VpcName, "vpc", "", "[optional] name of the vpc to use")
	cmd.Flags().StringVar(&opt.ClusterSuffix, "cluster-suffix", "cloud-platform.service.justice.gov.uk", "[optional] suffix to append to the cluster name")
	cmd.Flags().StringVar(&kubePath, "kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "[optional] absolute path to the desired kubeconfig location")

	// Terraform options
	cmd.Flags().StringVar(&opt.TfVersion, "terraform-version", "0.14.8", "[optional] the terraform version to use. [default] 0.14.8")
	return opt.validateClusterOpts(cmd, date, minHour)
}

func getCredentials(awsRegion string) (*client.AwsCredentials, error) {
	creds, err := client.NewAwsCreds(awsRegion)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

func (o *clusterOptions) validateClusterOpts(cmd *cobra.Command, date, minHour string) error {
	if err := o.checkCreateFlags(); err != nil {
		return err
	}

	if err := o.checkClusterName(date, minHour); err != nil {
		return err
	}

	return nil
}

func (o *clusterOptions) checkClusterName(date, minHour string) error {
	if o.Name == "" {
		name := fmt.Sprintf("cp-%s-%s", date, minHour)

		o.Name = name
		o.VpcName = name
	}

	if len(o.Name) > o.MaxNameLength {
		return errors.New("cluster name is too long, please use a shorter name")
	}

	if strings.Contains(o.Name, "live") || strings.Contains(o.Name, "manager") {
		return errors.New("cluster name cannot contain the words 'live' or 'manager'")
	}
	return nil
}

func (o *clusterOptions) checkCreateFlags() error {
	if awsProfile == "" && awsAccessKey == "" && awsSecret == "" {
		return errors.New("AWS credentials are required, please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or an AWS_PROFILE")
	}

	if o.Auth0.ClientSecret == "" {
		o.Auth0.ClientSecret = os.Getenv("AUTH0_CLIENT_SECRET")
	}

	if o.Auth0.ClientId == "" || o.Auth0.ClientSecret == "" || o.Auth0.Domain == "" {
		return errors.New("auth0 credentials are required, please set AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET and AUTH0_DOMAIN")
	}

	return nil
}

func checkCreateDirectory() error {
	// Ensure the executor is running the command in the correct directory.
	repoName, err := findTopLevelGitDir(".")
	if err != nil {
		return fmt.Errorf("cannot find top level git dir: %s", err)
	}

	if !strings.Contains(repoName, "cloud-platform-infrastructure") {
		return errors.New("must be run from the cloud-platform-infrastructure repository")
	}

	return nil
}

func findTopLevelGitDir(workingDir string) (string, error) {
	dir, err := filepath.Abs(workingDir)
	if err != nil {
		return "", fmt.Errorf("invalid working dir %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no git repository found")
		}
		dir = parent
	}
}

var opt recycle.Options

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)

	// sub cobra commands
	clusterCmd.AddCommand(clusterRecycleNodeCmd)

	// recycle node flags
	clusterRecycleNodeCmd.Flags().StringVarP(&opt.ResourceName, "name", "n", "", "name of the resource to recycle")
	clusterRecycleNodeCmd.Flags().BoolVarP(&opt.Force, "force", "f", true, "force the pods to drain")
	clusterRecycleNodeCmd.Flags().BoolVarP(&opt.IgnoreLabel, "ignore-label", "i", false, "whether to ignore the labels on the resource")
	clusterRecycleNodeCmd.Flags().IntVarP(&opt.TimeOut, "timeout", "t", 360, "amount of time to wait for the drain command to complete")
	clusterRecycleNodeCmd.Flags().BoolVar(&opt.Oldest, "oldest", false, "whether to recycle the oldest node")
	clusterRecycleNodeCmd.Flags().StringVar(&opt.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	clusterRecycleNodeCmd.Flags().StringVar(&awsAccessKey, "aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "aws access key to use")
	clusterRecycleNodeCmd.Flags().StringVar(&awsSecret, "aws-secret-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "aws secret to use")
	clusterRecycleNodeCmd.Flags().StringVar(&awsProfile, "aws-profile", os.Getenv("AWS_PROFILE"), "aws profile to use")
	clusterRecycleNodeCmd.Flags().StringVar(&opt.AwsRegion, "aws-region", "eu-west-2", "aws region to use")
	clusterRecycleNodeCmd.Flags().BoolVar(&opt.Debug, "debug", false, "enable debug logging")
}

var clusterCmd = &cobra.Command{
	Use:    "cluster",
	Short:  `Cloud Platform cluster actions`,
	PreRun: upgradeIfNotLatest,
}

var clusterRecycleNodeCmd = &cobra.Command{
	Use:   "recycle-node",
	Short: `recycle a node`,
	Example: heredoc.Doc(`
	$ cloud-platform cluster recycle-node
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		contextLogger := log.WithFields(log.Fields{"subcommand": "recycle-node"})
		// Check for missing name argument. You must define either a resource
		// or specify the --oldest flag.
		if opt.ResourceName == "" && !opt.Oldest {
			contextLogger.Fatal("--name or --oldest is required")
		}

		if awsProfile == "" && awsAccessKey == "" && awsSecret == "" {
			contextLogger.Fatal("AWS credentials are required, please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or an AWS_PROFILE")
		}

		clientset, err := client.GetClientset(opt.KubecfgPath)
		if err != nil {
			contextLogger.Fatal(err)
		}

		recycle := &recycle.Recycler{
			Client:  &client.KubeClient{Clientset: clientset},
			Options: &opt,
		}

		recycle.Cluster, err = cluster.NewCluster(recycle.Client)
		if err != nil {
			contextLogger.Fatal(err)
		}

		// Create a snapshot for comparison later.
		recycle.Snapshot = recycle.Cluster.NewSnapshot()

		recycle.AwsCreds, err = cluster.NewAwsCreds(opt.AwsRegion)
		if err != nil {
			contextLogger.Fatal(err)
		}

		err = recycle.Node()
		if err != nil {
			// Fail hard so we get an non-zero exit code.
			// This is mainly for when this is run in a pipeline.
			contextLogger.Fatal(err)
		}
	},
}
