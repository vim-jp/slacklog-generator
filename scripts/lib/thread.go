package slacklog

import (
	"time"
)

// Thread : スレッド
// rootMsgはスレッドの先頭メッセージを表わす。
// repliesにはそのスレッドへの返信メッセージが入る。先頭メッセージは含まない。
type Thread struct {
	rootMsg *Message
	replies []Message
}

func (t Thread) LastReplyTime() time.Time {
	return TsToDateTime(t.replies[len(t.replies)-1].Ts)
}

func (t Thread) ReplyCount() int {
	return len(t.replies)
}

func (t Thread) RootText() string {
	return t.rootMsg.Text
}

func (t Thread) Replies() []Message {
	return t.replies
}
