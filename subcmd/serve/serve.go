package serve

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	cli "github.com/urfave/cli/v2"
)

// Command provides "serve" sub-command. It starts HTML server for slacklog for
// development.
var Command = &cli.Command{
	Name:   "serve",
	Usage:  "serve a generated HTML with files proxy",
	Action: run,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "addr",
			Usage: "address for serve",
			Value: ":8081",
		},
		&cli.StringFlag{
			Name:  "htdocs",
			Usage: "root of files",
			Value: "_site",
		},
		&cli.StringFlag{
			Name:  "target",
			Usage: "proxy target endpoint",
			Value: "https://vim-jp.org/slacklog/",
		},
	},
}

func newReverseProxy(targetURL string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	rp := httputil.NewSingleHostReverseProxy(u)
	orig := rp.Director
	rp.Director = func(r *http.Request) {
		orig(r)
		r.Host = "" //u.Host
	}
	return rp, nil
}

// run starts combined HTTP server (file + reverse proxy) to serve slacklog for
// development.
func run(c *cli.Context) error {
	addr := c.String("addr")
	htdocs := c.String("htdocs")
	target := c.String("target")

	proxy, err := newReverseProxy(target)
	if err != nil {
		return err
	}
	http.Handle("/files/", proxy)
	http.Handle("/emojis/", proxy)
	http.Handle("/", http.FileServer(http.Dir(htdocs)))

	return http.ListenAndServe(addr, nil)
}
