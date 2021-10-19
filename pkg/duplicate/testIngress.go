package duplicate

import (
	"context"
	"fmt"
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

	fmt.Println(ing)

	// Copy the ingress resource and iterate, ready for apply
	duplicateIngress, err := copyAndChangeIngress(ing)
	if err != nil {
		return err
	}
	fmt.Println(duplicateIngress)

	err = applyIngress(clientset, duplicateIngress)

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
	outIngress.Status = networkingv1beta1.IngressStatus{}
	// TODO Should we nulify all metadata or specific ones?
	outIngress.ObjectMeta = metav1.ObjectMeta{}
	outIngress.Annotations = map[string]string{}

	outIngress.ObjectMeta.Name = inIngress.ObjectMeta.Name + "-second"
	outIngress.ObjectMeta.Namespace = inIngress.ObjectMeta.Namespace
	outIngress.Annotations["external-dns.alpha.kubernetes.io/set-identifier"] = outIngress.ObjectMeta.Name + "-" + outIngress.ObjectMeta.Namespace + "-green"
	outIngress.Annotations["external-dns.alpha.kubernetes.io/aws-weight"] = "100"

	// m = ingress.fetch("metadata")
	// %w[creationTimestamp generation resourceVersion selfLink uid].each { |key| m.delete(key) }
	// m.fetch("annotations").delete("kubectl.kubernetes.io/last-applied-configuration") if m.has_key?("annotations")

	return outIngress, nil
}

func applyIngress(clientset *kubernetes.Clientset, duplicateIngress *networkingv1beta1.Ingress) error {
	ingress, err := clientset.NetworkingV1beta1().Ingresses(duplicateIngress.Namespace).Create(context.TODO(), duplicateIngress, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Println(ingress)
	return nil
}
