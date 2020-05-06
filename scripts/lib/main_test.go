package slacklog_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"testing"
)

var usedTmpNums sync.Map

type tmpDir string

func (td tmpDir) cleanup(t *testing.T) {
	path := string(td)
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
	ind, err := strconv.Atoi(filepath.Base(path)[3:])
	if err != nil {
		t.Fatal(err)
	}
	usedTmpNums.Delete(ind)
}

func createTmpDir(t *testing.T) tmpDir {
	t.Helper()
	path := "testdata/tmp"
	var ind int
	for {
		ind = rand.Intn(100)
		_, ok := usedTmpNums.Load(ind)
		if !ok {
			break
		}
	}
	path += strconv.Itoa(ind)
	usedTmpNums.Store(ind, struct{}{})

	err := os.MkdirAll(path, 0777)
	if err != nil {
		t.Fatal(err)
	}
	return tmpDir(path)
}

func dirDiff(a, b string) error {
	aInfos, err := ioutil.ReadDir(a)
	if err != nil {
		return err
	}
	bInfos, err := ioutil.ReadDir(b)
	if err != nil {
		return err
	}

	if len(aInfos) != len(bInfos) {
		return fmt.Errorf(
			"the number of files in the directory is different: (%s: %d) (%s: %d)",
			a, len(aInfos),
			b, len(bInfos),
		)
	}

	sort.Slice(aInfos, func(i, j int) bool {
		return aInfos[i].Name() >= aInfos[i].Name()
	})
	sort.Slice(bInfos, func(i, j int) bool {
		return bInfos[i].Name() >= bInfos[i].Name()
	})

	for i := range aInfos {
		if aInfos[i].Name() != bInfos[i].Name() {
			return fmt.Errorf(
				"the file name is different: %s != %s",
				filepath.Join(a, aInfos[i].Name()),
				filepath.Join(b, bInfos[i].Name()),
			)
		}
		if aInfos[i].Size() != bInfos[i].Size() {
			return fmt.Errorf(
				"the file size is different: (%s: %d) (%s: %d)",
				filepath.Join(a, aInfos[i].Name()), aInfos[i].Size(),
				filepath.Join(b, bInfos[i].Name()), bInfos[i].Size(),
			)
		}
	}
	return nil
}
