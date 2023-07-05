package environment

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment/mocks"
	"github.com/stretchr/testify/assert"
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
				Dir: "/root/foo",
			},
			TerraformOutputs: "foo",
			checkExpectations: func(t *testing.T, apply *mocks.Applier, outputs string, err error) {
				apply.AssertCalled(t, "TerraformInitAndApply", "foobar", "/root/foo/resources")
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
				Dir: "/root/foo",
			},
			KubectlOutputs: "/root/foo",
			checkExpectations: func(t *testing.T, apply *mocks.Applier, outputs string, err error) {
				apply.AssertCalled(t, "KubectlApply", "foobar", "/root/foo", false)
				assert.Nil(t, err)
				assert.Len(t, outputs, 9)
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

func Test_createNamespaceforDestroy(t *testing.T) {
	type args struct {
		namespaces []string
		cluster    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createNamespaceforDestroy(tt.args.namespaces, tt.args.cluster); (err != nil) != tt.wantErr {
				t.Errorf("createNamespaceforDestroy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
