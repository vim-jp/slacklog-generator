package slacklog

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

	channels, _, err := readChannels(filepath.Join(logDir, "channels.json"), []string{"*"})
	if err != nil {
		return fmt.Errorf("could not read channels.json: %s", err)
	}

	if err := mkdir(filesDir); err != nil {
		return fmt.Errorf("could not create %s directory: %s", filesDir, err)
	}

	ch := make(chan *messageFile, downloadWorkerNum)
	wg := new(sync.WaitGroup)
	failed := false

	for i := 0; i < cap(ch); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range ch {
				errs := m.downloadAll(filesDir, slackToken)
				for i := range errs {
					failed = true
					fmt.Fprintf(os.Stderr, "[error] Download failed: %s\n", errs[i])
				}
			}
		}()
	}

	for _, channel := range channels {
		messages, err := readAllMessages(filepath.Join(logDir, channel.Id))
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

func (f *messageFile) downloadURLs() []string {
	return []string{
		f.UrlPrivate,
		f.Thumb64,
		f.Thumb80,
		f.Thumb160,
		f.Thumb360,
		f.Thumb480,
		f.Thumb720,
		f.Thumb800,
		f.Thumb960,
		f.Thumb1024,
		f.Thumb360Gif,
		f.Thumb480Gif,
		f.DeanimateGif,
		f.ThumbVideo,
	}
}

func (f *messageFile) downloadAll(outDir string, slackToken string) []error {
	fileBaseDir := path.Join(outDir, f.Id)
	err := mkdir(fileBaseDir)
	if err != nil {
		return []error{err}
	}

	var errs []error

	for _, url := range f.downloadURLs() {
		err = f.downloadFile(fileBaseDir, url, slackToken)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (f *messageFile) downloadFile(outDir, url, slackToken string) error {
	if url == "" {
		return nil
	}

	filename := urlToFilename(url)
	destFile := filepath.Join(outDir, filename)
	if _, err := os.Stat(destFile); err == nil {
		// Just skip already downloaded file
		return nil
	}

	fmt.Printf("Downloading: %s/%s [%s]\n", f.Id, filename, f.PrettyType)
	return downloadFile(url, destFile, slackToken)
}

func downloadFile(url, destFile, slackToken string) error {
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
