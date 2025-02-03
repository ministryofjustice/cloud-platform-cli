package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

func post(user, ts, webhookUrl, buildUrl string) error {
	// https://pkg.go.dev/github.com/slack-go/slack#PostWebhook
	message := fmt.Sprintf("<@%s> <%s|your build failed>, please review and resolve issues, or add an <https://user-guide.cloud-platform.service.justice.gov.uk/documentation/other-topics/concourse-pipelines.html#apply-live-pipelines-and-apply-pipeline-skip-this-namespace|APPLY_PIPELINE_SKIP_THIS_NAMESPACE> to your namespace.", user, buildUrl)

	webhookMsg := slack.WebhookMessage{
		Channel:         "ask-cloud-platform",
		Text:            message,
		ThreadTimestamp: ts,
		ReplyBroadcast:  true,
	}

	return slack.PostWebhook(webhookUrl, &webhookMsg)
}
