package cluster

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (c *Cluster) ApplyVpc(tf *terraform.Options, creds *client.AwsCredentials, dir string) error {
	exec, err := tfexec.NewTerraform(dir, tf.ExecPath)
	if err != nil {
		return fmt.Errorf("failed to create terraform object: %w", err)
	}
	// Start fresh and remove any local state.
	if err := deleteLocalState(dir, ".terraform", ".terraform.lock.hcl"); err != nil {
		return fmt.Errorf("failed to delete temp directory: %w", err)
	}
	output, err := tf.Apply(exec, creds)
	if err != nil {
		return err
	}

	vpcID := output["vpc_id"]
	if vpcID.Value == nil {
		return errors.New("vpc_id not found in terraform output")
	}

	fmt.Println("Starting to check vpc")
	// Trim the vpcId to remove quotes
	vpc := strings.Trim(string(vpcID.Value), "\"")
	return checkVpc(*tf, vpc, creds.Session)
}

func (c *Cluster) ApplyEks(tf *terraform.Options, creds *client.AwsCredentials, dir string) error {
	fmt.Println("Applying Terraform against EKS")
	exec, err := tfexec.NewTerraform(dir, tf.ExecPath)
	if err != nil {
		return fmt.Errorf("failed to create terraform object: %w", err)
	}

	// Start fresh and remove any local state.
	if err := deleteLocalState(dir, ".terraform", ".terraform.lock.hcl"); err != nil {
		return fmt.Errorf("failed to delete temp directory: %w", err)
	}

	_, err = tf.Apply(exec, creds)
	if err != nil {
		return err
	}

	if err := checkCluster(tf.Workspace, creds.Eks); err != nil {
		return err
	}

	c.HealthStatus = "Good"

	return nil
}

func (c *Cluster) ApplyComponents(tf *terraform.Options, awsCreds *client.AwsCredentials, clientset *kubernetes.Interface, dir string) error {
	// There is a requirement for the aws binary to exist at this point.
	if err := authToCluster(tf.Workspace); err != nil {
		return err
	}

	exec, err := tfexec.NewTerraform(dir, tf.ExecPath)
	if err != nil {
		return fmt.Errorf("failed to create terraform object: %w", err)
	}
	// Start fresh and remove any local state.
	if err := deleteLocalState(dir, ".terraform", ".terraform.lock.hcl"); err != nil {
		return fmt.Errorf("failed to delete temp directory: %w", err)
	}
	_, err = tf.Apply(exec, awsCreds)
	if err != nil {
		return err
	}

	if err := applyTacticalPspFix(*clientset); err != nil {
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

// CheckVpc asserts that the vpc is up and running. It tests the vpc state and id.
func checkVpc(tf terraform.Options, vpcId string, sess *session.Session) error {
	svc := ec2.New(sess)

	vpc, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Cluster"),
				Values: []*string{aws.String(tf.Workspace)},
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
