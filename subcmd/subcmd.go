package subcmd

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/subcmd/serve"
)

var commands = []*cli.Command{
	{
		Name: "convert-exported-logs",
		Usage: "convert slack exported logs to API download logs",
		Action: ConvertExportedLogs,
		Flags: ConvertExportedLogsFlags,
	},
	{
		Name: "download-emoji",
		Usage: "download custamized emoji from slack",
		Action: DownloadEmoji,
		Flags: EmojisFlags,
	},
	{
		Name: "download-files",
		Usage: "download files from slack.com",
		Action: DownloadFiles,
		Flags: FilesFlags,
	},
	{
		Name: "generate-html",
		Usage: "generate html from slacklog_data",
		Action: GenerateHTML,
		Flags: GenerateHTMLFlags,
	},
	{
		Name: "serve",
		Usage: "serve a generated HTML with files proxy",
		Action: serve.Run,
		Flags: serve.Flags,
	},
}

// Run runs one of sub-commands.
func Run() error {
	fmt.Println(os.Args)
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "slacklog-generator"
	app.Usage = "generate slacklog HTML"
	app.Version = "0.0.0"
	app.Commands = commands

	return app.Run(os.Args)
}
