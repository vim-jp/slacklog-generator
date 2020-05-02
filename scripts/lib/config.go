package slacklog

// Config : ログ出力時の設定を保持する。
type Config struct {
	EditedSuffix  string   `json:"edited_suffix"`
	Channels      []string `json:"channels"`
	EmojiJSONPath string   `json:"emoji_json_path"`
}

// ReadConfig : pathに指定したファイルからコンフィグを読み込む。
func ReadConfig(path string) (*Config, error) {
	var cfg Config
	if err := ReadFileAsJSON(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
