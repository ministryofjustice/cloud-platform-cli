package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

func Notify(prNumber, token, webhookUrl, buildUrl string) error {
	slackClient := initSlack(token)

	defaultSearchParams := slack.NewSearchParameters()

	results, searchErr := search(slackClient, defaultSearchParams, prNumber)

	if searchErr != nil {
		fmt.Printf("Failed to find pr in slack %v\n", searchErr)
		return nil
	}

	// get the user who posted
	user := results.Matches[0].User
	ts := results.Matches[0].Timestamp

	return post(user, ts, webhookUrl, buildUrl)
}

func PostToAsk(prUrl, webhookUrl string) error {
	webhookMsg := slack.WebhookMessage{
		Channel: "ask-cloud-platform",
		Text:    "PR for review please: " + prUrl,
	}

	return slack.PostWebhook(webhookUrl, &webhookMsg)
}
