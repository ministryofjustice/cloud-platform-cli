package environment

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment/mocks"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApply_ApplyTerraform(t *testing.T) {
	type fields struct {
		Options         *Options
		RequiredEnvVars RequiredEnvVars
		Applier         Applier
		Dir             string
	}
	tests := []struct {
		name              string
		fields            fields
		TerraformOutputs  string
		checkExpectations func(t *testing.T, terraform *mocks.Applier, outputs string, err error)
	}{
		{
			name: "Apply foo namespace",
			fields: fields{
				Options: &Options{
					Namespace:   "foobar",
					KubecfgPath: "/root/.kube/config",
					ClusterCtx:  "testctx",
				},
				RequiredEnvVars: RequiredEnvVars{
					clustername:        "cluster01",
					clusterstatebucket: "clusterstatebucket",
					kubernetescluster:  "kubernetescluster01",
					githubowner:        "githubowner",
					githubtoken:        "githubtoken",
					pingdomapitoken:    "pingdomApikey",
				},
				Dir: "/root/foobar",
			},
			TerraformOutputs: "foo",
			checkExpectations: func(t *testing.T, apply *mocks.Applier, outputs string, err error) {
				apply.AssertCalled(t, "TerraformInitAndApply", "foobar", "/root/foobar/resources")
				assert.Nil(t, err)
				assert.Len(t, outputs, 3)
			},
		},
	}
	for i := range tests {
		terraform := new(mocks.Applier)
		tfFolder := tests[i].fields.Dir + "/resources"
		terraform.On("TerraformInitAndApply", tests[i].fields.Options.Namespace, tfFolder).Return(tests[i].TerraformOutputs, nil)
		a := Apply{
			RequiredEnvVars: tests[i].fields.RequiredEnvVars,
			Applier:         terraform,
			Dir:             tests[i].fields.Dir,
			Options:         tests[i].fields.Options,
		}
		outputs, err := a.applyTerraform()
		t.Run(tests[i].name, func(t *testing.T) {
			tests[i].checkExpectations(t, terraform, outputs, err)
		})
	}
}

func TestApply_PlanKubectl(t *testing.T) {
	type fields struct {
		Options         *Options
		RequiredEnvVars RequiredEnvVars
		Applier         Applier
		Dir             string
	}
	tests := []struct {
		name              string
		fields            fields
		KubectlOutputs    string
		checkExpectations func(t *testing.T, kubectl *mocks.Applier, outputs string, err error)
	}{
		{
			name: "Apply foo namespace",
			fields: fields{
				Options: &Options{
					Namespace:   "foobar",
					KubecfgPath: "/root/.kube/config",
					ClusterCtx:  "testctx",
				},
				RequiredEnvVars: RequiredEnvVars{
					clustername:        "cluster01",
					clusterstatebucket: "clusterstatebucket",
					kubernetescluster:  "kubernetescluster01",
					githubowner:        "githubowner",
					githubtoken:        "githubtoken",
					pingdomapitoken:    "pingdomApikey",
				},
				Dir: "/root/foobar",
			},
			KubectlOutputs: "/root/foobar",
			checkExpectations: func(t *testing.T, apply *mocks.Applier, outputs string, err error) {
				apply.AssertCalled(t, "KubectlApply", "foobar", "/root/foobar", false)
				assert.Nil(t, err)
				assert.Len(t, outputs, 12)
			},
		},
	}
	for i := range tests {
		kubectl := new(mocks.Applier)
		kubectl.On("KubectlApply", "foobar", tests[i].fields.Dir, false).Return(tests[i].KubectlOutputs, nil)
		a := Apply{
			RequiredEnvVars: tests[i].fields.RequiredEnvVars,
			Applier:         kubectl,
			Dir:             tests[i].fields.Dir,
			Options:         tests[i].fields.Options,
		}
		outputs, err := a.applyKubectl()
		t.Run(tests[i].name, func(t *testing.T) {
			tests[i].checkExpectations(t, kubectl, outputs, err)
		})
	}
}

