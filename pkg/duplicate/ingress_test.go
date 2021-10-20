package duplicate

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
