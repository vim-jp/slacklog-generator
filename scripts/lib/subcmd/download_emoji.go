/*
リファクタリング中
処理をslacklog packageに移動していく。
一旦、必要な処理はすべてslacklog packageから一時的にエクスポートするか、このファ
イル内で定義している。
*/

package subcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/slack-go/slack"
	slacklog "github.com/vim-jp/slacklog/lib"
)

func DownloadEmoji(args []string) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return fmt.Errorf("$SLACK_TOKEN required")
	}

	if len(args) < 1 {
		fmt.Println("Usage: go run scripts/main.go download_emoji {emojis-dir}")
		return nil
	}

	emojisDir := filepath.Clean(args[0])
	emojiJSONPath := filepath.Join(emojisDir, "emoji.json")
	if 1 < len(args) {
		emojiJSONPath = filepath.Clean(args[1])
	}

	api := slack.New(slackToken)
	d := slacklog.NewDownloader(slackToken)
	go slacklog.GenerateEmojiFileTargets(d, api, emojisDir, emojiJSONPath)
	err := d.DownloadAll()
	if err != nil {
		return err
	}
	return nil
}
