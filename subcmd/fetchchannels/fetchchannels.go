package fetchchannels

import (
	"context"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/internal/jsonwriter"
	"github.com/vim-jp/slacklog-generator/internal/slackadapter"
)

func run(token, datadir string, excludeArchived, verbose bool) error {
	outfile := filepath.Join(datadir, "channels.json")
	fw, err := jsonwriter.CreateFile(outfile, true)
	if err != nil {
		return err
	}
	err = slackadapter.IterateCursor(context.Background(),
		slackadapter.CursorIteratorFunc(func(ctx context.Context, c slackadapter.Cursor) (slackadapter.Cursor, error) {
			r, err := slackadapter.Conversations(ctx, token, slackadapter.ConversationsParams{
				Cursor:          c,
				Limit:           100,
				ExcludeArchived: excludeArchived,
				Types:           []string{"public_channel"},
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
		token           string
		datadir         string
		excludeArchived bool
		verbose         bool
	)
	return &cli.Command{
		Name:  "fetch-channels",
		Usage: "fetch channels in the workspace",
		Action: func(c *cli.Context) error {
			return run(token, datadir, excludeArchived, verbose)
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
				Name:        "exclude-archived",
				Usage:       "exclude archived channesls",
				Destination: &excludeArchived,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "verbose log",
				Destination: &verbose,
			},
		},
	}
}
