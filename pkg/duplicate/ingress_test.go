package duplicate

import (
	"context"
	"reflect"
	"testing"

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
	dup, err := NewIngress(&fakeClient.Clientset, "Namespace2", "test")
	if !reflect.DeepEqual(ingress.Name, dup.Resource.Name) {
		t.Errorf("Expected %v, got %v", ingress, dup.Resource.Name)
	}
	if err != nil {
		t.Errorf("Expected %v, got %v", nil, err)
	}

	// Pass bad ingress to NewIngress. Expected to return error.
	_, err = NewIngress(&fakeClient.Clientset, "Namespace1", "bad")
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
	dup := &Ingress{
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
	goodDup := &Ingress{
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

func Test_copyAndChangeIngress(t *testing.T) {
	type args struct {
		inIngress *v1.Ingress
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.Ingress
		wantErr bool
	}{
		{
			name: "Change name, set-identifier and Hosts",
			args: args{
				inIngress: &v1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ing-01",
						Namespace: "ns-01",
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
				},
			},
			want: &v1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ing-01-duplicate",
					Namespace: "ns-01",
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/aws-weight":     "100",
						"external-dns.alpha.kubernetes.io/set-identifier": "ing-01-duplicate-ns-01-green",
					},
				},
				Spec: v1.IngressSpec{
					Rules: []v1.IngressRule{
						{
							Host: "example-ingress-duplicate.domain.com",
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
							Hosts: []string{"example-ingress-duplicate.domain.com"},
						},
					},
				},
				Status: v1.IngressStatus{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := copyAndChangeIngress(tt.args.inIngress)
			if (err != nil) != tt.wantErr {
				t.Errorf("copyAndChangeIngress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("copyAndChangeIngress() = %v, want %v", got, tt.want)
			}
		})
	}
}
