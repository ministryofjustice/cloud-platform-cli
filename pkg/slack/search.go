package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

func search(client *slack.Client, params slack.SearchParameters, query string) (*slack.SearchMessages, error) {
	results, searchErr := client.SearchMessages(query, params)
	if searchErr != nil {
		fmt.Printf("Error: searching for slack messages: %v\n", searchErr.Error())
		return nil, searchErr
	}

	if results == nil {
		fmt.Printf("Failed to find a slack message with the PR number %s\n", query)
		return nil, nil
	}

	return results, nil
}
