package subcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

func UpdateUserList(args []string) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return fmt.Errorf("$SLACK_TOKEN required")
	}

	if len(args) < 1 {
		fmt.Println("Usage: go run scripts/main.go update-user-list {out-file}")
		return nil
	}

	outFile := filepath.Clean(args[0])

	err := slacklog.DownloadEntitiesToFile(slackToken, "users.list", nil, "members", false, outFile)
	if err != nil {
		return err
	}

	return nil
}
