package environment

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment/mocks"
	ghMock "github.com/ministryofjustice/cloud-platform-cli/pkg/mocks/github"
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

func TestApply_nsChangedInPR(t *testing.T) {
	type args struct {
		cluster  string
		prNumber int
	}
	tests := []struct {
		name                   string
		GetChangedFilesOutputs []*github.CommitFile
		args                   args
		want                   []string
		wantErr                bool
	}{
		{
			name: "pr with one namespace",
			GetChangedFilesOutputs: []*github.CommitFile{
				{
					SHA:       github.String("6dcb09b5b57875f334f61aebed695e2e4193db5e"),
					Filename:  github.String("namespaces/testctx/ns1/file1.txt"),
					Additions: github.Int(103),
					Deletions: github.Int(21),
					Changes:   github.Int(124),
					Status:    github.String("added"),
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
			args: args{
				cluster:  "testctx",
				prNumber: 8834,
			},
			want:    []string{"ns1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		ghClient := new(ghMock.GithubIface)
		ghClient.On("GetChangedFiles", 8834).Return(tt.GetChangedFilesOutputs, nil)
		t.Run(tt.name, func(t *testing.T) {
			a := &Apply{
				GithubClient: ghClient,
			}
			got, err := a.nsChangedInPR(tt.args.cluster, tt.args.prNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("Apply.nsChangedInPR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Apply.nsChangedInPR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecretBlockerExists(t *testing.T) {
	tempDir := "namespaces/testCluster/testNamespace"
	tempFile := tempDir + "/SECRET_ROTATE_BLOCK"

	if err := os.MkdirAll(tempFile, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tempFile); err != nil {
			t.Fatal(err)
		}
	}()

	if secretBlockerExists(tempDir) != true {
		t.Errorf("secretBlocker should return true as the file does exist")
	}

}
