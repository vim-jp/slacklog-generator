package slacklog

import (
	"encoding/json"
	"io/ioutil"
)

type EmojiTable struct {
	m map[string]string
}

func NewEmojiTable(path string) *EmojiTable {
	emojis := &EmojiTable{
		m: map[string]string{},
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return emojis
	}

	json.Unmarshal(content, emojis)

	return emojis
}
