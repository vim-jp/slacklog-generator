/*
Package store defines types which be used in storage interface and
implementations.
*/
package store

import (
	"errors"

	"github.com/slack-go/slack"
)

// ErrIterateAbort is returned when iteration "break"ed.
var ErrIterateAbort = errors.New("iteration aborted")

// Channel represents channel object in Slack.
type Channel struct {
	slack.Channel

	Pins []Pin `json:"pins"`
}

// Tidy removes sensitive data from channel.
func (c *Channel) Tidy() {
	// nothing to do for now.
}

// Pin represents a pinned message for a channel.
type Pin struct {
	ID      string `json:"id"`
	Typ     string `json:"type"`
	Created int64  `json:"created"`
	User    string `json:"user"`
	Owner   string `json:"owner"`
}

// ChannelIterator is callback for channel iteration.
type ChannelIterator interface {
	Iterate(*Channel) bool
}

// ChannelIterateFunc is a function wrapper for ChannelIterator.
type ChannelIterateFunc func(*Channel) bool

// Iterate implements ChannelIterator.
func (fn ChannelIterateFunc) Iterate(c *Channel) bool {
	return fn(c)
}

var _ ChannelIterator = ChannelIterateFunc(nil)

// User represents user object in Slack.
type User slack.User

// Tidy removes sensitive data from user.
func (u *User) Tidy() {
	// nothing to do for now.
}

// Emoji represents an emoji object in Slack.
type Emoji struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Tidy removes sensitive data from emoji.
func (e *Emoji) Tidy() {
	// nothing to do for now.
}

// Message represents message object in Slack.
type Message slack.Message

// MessageTx defines transactional object for messages.
type MessageTx interface {
	Upsert(Message) (bool, error)

	Iterate(key TimeKey, iter MessageIterator) error

	Count(key TimeKey) (int, error)

	Commit() error

	Rollback() error
}

// MessageIterator is callback for message iteration.
type MessageIterator interface {
	Iterate(*Message) bool
}
