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
		if err := os.WriteFile(localCSV, data, 0o644); err != nil {
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

	nsMap := make(map[string][]string)
	for _, record := range records {
		if len(record) < 2 {
			log.Printf("Skipping incomplete record: %v\n", record)
			continue
		}
		namespace := strings.TrimSpace(record[0])
		errorMsg := strings.TrimSpace(strings.Join(record[1:], " "))
		nsMap[namespace] = append(nsMap[namespace], errorMsg)
	}

	ghClient := github.NewGithubClient(&github.GithubClientConfig{
		Repository: "cloud-platform-environments",
		Owner:      "ministryofjustice",
	}, os.Getenv("TF_VAR_github_token"))

	successes := make(map[string]string)
	failures := make(map[string]string)

	for namespace, errorsList := range nsMap {
		combinedErrorMsg := strings.Join(errorsList, "\n")
		prURL, err := processRecord(namespace, combinedErrorMsg, ghClient)
		if err != nil {
			log.Printf("Failed to process namespace %s: %v\n\n", namespace, err)
			failures[namespace] = err.Error()
			continue
		}
		successes[namespace] = prURL
	}

	dupNsCount := 0
	totalRdsDowngrade := 0
	errInfo := []string{}
	for ns, msgs := range nsMap {
		if len(msgs) > 1 {
			dupNsCount++
			totalRdsDowngrade += len(msgs)
			errInfo = append(errInfo, fmt.Sprintf("%s (%d downgrade)", ns, len(msgs)))
		}
	}

	log.Println("\n\n==================== SUMMARY ====================")
	fmt.Printf("Total CSV records: %d\n", len(records))
	fmt.Printf("Unique namespaces processed: %d\n", len(nsMap))
	if dupNsCount > 0 {
		fmt.Printf("\nNamespaces with multiple RDS downgrade: %d (containing %d RDS downgrade records)\n", dupNsCount, totalRdsDowngrade)
		for _, info := range errInfo {
			fmt.Printf("  - %s\n", info)
		}
	}
	if len(successes) > 0 {
		fmt.Printf("\nSuccessfully created PRs (%d):\n", len(successes))
		for ns, url := range successes {
			fmt.Printf("  - %s: %s\n", ns, url)
		}
	}
	if len(failures) > 0 {
		fmt.Printf("\nFailed to process namespaces (%d):\n", len(failures))
		for ns, reason := range failures {
			fmt.Printf("  - %s: %s\n", ns, reason)
		}
	}

	criticalFailureCount := 0
	for _, reason := range failures {
		if strings.Contains(reason, "PR creation failed") {
			criticalFailureCount++
		}
	}

	if criticalFailureCount > 0 {
		return fmt.Errorf("failed to create PR for %d namespaces", criticalFailureCount)
	}

	return nil
}

func downloadFromS3(s3Location, localPath string) error {
	cmd := exec.Command("aws", "s3", "cp", s3Location, localPath)
	return cmd.Run()
}

func processRecord(namespace, csvErr string, ghClient github.GithubIface) (string, error) {
	log.Printf("Processing namespace: %s", namespace)

	results, err := IsRdsVersionMismatched(csvErr)
	if err != nil {
		return "", fmt.Errorf("parsing terraform error failed: %v", err)
	}

	tfDir := "namespaces/live.cloud-platform.service.justice.gov.uk/" + namespace + "/resources"

	var filesChanged []string
	versionDescription := fmt.Sprintf("- Fix Terraform RDS version drift for namespace: `%s`\n\n```", namespace)

	for i, versions := range results.Versions {
		moduleName := results.ModuleNames[i][0]
		actualVersion := versions[0]
		tfVersion := versions[1]

		files, updateErr := updateVersion(moduleName, actualVersion, tfVersion, tfDir)
		if updateErr != nil {
			return "", fmt.Errorf("error updating Terraform: %v", updateErr)
		}

		for _, f := range files {
			duplicate := false
			for _, v := range filesChanged {
				if f == v {
					duplicate = true
					break
				}
			}
			if !duplicate {
				filesChanged = append(filesChanged, f)
			}
		}
		versionDescription += fmt.Sprintf("\nmodule.%s: downgrade from %s to %s", moduleName, actualVersion, tfVersion)
	}

	versionDescription += "\n```"
	description := versionDescription
	prCreator := createPR(description, namespace, os.Getenv("TF_VAR_github_token"), "cloud-platform-environments")
	prUrl, err := prCreator(ghClient, filesChanged)
	if err != nil {
		return "", fmt.Errorf("PR creation failed: %v", err)
	}

	log.Printf("Successfully created PR: %s\n\n", prUrl)
	postPR(prUrl, os.Getenv("SLACK_WEBHOOK_URL"))
	return prUrl, nil
}
