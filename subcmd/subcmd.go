package subcmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

var commands = []*cli.Command{
	{
		Name: "convert-exported-logs",
		Usage: "convert-exported-logs {indir} {outdir}",
		Action: ConvertExportedLogs,
	},
	{
		Name: "download-emoji",
		Usage: "download-emoji {outdir}",
		Action: DownloadEmoji,
	},
	{
		Name: "download-files",
		Usage: "download-files {outdir}",
		Action: DownloadFiles,
	},
	{
		Name: "generate-html",
		Usage: "generate-html {config.json} {templatedir} {filesdir} {indir} {outdir}",
		Action: GenerateHTML,
	},
}

// Run runs one of sub-commands.
func Run() error {
	fmt.Println(os.Args)
	app := cli.NewApp()
	app.Name = "slacklog-generator"
	app.Usage = "generate slacklog HTML"
	app.Version = "0.0.0"
	app.Commands = commands

	return app.Run(os.Args)
}
