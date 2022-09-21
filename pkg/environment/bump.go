package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// BumpModule takes a module name, walks the environments repository and
// changes the version of all modules with that name.
func BumpModule(m, v string) error {
	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".tf" {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s", err)
			}

			f, diags := hclwrite.ParseConfig(data, path, hcl.Pos{
				Line:   1,
				Column: 1,
			})

			if diags.HasErrors() {
				return fmt.Errorf("error getting TF resource: %s", diags)
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
				valueTokens = append(valueTokens, exprTokens...)

				blockSource := strings.TrimSpace(string(valueTokens.Bytes()))

				if strings.Contains(blockSource, m) {
					src := strings.SplitAfter(blockSource, "=")
					s := strings.Split(src[0], "\"")
					val := s[1] + v

					blockBody.SetAttributeValue("source", cty.StringVal(val))
				}
				err = os.WriteFile(path, f.Bytes(), 0o644)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
