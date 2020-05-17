package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/subcmd"
	"github.com/vim-jp/slacklog-generator/subcmd/fetchmessages"
	"github.com/vim-jp/slacklog-generator/subcmd/serve"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("[WARN] failed to load .env files")
	}

	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "slacklog-generator"
	app.Usage = "generate slacklog HTML"
	app.Version = "0.0.0"
	app.Commands = []*cli.Command{
		subcmd.ConvertExportedLogsCommand, // "convert-exported-logs"
		subcmd.DownloadEmojiCommand,       // "download-emoji"
		subcmd.DownloadFilesCommand,       // "download-files"
		subcmd.GenerateHTMLCommand,        // "generate-html"
		serve.Command,                     // "serve"
		fetchmessages.NewCLICommand(),     // "fetch-messages"
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		os.Exit(1)
	}
}
