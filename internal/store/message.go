package store

import (
	"time"

	"github.com/slack-go/slack"
)

// Message represents message object in Slack.
type Message slack.Message

// TimestampTime converts "Timestamp" to `time.Time`.
func (m Message) TimestampTime() (time.Time, error) {
	tv, err := TimestampToTime(m.Timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return tv, nil
}

// Tidy removes sensitive data from a message.
func (m *Message) Tidy() {
	// nothing to do for now.
}

// Before returns true when `m` is older than `b`
func (m Message) Before(b Message) bool {
	ta, _ := m.TimestampTime()
	tb, _ := b.TimestampTime()
	return ta.Before(tb)
}

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
