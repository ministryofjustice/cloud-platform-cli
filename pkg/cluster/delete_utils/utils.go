package deleteutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type EKSClient interface {
	ListClusters(*eks.ListClustersInput) (*eks.ListClustersOutput, error)
}

type EC2Client interface {
	DescribeVpcs(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error)
}

// If someone has deployed something into this cluster, there might be
// associated AWS resources which would be left orphaned if the cluster were
// destroyed. So, we check for any unexpected namespaces, and abort if we find
// any.
func AbortIfUserNamespacesExist(namespaces []v1.Namespace, systemNamespaces []string) error {
	var userNamespaces []string

	for _, ns := range namespaces {
		isStarterPackNamespace := strings.Contains(ns.Name, "starter-pack-")

		isSmokeTestNamespace := strings.Contains(ns.Name, "smoketest-")

		if isSmokeTestNamespace || isStarterPackNamespace {
			continue
		}

		isUserNamespace := !util.Contains(systemNamespaces, ns.Name)

		if isUserNamespace {
			userNamespaces = append(userNamespaces, ns.Name)
		}
	}

	if len(userNamespaces) > 0 {
		return fmt.Errorf("\nPlease delete these namespaces, and any associated AWS resources, before destroying this cluster: %s", userNamespaces)
	}

	return nil
}

func GetNamespaces(clientset kubernetes.Interface) ([]v1.Namespace, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	return namespaces.Items, nil
}

func CheckClusterIsDestroyed(clusterName string, eksAPI EKSClient) error {
	clusters, err := eksAPI.ListClusters(&eks.ListClustersInput{})
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	// check that workspace is not in the list
	var clusterValues []string

	for _, val := range clusters.Clusters {
		clusterValues = append(clusterValues, *val)
	}

	isDeleted := !util.Contains(clusterValues, clusterName)

	if !isDeleted {
		return fmt.Errorf("cluster has not successfully deleted")
	}

	return nil
}

func CheckVpcIsDestroyed(vpcID string, ec2API EC2Client) error {
	vpcs, err := ec2API.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{Name: aws.String("tag:owner"), Values: []*string{aws.String("Cloud Platform: platforms@digital.justice.gov.uk")}},
			{Name: aws.String("tag:application"), Values: []*string{aws.String("cloud-platform-aws/vpc")}},
			{Name: aws.String("tag:Terraform"), Values: []*string{aws.String("true")}},
			{Name: aws.String("tag:business-unit"), Values: []*string{aws.String("Platforms")}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to list vpcs: %w", err)
	}

	var vpcValues []string

	for _, val := range vpcs.Vpcs {
		vpcValues = append(vpcValues, *val.VpcId)
	}

	isDeleted := !util.Contains(vpcValues, vpcID)

	if !isDeleted {
		return fmt.Errorf("vpc has not successfully deleted")
	}

	return nil
}
