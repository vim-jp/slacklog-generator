package slackadapter

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
