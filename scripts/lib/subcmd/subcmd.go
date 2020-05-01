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
		fmt.Println("    download_files")
		fmt.Println("    download_emoji")
		fmt.Println("    generate_html")
		return nil
	}

	subCmdName := os.Args[1]
	switch subCmdName {
	case "convert_exported_logs":
		return doConvertExportedLogs()
	case "download_emoji":
		return doDownloadEmoji()
	case "download_files":
		return doDownloadFiles()
	case "generate_html":
		return GenerateHTML(os.Args[2:])
	}

	return fmt.Errorf("Unknown subcmd: %s", subCmdName)
}
