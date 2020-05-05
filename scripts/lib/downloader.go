package slacklog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Downloader : slack.Client/LogStoreを用いてダウンロードURLを生成し、そこから
// ファイルを並行にダウンロードするための構造体。
// 並行処理時に発生したエラーはerrsに蓄えられ、然るべき後に戻り値として返される。
// errsにerrorを加える際はスレッドセーフにするためerrsMuでロックを取る。
type Downloader struct {
	token string

	errs   []error
	errsMu sync.Mutex
}

func NewDownloader(token string) *Downloader {
	return &Downloader{
		token: token,
		errs:  []error{},
	}
}

// DownloadTarget : ダウンロードするURLとダウンロード先パスOutputPathのペア
// Downloaderにダウンロードする対象を指定するために使う
type DownloadTarget struct {
	URL        string
	OutputPath string
}

func GenerateMessageFileTargets(s *LogStore, outputDir string) <-chan DownloadTarget {
	targetCh := make(chan DownloadTarget)

	go func() {
		defer close(targetCh)
		channels := s.GetChannels()
		for _, channel := range channels {
			msgs, err := s.GetAllMessages(channel.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get messages on %s channel: %s", channel.Name, err)
				return
			}

			for _, msg := range msgs {
				for _, f := range msg.Files {
					targetDir := filepath.Join(outputDir, f.ID)
					err := os.MkdirAll(targetDir, 0777)
					if err != nil {
						fmt.Fprintf(os.Stderr, "failed to create %s directory: %s", targetDir, err)
						return
					}

					for url, suffix := range f.DownloadURLsAndSuffixes() {
						if url == "" {
							continue
						}
						targetCh <- DownloadTarget{
							URL:        url,
							OutputPath: filepath.Join(targetDir, f.DownloadFilename(url, suffix)),
						}
					}
				}
			}
		}
	}()

	return targetCh
}

var downloadWorkerNum = 8

// DownloadAll : targetChに届いたDownloadTargetを並行にダウンロードする。
// withTokenはHTTPリクエストにSlack API tokenを用いるかを指定する。
// 同時ダウンロード数をlimitChによりdownloadWorkerNumだけに制限している。
func (d *Downloader) DownloadAll(targetCh <-chan DownloadTarget, withToken bool) error {
	limitCh := make(chan struct{}, downloadWorkerNum)
	var wg sync.WaitGroup
	for target := range targetCh {
		wg.Add(1)
		go func(t DownloadTarget) {
			limitCh <- struct{}{}
			defer func() {
				wg.Done()
				<-limitCh
			}()

			err := d.Download(t, withToken)
			if err != nil {
				d.errsMu.Lock()
				d.errs = append(d.errs, err)
				d.errsMu.Unlock()
			}
		}(target)
	}

	wg.Wait()

	if len(d.errs) > 0 {
		return d.errs[0]
	}
	return nil
}

func (d *Downloader) Download(t DownloadTarget, withToken bool) error {
	_, err := os.Stat(t.OutputPath)
	if err == nil {
		// Just skip already downloaded file
		fmt.Printf("already exist: %s\n", t.OutputPath)
		return nil
	}
	// `err != nil` has two cases at here. first is "not exist" as expected.
	// and second is I/O error as unexpected.
	if !os.IsNotExist(err) {
		return err
	}
	fmt.Printf("Downloading: %s\n", t.OutputPath)
	httpClient := &http.Client{}
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest("GET", t.URL, nil)
	if err != nil {
		return err
	}

	if withToken {
		req.Header.Add("Authorization", "Bearer "+d.token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("[%s]: %s", resp.Status, t.URL)
	}

	w, err := os.Create(t.OutputPath)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
