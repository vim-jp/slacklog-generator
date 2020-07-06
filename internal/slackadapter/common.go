package slackadapter

import (
	"fmt"
	"time"
)

// Error represents error response of Slack.
type Error struct {
	Ok  bool   `json:"ok"`
	Err string `json:"error"`
}

// Error returns error message.
func (err *Error) Error() string {
	return err.Err
}

// NextCursor is cursor for next request.
type NextCursor struct {
	NextCursor Cursor `json:"next_cursor"`
}

// Cursor is type of cursor of Slack API.
type Cursor string

// Timestamp converts time.Time to timestamp formed for
// Slack API (<UNIX seconds>.<microseconds>)
func Timestamp(t *time.Time) string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%d.%6d", t.Unix(), t.Nanosecond()/1000)
}
