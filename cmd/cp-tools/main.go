package main

import (
	"fmt"
	"os"

	actions "github.com/ministryofjustice/cloud-platform-tools/pkg/actions"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

func main() {

	log.SetOutput(os.Stdout)

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "terraform",
				Aliases: []string{"t"},
				Usage:   "Terraform actions",
				Subcommands: []*cli.Command{
					{
						Name:   "check-divergence",
						Usage:  "Check if there are divergences in terraform",
						Action: actions.TerraformCheckDivergence,
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "workspace", Value: "default"},
							&cli.StringFlag{Name: "var-file"},
						},
					},
					{
						Name:  "apply",
						Usage: "Execute terraform apply",
						Action: func(c *cli.Context) error {
							fmt.Println("This is going to execute terraform apply - This feature is coming soon...", c.Args().First())
							return nil
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "aws-access-key-id",
				EnvVars:     []string{"AWS_ACCESS_KEY_ID"},
				Usage:       "Access key required to execute terraform against AWS",
				Required:    true,
				DefaultText: "NO_DEFAULT",
			},
			&cli.StringFlag{
				Name:        "aws-secret-access-key",
				EnvVars:     []string{"AWS_SECRET_ACCESS_KEY"},
				Usage:       "Access key required to execute terraform against AWS",
				Required:    true,
				DefaultText: "NO_DEFAULT",
			},
		},
		Action: func(c *cli.Context) error {
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
