package slacklog

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// Downloader : ダウンロード処理をするワーカを管理するための構造体。
// TODO: 今のところ、ダウンロード処理中にエラーが発生してもキューに積まれたタス
// クが全て完了するまでは分からない(Wait()の返り値として見るまでは)。
type Downloader struct {
	token string

	httpClient *http.Client
	targetCh   chan downloadTarget
	workerWg   sync.WaitGroup

	// firstError stores the first error raised in workers.
	firstError error

	// errsMu: firstError に触る際は必ずコレでロックを取る
	errsMu sync.Mutex
}

var downloadWorkerNum = 8

func NewDownloader(token string) *Downloader {
	// http.DefaultTransportの値からMaxConnsPerHostのみ修正
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2: true,
		MaxIdleConns:      100,
		// ワーカ起動のロジックにバグがあったとしてもこのhttp.Transportを利用してい
		// る限りは多量のリクエストが飛ばないように念の為downloadWorkerNumでコネク
		// ション数を制限しておく
		MaxConnsPerHost:       downloadWorkerNum,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	cli := &http.Client{Transport: t}
	// 無効なSlack API tokenを食わせても、リダイレクトされ、200が返ってきてエラー
	// かどうか判別できないためリダイレクトをしないように制御しておく
	cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	d := &Downloader{
		token:      token,
		httpClient: cli,
		targetCh:   make(chan downloadTarget),
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

// QueueDownloadRequest : ダウンロード処理をqueueに積む
func (d *Downloader) QueueDownloadRequest(url, outputPath string, withToken bool) {
	d.targetCh <- downloadTarget{
		url:        url,
		outputPath: outputPath,
		withToken:  withToken,
	}
}

// Wait : ワーカが全て実行終了するまで待つ。
// ダウンロード処理中にエラーが発生していた場合は最初に発生した1つを返す。
// 他のエラーはログに出力している。
func (d *Downloader) Wait() error {
	d.workerWg.Wait()
	d.errsMu.Lock()
	defer d.errsMu.Unlock()
	return d.firstError
}

// CloseQueue : ダウンロードキューへの追加が完了したことをDownloaderに通知する
// ために実行する。
// TODO: 2回実行するとpanicしてしまうのを修正する。Downloaderに状態でも持たせる
// とよいだろうか。
func (d *Downloader) CloseQueue() {
	close(d.targetCh)
}

func (d *Downloader) runWorker() {
	for t := range d.targetCh {
		err := d.download(t)
		if err != nil {
			d.errsMu.Lock()
			if d.firstError == nil {
				d.firstError = err
			}
			d.errsMu.Unlock()
			fmt.Printf("download failed for url=%s: %s\n", t.url, err)
		}
	}
}

// downloadTarget : Downloaderにダウンロードする対象を指定するために使う。
type downloadTarget struct {
	url        string
	outputPath string
	// ダウンロード時にSlack API tokenを利用するかどうかを指定する
	withToken bool
}

func (d *Downloader) download(t downloadTarget) error {
	_, err := os.Stat(t.outputPath)
	if err == nil {
		// Just skip already downloaded file
		fmt.Printf("already exist: %s\n", t.outputPath)
		return nil
	}
	// `err != nil` has two cases at here. first is "not exist" as expected.
	// and second is I/O error as unexpected.
	if !os.IsNotExist(err) {
		return err
	}

	fmt.Printf("Downloading: %s\n", t.outputPath)

	req, err := http.NewRequest("GET", t.url, nil)
	if err != nil {
		return err
	}

	if t.withToken {
		req.Header.Add("Authorization", "Bearer "+d.token)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("[%s]: %s", resp.Status, t.url)
	}

	w, err := os.Create(t.outputPath)
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
