package slacklog

import (
	"time"
)

type Thread struct {
	msgs []Message
}

func (t Thread) LastReplyTime() time.Time {
	return TsToDateTime(t.msgs[len(t.msgs)-1].Ts)
}

func (t Thread) ReplyNum() int {
	return len(t.msgs) - 1
}

func (t Thread) RootText() string {
	return t.msgs[0].Text
}

func (t Thread) Replies() []Message {
	return t.msgs[1:]
}
