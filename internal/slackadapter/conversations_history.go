package slackadapter

import (
	"context"
	"time"

	"github.com/slack-go/slack"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

// ConversationsHistoryParams is optional parameters for ConversationsHistory
type ConversationsHistoryParams struct {
	Cursor    Cursor     `json:"cursor,omitempty"`
	Inclusive bool       `json:"inclusive,omitempty"`
	Latest    *time.Time `json:"latest,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Oldest    *time.Time `json:"oldest,omitempty"`
}

// ConversationsHistoryResponse is response for ConversationsHistory
type ConversationsHistoryResponse struct {
	Ok               bool                `json:"ok"`
	Messages         []*slacklog.Message `json:"messages,omitempty"`
	HasMore          bool                `json:"has_more"`
	PinCount         int                 `json:"pin_count"`
	ResponseMetadata *NextCursor         `json:"response_metadata"`
}

// ConversationsHistory gets conversation messages in a channel.
func ConversationsHistory(ctx context.Context, token, channel string, params ConversationsHistoryParams) (*ConversationsHistoryResponse, error) {
	client := slack.New(token)
	res, err := client.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
		ChannelID: channel,
		Cursor:    string(params.Cursor),
		Limit:     params.Limit,
		Oldest:    Timestamp(params.Oldest),
		Latest:    Timestamp(params.Latest),
		Inclusive: params.Inclusive,
	})
	if err != nil {
		return nil, err
	}

	var messages []*slacklog.Message
	for _, m := range res.Messages {
		messages = append(messages, &slacklog.Message{
			Message: m,
		})
	}

	return &ConversationsHistoryResponse{
		Ok:       true,
		Messages: messages,
		HasMore:  res.HasMore,
		PinCount: res.PinCount,
		ResponseMetadata: &NextCursor{
			NextCursor: Cursor(res.ResponseMetaData.NextCursor),
		},
	}, nil
}
