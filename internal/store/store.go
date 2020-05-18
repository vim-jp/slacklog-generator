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
