package subcmd

import (
	"fmt"
	"path/filepath"

	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

// GenerateHTML : SlackからエクスポートしたデータをHTMLに変換して出力する。
func GenerateHTML(args []string) error {
	if len(args) < 4 {
		fmt.Println("Usage: go run scripts/main.go generate_html {config.json} {templatedir} {indir} {outdir}")
		return nil
	}
	configJSONPath := filepath.Clean(args[0])
	templateDir := filepath.Clean(args[1])
	inDir := filepath.Clean(args[2])
	outDir := filepath.Clean(args[3])

	cfg, err := slacklog.ReadConfig(configJSONPath)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	s, err := slacklog.NewLogStore(inDir, cfg)
	if err != nil {
		return err
	}

	g := slacklog.NewHTMLGenerator(templateDir, s)
	return g.Generate(outDir)
}
