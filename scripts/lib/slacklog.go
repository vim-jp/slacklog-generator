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
		return nil
	}

	subCmdName := os.Args[1]
	switch subCmdName {
	case "update":
		return doUpdate()
	}

	return nil
}
