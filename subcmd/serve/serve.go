package serve

import (
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
)

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

// Run starts combined HTTP server (file + reverse proxy) to serve slacklog for
// development.
func Run(args []string) error {
	var (
		addr   string
		htdocs string
		target string
	)
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.StringVar(&addr, "addr", ":8001", "address for serve")
	fs.StringVar(&htdocs, "htdocs", "_site", "root of files")
	fs.StringVar(&target, "target", "https://vim-jp.org/slacklog/", "proxy target endpoint")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	proxy, err := newReverseProxy(target)
	if err != nil {
		return err
	}
	http.Handle("/files/", proxy)
	http.Handle("/emojis/", proxy)
	http.Handle("/", http.FileServer(http.Dir(htdocs)))

	return http.ListenAndServe(addr, nil)
}
