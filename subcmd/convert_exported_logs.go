/*
リファクタリング中
処理をslacklog packageに移動していく。
一旦、必要な処理はすべてslacklog packageから一時的にエクスポートするか、このファ
イル内で定義している。
*/

package subcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

// ConvertExportedLogsCommand provides "convert-exported-logs". It converts log
// which exported from Slack, to generator can treat.
var ConvertExportedLogsCommand = &cli.Command{
	Name:   "convert-exported-logs",
	Usage:  "convert slack exported logs to API download logs",
	Action: convertExportedLogs,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "indir",
			Usage: "exported log dir",
			Value: "_old_logs",
		},
		&cli.StringFlag{
			Name:  "outdir",
			Usage: "slacklog_data dir",
			Value: filepath.Join("_logdata", "slacklog_data"),
		},
	},
}

// convertExportedLogs converts log which exported from Slack, to generator can
// treat.
func convertExportedLogs(c *cli.Context) error {
	inDir := filepath.Clean(c.String("indir"))
	outDir := filepath.Clean(c.String("outdir"))
	inChannelsFile := filepath.Join(inDir, "channels.json")

	channels, _, err := readChannels(inChannelsFile, []string{"*"})
	if err != nil {
		return fmt.Errorf("could not read channels.json: %w", err)
	}

	if err := os.MkdirAll(outDir, 0777); err != nil {
		return fmt.Errorf("could not create %s directory: %w", outDir, err)
	}

	err = copyFile(filepath.Join(inDir, "channels.json"), filepath.Join(outDir, "channels.json"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(inDir, "users.json"), filepath.Join(outDir, "users.json"))
	if err != nil {
		return err
	}

	for _, channel := range channels {
		messages, err := ReadAllMessages(filepath.Join(inDir, channel.Name))
		if err != nil {
			return err
		}

		for _, message := range messages {
			message.RemoveTokenFromURLs()
		}

		channelDir := filepath.Join(outDir, channel.ID)
		if err := os.MkdirAll(channelDir, 0777); err != nil {
			return fmt.Errorf("could not create %s directory: %w", channelDir, err)
		}

		messagesPerDay := groupMessagesByDay(messages)
		for key, msgs := range messagesPerDay {
			err = writeMessages(filepath.Join(channelDir, key+".json"), msgs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(from string, to string) error {
	r, err := os.Open(from)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)

	return err
}

func readChannels(channelJSONPath string, cfgChannels []string) ([]slacklog.Channel, map[string]*slacklog.Channel, error) {
	var channels []slacklog.Channel
	err := slacklog.ReadFileAsJSON(channelJSONPath, true, &channels)
	if err != nil {
		return nil, nil, err
	}

	channels = slacklog.FilterChannel(channels, cfgChannels)
	slacklog.SortChannel(channels)
	channelMap := make(map[string]*slacklog.Channel, len(channels))
	for i, ch := range channels {
		channelMap[ch.ID] = &channels[i]
	}
	return channels, channelMap, err
}

// ReadAllMessages reads all message files in the directory.
func ReadAllMessages(inDir string) ([]*slacklog.Message, error) {
	dir, err := os.Open(inDir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	r, err := regexp.Compile(`\d{4}-\d{2}-\d{2}\.json`)
	if err != nil {
		return nil, err
	}

	var messages []*slacklog.Message
	for _, name := range names {
		if !r.MatchString(name) {
			continue
		}
		var msgs []*slacklog.Message
		err := slacklog.ReadFileAsJSON(filepath.Join(inDir, name), false, &msgs)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msgs...)
	}

	return messages, nil
}

func groupMessagesByDay(messages []*slacklog.Message) map[string][]*slacklog.Message {
	messagesPerDay := map[string][]*slacklog.Message{}
	for _, msg := range messages {
		time := slacklog.TsToDateTime(msg.Timestamp).Format("2006-01-02")
		messagesPerDay[time] = append(messagesPerDay[time], msg)
	}
	return messagesPerDay
}

func writeMessages(filename string, messages []*slacklog.Message) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(messages)
	if err != nil {
		return err
	}
	return nil
}
