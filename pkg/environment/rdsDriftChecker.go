package environment

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
	"github.com/spf13/cobra"
)

func RdsDriftChecker(cmd *cobra.Command, args []string) error {
	sourceLocation := args[0]
	localCSV := "merged-rds-errored-namespaces.csv"

	if strings.HasPrefix(sourceLocation, "file://") {
		localFilePath := strings.TrimPrefix(sourceLocation, "file://")
		fmt.Printf("Using local CSV file: %s\n", localFilePath)

		data, err := os.ReadFile(localFilePath)
		if err != nil {
			return fmt.Errorf("error reading local CSV: %v", err)
		}
		if err := os.WriteFile(localCSV, data, 0644); err != nil {
			return fmt.Errorf("error copying local CSV: %v", err)
		}
	} else {
		fmt.Printf("Downloading CSV from S3: %s\n", sourceLocation)
		if err := downloadFromS3(sourceLocation, localCSV); err != nil {
			return fmt.Errorf("error downloading CSV: %v", err)
		}
	}

	file, err := os.Open(localCSV)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	log.Printf("📥 Loaded CSV file with %d records\n", len(records))

	namespaceMap := make(map[string][]string)
	for _, record := range records {
		if len(record) < 2 {
			log.Printf("⚠️ Skipping incomplete record: %v\n", record)
			continue
		}
		namespace := strings.TrimSpace(record[0])
		errorMsg := strings.TrimSpace(strings.Join(record[1:], " "))
		log.Printf("🧩 Grouping error under namespace '%s'\n", namespace)
		namespaceMap[namespace] = append(namespaceMap[namespace], errorMsg)
	}

	log.Printf("🗂️ Total unique namespaces found: %d\n", len(namespaceMap))

	ghClient := github.NewGithubClient(&github.GithubClientConfig{
		Repository: "cloud-platform-environments",
		Owner:      "ministryofjustice",
	}, os.Getenv("TF_VAR_github_token"))

	for namespace, errorsList := range namespaceMap {
		combinedErrorMsg := strings.Join(errorsList, "\n")
		if err := processRecord(namespace, combinedErrorMsg, ghClient); err != nil {
			log.Printf("❌ Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
	}

	return nil
}

func downloadFromS3(s3Location, localPath string) error {
	cmd := exec.Command("aws", "s3", "cp", s3Location, localPath)
	return cmd.Run()
}

func processRecord(namespace, tfErr string, ghClient github.GithubIface) error {
	log.Printf("🚀 Processing namespace: %s", namespace)

	results, err := IsRdsVersionMismatched(tfErr)
	if err != nil {
		return fmt.Errorf("parsing terraform error failed: %v", err)
	}

	tfDir := "namespaces/live.cloud-platform.service.justice.gov.uk/" + namespace + "/resources"
	log.Printf("📂 Target Terraform directory: %s", tfDir)

	var filesChanged []string
	versionDescription := fmt.Sprintf("- Fix Terraform RDS version drift for namespace: `%s`\n\n```", namespace)

	for i, versions := range results.Versions {
		moduleName := results.ModuleNames[i][0]
		actualVersion := versions[0]
		tfVersion := versions[1]

		log.Printf("🔧 Updating module '%s': Terraform version %s → Actual version %s", moduleName, tfVersion, actualVersion)

		file, updateErr := updateVersion(moduleName, actualVersion, tfVersion, tfDir)
		if updateErr != nil {
			return fmt.Errorf("error updating Terraform: %v", updateErr)
		}

		filesChanged = append(filesChanged, file)
		versionDescription += fmt.Sprintf("\nmodule.%s: downgrade from %s to %s", moduleName, tfVersion, actualVersion)
	}

	versionDescription += "\n```"

	description := versionDescription
	prCreator := createPR(description, namespace, os.Getenv("TF_VAR_github_token"), "cloud-platform-environments")
	prUrl, err := prCreator(ghClient, filesChanged)
	if err != nil {
		return fmt.Errorf("PR creation failed: %v", err)
	}

	log.Printf("✅ Successfully created PR: %s\n", prUrl)
	postPR(prUrl, os.Getenv("SLACK_WEBHOOK_URL"))
	return nil
}
