package slacklog_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	slacklog "github.com/vim-jp/slacklog/lib"
)

func TestDownloader(t *testing.T) {
	tmpPath := createTmpDir(t)
	defer t.Cleanup(func() {
		cleanupTmpDir(t, tmpPath)
	})

	ts := httptest.NewServer(http.FileServer(http.Dir("testdata/downloader")))
	defer ts.Close()

	fileInfos, err := ioutil.ReadDir("testdata/downloader")
	if err != nil {
		t.Fatal(err)
	}

	d := slacklog.NewDownloader("dummyToken")

	for _, fileInfo := range fileInfos {
		url := ts.URL + "/" + fileInfo.Name()
		path := filepath.Join(tmpPath, fileInfo.Name())
		d.QueueDownloadRequest(
			url,
			path,
			false,
		)
	}
	d.CloseQueue()
	err = d.Wait()
	if err != nil {
		t.Fatal(err)
	}

	err = dirDiff("testdata/downloader", tmpPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloader_usingToken(t *testing.T) {
	tmpPath := createTmpDir(t)
	defer t.Cleanup(func() {
		cleanupTmpDir(t, tmpPath)
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		res := map[string]interface{}{}
		res["token"] = r.Header.Get("Authorization")[7:]
		res["path"] = r.URL.Path

		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	testToken := "dummyToken"
	d := slacklog.NewDownloader(testToken)

	testFileName := "test.json"
	url := ts.URL + "/" + testFileName
	path := filepath.Join(tmpPath, testFileName)
	d.QueueDownloadRequest(url, path, true)
	d.CloseQueue()

	err := d.Wait()
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	err = json.NewDecoder(f).Decode(&got)
	if err != nil {
		t.Fatal(err)
	}
	gotPath, ok := got["path"]
	if !ok {
		t.Fatal("not found got[\"path\"]")
	}
	if gotPath != "/"+testFileName {
		t.Fatalf("want %s, but got %s", "/"+testFileName, gotPath)
	}
	gotToken, ok := got["token"]
	if !ok {
		t.Fatal("not found got[\"token\"]")
	}
	if gotToken != testToken {
		t.Fatalf("want %s, but got %s", testToken, gotToken)
	}
}
