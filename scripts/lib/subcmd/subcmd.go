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
		fmt.Println("    convert_exported_logs")
		fmt.Println("    download_emoji")
		fmt.Println("    download_files")
		fmt.Println("    generate_html")
		return nil
	}

	args := os.Args[2:]
	subCmdName := os.Args[1]
	switch subCmdName {
	case "convert_exported_logs":
		return ConvertExportedLogs(args)
	case "download_emoji":
		return DownloadEmoji(args)
	case "download_files":
		return DownloadFiles(args)
	case "generate_html":
		return GenerateHTML(args)
	}

	return fmt.Errorf("Unknown subcmd: %s", subCmdName)
}
