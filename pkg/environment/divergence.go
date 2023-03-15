package environment

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
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
	kubeClient, err := createKubeClient(kubeconfig)
	if err != nil {
		return nil, err
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
	if clusterNamespacesSet.Cardinality() > 0 {
		fmt.Println("The following namespaces are in github but not in the cluster:")
		for ns := range clusterNamespacesSet.Iter() {
			fmt.Println(ns)
		}
	}

	return nil
}

func compareNamespaces(clusterNamespaces, githubNamespaces, excludedNamespaces []string) mapset.Set[string] {
	clusterNamespacesSet := mapset.NewSet[string]()
	for _, ns := range clusterNamespaces {
		if !mapset.NewSet(githubNamespaces...).Contains(ns) && !mapset.NewSet(excludedNamespaces...).Contains(ns) {
			clusterNamespacesSet.Add(ns)
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
