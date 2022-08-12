package environment

import (
	"testing"

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
					clusterstatebucket: "clusterStateBucket",
					clusterstatekey:    "clusterStateKey",
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
					clusterstatebucket: "clusterStateBucket",
					clusterstatekey:    "clusterStateKey",
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
