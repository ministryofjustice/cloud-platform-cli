package duplicate

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// Ingress is a struct that allows us to store the ingress resource and the clientset
type Ingress struct {
	Clientset *kubernetes.Interface
	Namespace string
	Resource  *v1.Ingress
}

// NewIngress takes a clientset, namespace and ingress resource and returns a new ingress resource.
func NewIngress(clientset *kubernetes.Interface, namespace, resourceName string) (*Ingress, error) {
	ing := &Ingress{
		Clientset: clientset,
		Namespace: namespace,
	}

	ingress, err := getIngressResource(*clientset, namespace, resourceName)
	if err != nil {
		return nil, err
	}

	ing.Resource = ingress

	return ing, nil
}

// CreateDuplicateIngress takes an ingress resource and creates the Kubernetes Resource Object in the same namespace.
func (ing *Ingress) CreateDuplicate() error {
	duplicateIngress, err := copyAndChangeIngress(ing.Resource)
	if err != nil {
		return err
	}

	err = applyIngress(*ing.Clientset, duplicateIngress)
	if err != nil {
		return err
	}

	return nil
}

// getIngressResource takes a Kubernetes clientset, the name of the users namespace and ingress resource and inspect the
// cluster, returning a v1 ingress resource. There is no need to search for v1beta1 Ingress as the v1 API is backward
// compatible and would also fetch ingress resource with API version v1beta1.
func getIngressResource(clientset kubernetes.Interface, namespace, resourceName string) (*v1.Ingress, error) {
	ingress, err := clientset.NetworkingV1().Ingresses(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ingress, nil
}

// copyAndChangeIngress gets an ingress, do a deep copy and change the values to the one needed for duplicating
func copyAndChangeIngress(inIngress *v1.Ingress) (*v1.Ingress, error) {
	if inIngress == nil {
		return nil, fmt.Errorf("Ingress is nil")
	}

	if inIngress.Spec.TLS == nil {
		return nil, fmt.Errorf("Ingress does not have TLS")
	}

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
func applyIngress(clientset kubernetes.Interface, duplicateIngress *v1.Ingress) error {
	_, err := clientset.NetworkingV1().Ingresses(duplicateIngress.Namespace).Create(context.TODO(), duplicateIngress, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("ingress \"%v\" created\n", duplicateIngress.Name)

	return nil
}
