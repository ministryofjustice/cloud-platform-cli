package environment

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var file *os.File

func TestMain(m *testing.M) {
	var err error
	file, err = createMockKubeConfigFile("temp")
	if err != nil {
		log.Fatalln(err)
	}
	code := m.Run()

	defer file.Close()
	defer os.Remove(file.Name())
	os.Exit(code)
}

func createMockKubeConfigFile(path string) (*os.File, error) {
	data := []byte(`
apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:55171
  name: kind-kind
- cluster:
    server: https://127.0.0.1:55902
  name: kind-kind2
contexts:
- context:
    cluster: kind-kind
    user: kind-kind
  name: kind-kind
- context:
    cluster: kind-kind2
    user: kind-kind2
  name: kind-kind2
current-context: kind-kind2
kind: Config
preferences: {}
users:
- name: kind-kind
  user:
- name: kind-kind2
  user:
`)

	file, err := os.CreateTemp("", "temp")
	if err != nil {
		return nil, err
	}

	if _, err := file.Write(data); err != nil {
		return nil, err
	}

	return file, nil
}

func TestNewDivergence(t *testing.T) {
	divergence, err := NewDivergence("kind", file.Name(), "ghp_fake", nil)
	if err != nil {
		t.Fatalf("error creating divergence object, when it should have created: %v", err)
	}

	if divergence == nil {
		t.Fatalf("divergence is nil, when it should have created")
	}

	if divergence.ClusterName != "kind" {
		t.Fatalf("divergence cluster name is %s, when it should be kind", divergence.ClusterName)
	}
}

func TestNewDivergenceWithInvalidKubeConfigFile(t *testing.T) {
	_, err := NewDivergence("kind", "invalid", "ghp_fake", nil)
	if err == nil {
		t.Fatalf("error is nil, when it should have created")
	}
}

func TestNewDivergenceWithInvalidGitHubToken(t *testing.T) {
	_, err := NewDivergence("", file.Name(), "", nil)
	if err == nil {
		t.Fatalf("error is nil, when it should have created")
	}
}

func TestCheck(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetUsersByUsername,
			github.User{
				Name: github.String("foobar"),
			},
		),
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{
				{
					Name: github.String("test"),
					Type: github.String("dir"),
				},
			},
		),
	)
	testDivergence := &Divergence{
		ClusterName: "kind",
		KubeClient: fake.NewSimpleClientset(
			&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
		),
		GitHubClient:       github.NewClient(mockedHTTPClient),
		ExcludedNamespaces: nil,
	}

	err := testDivergence.Check()
	if err != nil {
		t.Fatalf("error checking divergence, when it should have succeeded: %v", err)
	}
}

func Test_getGithubNamespaces(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetUsersByUsername,
			github.User{
				Name: github.String("foobar"),
			},
		),
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{
				{
					Name: github.String("namespace1"),
					Type: github.String("dir"),
				},
				{
					Name: github.String("namespace2"),
					Type: github.String("dir"),
				},
			},
		),
	)
	type args struct {
		client  *github.Client
		cluster string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "get github namespaces",
			args: args{
				client: github.NewClient(mockedHTTPClient),
			},
			want:    []string{"namespace1", "namespace2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getGithubNamespaces(tt.args.client, tt.args.cluster)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGithubNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getGithubNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createGitHubClient(t *testing.T) {
	t.Parallel()
	type args struct {
		pass string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create a github client with a token",
			args: args{
				pass: "FALSE",
			},
			wantErr: false,
		},
		{
			name: "create a github client without a token",
			args: args{
				pass: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createGitHubClient(tt.args.pass)
			if (err != nil) != tt.wantErr {
				t.Errorf("createGitHubClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_createKubeClient(t *testing.T) {
	t.Parallel()
	type args struct {
		kubeconfig string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create a kube client entering a valid kubeconfig file",
			args: args{
				kubeconfig: file.Name(),
			},
			wantErr: false,
		},
		{
			name: "create a kube client entering an invalid kubeconfig file",
			args: args{
				kubeconfig: "invalid",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createKubeClient(tt.args.kubeconfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("createKubeClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_getClusterNamespaces(t *testing.T) {
	type args struct {
		client kubernetes.Interface
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "get cluster namespaces",
			args: args{
				client: fake.NewSimpleClientset(&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				}, &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "kube-system",
					},
				}),
			},
			want:    []string{"default", "kube-system"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getClusterNamespaces(tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getClusterNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getClusterNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compareNamespaces(t *testing.T) {
	type args struct {
		clusterNamespaces  []string
		githubNamespaces   []string
		excludedNamespaces []string
	}
	tests := []struct {
		name string
		args args
		want sets.String
	}{
		{
			name: "compare namespaces",
			args: args{
				clusterNamespaces:  []string{"default", "kube-system"},
				githubNamespaces:   []string{"default"},
				excludedNamespaces: nil,
			},
			want: sets.NewString("kube-system"),
		},
		{
			name: "compare namespaces with excluded namespaces",
			args: args{
				clusterNamespaces:  []string{"default", "kube-system"},
				githubNamespaces:   []string{"default"},
				excludedNamespaces: []string{"kube-system"},
			},
			want: sets.NewString(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareNamespaces(tt.args.clusterNamespaces, tt.args.githubNamespaces, tt.args.excludedNamespaces); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compareNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
