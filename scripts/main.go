package main

import (
	"fmt"
	"os"

	"github.com/vim-jp/slacklog/lib/subcmd"
)

func main() {
	if err := subcmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %s\n", err)
		os.Exit(1)
	}
}
