package clustercmd

import (
	"os"
	"path/filepath"

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
	// Fast is a flag that will skip the creation non-essential components of the cloud-platform cluster.
	Fast bool

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

func AddClusterCmd(topLevel *cobra.Command) {
	// top level command
	topLevel.AddCommand(clusterCmd)

	// sub commands
	clusterCmd.AddCommand(clusterRecycleNodeCmd)
	addCreateClusterCmd(clusterCmd)
	addDestroyClusterCmd(clusterCmd)

	// cluster level flags
	clusterCmd.Flags().StringVar(&awsAccessKey, "aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "[required] aws access key to use")
	clusterCmd.Flags().StringVar(&awsSecret, "aws-secret-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "[required] aws secret to use")
	clusterCmd.Flags().StringVar(&awsProfile, "aws-profile", os.Getenv("AWS_PROFILE"), "[required] aws profile to use")
	clusterCmd.Flags().StringVar(&awsRegion, "aws-region", os.Getenv("AWS_REGION"), "[required] aws region to use")
	clusterCmd.Flags().StringVar(&kubePath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")

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
	Use:   "cluster",
	Short: `Cloud Platform cluster actions`,
}
