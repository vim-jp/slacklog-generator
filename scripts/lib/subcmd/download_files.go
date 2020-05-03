/*
リファクタリング中
処理をslacklog packageに移動していく。
一旦、必要な処理はすべてslacklog packageから一時的にエクスポートするか、このファ
イル内で定義している。
*/

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

func DownloadFiles(args []string) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return fmt.Errorf("$SLACK_TOKEN required")
	}

	if len(args) < 2 {
		fmt.Println("Usage: go run scripts/main.go download_files {log-dir} {files-dir}")
		return nil
	}

	logDir := filepath.Clean(args[0])
	filesDir := filepath.Clean(args[1])

	s, err := slacklog.NewLogStore(logDir, &slacklog.Config{Channels: []string{"*"}})
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filesDir, 0777); err != nil {
		return fmt.Errorf("could not create %s directory: %w", filesDir, err)
	}

	// start download workers.
	ch := make(chan *slacklog.MessageFile, downloadWorkerNum)
	wg := new(sync.WaitGroup)
	failed := false
	for i := 0; i < cap(ch); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range ch {
				errs := downloadAll(m, filesDir, slackToken)
				if len(errs) > 0 {
					failed = true
					for i := range errs {
						fmt.Fprintf(os.Stderr, "[error] Download failed: %s\n", errs[i])
					}
				}
			}
		}()
	}

	// request to download files in messages.
	channels := s.GetChannels()
	for _, channel := range channels {
		messages, err := ReadAllMessages(filepath.Join(logDir, channel.ID))
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
	err := os.MkdirAll(fileBaseDir, 0777)
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
	_, err := os.Stat(destFile)
	if err == nil {
		// Just skip already downloaded file
		return nil
	}
	// `err != nil` has two cases at here. first is "not exist" as expected.
	// and second is I/O error as unexpected.
	if !os.IsNotExist(err) {
		return err
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
