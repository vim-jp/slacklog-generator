package subcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

func FetchMessages(args []string) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return fmt.Errorf("$SLACK_TOKEN required")
	}

	if len(args) < 2 {
		fmt.Println("Usage: go run scripts/main.go update-messages {data-dir} {yyyy-mm-dd}")
		return nil
	}

	dataDir := filepath.Clean(args[0])

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return err
	}

	date, err := time.ParseInLocation("2006-01-02", args[1], loc)
	if err != nil {
		return err
	}
	dateEndTime := date.AddDate(0, 0, 1)

	channelsJSONPath := filepath.Join(dataDir, "channels.json")
	channelTable, err := slacklog.NewChannelTable(channelsJSONPath, []string{"*"})
	if err != nil {
		return err
	}

	for _, channel := range channelTable.Channels {
		channelDir := filepath.Join(dataDir, channel.ID)

		extraParams := map[string]string{
			"channel": channel.ID,
			"oldest":  strconv.FormatInt(date.Unix(), 10) + ".000000",
			"latest":  strconv.FormatInt(dateEndTime.Unix(), 10) + ".000000",
			"limit":   "200",
		}
		outFile := filepath.Join(channelDir, date.Format("2006-01-02")+".json")
		err := slacklog.DownloadEntitiesToFile(slackToken, "conversations.history", extraParams, "messages", true, outFile)
		if err != nil {
			return err
		}
	}

	return nil
}
