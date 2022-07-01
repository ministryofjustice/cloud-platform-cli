package environment

import (
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment/mocks"

	"github.com/stretchr/testify/assert"
)

func TestApply_applyNamespace(t *testing.T) {
	type fields struct {
		RequiredEnvVars RequiredEnvVars
		Dir             string
		Namespace       string
	}
	tests := []struct {
		name              string
		fields            fields
		checkExpectations func(t *testing.T, terraform *mocks.Terraformer, outputs map[string]string, err error)
	}{
		{
			name: "Apply foo namespace",
			fields: fields{
				RequiredEnvVars: RequiredEnvVars{
					clustername:        "cluster01",
					clusterstatebucket: "clusterStateBucket",
					clusterstatekey:    "clusterStateKey",
					githubowner:        "githubowner",
					githubtoken:        "githubtoken",
					pingdomapitoken:    "pingdomApikey",
				},
				Dir:       "/root/foo",
				Namespace: "foobar",
			},
			checkExpectations: func(t *testing.T, terraform *mocks.Terraformer, outputs map[string]string, err error) {
				terraform.AssertCalled(t, "TerraformInitAndApply", "/root/foo")
				assert.Nil(t, err)
				assert.Len(t, outputs, 1)
				assert.Equal(t, "old", outputs["myoutput"])
			},
		},
	}
	for _, tt := range tests {
		terraform := new(mocks.Terraformer)
		a := &Apply{
			RequiredEnvVars: tt.fields.RequiredEnvVars,
			Terraformer:     terraform,
			Dir:             tt.fields.Dir,
			Namespace:       tt.fields.Namespace,
		}
		outputs, err := a.applyNamespace()
		t.Run(tests[tt].name, func(t *testing.T) { tests[tt].CheckExpectations(t, terraform, state, outputs, err) })
	}
}
