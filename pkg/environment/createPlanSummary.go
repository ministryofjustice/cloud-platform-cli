package environment

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
)

func CreateCommentBody(tfPlan *tfjson.Plan) string {
	resourcesToUpdate := []string{}
	resourcesToDelete := []string{}
	resourcesToReplace := []string{}
	resourcesUnchanged := []string{}
	resourcesToCreate := []string{}
	body := `
<h1>Terraform Plan Summary</h1>
`

	for _, resource := range tfPlan.ResourceChanges {
		changes := resource.Change

		if len(changes.Actions) > 0 {
			address := resource.Address
			switch action := changes.Actions[0]; action {
			case "no-op":
				resourcesUnchanged = append(resourcesUnchanged, address)
			case "create":
				resourcesToCreate = append(resourcesToCreate, address)
			case "delete":
				if len(changes.Actions) > 1 {
					resourcesToReplace = append(resourcesToReplace, address)
				} else {
					resourcesToDelete = append(resourcesToDelete, address)
				}
			case "update":
				resourcesToUpdate = append(resourcesToUpdate, address)
			default:
				break
			}

		}
	}

	body += `
<details "open">
	<summary>
		<b>Terraform Plan: %d to be created, %d to be destroyed, %d to be updated, %d to be replaced and %d unchanged.</b>
	</summary>
%s
%s
%s
%s
</details>
`
	body = fmt.Sprintf(body, len(resourcesToCreate), len(resourcesToDelete), len(resourcesToUpdate), len(resourcesToReplace), len(resourcesUnchanged), details("create", "+", resourcesToCreate), details("destroy", "-", resourcesToDelete), details("update", "!", resourcesToUpdate), details("replace", "-+", resourcesToReplace))

	if len(resourcesToCreate) == 0 && len(resourcesToReplace) == 0 && len(resourcesToUpdate) == 0 && len(resourcesToDelete) == 0 {
		body = "\n```diff\n+ There are no terraform changes to apply\n```\n"
	}

	return body
}

func details(action, operator string, resources []string) string {
	var str string

	if len(resources) > 0 {
		str = `
#### Resources to %s:
`
		str = fmt.Sprintf(str, action)

		str += "```diff\n"
		for _, el := range resources {
			str += operator + " " + el + "\n"
		}

		str += "```\n"
	}

	return str
}

func CreateComment(gh github.GithubIface, tfplan *tfjson.Plan, prNum int) error {
	body := CreateCommentBody(tfplan)

	return gh.CreateComment(prNum, body)
}
