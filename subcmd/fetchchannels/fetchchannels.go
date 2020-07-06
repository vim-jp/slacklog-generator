package fetchchannels

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/internal/jsonwriter"
	"github.com/vim-jp/slacklog-generator/internal/slackadapter"
)

// Run runs "fetch-channels" sub-command. It fetch channels in the workspace.
func Run(args []string) error {
	var (
		token   string
		datadir string
		verbose bool
	)
	fs := flag.NewFlagSet("fetch-channels", flag.ExitOnError)
	fs.StringVar(&token, "token", os.Getenv("SLACK_TOKEN"), `slack token. can be set by SLACK_TOKEN env var`)
	fs.StringVar(&datadir, "datadir", "_logdata", `directory to load/save data`)
	fs.BoolVar(&verbose, "verbose", false, "verbose log")
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if token == "" {
		return errors.New("SLACK_TOKEN environment variable requied")
	}
	return run(token, datadir, verbose)
}

func run(token, datadir string, verbose bool) error {
	outfile := filepath.Join(datadir, "channels.json")
	fw, err := jsonwriter.CreateFile(outfile, true)
	if err != nil {
		return err
	}
	err = slackadapter.IterateCursor(context.Background(),
		slackadapter.CursorIteratorFunc(func(ctx context.Context, c slackadapter.Cursor) (slackadapter.Cursor, error) {
			r, err := slackadapter.Conversations(ctx, token, slackadapter.ConversationsParams{
				Cursor: c,
				Limit:  100,
			})
			if err != nil {
				return "", err
			}
			for _, c := range r.Channels {
				err := fw.Write(c)
				if err != nil {
					return "", err
				}
			}
			if m := r.ResponseMetadata; m != nil {
				return m.NextCursor, nil
			}
			return "", nil
		}))
	if err != nil {
		// ロールバック相当が好ましいが今はまだその時期ではない
		fw.Close()
		return err
	}
	if err := fw.Close(); err != nil {
		return err
	}

	return nil
}

// NewCLICommand creates a cli.Command, which provides "fetch-channels"
// sub-command.
func NewCLICommand() *cli.Command {
	var (
		token   string
		datadir string
		verbose bool
	)
	return &cli.Command{
		Name:  "fetch-channels",
		Usage: "fetch channels in the workspace",
		Action: func(c *cli.Context) error {
			return run(token, datadir, verbose)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "token",
				Usage:       "slack token",
				EnvVars:     []string{"SLACK_TOKEN"},
				Destination: &token,
			},
			&cli.StringFlag{
				Name:        "datadir",
				Usage:       "directory to load/save data",
				Value:       "_logdata",
				Destination: &datadir,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "verbose log",
				Destination: &verbose,
			},
		},
	}
}
