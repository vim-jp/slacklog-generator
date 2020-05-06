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

	slacklog "github.com/vim-jp/slacklog/lib"
)

func DownloadFiles(args []string) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return fmt.Errorf("$SLACK_TOKEN required")
	}

	if len(args) < 2 {
		fmt.Println("Usage: go run scripts/main.go download_files {log-dir} {files-dir}")
		return nil
	}

	logDir := filepath.Clean(args[0])
	filesDir := filepath.Clean(args[1])

	s, err := slacklog.NewLogStore(logDir, &slacklog.Config{Channels: []string{"*"}})
	if err != nil {
		return err
	}

	d := slacklog.NewDownloader(slackToken)

	go slacklog.GenerateMessageFileTargets(d, s, filesDir)
	err = d.DownloadAll()
	if err != nil {
		return err
	}
	return nil
}
