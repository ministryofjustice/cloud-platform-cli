package duplicate_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/duplicate"
	"github.com/ministryofjustice/cloud-platform-go-library/client"
	"github.com/ministryofjustice/cloud-platform-go-library/mock"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	fakeCluster = mock.NewCluster(
		mock.WithNamespaces(),
	)
	// fakeClient is a fake cloud platform kubernetes cluster
	// containing some namespace objects.
	fakeClient = client.KubeClient{
		Clientset: fake.NewSimpleClientset(&fakeCluster.Cluster.Namespaces),
	}
)

func TestNewIngress(t *testing.T) {
	ingress := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "Namespace2",
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					Host: "test",
				},
			},
		},
	}

	_, err := fakeClient.Clientset.NetworkingV1().Ingresses("Namespace2").Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		t.Errorf("Error creating ingress: %v", err)
	}

	// Pass good ingress to NewIngress. Expected to return nil error.
	dup, err := duplicate.NewIngress(&fakeClient.Clientset, "Namespace2", "test")
	if !reflect.DeepEqual(ingress.Name, dup.Resource.Name) {
		t.Errorf("Expected %v, got %v", ingress, dup.Resource.Name)
	}
	if err != nil {
		t.Errorf("Expected %v, got %v", nil, err)
	}

	// Pass bad ingress to NewIngress. Expected to return error.
	_, err = duplicate.NewIngress(&fakeClient.Clientset, "Namespace1", "bad")
	if err == nil {
		t.Errorf("Expected %v, got %v", "error", nil)
	}
}

func TestBadCreateDuplicate(t *testing.T) {
	badIngress := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bad",
			Namespace: "Namespace1",
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					Host: "bad",
				},
			},
		},
	}

	// Pass bad ingress to CreateDuplicate. Expected to return error.
	dup := &duplicate.Ingress{
		Resource:  badIngress,
		Clientset: &fakeClient.Clientset,
	}
	_, err := fakeClient.Clientset.NetworkingV1().Ingresses("Namespace1").Create(context.TODO(), badIngress, metav1.CreateOptions{})
	if err != nil {
		t.Errorf("Error creating ingress: %v", err)
	}

	err = dup.CreateDuplicate()
	if err == nil {
		t.Errorf("Expected %v, got %v", "error", nil)
	}
	defer func() {
		err := fakeClient.Clientset.NetworkingV1().Ingresses("Namespace1").Delete(context.TODO(), "bad", metav1.DeleteOptions{})
		if err != nil {
			t.Errorf("Error deleting ingress: %v", err)
		}
	}()
}

func TestGoodCreateDuplicate(t *testing.T) {
	goodIngress := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ing-01",
			Namespace: "Namespace2",
			Annotations: map[string]string{
				"external-dns.alpha.kubernetes.io/aws-weight":     "100",
				"external-dns.alpha.kubernetes.io/set-identifier": "ing-01-ns-01-green",
			},
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					Host: "example-ingress.domain.com",
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: func() *v1.PathType { p := v1.PathTypeImplementationSpecific; return &p }(),
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: "svc-01",
											Port: v1.ServiceBackendPort{
												Number: 1234,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []v1.IngressTLS{
				{
					Hosts: []string{"example-ingress.domain.com"},
				},
			},
		},
		Status: v1.IngressStatus{},
	}

	_, err := fakeClient.Clientset.NetworkingV1().Ingresses("Namespace2").Create(context.TODO(), goodIngress, metav1.CreateOptions{})
	if err != nil {
		t.Errorf("Error creating ingress: %v", err)
	}

	// Pass good ingress to CreateDuplicate. Expected to return nil error.
	goodDup := &duplicate.Ingress{
		Resource:  goodIngress,
		Clientset: &fakeClient.Clientset,
	}

	err = goodDup.CreateDuplicate()
	if err != nil {
		t.Errorf("Expected %v, got %v", nil, err)
	}

	// Check that the duplicate ingress was created.
	_, err = fakeClient.Clientset.NetworkingV1().Ingresses("Namespace2").Get(context.TODO(), "ing-01", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Error getting ingress: %v", err)
	}

	defer func() {
		err := fakeClient.Clientset.NetworkingV1().Ingresses("Namespace2").Delete(context.TODO(), "ing-01", metav1.DeleteOptions{})
		if err != nil {
			t.Errorf("Error deleting ingress: %v", err)
		}
	}()
}
