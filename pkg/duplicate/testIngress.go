package duplicate

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	networkingv1beta1 "k8s.io/api/networking/v1beta1"
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
func DuplicateTestIngress(namespace, resourceName string) error {
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

	// Get the ingress resource
	ing, err := getIngressResource(clientset, namespace, resourceName)
	if err != nil {
		return err
	}

	// Copy the ingress resource and iterate, ready for apply
	_, err = copyAndChangeIngress(ing)
	if err != nil {
		return err
	}

	return nil
}

// getIngressResource takes a Kubernetes clientset, the name of the users namespace and ingress resource and inspect the
// cluster, returning a v1beta ingress resource only. At this point there is no need to search for v1Ingress as the majority
// of Cloud Platform users don't have the API set. This may change.
func getIngressResource(clientset *kubernetes.Clientset, namespace, resourceName string) (*networkingv1beta1.Ingress, error) {
	ingress, err := clientset.NetworkingV1beta1().Ingresses(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ingress, nil
}

// copyAndChangeIngress gets an ingress, do a deep copy and change the values to the one needed for duplicating
// copyAndChangeIngress gets an ingress, do a deep copy and change the values to the one needed for duplicating
func copyAndChangeIngress(inIngress *networkingv1beta1.Ingress) (*networkingv1beta1.Ingress, error) {
	outIngress := inIngress.DeepCopy()
	outIngress.ObjectMeta.Name = inIngress.ObjectMeta.Name + "-second"
	outIngress.Annotations["external-dns.alpha.kubernetes.io/set-identifier"] = outIngress.ObjectMeta.Name + "-" + outIngress.ObjectMeta.Namespace + "-green"

	for i := 0; i < len(inIngress.Spec.Rules); i++ {
		subDomain := strings.SplitN(inIngress.Spec.Rules[i].Host, ".", 2)
		outIngress.Spec.Rules[i].Host = subDomain[0] + "-second." + subDomain[1]
	}

	for i := 0; i < len(inIngress.Spec.TLS[0].Hosts); i++ {
		subDomain := strings.SplitN(inIngress.Spec.TLS[0].Hosts[i], ".", 2)
		outIngress.Spec.TLS[0].Hosts[i] = subDomain[0] + "-second." + subDomain[1]
	}
	return outIngress, nil
}
