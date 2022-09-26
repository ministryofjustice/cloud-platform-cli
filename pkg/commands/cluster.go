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
	"github.com/ministryofjustice/cloud-platform-cli/pkg/recycle"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// CreateOptions struct represents the options passed to the Create method
// by the `cloud-platform create cluster` command.
type createCluster struct {
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
	Auth0 cluster.AuthOpts
}

func addCreateClusterCmd(toplevel *cobra.Command) {
	var (
		opt = createCluster{
			MaxNameLength: 12,
		}
		auth    = &cluster.AuthOpts{}
		date    = time.Now().Format("0201")
		minHour = time.Now().Format("1504")
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: `create a new cloud-platform cluster`,
		Long: heredoc.Doc(`
Running this command will create a new eks cluster in the cloud-platform aws account. It will create you a VPC, a cluster and the components the cloud-platform team defines as being required for a cluster to function.

			You must have the following environment variables set, or passed via arguments:
			- a valid AWS profile or access key and secret.
			- a valid auth0 client id and secret.
			- a valid auth0 domain.

			You must also be in the infrastructure repository, and have decrypted the repository before running this command.
`),
		Example: heredoc.Doc(`
		$ cloud-platform cluster create --name jb-test-01
	`),
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "create-cluster"})
			opt.Auth0 = *auth

			c := cluster.Cluster{
				Name:         opt.Name,
				VpcId:        opt.VpcName,
				HealthStatus: "Creating",
			}

			tf, err := cluster.NewTerraformOptions(opt.TfVersion, c.Name, nil)
			if err != nil {
				contextLogger.Fatal(err)
			}

			awsCreds, err := getCredentials(awsRegion)
			if err != nil {
				contextLogger.Fatal(err)
			}

			if err := c.ApplyVpc(tf, awsCreds); err != nil {
				contextLogger.Fatal(err)
			}

			if err := c.ApplyEks(tf, awsCreds, opt.Fast); err != nil {
				contextLogger.Fatal(err)
			}

			clientset, err := client.GetClientset(kubePath)
			if err != nil {
				contextLogger.Fatal(err)
			}

			if err := c.ApplyComponents(tf, awsCreds, &clientset); err != nil {
				contextLogger.Fatal(err)
			}

			// TODO: Display a nice table of the cluster status.
		},
	}
	if err := opt.addCreateClusterFlags(cmd, auth); err != nil {
		log.Fatal(err)
	}

	// Ensure the flags passed to the command are valid
	if err := opt.validateClusterOpts(cmd, date, minHour); err != nil {
		log.Fatal(err)
	}

	toplevel.AddCommand(cmd)
}

var recycleOptions recycle.Options

var (
	awsSecret, awsAccessKey, awsProfile, awsRegion string
)

var kubePath string