func TestApply_ApplyKubectl(t *testing.T) {
	type fields struct {
		Options         *Options
		RequiredEnvVars RequiredEnvVars
		Applier         Applier
		Dir             string
	}
	tests := []struct {
		name              string
		fields            fields
		KubectlOutputs    string
		checkExpectations func(t *testing.T, kubectl *mocks.Applier, outputs string, err error)
	}{
		{
			name: "Apply foo namespace",
			fields: fields{
				Options: &Options{
					Namespace:   "foobar",
					KubecfgPath: "/root/.kube/config",
					ClusterCtx:  "testctx",
				},
				RequiredEnvVars: RequiredEnvVars{
					clustername:        "cluster01",
					clusterstatebucket: "clusterstatebucket",
					kubernetescluster:  "kubernetescluster01",
					githubowner:        "githubowner",
					githubtoken:        "githubtoken",
					pingdomapitoken:    "pingdomApikey",
				},
				Dir: "/root/foobar",
			},
			KubectlOutputs: "/root/foobar",
			checkExpectations: func(t *testing.T, apply *mocks.Applier, outputs string, err error) {
				apply.AssertCalled(t, "KubectlApply", "foobar", "/root/foobar", false)
				assert.Nil(t, err)
				assert.Len(t, outputs, 12)
			},
		},
	}
	for i := range tests {
		kubectl := new(mocks.Applier)
		kubectl.On("KubectlApply", "foobar", tests[i].fields.Dir, false).Return(tests[i].KubectlOutputs, nil)
		a := Apply{
			RequiredEnvVars: tests[i].fields.RequiredEnvVars,
			Applier:         kubectl,
			Dir:             tests[i].fields.Dir,
			Options:         tests[i].fields.Options,
		}
		outputs, err := a.applyKubectl()
		t.Run(tests[i].name, func(t *testing.T) {
			tests[i].checkExpectations(t, kubectl, outputs, err)
		})
	}
}

func TestSecretBlockerExists(t *testing.T) {
	tempDir := "namespaces/testCluster/testNamespace"
	tempFile := tempDir + "/SECRET_ROTATE_BLOCK"

	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Create(tempFile); err != nil {
		t.Fatal(err)
	}

	if secretBlockerExists(tempDir) != true {
		t.Errorf("secretBlocker should return true as the file does exist")
	}
	defer os.RemoveAll("namespaces")
}

func Test_applySkipExists(t *testing.T) {
	tempDir := "namespaces/testCluster/testNamespace"
	tempFile := tempDir + "/APPLY_PIPELINE_SKIP_THIS_NAMESPACE"

	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Create(tempFile); err != nil {
		t.Fatal(err)
	}

	if applySkipExists(tempDir) != true {
		t.Errorf("secretBlocker should return true as the file does exist")
	}
	defer os.RemoveAll("namespaces")
}

