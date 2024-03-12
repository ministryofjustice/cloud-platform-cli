package slack

import "github.com/slack-go/slack"

func initSlack(token string) *slack.Client {
	return slack.New(token)
}
