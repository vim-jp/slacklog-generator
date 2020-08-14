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
	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

// DownloadFilesCommand provides "downloads-files" sub-command, it downloads
// and saves files which attached to message.
var DownloadFilesCommand = &cli.Command{
	Name:   "download-files",
	Usage:  "download files from slack.com",
	Action: downloadFiles,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "indir",
			Usage: "slacklog_data dir",
			Value: filepath.Join("_logdata", "slacklog_data"),
		},
		&cli.StringFlag{
			Name:  "outdir",
			Usage: "files download target dir",
			Value: filepath.Join("_logdata", "files"),
		},
	},
}

// downloadFiles downloads and saves files which attached to message.
func downloadFiles(c *cli.Context) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return fmt.Errorf("$SLACK_TOKEN required")
	}

	logDir := filepath.Clean(c.String("indir"))
	filesDir := filepath.Clean(c.String("outdir"))

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

func urlAndSuffixes(f slack.File) map[string]string {
	return map[string]string{
		f.URLPrivate:   "",
		f.Thumb64:      "_64",
		f.Thumb80:      "_80",
		f.Thumb160:     "_160",
		f.Thumb360:     "_360",
		f.Thumb480:     "_480",
		f.Thumb720:     "_720",
		f.Thumb800:     "_800",
		f.Thumb960:     "_960",
		f.Thumb1024:    "_1024",
		f.Thumb360Gif:  "_360_gif",
		f.Thumb480Gif:  "_480_gif",
		f.DeanimateGif: "_deanimate_gif",
		f.ThumbVideo:   "_video",
	}
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
				if !slacklog.HostBySlack(f) {
					continue
				}

				targetDir := filepath.Join(outputDir, f.ID)
				err := os.MkdirAll(targetDir, 0777)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to create %s directory: %s", targetDir, err)
					return
				}

				for url, suffix := range urlAndSuffixes(f) {
					if url == "" {
						continue
					}
					d.QueueDownloadRequest(
						url,
						filepath.Join(targetDir, slacklog.LocalName(f, url, suffix)),
						true,
					)
				}
			}
		}
	}
}
