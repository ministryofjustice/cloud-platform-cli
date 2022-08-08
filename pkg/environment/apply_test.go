package environment

import (
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment/mocks"
	"github.com/stretchr/testify/assert"
)

// func TestApply_ApplyNamespace(t *testing.T) {
// 	type fields struct {
// 		RequiredEnvVars RequiredEnvVars
// 		Dir             string
// 		Namespace       string
// 	}
// 	tests := []struct {
// 		name              string
// 		fields            fields
// 		TerraformOutputs  map[string]string
// 		checkExpectations func(t *testing.T, terraform *mocks.Terraformer, outputs map[string]string, err error)
// 	}{
// 		{
// 			name: "Apply foo namespace",
// 			fields: fields{
// 				RequiredEnvVars: RequiredEnvVars{
// 					clustername:        "cluster01",
// 					clusterstatebucket: "clusterStateBucket",
// 					clusterstatekey:    "clusterStateKey",
// 					githubowner:        "githubowner",
// 					githubtoken:        "githubtoken",
// 					pingdomapitoken:    "pingdomApikey",
// 				},
// 				Dir:       "/root/foo",
// 				Namespace: "foobar",
// 			},
// 			TerraformOutputs: map[string]string{"myoutput": "foo"},
// 			checkExpectations: func(t *testing.T, terraform *mocks.Terraformer, outputs map[string]string, err error) {
// 				terraform.AssertCalled(t, "TerraformInitAndApply", "foobar", "/root/foo")
// 				assert.Nil(t, err)
// 				assert.Len(t, outputs, 1)
// 			},
// 		},
// 	}
// 	for i := range tests {
// 		terraform := new(mocks.Terraformer)
// 		terraform.On("TerraformInitAndApply", tests[i].fields.Namespace, tests[i].fields.Dir).Return(tests[i].TerraformOutputs, nil)
// 		a := Apply{
// 			RequiredEnvVars: tests[i].fields.RequiredEnvVars,
// 			Terraformer:     terraform,
// 			Dir:             tests[i].fields.Dir,
// 			Namespace:       tests[i].fields.Namespace,
// 		}
// 		outputs, err := a.ApplyNamespace()
// 		t.Run(tests[i].name, func(t *testing.T) {
// 			tests[i].checkExpectations(t, terraform, outputs, err)
// 		})
// 	}
// }

func TestApply_ApplyNamespace(t *testing.T) {
	type fields struct {
		Options         *Options
		RequiredEnvVars RequiredEnvVars
		Terraformer     Terraformer
		Dir             string
	}
	tests := []struct {
		name              string
		fields            fields
		TerraformOutputs  map[string]string
		checkExpectations func(t *testing.T, terraform *mocks.Terraformer, outputs map[string]string, err error)
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
			TerraformOutputs: map[string]string{"myoutput": "foo"},
			checkExpectations: func(t *testing.T, terraform *mocks.Terraformer, outputs map[string]string, err error) {
				terraform.AssertCalled(t, "TerraformInitAndApply", "foobar", "/root/foo")
				assert.Nil(t, err)
				assert.Len(t, outputs, 1)
			},
		},
	}
	for i := range tests {
		terraform := new(mocks.Terraformer)
		terraform.On("TerraformInitAndApply", tests[i].fields.Options.Namespace, tests[i].fields.Dir).Return(tests[i].TerraformOutputs, nil)
		a := Apply{
			RequiredEnvVars: tests[i].fields.RequiredEnvVars,
			Terraformer:     terraform,
			Dir:             tests[i].fields.Dir,
			Options:         tests[i].fields.Options,
		}
		outputs, err := a.ApplyNamespace()
		t.Run(tests[i].name, func(t *testing.T) {
			tests[i].checkExpectations(t, terraform, outputs, err)
		})
	}
}
