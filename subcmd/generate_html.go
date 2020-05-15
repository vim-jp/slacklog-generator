package subcmd

import (
	"fmt"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

var GenerateHTMLFlags =  []cli.Flag{
	&cli.StringFlag{
		Name: "config",
		Usage: "config.json path",
		Value: filepath.Join("scripts", "config.json"),
	},
	&cli.StringFlag{
		Name: "templatedir",
		Usage: "templates dir",
		Value: "templates",
	},
	&cli.StringFlag{
		Name: "filesdir",
		Usage: "files downloaded dir",
		Value: filepath.Join("_logdata", "files"),
	},
	&cli.StringFlag{
		Name: "indir",
		Usage: "slacklog_data dir",
		Value: filepath.Join("_logdata", "slacklog_data"),
	},
	&cli.StringFlag{
		Name: "outdir",
		Usage: "generated html target dir",
		Value: "_site",
	},
}

// GenerateHTML : SlackからエクスポートしたデータをHTMLに変換して出力する。
func GenerateHTML(c *cli.Context) error {
	configJSONPath := filepath.Clean(c.String("config"))
	templateDir := filepath.Clean(c.String("templatedir"))
	filesDir := filepath.Clean(c.String("filesdir"))
	inDir := filepath.Clean(c.String("indir"))
	outDir := filepath.Clean(c.String("outdir"))

	cfg, err := slacklog.ReadConfig(configJSONPath)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	s, err := slacklog.NewLogStore(inDir, cfg)
	if err != nil {
		return err
	}

	g := slacklog.NewHTMLGenerator(templateDir, filesDir, s)
	return g.Generate(outDir)
}
