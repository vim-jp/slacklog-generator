package slacklog

import (
	"time"
)

// Thread : スレッド
// rootMsgはスレッドの先頭メッセージを表わす。
// repliesにはそのスレッドへの返信メッセージが入る。先頭メッセージは含まない。
type Thread struct {
	rootMsg *Message
	replies Messages
}

// LastReplyTime returns last replied time for the thread.
func (th Thread) LastReplyTime() time.Time {
	return TsToDateTime(th.replies[len(th.replies)-1].Timestamp)
}

// ReplyCount return counts of replied messages.
func (th Thread) ReplyCount() int {
	return len(th.replies)
}

// RootText returns text of root message of the thread.
func (th Thread) RootText() string {
	return th.rootMsg.Text
}

// Replies returns replied messages for the thread.
func (th Thread) Replies() Messages {
	return th.replies
}

// Put puts a message to the thread as "root" or "reply".
func (th *Thread) Put(m *Message) {
	if m.IsRootOfThread() {
		th.rootMsg = m
	} else {
		th.replies = append(th.replies, m)
	}
}
