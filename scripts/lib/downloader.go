package slacklog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// Downloader : slack.Client/LogStoreを用いてダウンロードURLを生成し、そこから
// ファイルを並行にダウンロードするための構造体。
// 並行処理時に発生したエラーはerrsに蓄えられ、然るべき後に戻り値として返される。
// errsにerrorを加える際はスレッドセーフにするためerrsMuでロックを取る。
type Downloader struct {
	token string

	targetCh chan DownloadTarget
	workerWg sync.WaitGroup

	errs   []error
	errsMu sync.Mutex
}

var downloadWorkerNum = 8

func NewDownloader(token string) *Downloader {
	d := &Downloader{
		token:    token,
		targetCh: make(chan DownloadTarget),
		errs:     []error{},
	}

	for i := 0; i < downloadWorkerNum; i++ {
		d.workerWg.Add(1)
		go func() {
			defer d.workerWg.Done()
			d.runWorker()
		}()
	}

	return d
}

func (d *Downloader) QueueDownloadRequest(url, outputPath string, withToken bool) {
	d.targetCh <- DownloadTarget{
		URL:        url,
		OutputPath: outputPath,
		WithToken:  withToken,
	}
}

func (d *Downloader) Wait() error {
	d.workerWg.Wait()
	if len(d.errs) != 0 {
		return d.errs[0]
	}
	return nil
}

func (d *Downloader) CloseQueue() {
	close(d.targetCh)
}

func (d *Downloader) runWorker() {
	for t := range d.targetCh {
		err := d.Download(t)
		if err != nil {
			d.errsMu.Lock()
			d.errs = append(d.errs, err)
			d.errsMu.Unlock()
		}
	}
}

// DownloadTarget : ダウンロードするURLとダウンロード先パスOutputPathのペア
// Downloaderにダウンロードする対象を指定するために使う
type DownloadTarget struct {
	URL        string
	OutputPath string
	WithToken  bool
}

func (d *Downloader) Download(t DownloadTarget) error {
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

	if t.WithToken {
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
