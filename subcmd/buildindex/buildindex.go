package buildindex

import (
	"fmt"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

func run(datadir, outdir, config string) error {
	configJSONPath := filepath.Clean(config)
	cfg, err := slacklog.ReadConfig(configJSONPath)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}
	s, err := slacklog.NewLogStore(datadir, cfg)
	if err != nil {
		return err
	}

	i := slacklog.NewIndexer(s)
	err = i.Build()
	if err != nil {
		return err
	}

	err = i.Output(outdir)
	if err != nil {
		return err
	}

	return nil
}

func NewCLICommand() *cli.Command {
	var (
		datadir string
		outdir  string
		config  string
	)
	return &cli.Command{
		Name:  "build-index",
		Usage: "build index for searching",
		Action: func(c *cli.Context) error {
			return run(datadir, outdir, config)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "config.json path",
				Value:       filepath.Join("scripts", "config.json"),
				Destination: &config,
			},
			&cli.StringFlag{
				Name:        "datadir",
				Usage:       "directory to load/save data",
				Value:       "_logdata",
				Destination: &datadir,
			},
			&cli.StringFlag{
				Name:        "outdir",
				Usage:       "directory to output result",
				Destination: &outdir,
			},
		},
	}
}
