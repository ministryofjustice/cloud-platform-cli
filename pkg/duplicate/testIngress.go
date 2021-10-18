package duplicate

import (
	"fmt"

	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func DuplicateTestIngress() error {
	inIngress, _ := getIngressJson()
	fmt.Printf("****In Ingress\n %s", inIngress)

	copyIngress, err := copyAndChangeIngress(inIngress)
	if err != nil {
		return err
	}
	fmt.Printf("\n***Out ingress\n %s", copyIngress)

	return nil
}

// getIngressJson takes a Kubernetes clientset, ingress name and a namespace and ingress resource of type *networkingv1beta1.Ingress.
func getIngressJson() (*networkingv1beta1.Ingress, error) {

	inIngress := &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "helloworld-rubyapp-ingress",
			Namespace: "poornima-staging",
			Annotations: map[string]string{
				"external-dns.alpha.kubernetes.io/aws-weight":     "100",
				"external-dns.alpha.kubernetes.io/set-identifier": "helloworld-rubyapp-ingress-poornima-staging-blue",
			},
		},
		Spec: networkingv1beta1.IngressSpec{
			Backend: &networkingv1beta1.IngressBackend{
				ServiceName: "rubyapp-service",
				ServicePort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 4567,
				},
			},
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: "poornima-staging-app.apps.live-1.cloud-platform.service.justice.gov.uk",
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: func() *networkingv1beta1.PathType { p := networkingv1beta1.PathTypeImplementationSpecific; return &p }(),
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: "rubyapp-service",
										ServicePort: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 4567,
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1beta1.IngressTLS{
				{
					Hosts: []string{"poornima-staging-app.apps.live-1.cloud-platform.service.justice.gov.uk"},
				},
			},
		},
		Status: networkingv1beta1.IngressStatus{},
	}
	return inIngress, nil

}

// copyAndChangeIngress gets an ingress, do a deep copy and change the values to the one needed for duplicating
func copyAndChangeIngress(inIngress *networkingv1beta1.Ingress) (*networkingv1beta1.Ingress, error) {
	outIngress := inIngress.DeepCopy()
	outIngress.ObjectMeta.Name = inIngress.ObjectMeta.Name + "-second"
	outIngress.Annotations["external-dns.alpha.kubernetes.io/set-identifier"] = outIngress.ObjectMeta.Name + "-" + outIngress.ObjectMeta.Namespace + "-blue"
	if inIngress.Spec.Rules != nil {
		outIngress.Spec.Rules[0].Host = outIngress.ObjectMeta.Name + "-" + outIngress.ObjectMeta.Namespace + ".apps.live.cloud-platform.service.justice.gov.uk"

	} else {
		return nil, fmt.Errorf("no Ingress Rules to duplicate for ingress %s", outIngress.ObjectMeta.Name)
	}

	// TODO Need to null other Rules

	if inIngress.Spec.TLS != nil {
		outIngress.Spec.TLS[0].Hosts[0] = outIngress.ObjectMeta.Name + "-" + outIngress.ObjectMeta.Namespace + ".apps.live.cloud-platform.service.justice.gov.uk"

	} else {
		return nil, fmt.Errorf("no Ingress TLS Hosts to duplicate for ingress %s", outIngress.ObjectMeta.Name)
	}
	return outIngress, nil
}
