package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/vim-jp/slacklog-generator/subcmd"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[warning] failed to load .env file\n")
	}
	if err := subcmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %s\n", err)
		os.Exit(1)
	}
}
