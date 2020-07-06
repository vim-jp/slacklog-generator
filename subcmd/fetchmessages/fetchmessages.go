package fetchmessages

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"time"

	cli "github.com/urfave/cli/v2"
	"github.com/vim-jp/slacklog-generator/internal/jsonwriter"
	"github.com/vim-jp/slacklog-generator/internal/slackadapter"
	"github.com/vim-jp/slacklog-generator/internal/slacklog"
)

const dateFormat = "2006-01-02"

func toDateString(ti time.Time) string {
	return ti.Format(dateFormat)
}

func parseDateString(s string) (time.Time, error) {
	l, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, err
	}
	ti, err := time.ParseInLocation(dateFormat, s, l)
	if err != nil {
		return time.Time{}, err
	}
	return ti, nil
}

// Run runs "fetch-messages" sub-command. It fetch messages of a channel by a
// day.
func Run(args []string) error {
	var (
		token   string
		datadir string
		date    string
		verbose bool
	)
	fs := flag.NewFlagSet("fetch-messages", flag.ExitOnError)
	fs.StringVar(&token, "token", os.Getenv("SLACK_TOKEN"), `slack token. can be set by SLACK_TOKEN env var`)
	fs.StringVar(&datadir, "datadir", "_logdata", `directory to load/save data`)
	fs.StringVar(&date, "date", toDateString(time.Now()), `target date to get`)
	fs.BoolVar(&verbose, "verbose", false, "verbose log")
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if token == "" {
		return errors.New("SLACK_TOKEN environment variable requied")
	}
	return run(token, datadir, date, verbose)
}

func run(token, datadir, date string, verbose bool) error {
	oldest, err := parseDateString(date)
	if err != nil {
		return err
	}
	latest := oldest.AddDate(0, 0, 1)

	ct, err := slacklog.NewChannelTable(filepath.Join(datadir, "channels.json"), []string{"*"})
	if err != nil {
		return err
	}

	for _, sch := range ct.Channels {
		outfile := filepath.Join(datadir, sch.ID, toDateString(oldest)+".json")
		fw, err := jsonwriter.CreateFile(outfile, true)
		if err != nil {
			return err
		}
		err = slackadapter.IterateCursor(context.Background(),
			slackadapter.CursorIteratorFunc(func(ctx context.Context, c slackadapter.Cursor) (slackadapter.Cursor, error) {
				r, err := slackadapter.ConversationsHistory(ctx, token, sch.ID, slackadapter.ConversationsHistoryParams{
					Cursor: c,
					Limit:  100,
					Oldest: &oldest,
					Latest: &latest,
				})
				if err != nil {
					return "", err
				}
				for _, m := range r.Messages {
					err := fw.Write(m)
					if err != nil {
						return "", err
					}
				}
				if m := r.ResponseMetadata; r.HasMore && m != nil {
					return m.NextCursor, nil
				}
				// HasMore && ResponseMetadata == nil は明らかにエラーだがいま
				// は握りつぶしてる
				return "", nil
			}))
		if err != nil {
			// ロールバック相当が好ましいが今はまだその時期ではない
			fw.Close()
			return err
		}
		err = fw.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// NewCLICommand creates a cli.Command, which provides "fetch-messages"
// sub-command.
func NewCLICommand() *cli.Command {
	var (
		token   string
		datadir string
		date    string
		verbose bool
	)
	return &cli.Command{
		Name:  "fetch-messages",
		Usage: "fetch messages of channel by day",
		Action: func(c *cli.Context) error {
			return run(token, datadir, date, verbose)
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
			&cli.StringFlag{
				Name:        "date",
				Usage:       "target date to get",
				Value:       toDateString(time.Now()),
				Destination: &date,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "verbose log",
				Destination: &verbose,
			},
		},
	}
}
