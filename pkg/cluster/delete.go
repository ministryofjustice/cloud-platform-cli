package cluster

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	deleteutils "github.com/ministryofjustice/cloud-platform-cli/pkg/cluster/delete_utils"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
)

var EKS_SYSTEM_NAMESPACES = []string{
	"cert-manager",
	"default",
	"external-secrets-operator",
	"ingress-controllers",
	"kube-node-lease",
	"kube-public",
	"kube-system",
	"kuberos",
	"logging",
	"monitoring",
	"velero",
	"kuberhealthy",
	"trivy-system",
	"calico-apiserver",
	"calico-system",
	"tigera-operator",
	"gatekeeper-system",
	"cloud-platform-label-pods",
	"cloud-platform-github-teams-filter",
}

// DestroyComponents will destroy the Cloud Platform specific components on top of a running cluster. At this point your
// cluster should be up and running and you should be able to connect to it.
func (c *Cluster) DestroyComponents(tf *terraform.TerraformCLIConfig, awsCreds *client.AwsCredentials, dir, kubeconf string, dryRun bool) error {
	// Reset any previous variables that might've been set.
	tf.DestroyVars = nil
	tf.WorkingDir = dir

	clientset, err := AuthToCluster(tf.Workspace, awsCreds.Eks, kubeconf, awsCreds.Profile)
	if err != nil {
		return fmt.Errorf("failed to auth to cluster: %w", err)
	}

	namespaces, err := deleteutils.GetNamespaces(clientset)
	if err != nil {
		return err
	}

	err = deleteutils.AbortIfUserNamespacesExist(namespaces, EKS_SYSTEM_NAMESPACES)
	if err != nil {
		return err
	}

	err = deleteutils.TerraformDestroyLayer(tf, dryRun)
	if err != nil {
		return err
	}

	return nil
}

// DestroyCore will destroy the Cloud Platform specific core components on top of a running cluster.
func (c *Cluster) DestroyCore(tf *terraform.TerraformCLIConfig, awsCreds *client.AwsCredentials, dir, kubeconf string, dryRun bool) error {
	// Reset any previous variables that might've been set.
	tf.DestroyVars = nil
	tf.WorkingDir = dir

	err := deleteutils.TerraformDestroyLayer(tf, dryRun)
	if err != nil {
		return err
	}

	return nil
}

// DestroyEks destroys the terraform for a Cloud Platform EKS cluster
// It will return an error if the EKS cluster is not destroyed or the terraform commands fail.
func (c *Cluster) DestroyEks(tf *terraform.TerraformCLIConfig, creds *client.AwsCredentials, dir string, dryRun bool) error {
	tf.DestroyVars = nil
	tf.WorkingDir = dir

	err := deleteutils.TerraformDestroyLayer(tf, dryRun)
	if err != nil {
		return err
	}

	if !dryRun {
		if err := deleteutils.CheckClusterIsDestroyed(tf.Workspace, creds.Eks); err != nil {
			return err
		}

		c.HealthStatus = "Deleted"
	}

	return nil
}

// DestroyVpc destroys terraform code for the Cloud Platform VPC and ensures it's deleted.
// It will return an error if the VPC is still up and running or the terraform commands fail.
func (c *Cluster) DestroyVpc(tf *terraform.TerraformCLIConfig, creds *client.AwsCredentials, dir string, dryRun bool) error {
	tf.DestroyVars = nil
	tf.WorkingDir = dir

	err := deleteutils.TerraformDestroyLayer(tf, dryRun)
	if err != nil {
		return err
	}

	if !dryRun {
		if err := deleteutils.CheckVpcIsDestroyed(tf.Workspace, creds.Ec2); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) DeleteTfWorkspace(tf *terraform.TerraformCLIConfig, dirs []string, dryRun bool) error {
	if !dryRun {
		for _, dir := range dirs {
			tf.WorkingDir = dir

			tfCli, err := deleteutils.InitTfCLI(tf, dryRun)
			if err != nil {
				return err
			}

			if err := tfCli.WorkspaceDelete(context.TODO(), tf.Workspace); err != nil {
				return err
			}

		}
	}

	return nil
}
