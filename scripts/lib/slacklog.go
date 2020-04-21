package slacklog

import (
	"fmt"
	"os"
)

func Run() error {
	fmt.Println(os.Args)
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/main.go {subcmd}")
		fmt.Println("  Subcmd:")
		fmt.Println("    - update")
		fmt.Println("    - convert")
		return nil
	}

	subCmdName := os.Args[1]
	switch subCmdName {
	case "convert":
		return doConvert()
	case "update":
		return doUpdate()
	}

	return fmt.Errorf("Unknown subcmd: %s", subCmdName)
}
