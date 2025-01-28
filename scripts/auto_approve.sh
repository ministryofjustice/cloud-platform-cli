#!/bin/sh
set -eu

PLAN_FILE=$1
PLAN_DIR=$2
PLAN_NAME=$3
PR=$4

JSON_FILE="${PLAN_NAME%.out}.json"

terraform -chdir="$PLAN_DIR" show -json "$PLAN_NAME" > "$JSON_FILE"

OUTPUT=$(opa exec --decision terraform/analysis/allow --bundle opa-auto-approve-policy/ "$JSON_FILE")
OPA_RESULT=$(echo "$OUTPUT" | jq -r '.result[0].result')

CHANGED_FILES=$(curl -L \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GITHUB_TOKEN" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/ministryofjustice/cloud-platform-environments/pulls/$PR/files" |  jq -r '.[].filename')

YAML_CHANGES=0
for f in $CHANGED_FILES; do
    if [[ "$f" != "namespaces/live.cloud-platform.service.justice.gov.uk/"*"/resources/"* ]]; then
        YAML_CHANGES=1
        break
    fi
done

if [ "$OPA_RESULT" == "true" ] && [ "$YAML_CHANGES" -eq 0 ]; then
    curl -L \
    -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GITHUB_TOKEN" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/ministryofjustice/cloud-platform-environments/pulls/$PR/reviews" \
    -d '{
        "body": "Automatically approving PR",
        "event": "APPROVE"
    }'
else
    REASON=""
    if [ "$OPA_RESULT" != "true" ]; then
    REASON="OPA auto approval policy check failed"
    fi

    if [ "$YAML_CHANGES" -ne 0 ]; then
    if [ -n "$REASON" ]; then
        REASON="$REASON and changes exist outside the 'resources/' folder"
    else
        REASON="Changes exist outside the 'resources/' folder"
    fi
    fi

    curl -L \
    -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GITHUB_TOKEN" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/ministryofjustice/cloud-platform-environments/issues/$PR/comments" \
    -d '{
        "body": "This PR requires approval from the Cloud Platform team.\n Reason: '"$REASON"'.\n Please raise it in #ask-cloud-platform Slack channel."
    }'
fi

exit 0