package environment

import (
	"fmt"
	"log"
	"os"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

// planKubectl calls the applier -> applyKubectl with dry-run enabled and return the output from applier
func (a *Apply) planKubectl() (string, error) {
	log.Printf("Running kubectl dry-run for namespace: %v in directory %v", a.Options.Namespace, a.Dir)

	outputKubectl, err := a.Applier.KubectlApply(a.Options.Namespace, a.Dir, true)
	if err != nil {
		err := fmt.Errorf("error running kubectl on namespace %s: in directory: %v, %v\n %v", a.Options.Namespace, a.Dir, err, outputKubectl)
		return "", err
	}

	return outputKubectl, nil
}

// Plan is the entry point for performing a namespace plan.
// It checks if the working directory is in cloud-platform-environments, checks if a PR number or a namespace is given
// If a namespace is given, it perform a `kubectl apply --dry-run=client` and a terraform init and plan of that namespace
// else checks for PR number and get the list of changed namespaces in the PR. Then does the `kubectl apply --dry-run=client` and
// terraform init and plan of all the namespaces changed in the PR
func (a *Apply) Plan() error {
	if a.Options.PRNumber == 0 && a.Options.Namespace == "" {
		return fmt.Errorf("either a PR Id/Number or a namespace is required to perform plan")
	}

	// If a namespace is given as a flag, then perform a plan for the given namespace.
	if a.Options.Namespace != "" {
		err := a.planNamespace()
		if err != nil {
			return err
		}
		return nil
	} else {
		files, err := a.GithubClient.GetChangedFiles(a.Options.PRNumber)
		if err != nil {
			return fmt.Errorf("failed to fetch list of changed files: %s in PR %v", err, a.Options.PRNumber)
		}

		changedNamespaces, err := nsChangedInPR(files, a.Options.ClusterDir, false)
		if err != nil {
			fmt.Println("failed to get list of changed namespaces in PR:", err)
			return err
		}
		for _, namespace := range changedNamespaces {
			a.Options.Namespace = namespace
			err = a.planNamespace()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// planTerraform calls applier -> TerraformInitAndPlan and prints the output from applier
func (a *Apply) planTerraform() (*tfjson.Plan, string, error) {
	log.Printf("Running Terraform Plan for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	tfPlan, outputTerraform, err := a.Applier.TerraformInitAndPlan(a.Options.Namespace, tfFolder)
	if err != nil {
		err := fmt.Errorf("error running terraform on namespace %s: %v \n %v", a.Options.Namespace, err, outputTerraform)
		return nil, "", err
	}
	return tfPlan, outputTerraform, nil
}

// planNamespace intiates a new Apply object with options and env variables, and calls the
// applyKubectl with dry-run enabled and calls applier TerraformInitAndPlan and prints the output
func (a *Apply) planNamespace() error {
	applier := NewApply(*a.Options)
	repoPath := "namespaces/" + a.Options.ClusterDir + "/" + a.Options.Namespace

	if util.IsYamlFileExists(repoPath) {
		outputKubectl, err := applier.planKubectl()
		if err != nil {
			return err
		}

		fmt.Println("\nOutput of kubectl:", outputKubectl)
	} else {
		fmt.Printf("Namespace %s does not have yaml resources folder, skipping kubectl apply --dry-run\n", a.Options.Namespace)
	}

	exists, err := util.IsFilePathExists(repoPath + "/resources")
	if err == nil && exists {
		tfPlan, outputTerraform, err := applier.planTerraform()
		if err != nil {
			return err
		}

		fmt.Println("\nOutput of terraform:")

		commentErr := CreateComment(a.GithubClient, tfPlan, a.Options.PRNumber)
		if commentErr != nil {
			fmt.Printf("\nError posting comment: %v", commentErr)
		}
		util.RedactedEnv(os.Stdout, outputTerraform, a.Options.RedactedEnv)
	} else {
		fmt.Printf("Namespace %s does not have terraform resources folder, skipping terraform plan\n", a.Options.Namespace)
	}
	return nil
}
