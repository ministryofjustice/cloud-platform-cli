package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	awsProfile   = flag.String("aws-profile", "", "AWS Profile to use")
	account      = flag.String("account", "754256621582", "AWS Account ID")
	namespace    = flag.String("namespace", "", "Namespace to set the environment for")
	clusterArray = []string{"live", "manager", "live-2"}
	home, _      = os.UserHomeDir()
	colourCyan   = "\033[36m"
	colourReset  = "\033[0m"
	colourYellow = "\033[33m"
	colourRed    = "\033[31m"
)

func setAWSEnv(ns string) {
	fmt.Println(string(colourYellow), "\nSetting AWS Configuration", string(colourReset))

	os.Setenv("AWS_PROFILE", ns)

	fmt.Println(string(colourCyan), "AWS_PROFILE:", string(colourReset), os.Getenv("AWS_PROFILE"))

}

func setKubeEnv() error {
	fmt.Println(string(colourYellow), "\nSetting Kube Configuration", string(colourReset))

	kubeConfig := home + "/.kube/config"

	// these are the three kube variables expected by kubectl
	os.Setenv("KUBECONFIG", kubeConfig)
	// this is needed for kubectl provider
	os.Setenv("KUBE_CONFIG", os.Getenv("KUBECONFIG"))
	os.Setenv("KUBE_CONFIG_PATH", os.Getenv("KUBECONFIG"))

	fmt.Println(string(colourCyan), "KUBECONFIG:", string(colourReset), os.Getenv("KUBECONFIG"))
	fmt.Println(string(colourCyan), "KUBE_CONFIG:", string(colourReset), os.Getenv("KUBE_CONFIG"))
	fmt.Println(string(colourCyan), "KUBE_CONFIG_PATH:", string(colourReset), os.Getenv("KUBE_CONFIG_PATH"))

	return nil
}

func setTFWksp(namespace string) error {
	// tf workspace to the cluster name
	fmt.Println(string(colourYellow), "\nUpdating Terraform Workspace")
	err := os.Setenv("TF_WORKSPACE", namespace)
	if err != nil {
		return err
	} else {
		fmt.Println(string(colourCyan), "TF_WORKSPACE:", string(colourReset), os.Getenv("TF_WORKSPACE"))
	}
	return nil
}

func setNamespaceTF(namespace, clusterName string) {
	getSSMParam := func(paramName string) string {
		cmd := exec.Command("aws", "ssm", "get-parameter", "--name", paramName, "--with-decryption", "--profile", *awsProfile, "--query", "Parameter.Value", "--output", "text")
		out, err := cmd.Output()
		if err != nil {
			log.Fatalf("Failed to get SSM parameter %s: %v", paramName, err)
		}
		return strings.TrimSpace(string(out))
	}

	os.Setenv("TF_VAR_github_cloud_platform_concourse_bot_app_id", getSSMParam("/cloud-platform/infrastructure/components/github_cloud_platform_concourse_bot_app_id"))
	os.Setenv("TF_VAR_github_cloud_platform_concourse_bot_installation_id", getSSMParam("/cloud-platform/infrastructure/components/github_cloud_platform_concourse_bot_installation_id"))
	os.Setenv("TF_VAR_github_cloud_platform_concourse_bot_pem_file", getSSMParam("/cloud-platform/infrastructure/components/github_cloud_platform_concourse_bot_pem_file"))
	os.Setenv("PINGDOM_API_TOKEN", getSSMParam("/cloud-platform/infrastructure/components/pingdom_api_token"))
	os.Setenv("PIPELINE_STATE_BUCKET", "cloud-platform-terraform-state")
	os.Setenv("PIPELINE_STATE_KEY_PREFIX", "cloud-platform-environments/")
	os.Setenv("PIPELINE_TERRAFORM_STATE_LOCK_TABLE", "cloud-platform-environments-terraform-lock")
	os.Setenv("PIPELINE_STATE_REGION", "eu-west-1")
	os.Setenv("PIPELINE_CLUSTER", fmt.Sprintf("arn:aws:eks:eu-west-2:%v:cluster/%v", *account, clusterName))
	os.Setenv("PIPELINE_CLUSTER_DIR", fmt.Sprintf("%v.cloud-platform.service.justice.gov.uk", clusterName))
	os.Setenv("TF_VAR_eks_cluster_name", fmt.Sprintf("%v", clusterName))

	os.Setenv("TF_VAR_cluster_state_bucket", "cloud-platform-terraform-state")
	os.Setenv("TF_VAR_github_owner", "ministryofjustice")
	os.Setenv("TF_VAR_kubernetes_cluster", "DF366E49809688A3B16EEC29707D8C09.gr7.eu-west-2.eks.amazonaws.com")

	if clusterName == "live" {
		clusterName = "live-1"
	}
	os.Setenv("PIPELINE_CLUSTER_STATE", fmt.Sprintf("%v.cloud-platform.service.justice.gov.uk", clusterName))
	os.Setenv("TF_VAR_cluster_name", fmt.Sprintf("%v", clusterName))
	os.Setenv("TF_VAR_vpc_name", fmt.Sprintf("%v", clusterName))
}

func setTerm(clusterName string) {
	fmt.Println(string(colourYellow), "Updating Kube Context")
	cmd := exec.Command("aws", "eks", "update-kubeconfig", "--name", clusterName, "--region", "eu-west-2")
	cmd.Run()

	log.Println(string(colourYellow), "\nSetting Terminal Context", string(colourReset))
	cmd = exec.Command("kubectl", "config", "current-context")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(string(colourRed), "(setTerm)CombinedOutput: ", err, string(colourReset))
	}

	kconfig := string(out)
	fmt.Println(string(colourYellow), "\nSetting Terminal Environment")
	os.Setenv("KUBE_PS1", kconfig+": ")
	errsys := syscall.Exec(os.Getenv("SHELL"), []string{os.Getenv("SHELL")}, os.Environ())
	if errsys != nil {
		log.Fatal(string(colourRed), "(setTerm)errsys: ", errsys, string(colourReset))
	}
}

func contains(arg string) bool {
	for _, cluster := range clusterArray {
		if cluster == arg {
			return true
		}
	}
	return false
}

func namespaceEnv(namespace *string) {
	var arg string
	setAWSEnv(*awsProfile)
	fmt.Println("Please enter cluster name:")
	fmt.Scanln(&arg)
	// retry loop with max three attempts
	for i := 0; i < 3; i++ {
		if contains(arg) {
			break
		} else {
			fmt.Println("Please select a cluster from the list:")
			fmt.Scanln(&arg)
		}
	}
	if !contains(arg) {
		log.Fatalf(string(colourRed), "Cluster name is incorrect", string(colourReset))
	}

	clusterName := arg

	err := setKubeEnv()
	if err != nil {
		log.Fatalf(string(colourRed), "Error setting kube config: %v", err, string(colourReset))
	}

	err = setTFWksp(*namespace)
	if err != nil {
		log.Fatalf(string(colourRed), "Error setting Terraform workspace: %v", err, string(colourReset))
	}

	setNamespaceTF(*namespace, clusterName)

	setTerm(clusterName)
}

func main() {
	flag.Parse()
	if *namespace == "" {
		log.Fatalf(string(colourRed), "Namespace is required", string(colourReset))
	}

	if *awsProfile == "" {
		log.Fatalf(string(colourRed), "AWS Profile is required", string(colourReset))
	}

	h, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("User Home: %v", err)
	}
	home = h
	namespaceEnv(namespace)

}
