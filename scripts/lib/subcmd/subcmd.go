package subcmd

import (
	"fmt"
	"os"
)

func Run() error {
	fmt.Println(os.Args)
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/main.go {subcmd}")
		fmt.Println("  Subcmd:")
		fmt.Println("    convert-exported-logs")
		fmt.Println("    download-emoji")
		fmt.Println("    download-files")
		fmt.Println("    generate-html")
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
	}

	return fmt.Errorf("Unknown subcmd: %s", subCmdName)
}
