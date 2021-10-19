package duplicate

import (
	"reflect"
	"testing"

	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func Test_copyAndChangeIngress(t *testing.T) {
	type args struct {
		inIngress *networkingv1beta1.Ingress
	}
	tests := []struct {
		name    string
		args    args
		want    *networkingv1beta1.Ingress
		wantErr bool
	}{
		{
			name: "Change name, set-identifier and Hosts",
			args: args{
				inIngress: &networkingv1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ing-01",
						Namespace: "ns-01",
						Annotations: map[string]string{
							"external-dns.alpha.kubernetes.io/aws-weight":     "100",
							"external-dns.alpha.kubernetes.io/set-identifier": "ing-01-ns-01-green",
						},
					},
					Spec: networkingv1beta1.IngressSpec{
						Backend: &networkingv1beta1.IngressBackend{
							ServiceName: "svc-01",
							ServicePort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 1234,
							},
						},
						Rules: []networkingv1beta1.IngressRule{
							{
								Host: "example-ingress.domain.com",
								IngressRuleValue: networkingv1beta1.IngressRuleValue{
									HTTP: &networkingv1beta1.HTTPIngressRuleValue{
										Paths: []networkingv1beta1.HTTPIngressPath{
											{
												Path:     "/",
												PathType: func() *networkingv1beta1.PathType { p := networkingv1beta1.PathTypeImplementationSpecific; return &p }(),
												Backend: networkingv1beta1.IngressBackend{
													ServiceName: "svc-01",
													ServicePort: intstr.IntOrString{
														Type:   intstr.Int,
														IntVal: 1234,
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
								Hosts: []string{"example-ingress.domain.com"},
							},
						},
					},
					Status: networkingv1beta1.IngressStatus{},
				},
			},
			want: &networkingv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ing-01-second",
					Namespace: "ns-01",
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/aws-weight":     "100",
						"external-dns.alpha.kubernetes.io/set-identifier": "ing-01-second-ns-01-green",
					},
				},
				Spec: networkingv1beta1.IngressSpec{
					Backend: &networkingv1beta1.IngressBackend{
						ServiceName: "svc-01",
						ServicePort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 1234,
						},
					},
					Rules: []networkingv1beta1.IngressRule{
						{
							Host: "example-ingress-second.domain.com",
							IngressRuleValue: networkingv1beta1.IngressRuleValue{
								HTTP: &networkingv1beta1.HTTPIngressRuleValue{
									Paths: []networkingv1beta1.HTTPIngressPath{
										{
											Path:     "/",
											PathType: func() *networkingv1beta1.PathType { p := networkingv1beta1.PathTypeImplementationSpecific; return &p }(),
											Backend: networkingv1beta1.IngressBackend{
												ServiceName: "svc-01",
												ServicePort: intstr.IntOrString{
													Type:   intstr.Int,
													IntVal: 1234,
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
							Hosts: []string{"example-ingress-second.domain.com"},
						},
					},
				},
				Status: networkingv1beta1.IngressStatus{},
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

func Test_getIngressResource(t *testing.T) {
	type args struct {
		clientset    *kubernetes.Clientset
		namespace    string
		resourceName string
	}
	tests := []struct {
		name    string
		args    args
		want    *networkingv1beta1.Ingress
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIngressResource(tt.args.clientset, tt.args.namespace, tt.args.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIngressResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getIngressResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