func clusterFlags() {
	clusterCmd.Flags().StringVar(&awsAccessKey, "aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "[required] aws access key to use")
	clusterCmd.Flags().StringVar(&awsSecret, "aws-secret-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "[required] aws secret to use")
	clusterCmd.Flags().StringVar(&awsProfile, "aws-profile", os.Getenv("AWS_PROFILE"), "[required] aws profile to use")
	clusterCmd.Flags().StringVar(&awsRegion, "aws-region", os.Getenv("AWS_REGION"), "[required] aws region to use")
	clusterCmd.Flags().StringVar(&kubePath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
}

func (opt *createCluster) addCreateClusterFlags(cmd *cobra.Command, auth *cluster.AuthOpts) error {
	cmd.Flags().StringVar(&opt.Auth0.ClientId, "auth0-client-id", os.Getenv("AUTH0_CLIENT_ID"), "[required] auth0 client id to use")
	cmd.Flags().StringVar(&opt.Auth0.ClientSecret, "auth0-client-secret", os.Getenv("AUTH0_CLIENT_SECRET"), "[required] auth0 client secret to use")
	cmd.Flags().StringVar(&opt.Auth0.Domain, "auth0-domain", os.Getenv("AUTH0_DOMAIN"), "[required] auth0 domain to use")

	cmd.Flags().StringVar(&opt.Name, "name", "", "[optional] name of the cluster")
	cmd.Flags().StringVar(&opt.VpcName, "vpc", "", "[optional] name of the vpc to use")
	cmd.Flags().StringVar(&opt.ClusterSuffix, "cluster-suffix", "cloud-platform.service.justice.gov.uk", "[optional] suffix to append to the cluster name")
	cmd.Flags().BoolVar(&opt.Debug, "debug", false, "[optional] enable debug logging")
	cmd.Flags().IntVar(&opt.NodeCount, "nodes", 3, "[optional] number of nodes to create. [default] 3")
	cmd.Flags().BoolVar(&opt.Fast, "fast", false, "[optional] enable fast mode - this creates a cluster as quickly as possible. [default] false")

	// Terraform options
	cmd.Flags().StringVar(&opt.TfVersion, "terraformVersion", "0.14.8", "[optional] the terraform version to use. [default] 0.14.8")

	return nil
}

func recycleFlags() {
	// recycle node flags
	clusterRecycleNodeCmd.Flags().StringVarP(&recycleOptions.ResourceName, "name", "n", "", "name of the resource to recycle")
	clusterRecycleNodeCmd.Flags().BoolVarP(&recycleOptions.Force, "force", "f", true, "force the pods to drain")
	clusterRecycleNodeCmd.Flags().BoolVarP(&recycleOptions.IgnoreLabel, "ignore-label", "i", false, "whether to ignore the labels on the resource")
	clusterRecycleNodeCmd.Flags().IntVarP(&recycleOptions.TimeOut, "timeout", "t", 360, "amount of time to wait for the drain command to complete")
	clusterRecycleNodeCmd.Flags().BoolVar(&recycleOptions.Oldest, "oldest", false, "whether to recycle the oldest node")
	clusterRecycleNodeCmd.Flags().StringVar(&recycleOptions.AwsRegion, "aws-region", "eu-west-2", "aws region to use")
	clusterRecycleNodeCmd.Flags().BoolVar(&recycleOptions.Debug, "debug", false, "enable debug logging")
}

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)

	clusterFlags()
	// sub cobra commands
	recycleFlags()
	clusterCmd.AddCommand(clusterRecycleNodeCmd)

	// add flags to sub commands
	addCreateClusterCmd(clusterCmd)
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
		if recycleOptions.ResourceName == "" && !recycleOptions.Oldest {
			contextLogger.Fatal("--name or --oldest is required")
		}

		if awsProfile == "" && awsAccessKey == "" && awsSecret == "" {
			contextLogger.Fatal("AWS credentials are required, please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or an AWS_PROFILE")
		}

		recycleOptions.KubecfgPath = kubePath

		clientset, err := client.GetClientset(recycleOptions.KubecfgPath)
		if err != nil {
			contextLogger.Fatal(err)
		}

		recycle := &recycle.Recycler{
			Client:  &client.Client{Clientset: clientset},
			Options: &recycleOptions,
		}

		recycle.Cluster, err = cluster.NewCluster(recycle.Client)
		if err != nil {
			contextLogger.Fatal(err)
		}

		// Create a snapshot for comparison later.
		recycle.Snapshot = recycle.Cluster.NewSnapshot()

		recycle.AwsCreds, err = cluster.NewAwsCreds(recycleOptions.AwsRegion)
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

func getCredentials(awsRegion string) (*cluster.AwsCredentials, error) {
	creds, err := cluster.NewAwsCreds(awsRegion)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

func (o *createCluster) checkCreateFlags() error {
	if awsProfile == "" && awsAccessKey == "" && awsSecret == "" {
		return errors.New("AWS credentials are required, please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or an AWS_PROFILE")
	}

	if o.Auth0.ClientId == "" || o.Auth0.ClientSecret == "" || o.Auth0.Domain == "" {
		return errors.New("Auth0 credentials are required, please set AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET and AUTH0_DOMAIN")
	}

	return nil
}

func (o *createCluster) checkClusterName(date, minHour string) error {
	if o.Name == "" {
		name := fmt.Sprintf("cp-%s-%s", date, minHour)

		o.Name = name
		o.VpcName = name
	}

	if len(o.Name) > o.MaxNameLength {
		return errors.New("Cluster name is too long, please use a shorter name")
	}

	if strings.Contains(o.Name, "live") || strings.Contains(o.Name, "manager") {
		return errors.New("Cluster name cannot contain the words 'live' or 'manager'")
	}
	return nil
}

func (o *createCluster) checkCreateDirectory() error {
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

func (o *createCluster) validateClusterOpts(cmd *cobra.Command, date, minHour string) error {
	if err := o.checkCreateFlags(); err != nil {
		return err
	}

	if err := o.checkClusterName(date, minHour); err != nil {
		return err
	}

	if err := o.checkCreateDirectory(); err != nil {
		return err
	}
	return nil
}
