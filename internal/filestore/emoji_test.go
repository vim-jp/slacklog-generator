package filestore

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/vim-jp/slacklog-generator/internal/store"
	"github.com/vim-jp/slacklog-generator/internal/testassert"
)

func jsonToEmoji(t *testing.T, s string) *store.Emoji {
	t.Helper()
	var e store.Emoji
	err := json.Unmarshal([]byte(s), &e)
	if err != nil {
		t.Fatalf("failed to parse as Emoji: %s", err)
	}
	return &e
}

func jsonToEmojiItem(t *testing.T, s string) *emojiItem {
	t.Helper()
	var e emojiItem
	err := json.Unmarshal([]byte(s), &e)
	if err != nil {
		t.Fatalf("failed to parse as Emoji: %s", err)
	}
	return &e
}

func TestEmojiStore_Get(t *testing.T) {
	es := &emojiStore{dir: "testdata/emoji_read"}

	for _, tc := range []struct {
		name string
		exp  string
	}{
		{"emoji01", `{"name":"emoji01","url":"http://example.org/emoji/0001.png"}`},
		{"emoji02", `{"name":"emoji02","url":"http://example.org/emoji/0002.png"}`},
		{"emoji03", `{"name":"emoji03","url":"http://example.org/emoji/0003.png"}`},
		{"emoji04", `{"name":"emoji04","url":"http://example.org/emoji/0004.png"}`},
		{"emoji05", `{"name":"emoji05","url":"http://example.org/emoji/0005.png"}`},
	} {
		act, err := es.Get(tc.name)
		if err != nil {
			t.Fatalf("failed to get(%s): %s", tc.name, err)
		}
		exp := jsonToEmoji(t, tc.exp)
		testassert.Equal(t, exp, act, "name:"+tc.name)
	}

	e, err := es.Get("emoji99")
	if err == nil {
		t.Fatalf("should fail to get unknown name: %+v", e)
	}
	if !strings.HasPrefix(err.Error(), "emoji not found, ") {
		t.Fatalf("unexpected error for getting unknown name: %s", err)
	}
}

func TestEmojiStore_UpsertCommit(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "emoji_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	es := &emojiStore{dir: dir}

	for i, s := range []string{
		`{"name":"emoji01","url":"http://example.org/emoji/0001.png"}`,
		`{"name":"emoji09","url":"http://example.org/emoji/0009.png"}`,
		`{"name":"emoji05","url":"http://example.org/emoji/0005.png"}`,
		`{"name":"emoji01","url":"http://example.org/emoji/0001a.png"}`,
		`{"name":"emoji99","url":"alias:emoji09"}`,
	} {
		_, err := es.Upsert(*jsonToEmoji(t, s))
		if err != nil {
			t.Fatalf("upsert failed #%d: %s", i, err)
		}
	}
	err = es.Commit()
	if err != nil {
		t.Fatalf("commit failed: %s", err)
	}

	cs2 := &emojiStore{dir: dir}
	err = cs2.assureLoad()
	if err != nil {
		t.Fatalf("assureLoad failed: %s", err)
	}
	testassert.Equal(t, []emojiItem{
		*jsonToEmojiItem(t, `{"name":"emoji01","url":"http://example.org/emoji/0001a.png","path":"0e/389a8ff9cd74563efd4209753ef0760e..png"}`),
		*jsonToEmojiItem(t, `{"name":"emoji05","url":"http://example.org/emoji/0005.png","path":"31/055b8e6280bf3ce59f538118895d7831..png"}`),
		*jsonToEmojiItem(t, `{"name":"emoji09","url":"http://example.org/emoji/0009.png","path":"b4/a39a87aca4c8b9de99230e380d0b9ab4..png"}`),
		*jsonToEmojiItem(t, `{"name":"emoji99","url":"alias:emoji09","alias_to":"emoji09"}`),
	}, cs2.emojis, "wrote emojis.json")
}

func TestEmojiStore_Upsert_NoName(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "emoji_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	cs := &emojiStore{dir: dir}
	_, err = cs.Upsert(*jsonToEmoji(t, `{"url":"http://example.org/emoji/0000.png"}`))
	if err == nil {
		t.Fatal("upsert without name should be failed")
	}
	if err.Error() != "empty name is forbidden" {
		t.Fatalf("unexpected failure: %s", err)
	}
}

func TestEmojiStore_Commit_Empty(t *testing.T) {
	// 空のCommitは emojis.json を作らない
	dir, err := ioutil.TempDir("testdata", "emoji_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	cs := &emojiStore{dir: dir}

	err = cs.Commit()
	if err != nil {
		t.Fatalf("unexpted failure: %s", err)
	}
	fi, err := os.Stat(cs.path())
	if err == nil {
		t.Fatalf("emojis.json created unexpectedly: %s", fi.Name())
	}
	if !os.IsNotExist(err) {
		t.Fatalf("unexpected failure: %s", err)
	}
}
