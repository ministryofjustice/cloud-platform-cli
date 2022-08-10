package commands

import (
	"errors"
	"path/filepath"

	environment "github.com/ministryofjustice/cloud-platform-cli/pkg/environment"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/util/homedir"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

var module, moduleVersion string

var optFlags environment.Options

func addEnvironmentCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(environmentCmd)
	environmentCmd.AddCommand(environmentEcrCmd)
	environmentCmd.AddCommand(environmentRdsCmd)
	environmentCmd.AddCommand(environmentS3Cmd)
	environmentCmd.AddCommand(environmentSvcCmd)
	environmentCmd.AddCommand(environmentCreateCmd)
	environmentCmd.AddCommand(environmentApplyCmd)
	environmentEcrCmd.AddCommand(environmentEcrCreateCmd)
	environmentRdsCmd.AddCommand(environmentRdsCreateCmd)
	environmentS3Cmd.AddCommand(environmentS3CreateCmd)
	environmentSvcCmd.AddCommand(environmentSvcCreateCmd)
	environmentCmd.AddCommand(environmentPrototypeCmd)
	environmentPrototypeCmd.AddCommand(environmentPrototypeCreateCmd)
	environmentCmd.AddCommand(environmentBumpModuleCmd)

	// flags
	environmentApplyCmd.Flags().StringVarP(&optFlags.Namespace, "namespace", "n", "", "Namespace which you want to perform the apply")
	environmentApplyCmd.Flags().StringVar(&optFlags.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	environmentApplyCmd.Flags().StringVar(&optFlags.ClusterCtx, "cluster", "live.cloud-platform.service.justice.gov.uk", "path to kubeconfig file")
	environmentBumpModuleCmd.Flags().StringVarP(&module, "module", "m", "", "Module to upgrade the version")
	environmentBumpModuleCmd.Flags().StringVarP(&moduleVersion, "module-version", "v", "", "Semantic version to bump a module to")
}

var environmentCmd = &cobra.Command{
	Use:    "environment",
	Short:  `Cloud Platform Environment actions`,
	PreRun: upgradeIfNotLatest,
}

var environmentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create an environment`,
	Example: heredoc.Doc(`
	$ cloud-platform environment create
	`),
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateNamespace,
}

var environmentEcrCmd = &cobra.Command{
	Use:   "ecr",
	Short: `Add an ECR to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment ecr create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: `apply command to apply the given namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment apply -n <namespace>
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		contextLogger := log.WithFields(log.Fields{"subcommand": "apply"})
		applier := &environment.Apply{
			Options: &optFlags,
		}
		err := applier.Apply()
		if err != nil {
			contextLogger.Fatal(err)
		}
	},
}

var environmentEcrCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/ecr.tf" terraform file for an ECR`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateEcr,
}

var environmentRdsCmd = &cobra.Command{
	Use:   "rds",
	Short: `Add an RDS instance to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment rds create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentRdsCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/rds.tf" terraform file for an RDS instance`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateRds,
}

var environmentS3Cmd = &cobra.Command{
	Use:   "s3",
	Short: `Add a S3 bucket to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment s3 create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentS3CreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/s3.tf" terraform file for a S3 bucket`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateS3,
}

var environmentSvcCmd = &cobra.Command{
	Use:   "serviceaccount",
	Short: `Add a serviceaccount to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment serviceaccount
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentSvcCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Creates a serviceaccount in your chosen namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment serviceaccount create
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := environment.CreateTemplateServiceAccount(); err != nil {
			return err
		}

		return nil
	},
}

var environmentPrototypeCmd = &cobra.Command{
	Use:   "prototype",
	Short: `Create a gov.uk prototype kit site on the cloud platform`,
	Example: heredoc.Doc(`
	$ cloud-platform environment prototype
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentPrototypeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create an environment to host gov.uk prototype kit site on the cloud platform`,
	Long: `
Create a namespace folder and files in an existing prototype github repository to host a Gov.UK
Prototype Kit website on the Cloud Platform.

The namespace name should be your prototype github repository name:

  https://github.com/ministryofjustice/[repository name]
	`,
	Example: heredoc.Doc(`
	$ cloud-platform environment prototype create
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := environment.CreateTemplatePrototype(); err != nil {
			return err
		}

		return nil
	},
}

var environmentBumpModuleCmd = &cobra.Command{
	Use:   "bump-module",
	Short: `Bump all specified module versions`,
	Example: heredoc.Doc(`
cloud-platform environments bump-module --module serviceaccount --module-version 1.1.1

Would bump all users serviceaccount modules in the environments repository to the specified version.
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		if moduleVersion == "" || module == "" {
			return errors.New("--module and --module-version are required")
		}

		if err := environment.BumpModule(module, moduleVersion); err != nil {
			return err
		}
		return nil
	},
}
