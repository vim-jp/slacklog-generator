package slacklog

import (
	"encoding/json"
	"io/ioutil"
)

// Config : ログ出力時の設定を保持する。
type Config struct {
	EditedSuffix string   `json:"edited_suffix"`
	Channels     []string `json:"channels"`
	EmojiJson    string   `json:"emoji_json"`
}

// ReadConfig : pathに指定したファイルからコンフィグを読み込む。
func ReadConfig(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
