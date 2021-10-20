package duplicate

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// DuplicateTestIngress is the function run when a user performs
// cloud-platform duplicate ingress -n <namespace> <resource>
// It requires a namespace name, the name of an ingress resource
// and will:
//   - inspect the cluster
//   - copy the ingress rule defined
//   - make relevant changes
//   - create a new rule
func DuplicateIngress(namespace, resourceName string) error {
	// Set the filepath of the kubeconfig file. This assumes
	// the user has either the envname set or stores their config file
	// in the default location.
	configFile := os.Getenv("KUBECONFIG")
	if configFile == "" {
		configFile = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	// Build the Kubernetes client using the default context (user set).
	config, err := clientcmd.BuildConfigFromFlags("", configFile)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Get the specified ingress resource
	ing, err := getIngressResource(clientset, namespace, resourceName)
	if err != nil {
		return err
	}

	// Duplicate the specified ingress resource
	duplicateIngress, err := copyAndChangeIngress(ing)
	if err != nil {
		return err
	}

	// Apply the duplicate ingress resource to the same namespace
	err = applyIngress(clientset, duplicateIngress)
	if err != nil {
		return err
	}

	return nil
}

// getIngressResource takes a Kubernetes clientset, the name of the users namespace and ingress resource and inspect the
// cluster, returning a v1 ingress resource. There is no need to search for v1beta1 Ingress as the v1 API is backward
// compatible and would also fetch ingress resource with API version v1beta1.
func getIngressResource(clientset *kubernetes.Clientset, namespace, resourceName string) (*v1.Ingress, error) {
	ingress, err := clientset.NetworkingV1().Ingresses(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ingress, nil
}

// copyAndChangeIngress gets an ingress, do a deep copy and change the values to the one needed for duplicating
func copyAndChangeIngress(inIngress *v1.Ingress) (*v1.Ingress, error) {
	duplicateIngress := inIngress.DeepCopy()

	// loop over Spec.Rules from the original ingress, add -duplicate string to the sub-domain i.e first part of domain
	// and set that as the host rule for the duplicate ingress
	for i := 0; i < len(inIngress.Spec.Rules) && len(inIngress.Spec.Rules) > 0; i++ {
		subDomain := strings.SplitN(inIngress.Spec.Rules[i].Host, ".", 2)
		duplicateIngress.Spec.Rules[i].Host = subDomain[0] + "-duplicate." + subDomain[1]
	}

	// Check if there is any TLS configuration, loop over the list of hosts from Spec.TLS of the original ingress,
	// add -duplicate string to the sub-domain i.e first part of the domain
	// and set that as the TLS host for the duplicate ingress
	if len(inIngress.Spec.TLS) > 0 {
		for i := 0; i < len(inIngress.Spec.TLS[0].Hosts) && len(inIngress.Spec.TLS) > 0; i++ {
			subDomain := strings.SplitN(inIngress.Spec.TLS[0].Hosts[i], ".", 2)
			duplicateIngress.Spec.TLS[0].Hosts[i] = subDomain[0] + "-duplicate." + subDomain[1]
		}
	}

	// Discard the extra data returned by the k8s API which we don't need in the duplicate
	duplicateIngress.Status = v1.IngressStatus{}
	duplicateIngress.ObjectMeta.ResourceVersion = ""
	duplicateIngress.ObjectMeta.SelfLink = ""
	duplicateIngress.ObjectMeta.UID = ""

	// Discard unwanted annotations that are copied from original ingress that are not needs in the duplicate
	for k := range duplicateIngress.ObjectMeta.Annotations {
		if strings.Contains(k, "meta.helm.sh") || strings.Contains(k, "kubectl.kubernetes.io/last-applied-configuration") {
			delete(duplicateIngress.Annotations, k)
		}

	}

	duplicateIngress.ObjectMeta.Name = inIngress.ObjectMeta.Name + "-duplicate"
	duplicateIngress.ObjectMeta.Namespace = inIngress.ObjectMeta.Namespace
	duplicateIngress.ObjectMeta.Annotations["external-dns.alpha.kubernetes.io/set-identifier"] = duplicateIngress.ObjectMeta.Name + "-" + duplicateIngress.ObjectMeta.Namespace + "-green"
	duplicateIngress.ObjectMeta.Annotations["external-dns.alpha.kubernetes.io/aws-weight"] = "100"

	return duplicateIngress, nil
}

// applyIngress takes a clientset and an ingress resource (which has been duplicated) and applies
// to the current cluster and namespace.
func applyIngress(clientset *kubernetes.Clientset, duplicateIngress *v1.Ingress) error {
	_, err := clientset.NetworkingV1().Ingresses(duplicateIngress.Namespace).Create(context.TODO(), duplicateIngress, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
