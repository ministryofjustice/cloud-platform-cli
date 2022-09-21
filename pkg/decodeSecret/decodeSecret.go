package decodeSecret

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//	type Secret struct {
//		AccessKeyID     string
//		SecretAccessKey string
//	}
type Secret struct {
	APIVersion string   `json:"apiVersion"`
	Data       struct{} `json:"data"`
	Kind       string   `json:"kind"`
	Metadata   struct {
		Annotations       struct{}  `json:"annotations"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
	} `json:"metadata"`
	Type string `json:"type"`
}

type Options struct {
	// Clientset is the kubernetes clientset to speak with the API server.
	Clientset *kubernetes.Interface
	// ExportAwsCreds will output the AWS secrets in a namespace.
	ExportAwsCreds bool
	// Namespace is the namespace to retrieve the secret from.
	Namespace string
	// Formatting the json means we can't print unicode characters (that might exist in secrets).
	// Raw allows you to output the json without formatting it.
	Raw bool
	// Resource is the resource to retrieve the secret from.
	Resource *corev1.Secret
	// SecretName is the name of the secret to retrieve.
	SecretName string
}

func NewOptions(clientset kubernetes.Interface, secret, ns string, creds, raw bool) (*Options, error) {
	resource, err := getSecret(clientset, ns, secret)
	if err != nil {
		return nil, err
	}

	if resource == nil {
		return nil, fmt.Errorf("secret %s not found in namespace %s", secret, ns)
	}

	return &Options{
		Clientset:      &clientset,
		SecretName:     secret,
		Resource:       resource,
		Namespace:      ns,
		ExportAwsCreds: creds,
		Raw:            raw,
	}, nil
}

// DecodeSecret takes the decode receiver, passes the options to the decode function and will print the output.
func (s *Secret) DecodeSecret(opts *Options) error {
	// Print the secret as if it's a json object.
	err := json.Unmarshal(opts.Resource.Data["data"], &s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal secret: %v", err)
	}

	format, err := s.formatJson()
	if err != nil {
		return err
	}

	fmt.Println(format)

	// Print the secret as if it's a json object.

	return nil
}

func getSecret(clientset kubernetes.Interface, namespace, secret string) (*corev1.Secret, error) {
	return clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secret, metav1.GetOptions{})
}

func (s *Secret) formatJson() (string, error) {
	str, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return "", err
	}

	return string(str), nil
}
