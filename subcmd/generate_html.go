package subcmd

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

// GenerateHTML : SlackからエクスポートしたデータをHTMLに変換して出力する。
func GenerateHTML(c *cli.Context) error {
	var configJSONPath, templateDir, filesDir, inDir, outDir string
	if c.Args().Present() {
		configJSONPath = filepath.Clean(c.Args().Get(0))
		templateDir = filepath.Clean(c.Args().Get(1))
		filesDir = filepath.Clean(c.Args().Get(2))
		inDir = filepath.Clean(c.Args().Get(3))
		outDir = filepath.Clean(c.Args().Get(4))
	} else {
		configJSONPath = filepath.Clean(filepath.Join("scripts", "config.json"))
		templateDir = filepath.Clean("templates")
		filesDir = filepath.Clean(filepath.Join("_logdata", "files"))
		inDir = filepath.Clean(filepath.Join("_logdata", "slacklog_data"))
		outDir = filepath.Clean("_site")
	}

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
