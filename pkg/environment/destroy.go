package environment

import (
	"fmt"
	"log"
	"os"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

// Destroy is the entry point for performing a namespace destroy.
// It checks if the working directory is in cloud-platform-environments, checks if a PR number is given and merged
// The method get the list of namespaces that are deleted in that merger PR, and for all namespaces in the PR does the
// terraform init and destroy and do a kubectl delete
func (a *Apply) Destroy() error {
	fmt.Println("Destroying Namespaces in PR", a.Options.PRNumber)
	if a.Options.PRNumber == 0 {
		err := fmt.Errorf("a PR ID/Number is required to perform destroy")
		return err
	}
	isMerged, err := a.GithubClient.IsMerged(a.Options.PRNumber)
	if err != nil {
		return err
	}
	if isMerged {
		changedNamespaces, err := a.nsCreateRawChangedFilesInPR(a.Options.ClusterDir, a.Options.PRNumber)
		if err != nil {
			return err
		}
		if len(changedNamespaces) == 0 {
			fmt.Println("No namespaces to destroy")
			return nil
		}

		fmt.Println("Namespaces removed in PR", changedNamespaces)

		kubeClient, err := authenticate.CreateClientFromConfigFile(a.Options.KubecfgPath, a.Options.ClusterCtx)
		if err != nil {
			return err
		}

		// GetAllNamespacesFromCluster
		namespaces, err := namespace.GetAllNamespacesFromCluster(kubeClient)
		if err != nil {
			return err
		}

		for _, namespace := range changedNamespaces {
			if a.Options.SkipProdDestroy && isProductionNs(namespace, namespaces) {
				err := fmt.Errorf("cannot destroy production namespace with skip-prod-destroy flag set to true")
				return err
			}
			// Check if the namespace is present in the folder
			if _, err = os.Stat(namespace); err != nil {
				fmt.Println("Destroying Namespace:", namespace)
				err = a.destroyNamespace(namespace)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// destroyTerraform calls applier -> TerraformInitAndDestroy and prints the output from applier
func (a *Apply) destroyTerraform() (string, error) {
	log.Printf("Running Terraform Destroy for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	outputTerraform, err := a.Applier.TerraformInitAndDestroy(a.Options.Namespace, tfFolder)
	if err != nil {
		err := fmt.Errorf("error running terraform on namespace %s: %v \n %v", a.Options.Namespace, err, outputTerraform)
		return "", err
	}
	return outputTerraform, nil
}

// destroyNamespace intiates a apply object with options and env variables, and calls the
// calls applier TerraformInitAndDestroy, applyKubectl with dry-run disabled and prints the output
func (a *Apply) destroyNamespace(namespace string) error {
	repoPath := "namespaces/" + a.Options.ClusterDir + "/" + namespace

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		fmt.Printf("Namespace %s does not exist, skipping destroy\n", namespace)
		return nil
	}

	applier := NewApply(*a.Options, namespace)
	applier.Options.Namespace = namespace

	exists, err := util.IsFilePathExists(repoPath + "/resources")
	if err == nil && exists {
		outputTerraform, err := applier.destroyTerraform()
		if err != nil {
			return err
		}

		fmt.Println("\nOutput of terraform:")
		util.RedactedEnv(os.Stdout, outputTerraform, a.Options.RedactedEnv)

		if util.IsYamlFileExists(repoPath) {
			outputKubectl, err := applier.deleteKubectl()
			if err != nil {
				return err
			}

			fmt.Println("\nOutput of kubectl:", outputKubectl)
		} else {
			fmt.Printf("Namespace %s does not have yaml resources folder, skipping kubectl delete", namespace)
		}

	} else {
		fmt.Printf("Namespace %s does not have terraform resources folder, skipping terraform destroy", namespace)
	}
	return nil
}
