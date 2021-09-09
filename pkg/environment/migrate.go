package environment

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gookit/color"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	otiai10 "github.com/otiai10/copy"
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

	color.Info.Printf("\nNamespace %s was succesffully migrated to live folder\n", nsName)

	// recursive grep in Golang
	filepath.Walk(".", func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			envHasIAMannotation := grepFile(path, []byte("iam.amazonaws.com/permitted"))
			envHasElasticSearch := grepFile(path, []byte("github.com/ministryofjustice/cloud-platform-terraform-elasticsearch"))

			if envHasIAMannotation >= 1 {
				color.Error.Println("\nIMPORTANT: This namespace uses IAM policies - please contact Cloud-Platform team before proceeding")
			}

			if envHasElasticSearch >= 1 {
				// color.Error.Println("\nIMPORTANT: This namespace uses ElasticSearch module - please contact Cloud-Platform team before proceeding")
				err = changeElasticSearch(path)
				if err != nil {
					log.Println(err)
				}
			}
		}
		return nil
	})

	return nil
}

func changeElasticSearch(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
	}

	f, diags := hclwrite.ParseConfig(data, file, hcl.Pos{
		Line:   1,
		Column: 1,
	})

	if diags.HasErrors() {
		return fmt.Errorf("Error getting TF resource: %s", diags)
	}

	blocks := f.Body().Blocks()

	f.Body().RemoveBlock(blocks[0])

	block := f.Body().AppendBlock(blocks[0])

	blockBody := block.Body()

	blockBody.SetAttributeValue("irsa_enabled", cty.StringVal("true"))
	blockBody.SetAttributeValue("assume_enabled", cty.StringVal("false"))

	fmt.Println(string(f.Bytes()))

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
