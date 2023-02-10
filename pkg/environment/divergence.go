package environment

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Divergence is a struct that contains the information needed to check for divergence between a cluster and github
type Divergence struct {
	ClusterName        string
	KubeClient         kubernetes.Interface
	GitHubClient       *github.Client
	ExcludedNamespaces []string
}

// NewDivergence takes the name of a kubernetes cluster, the path to a kubeconfig file and a github personal
// access token, and returns a Divergence struct.
func NewDivergence(clusterName, kubeconfig, githubToken string, excludedNamespaces []string) (*Divergence, error) {
	if !strings.HasPrefix(githubToken, "ghp_") {
		return nil, fmt.Errorf("invalid github token")
	}

	kubeClient, err := createKubeClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	if clusterName == "" {
		clusterName = getClusterName(kubeClient)
	}

	githubClient, err := createGitHubClient(githubToken)
	if err != nil {
		return nil, err
	}

	return &Divergence{
		ClusterName:        clusterName,
		KubeClient:         kubeClient,
		GitHubClient:       githubClient,
		ExcludedNamespaces: excludedNamespaces,
	}, nil
}

func getClusterName(kubeClient kubernetes.Interface) string {
	clusterName, err := kubeClient.CoreV1().ConfigMaps("kube-public").Get(context.Background(), "cluster-info", metav1.GetOptions{})
	if err != nil {
		return ""
	}
	return clusterName.Data["cluster-name"]
}

func (d *Divergence) Check() error {
	// get all cluster namespaces
	clusterNamespaces, err := getClusterNamespaces(d.KubeClient)
	if err != nil {
		return fmt.Errorf("error getting cluster namespaces: %v", err)
	}

	// get all github namespaces
	githubNamespaces, err := getGithubNamespaces(d.GitHubClient, d.ClusterName)
	if err != nil {
		return fmt.Errorf("error getting github namespaces: %v", err)
	}

	// compare namespaces and print out differences
	clusterNamespacesSet := compareNamespaces(clusterNamespaces, githubNamespaces, d.ExcludedNamespaces)
	if len(clusterNamespacesSet.List()) > 0 {
		fmt.Println("The following namespaces are in the cluster but not in github:")
		for _, ns := range clusterNamespacesSet.List() {
			fmt.Println(ns)
		}
	}

	return nil
}

func compareNamespaces(clusterNamespaces, githubNamespaces, excludedNamespaces []string) sets.String {
	clusterNamespacesSet := sets.NewString()
	for _, ns := range clusterNamespaces {
		if !sets.NewString(githubNamespaces...).Has(ns) && !sets.NewString(excludedNamespaces...).Has(ns) {
			clusterNamespacesSet.Insert(ns)
		}
	}

	return clusterNamespacesSet
}

// getClusterNamespaces returns a set of namespaces in the cluster
func getClusterNamespaces(client kubernetes.Interface) ([]string, error) {
	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nsSet := sets.NewString()
	for _, ns := range namespaces.Items {
		nsSet.Insert(ns.Name)
	}
	return nsSet.List(), nil
}

// getGithubNamespaces returns a set of namespaces in githubToken
func getGithubNamespaces(client *github.Client, clusterName string) ([]string, error) {
	// get the list of all directories in https://github.com/ministryofjustice/cloud-platform-environments/namespaces
	opt := &github.RepositoryContentGetOptions{Ref: "main"}
	_, dir, _, err := client.Repositories.GetContents(context.TODO(), "ministryofjustice", "cloud-platform-environments", "namespaces/"+clusterName+".cloud-platform.service.justice.gov.uk", opt)
	if err != nil {
		return nil, err
	}

	nsSet := sets.NewString()
	for _, d := range dir {
		nsSet.Insert(*d.Name)
	}
	return nsSet.List(), nil
}

func createClients(kubeconfig, githubToken string) (kubernetes.Interface, *github.Client, error) {
	// create kube client
	kubeClient, err := createKubeClient(kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	// create github client
	gitHubClient, err := createGitHubClient(githubToken)
	if err != nil {
		return nil, nil, err
	}

	return kubeClient, gitHubClient, nil
}

func createGitHubClient(pass string) (*github.Client, error) {
	if pass == "" {
		return nil, fmt.Errorf("no github token provided")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pass},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client, nil
}

func createKubeClient(kubeconfig string) (kubernetes.Interface, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %v", err)
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}
