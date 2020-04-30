package subcmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	slacklog "github.com/vim-jp/slacklog/lib"
)

const downloadWorkerNum = 8

func doDownloadFiles() error {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return fmt.Errorf("$SLACK_TOKEN required")
	}

	if len(os.Args) < 4 {
		fmt.Println("Usage: go run scripts/main.go download_files {log-dir} {files-dir}")
		return nil
	}

	logDir := filepath.Clean(os.Args[2])
	filesDir := filepath.Clean(os.Args[3])

	s, err := slacklog.NewLogStore(logDir, &slacklog.Config{Channels: []string{"*"}})
	if err != nil {
		return err
	}

	channels := s.GetChannels()

	if err := slacklog.Mkdir(filesDir); err != nil {
		return fmt.Errorf("could not create %s directory: %s", filesDir, err)
	}

	ch := make(chan *slacklog.MessageFile, downloadWorkerNum)
	wg := new(sync.WaitGroup)
	failed := false

	for i := 0; i < cap(ch); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range ch {
				errs := downloadAll(m, filesDir, slackToken)
				for i := range errs {
					failed = true
					fmt.Fprintf(os.Stderr, "[error] Download failed: %s\n", errs[i])
				}
			}
		}()
	}

	for _, channel := range channels {
		messages, err := slacklog.ReadAllMessages(filepath.Join(logDir, channel.ID))
		if err != nil {
			close(ch)
			return err
		}
		for _, message := range messages {
			for i := range message.Files {
				ch <- &message.Files[i]
			}
		}
	}

	close(ch)
	wg.Wait()

	if failed {
		return errors.New("failed to download some file(s)")
	}
	return nil
}

func urlToFilename(url string) string {
	i := strings.LastIndex(url, "/")
	if i < 0 {
		return ""
	}
	return url[i+1:]
}

func downloadAll(f *slacklog.MessageFile, outDir string, slackToken string) []error {
	fileBaseDir := path.Join(outDir, f.ID)
	err := slacklog.Mkdir(fileBaseDir)
	if err != nil {
		return []error{err}
	}

	var errs []error

	for url, suffix := range f.DownloadURLsAndSuffixes() {
		err = downloadFile(f, fileBaseDir, url, suffix, slackToken)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func downloadFile(f *slacklog.MessageFile, outDir, url, suffix, slackToken string) error {
	if url == "" {
		return nil
	}

	filename := f.DownloadFilename(url, suffix)

	destFile := filepath.Join(outDir, filename)
	if _, err := os.Stat(destFile); err == nil {
		// Just skip already downloaded file
		return nil
	}

	fmt.Printf("Downloading: %s/%s [%s]\n", f.ID, filename, f.PrettyType)

	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+slackToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("[%s]: %s", resp.Status, url)
	}

	w, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, resp.Body)

	return err
}
