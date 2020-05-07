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

func (th Thread) LastReplyTime() time.Time {
	return TsToDateTime(th.replies[len(th.replies)-1].Ts)
}

func (th Thread) ReplyCount() int {
	return len(th.replies)
}

func (th Thread) RootText() string {
	return th.rootMsg.Text
}

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
