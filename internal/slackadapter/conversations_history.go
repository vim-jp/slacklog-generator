package slackadapter

import (
	"context"
	"errors"
	"time"

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

// ConversationsHistoryReponse is response for ConversationsHistory
type ConversationsHistoryReponse struct {
	Ok               bool                `json:"ok"`
	Messages         []*slacklog.Message `json:"messages,omitempty"`
	HasMore          bool                `json:"has_more"`
	PinCount         int                 `json:"pin_count"`
	ResponseMetadata *NextCursor         `json:"response_metadata"`
}

// ConversationsHistory gets conversation messages in a channel.
func ConversationsHistory(ctx context.Context, token, channel string, params ConversationsHistoryParams) (*ConversationsHistoryReponse, error) {
	// TODO: call Slack's conversations.history
	return nil, errors.New("not implemented yet")
}
