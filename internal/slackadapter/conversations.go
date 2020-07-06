package slackadapter

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

// ConversationsParams is optional parameters for Conversations
type ConversationsParams struct {
	Cursor          Cursor   `json:"cursor,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	ExcludeArchived string   `json:"excludeArchived,omitempty"`
	Types           []string `json:"types,omitempty"`
}

// ConversationsResponse is response for Conversations
type ConversationsResponse struct {
	Ok               bool                `json:"ok"`
	Channels         []*slacklog.Channel `json:"channels,omitempty"`
	ResponseMetadata *NextCursor         `json:"response_metadata"`
}

// Conversations gets conversation channels in a channel.
func Conversations(ctx context.Context, token string, params ConversationsParams) (*ConversationsResponse, error) {
	client := slack.New(token)
	channels, nextCursor, err := client.GetConversationsContext(ctx, &slack.GetConversationsParameters{
		Cursor:          string(params.Cursor),
		Limit:           params.Limit,
		ExcludeArchived: params.ExcludeArchived,
		Types:           params.Types,
	})
	if err != nil {
		return nil, err
	}

	var logChannels []*slacklog.Channel
	for _, c := range channels {
		logChannels = append(logChannels, &slacklog.Channel{
			Channel: c,
		})
	}

	res := &ConversationsResponse{
		Ok:       true,
		Channels: logChannels,
	}
	if nextCursor != "" {
		res.ResponseMetadata = &NextCursor{
			NextCursor: Cursor(nextCursor),
		}
	}
	return res, nil
}
