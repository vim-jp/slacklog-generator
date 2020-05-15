package subcmd

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/subcmd/serve"
)

var commands = []*cli.Command{
	ConvertExportedLogsCommand,
	DownloadEmojiCommand,
	DownloadFilesCommand,
	GenerateHTMLCommand,
	serve.Command,
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
