package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

func Notify(prNumber, token, webhookUrl, buildUrl string) error {
	slackClient := initSlack(token)

	defaultSearchParams := slack.NewSearchParameters()

	results, searchErr := search(slackClient, defaultSearchParams, prNumber)

	// Don't post if slack search failed
	if searchErr != nil {
		fmt.Printf("Failed to find pr in slack with error: %v\n", searchErr)
		return nil
	}

	// Don't post if no matching posts found
    if len(results.Matches) == 0 {
        fmt.Println("Failed to find pr in slack: no matches")
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
		Text:    "RDS Minor version mismatch. PR for review please: " + prUrl,
	}

	return slack.PostWebhook(webhookUrl, &webhookMsg)
}
