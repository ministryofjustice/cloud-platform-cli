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
// cluster, returning a v1beta ingress resource only. At this point there is no need to search for v1Ingress as the majority
// of Cloud Platform users don't have the API set. This may change.
func getIngressResource(clientset *kubernetes.Clientset, namespace, resourceName string) (*v1.Ingress, error) {
	ingress, err := clientset.NetworkingV1().Ingresses(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ingress, nil
}

// copyAndChangeIngress gets an ingress, do a deep copy and change the values to the one needed for duplicating
func copyAndChangeIngress(inIngress *v1.Ingress) (*v1.Ingress, error) {
	outIngress := inIngress.DeepCopy()

	for i := 0; i < len(inIngress.Spec.Rules) && len(inIngress.Spec.Rules) > 0; i++ {
		subDomain := strings.SplitN(inIngress.Spec.Rules[i].Host, ".", 2)
		outIngress.Spec.Rules[i].Host = subDomain[0] + "-second." + subDomain[1]
	}

	if len(inIngress.Spec.TLS) > 0 {
		for i := 0; i < len(inIngress.Spec.TLS[0].Hosts) && len(inIngress.Spec.TLS) > 0; i++ {
			subDomain := strings.SplitN(inIngress.Spec.TLS[0].Hosts[i], ".", 2)
			outIngress.Spec.TLS[0].Hosts[i] = subDomain[0] + "-second." + subDomain[1]
		}
	}

	// Discard the extra data returned by the k8s API which we don't need in the copy
	outIngress.Status = v1.IngressStatus{}
	// TODO Should we nulify all metadata or specific ones?
	outIngress.ObjectMeta = metav1.ObjectMeta{}
	outIngress.Annotations = map[string]string{}

	outIngress.ObjectMeta.Name = inIngress.ObjectMeta.Name + "-second"
	outIngress.ObjectMeta.Namespace = inIngress.ObjectMeta.Namespace
	outIngress.Annotations["external-dns.alpha.kubernetes.io/set-identifier"] = outIngress.ObjectMeta.Name + "-" + outIngress.ObjectMeta.Namespace + "-green"
	outIngress.Annotations["external-dns.alpha.kubernetes.io/aws-weight"] = "100"

	return outIngress, nil
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
