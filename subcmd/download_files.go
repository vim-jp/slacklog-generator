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

	"github.com/vim-jp/slacklog-generator/internal/slacklog"
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

	go generateMessageFileTargets(d, s, filesDir)

	err = d.Wait()
	if err != nil {
		return err
	}
	return nil
}

func generateMessageFileTargets(d *slacklog.Downloader, s *slacklog.LogStore, outputDir string) {
	defer d.CloseQueue()
	channels := s.GetChannels()
	for _, channel := range channels {
		msgs, err := s.GetAllMessages(channel.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get messages on %s channel: %s", channel.Name, err)
			return
		}

		for _, msg := range msgs {
			for _, f := range msg.Files {
				targetDir := filepath.Join(outputDir, f.ID)
				err := os.MkdirAll(targetDir, 0777)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to create %s directory: %s", targetDir, err)
					return
				}

				for url, suffix := range f.DownloadURLsAndSuffixes() {
					if url == "" {
						continue
					}
					d.QueueDownloadRequest(
						url,
						filepath.Join(targetDir, f.DownloadFilename(url, suffix)),
						true,
					)
				}
			}
		}
	}
}
