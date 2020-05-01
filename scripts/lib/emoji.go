package slacklog

import (
	"os"
)

// EmojiTable : 絵文字データを保持する。
// urlMapは絵文字名をキーとし、画像が置いてあるURLが値である。
type EmojiTable struct {
	URLMap map[string]string
}

// NewEmojiTable : pathに指定したJSON形式の絵文字データを読み込み、EmojiTableを
// 生成する。
func NewEmojiTable(path string) (*EmojiTable, error) {
	emojis := &EmojiTable{
		URLMap: map[string]string{},
	}

	if info, err := os.Stat(path); err != nil || info.IsDir() {
		// pathにディレクトリが存在しても、その場合は無視して、ファイル自体が存在し
		// なかったこととする。
		return nil, os.ErrNotExist
	}

	if err := ReadFileAsJSON(path, &emojis.URLMap); err != nil {
		return nil, err
	}

	return emojis, nil
}
