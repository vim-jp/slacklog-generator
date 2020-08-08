package slackadapter

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

// UsersResponse is response for Conversations
type UsersResponse struct {
	Ok    bool             `json:"ok"`
	Users []*slacklog.User `json:"users,omitempty"`
}

// Users gets users.
func Users(ctx context.Context, token string) ([]*slacklog.User, error) {
	client := slack.New(token)
	users, err := client.GetUsersContext(ctx)
	if err != nil {
		return nil, err
	}

	var logUsers []*slacklog.User
	for _, u := range users {
		lu := slacklog.User(u)
		logUsers = append(logUsers, &lu)
	}

	return logUsers, nil
}
