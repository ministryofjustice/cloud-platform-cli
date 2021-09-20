package environment

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
	otiai10 "github.com/otiai10/copy"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// Migrate subcommand copy a namespace folder from live-1 -> live directory.
// It also performs some basic checks (IAM roles & and ElasticSearch) to ensure
// the namespace will not have problem during the migration
func Migrate(skipWarning bool) error {
	re := RepoEnvironment{}

	// this already checks we are within the environment repo.
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	nsName, err := re.getNamespaceName()
	if err != nil {
		return err
	}

	if skipWarning == false {
		ann, err := hasExternalDNSAnnotations(nsName)
		if err != nil {
			return err
		}

		if ann == false {
			color.Error.Printf("Namespace: %s doesn't have the correct ingress annotation.", nsName)
		}
	}

	src := fmt.Sprintf("../%s", nsName)
	dst := fmt.Sprintf("../../live.cloud-platform.service.justice.gov.uk/%s", nsName)

	err = otiai10.Copy(src, dst)
	if err != nil {
		return err
	}

	// Required to skip the live-1 pipeline
	emptyFile, err := os.Create("APPLY_PIPELINE_SKIP_THIS_NAMESPACE")
	if err != nil {
		return err
	}
	emptyFile.Close()

	if err := os.Chdir(dst); err != nil {
		return err
	}

	// recursive grep in Golang
	filepath.Walk(".", func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			envHasIAMannotation := grepFile(path, []byte("iam.amazonaws.com/permitted"))
			envHasElasticSearch := grepFile(path, []byte("github.com/ministryofjustice/cloud-platform-terraform-elasticsearch"))
			envHasServiceAccount := grepFile(path, []byte("github.com/ministryofjustice/cloud-platform-terraform-serviceaccount"))

			if envHasServiceAccount >= 1 {
				color.Error.Println("\nIMPORTANT: This namespace uses ServiceAccount module - please contact Cloud-Platform team before proceeding")
			}

			if envHasIAMannotation >= 1 {
				color.Error.Println("\nIMPORTANT: This namespace uses IAM policies - please contact Cloud-Platform team before proceeding")
			}

			if envHasElasticSearch >= 1 {
				err = changeElasticSearch(path)
				if err != nil {
					color.Error.Println("\nIMPORTANT: There was an error changing your ElasticSearch module - please contact Cloud-Platform team before proceeding", err)
				}
			}
		}
		return nil
	})

	color.Info.Printf("\nNamespace %s was succesffully migrated to live folder\n", nsName)

	return nil
}

func grepFile(file string, pat []byte) int64 {
	patCount := int64(0)
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if bytes.Contains(scanner.Bytes(), pat) {
			patCount++
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return patCount
}

// changeElasticSearch takes the location of a Terraform HCL file (as a string)
// parses, then locates the ElasticSearch block. To add two additional attributes
// the block is removed and recreated at the bottom of the file.
func changeElasticSearch(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Error reading file %s", err)
	}

	f, diags := hclwrite.ParseConfig(data, file, hcl.Pos{
		Line:   1,
		Column: 1,
	})

	if diags.HasErrors() {
		return fmt.Errorf("Error getting TF resource: %s", diags)
	}

	// Grab slice of blocks in HCL file.
	blocks := f.Body().Blocks()

	for _, block := range blocks {
		blockBody := block.Body()
		if blockBody.Attributes()["source"] == nil {
			continue
		}
		expr := blockBody.Attributes()["source"].Expr()
		exprTokens := expr.BuildTokens(nil)
		var valueTokens hclwrite.Tokens
		for _, t := range exprTokens {
			valueTokens = append(valueTokens, t)
		}
		blockSource := strings.TrimSpace(string(valueTokens.Bytes()))
		if strings.HasPrefix(blockSource, "\"github.com/ministryofjustice/cloud-platform-terraform-elasticsearch") {
			blockBody.SetAttributeValue("irsa_enabled", cty.BoolVal(true))
			blockBody.SetAttributeValue("assume_enabled", cty.BoolVal(false))
		}
	}

	err = os.WriteFile(file, f.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
