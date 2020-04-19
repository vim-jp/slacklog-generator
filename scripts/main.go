package main

import (
	"fmt"
	"os"

	slacklog "github.com/vim-jp/slacklog/lib"
)

func main() {
	if err := slacklog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %s\n", err)
		os.Exit(1)
	}
}