func Test_canCreateNamespaces(t *testing.T) {
	repoPath := "namespaces/testCluster"
	existingNamespace := "namespaces/testCluster/testNamespaceCantDestroy"

	err := os.MkdirAll(repoPath, os.ModePerm)
	if err != nil {
		t.Errorf("Failed to create repo path: %s", err)
	}
	err = os.MkdirAll(existingNamespace, os.ModePerm)
	if err != nil {
		t.Errorf("Failed to create repo path for existingNamespace: %s", err)
	}

	type args struct {
		namespaces []string
		cluster    string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Create namespace for destroy",
			args: args{
				namespaces: []string{"testNamespace"},
				cluster:    "testCluster",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Create namespace for destroy with existing namespace",
			args: args{
				namespaces: []string{"testNamespaceCantDestroy"},
				cluster:    "testCluster",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := canCreateNamespaces(tt.args.namespaces, tt.args.cluster)
			if (err != nil) != tt.wantErr {
				t.Errorf("canCreateNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("canCreateNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
	defer os.RemoveAll("namespaces")
}

func Test_nsChangedInPR(t *testing.T) {
	type args struct {
		files     []*github.CommitFile
		cluster   string
		isDeleted bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Pr for namespace change",
			args: args{
				files: []*github.CommitFile{
					{
						SHA:       github.String("f61aebed695e2e4193db5euy686dcb09b5b57875f334"),
						Filename:  github.String("namespaces/testctx/ns1/file1.txt"),
						Additions: github.Int(5),
						Deletions: github.Int(3),
						Changes:   github.Int(103),
						Status:    github.String("modified"),
						Patch:     github.String("@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"),
					},
					{
						SHA:       github.String("f61aebed695e2e4193db5e6dcb09b5b57875f334"),
						Filename:  github.String("namespaces/testctx/ns1/file2.txt"),
						Additions: github.Int(5),
						Deletions: github.Int(3),
						Changes:   github.Int(103),
						Status:    github.String("modified"),
						Patch:     github.String("@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"),
					},
				},
				cluster:   "testctx",
				isDeleted: false,
			},
			want:    []string{"ns1"},
			wantErr: false,
		},
		{
			name: "Pr for namespace change with deleted file",
			args: args{
				files: []*github.CommitFile{
					{
						SHA:       github.String("f61aebed695e2e4193db5euy686dcb09b5b57875f334"),
						Filename:  github.String("namespaces/testctx/ns2/file1.txt"),
						Additions: github.Int(0),
						Deletions: github.Int(3),
						Changes:   github.Int(0),
						Status:    github.String("removed"),
						Patch:     github.String("@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"),
					},
					{
						SHA:       github.String("f61aebed695e2e4193db5e6dcb09b5b57875f334"),
						Filename:  github.String("namespaces/testctx/ns2/file2.txt"),
						Additions: github.Int(0),
						Deletions: github.Int(3),
						Changes:   github.Int(0),
						Status:    github.String("removed"),
						Patch:     github.String("@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"),
					},
				},
				cluster:   "testctx",
				isDeleted: true,
			},
			want:    []string{"ns2"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsChangedInPR(tt.args.files, tt.args.cluster, tt.args.isDeleted)
			if (err != nil) != tt.wantErr {
				t.Errorf("nsChangedInPR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nsChangedInPR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApply_destroyTerraform(t *testing.T) {
	type fields struct {
		Options         *Options
		RequiredEnvVars RequiredEnvVars
		Applier         Applier
		Dir             string
	}
	tests := []struct {
		name              string
		fields            fields
		TerraformOutputs  string
		checkExpectations func(t *testing.T, terraform *mocks.Applier, outputs string, err error)
	}{
		{
			name: "Destroy foobar namespace",
			fields: fields{
				Options: &Options{
					Namespace:   "foobar",
					KubecfgPath: "/root/.kube/config",
					ClusterCtx:  "testctx",
				},
				RequiredEnvVars: RequiredEnvVars{
					clustername:        "cluster01",
					clusterstatebucket: "clusterstatebucket",
					kubernetescluster:  "kubernetescluster01",
					githubowner:        "githubowner",
					githubtoken:        "githubtoken",
					pingdomapitoken:    "pingdomApikey",
				},
				Dir: "/root/foobar",
			},
			TerraformOutputs: "foobar",
			checkExpectations: func(t *testing.T, apply *mocks.Applier, outputs string, err error) {
				apply.AssertCalled(t, "TerraformInitAndDestroy", "foobar", "/root/foobar/resources")
				assert.Nil(t, err)
				assert.Len(t, outputs, 6)
			},
		},
	}
	for i := range tests {
		terraform := new(mocks.Applier)
		tfFolder := tests[i].fields.Dir + "/resources"
		terraform.On("TerraformInitAndDestroy", tests[i].fields.Options.Namespace, tfFolder).Return(tests[i].TerraformOutputs, nil)
		a := Apply{
			RequiredEnvVars: tests[i].fields.RequiredEnvVars,
			Applier:         terraform,
			Dir:             tests[i].fields.Dir,
			Options:         tests[i].fields.Options,
		}
		outputs, err := a.destroyTerraform()
		t.Run(tests[i].name, func(t *testing.T) {
			tests[i].checkExpectations(t, terraform, outputs, err)
		})
	}
}

func TestApply_deleteKubectl(t *testing.T) {
	type fields struct {
		Options         *Options
		RequiredEnvVars RequiredEnvVars
		Applier         Applier
		Dir             string
	}
	tests := []struct {
		name              string
		fields            fields
		KubectlOutputs    string
		checkExpectations func(t *testing.T, kubectl *mocks.Applier, outputs string, err error)
	}{
		{
			name: "Delete foo namespace",
			fields: fields{
				Options: &Options{
					Namespace:   "foobar",
					KubecfgPath: "/root/.kube/config",
					ClusterCtx:  "testctx",
				},
				RequiredEnvVars: RequiredEnvVars{
					clustername:        "cluster01",
					clusterstatebucket: "clusterstatebucket",
					kubernetescluster:  "kubernetescluster01",
					githubowner:        "githubowner",
					githubtoken:        "githubtoken",
					pingdomapitoken:    "pingdomApikey",
				},
				Dir: "/root/foobar",
			},
			KubectlOutputs: "/root/foobar",
			checkExpectations: func(t *testing.T, apply *mocks.Applier, outputs string, err error) {
				apply.AssertCalled(t, "KubectlDelete", "foobar", "/root/foobar", false)
				assert.Nil(t, err)
				assert.Len(t, outputs, 12)
			},
		},
	}
	for i := range tests {
		kubectl := new(mocks.Applier)
		kubectl.On("KubectlDelete", "foobar", tests[i].fields.Dir, false).Return(tests[i].KubectlOutputs, nil)
		a := Apply{
			RequiredEnvVars: tests[i].fields.RequiredEnvVars,
			Applier:         kubectl,
			Dir:             tests[i].fields.Dir,
			Options:         tests[i].fields.Options,
		}
		outputs, err := a.deleteKubectl()
		t.Run(tests[i].name, func(t *testing.T) {
			tests[i].checkExpectations(t, kubectl, outputs, err)
		})
	}
}

func Test_isProductionNs(t *testing.T) {
	type args struct {
		nsInPR     string
		namespaces []v1.Namespace
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Prod Namespace is not in PR",
			args: args{
				nsInPR: "foobar",
				namespaces: []v1.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foobar",
							Labels: map[string]string{
								"cloud-platform.justice.gov.uk/is-production": "false",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Prod Namespace is in PR",
			args: args{
				nsInPR: "foobar",
				namespaces: []v1.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foobar",
							Labels: map[string]string{
								"cloud-platform.justice.gov.uk/is-production": "true",
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isProductionNs(tt.args.nsInPR, tt.args.namespaces); got != tt.want {
				t.Errorf("isProductionNs() = %v, want %v", got, tt.want)
			}
		})
	}
}
