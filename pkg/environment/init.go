package environment

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

func (a *Apply) Initialize() {
	var reqEnvVars RequiredEnvVars
	err := envconfig.Process("", &reqEnvVars)
	if err != nil {
		log.Fatalln("Environment variables required to perform terraform operations not set:", err.Error())
	}
	a.RequiredEnvVars.clustername = reqEnvVars.clustername
	a.RequiredEnvVars.clusterstatebucket = reqEnvVars.clusterstatebucket
	a.RequiredEnvVars.kubernetescluster = reqEnvVars.kubernetescluster
	a.RequiredEnvVars.githubowner = reqEnvVars.githubowner
	a.RequiredEnvVars.githubtoken = reqEnvVars.githubtoken
	a.RequiredEnvVars.SlackBotToken = reqEnvVars.SlackBotToken
	a.RequiredEnvVars.SlackWebhookUrl = reqEnvVars.SlackWebhookUrl
	a.RequiredEnvVars.pingdomapitoken = reqEnvVars.pingdomapitoken

	// Set KUBE_CONFIG_PATH to the path of the kubeconfig file
	// This is needed for terraform to be able to connect to the cluster when a different kubecfg is passed
	if err := os.Setenv("KUBE_CONFIG_PATH", a.Options.KubecfgPath); err != nil {
		log.Fatalln("KUBE_CONFIG_PATH environment variable cant be set:", err.Error())
	}
}
