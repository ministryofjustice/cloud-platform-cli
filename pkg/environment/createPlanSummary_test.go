package environment_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment"
)

func readPlanFromJson(path string) (*tfjson.Plan, error) {
	var initPlan tfjson.Plan

	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteVal, readErr := io.ReadAll(jsonFile)

	if readErr != nil {
		return nil, readErr
	}

	unMarshallErr := initPlan.UnmarshalJSON(byteVal)

	if unMarshallErr != nil {
		return nil, unMarshallErr
	}

	return &initPlan, nil
}

func Test_CreateCommentBody(t *testing.T) {
	changesOutput := `
<h1>Terraform Plan Summary</h1>

<details "open">
	<summary>
		<b>Terraform Plan: %d to be created, %d to be destroyed, %d to be updated, %d to be replaced and %d unchanged.</b>
	</summary>

%s

</details>
`

	createChangesDiff := "#### Resources to create:\n```" + `diff
+ azurerm_linux_virtual_machine.calvinvm
+ azurerm_network_interface.calvin-nic
+ azurerm_network_interface_security_group_association.calvin-sg-nic
+ azurerm_network_security_group.calvin-security-group
+ azurerm_public_ip.calvin-ip
+ azurerm_resource_group.calvin
+ azurerm_storage_account.calvin-sa
+ azurerm_subnet.calvin-subnet
+ azurerm_virtual_network.calvin-vn
+ random_id.calvin-rid
+ tls_private_key.calvin_ssh
` + "```\n\n\n"

	mixedChangesDiff := "#### Resources to create:\n```" + `diff
+ azurerm_linux_virtual_machine.calvinvm
+ azurerm_network_security_group.calvin-security-group
+ azurerm_public_ip.calvin-ip
+ azurerm_resource_group.calvin
+ azurerm_storage_account.calvin-sa
+ azurerm_subnet.calvin-subnet
+ azurerm_virtual_network.calvin-vn
+ random_id.calvin-rid
+ tls_private_key.calvin_ssh
` + "```\n\n\n#### Resources to destroy:\n```" + `diff
- azurerm_network_interface.calvin-nic
` + "```\n\n\n#### Resources to update:\n```" + `diff
! azurerm_network_interface_security_group_association.calvin-sg-nic
` + "```\n"

	replacedChangesDiff := "#### Resources to create:\n```" + `diff
+ module.aks_lz.module.key_vault.azurerm_private_endpoint.key_vault
` + "```\n\n\n\n#### Resources to update:\n```" + `diff
! module.aks_lz.module.key_vault.azurerm_key_vault.key_vault
` + "```\n\n\n#### Resources to replace:\n```" + `diff
-+ module.aks_lz.azurerm_virtual_hub_connection.aks_vnet_hub_connection[0]
` + "```"

	createChangesExpected := fmt.Sprintf(changesOutput, 11, 0, 0, 0, 0, createChangesDiff)
	mixedChangesExpected := fmt.Sprintf(changesOutput, 9, 1, 1, 0, 0, mixedChangesDiff)
	replacedChangesExpected := fmt.Sprintf(changesOutput, 1, 0, 1, 1, 35, replacedChangesDiff)

	tfNoChangesPlan, _ := readPlanFromJson("./fixtures/tf_nochanges.json")
	tfChangesPlan01, _ := readPlanFromJson("./fixtures/tf_test01.json")
	tfChangesPlan02, _ := readPlanFromJson("./fixtures/tf_test02.json")
	tfChangesPlan03, _ := readPlanFromJson("./fixtures/tf_test03.json")

	type args struct {
		tfPlan *tfjson.Plan
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"GIVEN a terraform plan with no changes THEN return a comment body stating so", args{tfNoChangesPlan}, "\n```diff\n+ There are no terraform changes to apply```\n",
		},
		{
			"GIVEN a terraform plan with CREATE changes THEN return a comment body with correct changes", args{tfChangesPlan01}, createChangesExpected,
		},
		{
			"GIVEN a terraform plan with CREATE, UPDATE, DESTROY changes THEN return a comment body with correct changes", args{tfChangesPlan02}, mixedChangesExpected,
		},
		{
			"GIVEN a terraform plan with CREATE, DESTROY, REPLACE changes THEN return a comment body with correct changes", args{tfChangesPlan03}, replacedChangesExpected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := environment.CreateCommentBody(tt.args.tfPlan); got != tt.want {
				t.Errorf("createCommentBody() = %v, want %v", got, tt.want)
			}
		})
	}
}
