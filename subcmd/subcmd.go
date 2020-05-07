package subcmd

import (
	"fmt"
	"os"
)

func Run() error {
	fmt.Println(os.Args)
	if len(os.Args) < 2 {
		fmt.Println(`Usage: go run . {subcmd}
  Subcmd:
    convert-exported-logs
    download-emoji
    download-files
    generate-html
    update-channel-list
    update-user-list`)
		return nil
	}

	args := os.Args[2:]
	subCmdName := os.Args[1]
	switch subCmdName {
	case "convert-exported-logs":
		return ConvertExportedLogs(args)
	case "download-emoji":
		return DownloadEmoji(args)
	case "download-files":
		return DownloadFiles(args)
	case "generate-html":
		return GenerateHTML(args)
	case "update-channel-list":
		return UpdateChannelList(args)
	case "update-user-list":
		return UpdateUserList(args)
	}

	return fmt.Errorf("unknown subcmd: %s", subCmdName)
}
