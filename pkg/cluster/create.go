package cluster

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	clientset, err := AuthToCluster(tf.Workspace, awsCreds.Eks, kubeconf, awsCreds.Profile)
	if err != nil {
		return fmt.Errorf("failed to auth to cluster: %w", err)
	}

	tf.WorkingDir = dir

	if err := applyTacticalPspFix(clientset); err != nil {
		return err
	}
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
	terraform, error := terraform.NewTerraformCLI(tf)
	if error != nil {
		return nil, error
	}

	// Start fresh and remove any local state.
	if err := deleteLocalState(tf.WorkingDir, ".terraform", ".terraform.lock.hcl"); err != nil {
		fmt.Println("Failed to delete local state, continuing anyway")
	}

	err := terraform.Init(context.Background(), log.Writer())
	if err != nil {
		return nil, fmt.Errorf("failed to init terraform: %w", err)
	}

	if err = terraform.Apply(context.Background(), log.Writer()); err != nil {
		return nil, fmt.Errorf("failed to apply terraform: %w", err)
	}

	// We pass a nil writer to the output command as we don't want to print the output to the console.
	return terraform.Output(context.Background(), nil)
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
