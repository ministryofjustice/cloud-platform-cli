package terminalcfg

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

// global variables
var (
	clusterName  string
	kubeConfig   string
	clusterArray []string
	returnOuput  bool
	home, _      = os.UserHomeDir()
	colourCyan   = "\033[36m"
	colourReset  = "\033[0m"
	colourYellow = "\033[33m"
	colourRed    = "\033[31m"
)

// list eks clusters for aws profile
func ListEksClusters() {
	profile := os.Getenv("AWS_PROFILE")
	region := os.Getenv("AWS_REGION")
	awsconfig := os.Getenv("AWS_CONFIG_FILE")
	awscreds := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")

	// Create a session that will use the credentials and config to authenticate
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
		SharedConfigFiles: []string{awsconfig, awscreds},
	}))

	// Create a new instance of the service's client with a Session.
	// Optional aws.Config values can also be provided as variadic arguments
	// to the New function. This option allows you to provide service
	// specific configuration.
	svc := eks.New(sess)

	ctx := context.Background()

	// lists eks clusters
	result, err := svc.ListClustersWithContext(ctx, &eks.ListClustersInput{})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	fmt.Println(string(colourYellow), "\nEKS Test Clusters:", string(colourReset))
	// exclude live and manager from list
	for _, cluster := range result.Clusters {
		if *cluster != "live" && *cluster != "manager" {
			clusterArray = []string{*cluster}
			fmt.Println(string(colourCyan), "Cluster:", string(colourReset), clusterArray)
		}
	}

}

//setting AWS config
func SetAWSEnv() {
	awsRegion := "eu-west-2"
	awsConfig := home + "/.aws/config"
	awsCreds := home + "/.aws/credentials"
	fmt.Println(string(colourYellow), "\nSetting AWS Configuration", string(colourReset))
	// set aws_profile to correct profile
	// set aws_config_file to correct path
	// set aws_shared_credentials_file to correct path
	// set aws_region to correct region
	// set aws_default_region to correct region
	os.Setenv("AWS_PROFILE", "moj-cp")
	os.Setenv("AWS_CONFIG_FILE", awsConfig)
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", awsCreds)
	os.Setenv("AWS_REGION", awsRegion)
	os.Setenv("AWS_DEFAULT_REGION", os.Getenv("AWS_REGION"))

	ap, _ := os.LookupEnv("AWS_PROFILE")
	acf, _ := os.LookupEnv("AWS_CONFIG_FILE")
	ascf, _ := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE")
	ar, _ := os.LookupEnv("AWS_REGION")
	adr, _ := os.LookupEnv("AWS_DEFAULT_REGION")

	for _, env := range []string{ap, acf, ascf, ar, adr} {
		if env == "" {
			log.Fatal(string(colourRed), "Environment variable is empty", string(colourReset))
			os.Exit(1)
		}
	}
	fmt.Println(string(colourCyan), "AWS_PROFILE:", string(colourReset), ap)
	fmt.Println(string(colourCyan), "AWS_CONFIG_FILE:", string(colourReset), acf)
	fmt.Println(string(colourCyan), "AWS_SHARED_CREDENTIALS_FILE:", string(colourReset), ascf)
	fmt.Println(string(colourCyan), "AWS_REGION:", string(colourReset), ar)
	fmt.Println(string(colourCyan), "AWS_DEFAULT_REGION:", string(colourReset), adr)
}

// sets Kube config
func SetKubeEnv(clusterName string) {
	fmt.Println(string(colourYellow), "\nSetting Kube Configuration", string(colourReset))
	// set kubeconfig for live, manager or test
	// set kube_config_path to kubeconfig
	if clusterName == "live" {
		kubeConfig = home + "/.kube/" + clusterName + "/config"
	} else if clusterName == "manager" {
		kubeConfig = home + "/.kube/" + clusterName + "/config"
	} else {
		kubeConfig = home + "/.kube/test/" + clusterName + "/config"
	}
	if kubeConfig == "" {
		log.Fatal("kubeConfig is empty")
		os.Exit(1)
	}
	os.Setenv("KUBECONFIG", kubeConfig)
	os.Setenv("KUBE_CONFIG_PATH", os.Getenv("KUBECONFIG"))

	fmt.Println(string(colourCyan), "KUBECONFIG:", string(colourReset), os.Getenv("KUBECONFIG"))
	fmt.Println(string(colourCyan), "KUBE_CONFIG_PATH:", string(colourReset), os.Getenv("KUBE_CONFIG_PATH"))
}

// sets Terraform Workspace
func SetTFWksp(clusterName string) {
	// tf workspace to the cluster name
	fmt.Println(string(colourYellow), "\nUpdating Terraform Workspace")
	os.Setenv("TF_WORKSPACE", clusterName)

	fmt.Println(string(colourCyan), "TF_WORKSPACE:", string(colourReset), os.Getenv("TF_WORKSPACE"))
}

func Contains(args string) bool {
	for _, cluster := range clusterArray {
		if cluster == args {
			return true
		}
	}
	return false
}

// sets up test environment for eks cluster
func TestEnv() bool {
	var arg string
	SetAWSEnv()
	ListEksClusters()
	fmt.Println("Please select a cluster to use:")
	fmt.Scanln(&arg)
	if !Contains(arg) {
		log.Fatal(string(colourRed), "Invalid cluster name entered", string(colourReset))
		os.Exit(1)
		returnOuput = false
	} else {
		returnOuput = true
	}
	clusterName = arg
	SetKubeEnv(clusterName)
	// set kubecontext to correct context name
	fmt.Println(string(colourYellow), "Updating Kube Context")
	cmd := exec.Command("aws", "eks", "update-kubeconfig", "--name", clusterName)
	cmd.Run()
	// Set Terraform workspace to the cluster name
	SetTFWksp(clusterName)
	// set command line prompt to comtext name
	os.Setenv("PS1", "\\e[1;33m`kubectl config current-context`> \\e[m")
	// start shell with new environment variables
	errsys := syscall.Exec(os.Getenv("SHELL"), []string{os.Getenv("SHELL")}, os.Environ())
	if errsys != nil {
		log.Fatal(string(colourRed), errsys, string(colourReset))
	}
	return returnOuput
}

// sets up live environment for eks cluster
func LiveManagerEnv(env string) bool {
	SetAWSEnv()
	clusterName = env
	if clusterName != "live" && clusterName != "manager" {
		log.Fatal(string(colourRed), "Cluster name is incorrect", string(colourReset))
		returnOuput = false
	} else {
		returnOuput = true
	}
	SetKubeEnv(clusterName)
	// set kubecontext to correct context name
	fmt.Println(string(colourYellow), "Updating Kube Context")
	cmd := exec.Command("aws", "eks", "update-kubeconfig", "--name", clusterName)
	cmd.Run()
	// Set Terraform workspace to the cluster name
	SetTFWksp(clusterName)
	// set command line prompt to comtext name
	os.Setenv("PS1", "\\e[1;33m`kubectl config current-context`> \\e[m")
	// start shell with new environment variables
	errsys := syscall.Exec(os.Getenv("SHELL"), []string{os.Getenv("SHELL")}, os.Environ())
	if errsys != nil {
		log.Fatal(string(colourRed), errsys, string(colourReset))
	}
	return returnOuput
}
