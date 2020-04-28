package slacklog

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	EditedSuffix string   `json:"edited_suffix"`
	Channels     []string `json:"channels"`
	EmojiJson    string   `json:"emoji_json"`
}

func ReadConfig(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(content, &cfg)
	return &cfg, err
}
