package environment

import (
	"reflect"
	"testing"
)

func TestTerraformerImpl_TerraformInitAndApply(t *testing.T) {
	type fields struct {
		terraformBinaryPath string
		terraformVersion    string
		config              EnvBackendConfigVars
	}
	type args struct {
		namespace string
		directory string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "foo cluster and ns-01 namespace",
			fields: fields{
				terraformBinaryPath: "/usr/local/bon/terraform",
				terraformVersion:    "0.14.6",
				config:              EnvBackendConfigVars{},
			},
			args: args{
				namespace: "ns-01",
				directory: "foo",
			},
			want: map[string]string{
				"ns-01": "Apply completed",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &TerraformerImpl{
				terraformBinaryPath: tt.fields.terraformBinaryPath,
				terraformVersion:    tt.fields.terraformVersion,
				config:              tt.fields.config,
			}
			got, err := m.TerraformInitAndApply(tt.args.namespace, tt.args.directory)
			if (err != nil) != tt.wantErr {
				t.Errorf("TerraformerImpl.TerraformInitAndApply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TerraformerImpl.TerraformInitAndApply() = %v, want %v", got, tt.want)
			}
		})
	}
}
