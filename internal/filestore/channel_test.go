package filestore

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/vim-jp/slacklog-generator/internal/store"
	"github.com/vim-jp/slacklog-generator/internal/testassert"
)

func jsonToChannel(t *testing.T, s string) *store.Channel {
	t.Helper()
	var c store.Channel
	err := json.Unmarshal([]byte(s), &c)
	if err != nil {
		t.Fatalf("failed to parse as Channel: %s", err)
	}
	return &c
}

func TestChannelStore_Get(t *testing.T) {
	cs := &channelStore{dir: "testdata/channel_read"}

	for _, tc := range []struct {
		id  string
		exp string
	}{
		{"CXXXX0001", `{"id":"CXXXX0001","name":"channel01"}`},
		{"CXXXX0002", `{"id":"CXXXX0002","name":"channel02"}`},
		{"CXXXX0003", `{"id":"CXXXX0003","name":"channel03"}`},
		{"CXXXX0004", `{"id":"CXXXX0004","name":"channel04"}`},
		{"CXXXX0005", `{"id":"CXXXX0005","name":"channel05"}`},
	} {
		act, err := cs.Get(tc.id)
		if err != nil {
			t.Fatalf("failed to get(%s): %s", tc.id, err)
		}
		exp := jsonToChannel(t, tc.exp)
		testassert.Equal(t, exp, act, "id:"+tc.id)
	}

	c, err := cs.Get("CXXXX9999")
	if err == nil {
		t.Fatalf("should fail to get unknown ID: %+v", c)
	}
	if !strings.HasPrefix(err.Error(), "channel not found, ") {
		t.Fatalf("unexpected error for getting unknown ID: %s", err)
	}
}

func TestChannelStore_Iterate(t *testing.T) {
	cs := &channelStore{dir: "testdata/channel_read"}

	var act []*store.Channel
	err := cs.Iterate(store.ChannelIterateFunc(func(c *store.Channel) bool {
		act = append(act, c)
		return true
	}))
	if err != nil {
		t.Fatalf("iteration failed: %s", err)
	}
	exp := []*store.Channel{
		jsonToChannel(t, `{"id":"CXXXX0001","name":"channel01"}`),
		jsonToChannel(t, `{"id":"CXXXX0002","name":"channel02"}`),
		jsonToChannel(t, `{"id":"CXXXX0003","name":"channel03"}`),
		jsonToChannel(t, `{"id":"CXXXX0004","name":"channel04"}`),
		jsonToChannel(t, `{"id":"CXXXX0005","name":"channel05"}`),
	}
	testassert.Equal(t, exp, act, "simple iteration")
}

func TestChannelStore_Iterate_Break(t *testing.T) {
	cs := &channelStore{dir: "testdata/channel_read"}

	i := 0
	err := cs.Iterate(store.ChannelIterateFunc(func(_ *store.Channel) bool {
		i++
		return i > 2
	}))
	if !errors.Is(err, store.ErrIterateAbort) {
		t.Fatalf("iterate should be failed with:%s got:%s", store.ErrIterateAbort, err)
	}
}

func TestChannelStore_Write(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "channel_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	cs := &channelStore{dir: dir}

	for i, s := range []string{
		`{"id":"W0001","name":"channel01"}`,
		`{"id":"W0009","name":"channel09"}`,
		`{"id":"W0005","name":"channel05"}`,
		`{"id":"W0001","name":"channel01a"}`,
	} {
		_, err := cs.Upsert(*jsonToChannel(t, s))
		if err != nil {
			t.Fatalf("upsert failed #%d: %s", i, err)
		}
	}
	err = cs.Commit()
	if err != nil {
		t.Fatalf("commit failed: %s", err)
	}

	cs2 := &channelStore{dir: dir}
	err = cs2.assureLoad()
	if err != nil {
		t.Fatalf("assureLoad failed: %s", err)
	}
	testassert.Equal(t, []store.Channel{
		*jsonToChannel(t, `{"id":"W0001","name":"channel01a"}`),
		*jsonToChannel(t, `{"id":"W0005","name":"channel05"}`),
		*jsonToChannel(t, `{"id":"W0009","name":"channel09"}`),
	}, cs2.channels, "wrote channels.json")
}

func TestChannelStore_Upsert_NoID(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "channel_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	cs := &channelStore{dir: dir}
	_, err = cs.Upsert(*jsonToChannel(t, `{"name":"foobar"}`))
	if err == nil {
		t.Fatal("upsert without ID should be failed")
	}
	if err.Error() != "empty ID is forbidden" {
		t.Fatalf("unexpected failure: %s", err)
	}
}

func TestChannelStore_Commit_Empty(t *testing.T) {
	// 空のCommitは channels.json を作らない
	dir, err := ioutil.TempDir("testdata", "channel_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	cs := &channelStore{dir: dir}

	err = cs.Commit()
	if err != nil {
		t.Fatalf("unexpted failure: %s", err)
	}
	fi, err := os.Stat(cs.path())
	if err == nil {
		t.Fatalf("channels.json created unexpectedly: %s", fi.Name())
	}
	if !os.IsNotExist(err) {
		t.Fatalf("unexpected failure: %s", err)
	}
}
