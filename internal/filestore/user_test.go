package filestore

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/slack-go/slack"
	"github.com/vim-jp/slacklog-generator/internal/store"
	"github.com/vim-jp/slacklog-generator/internal/testassert"
)

func jsonToUser(t *testing.T, s string) *store.User {
	t.Helper()
	var u store.User
	err := json.Unmarshal([]byte(s), &u)
	if err != nil {
		t.Fatalf("failed to parse as User: %s", err)
	}
	return &u
}

var userCmpOpts = []cmp.Option{
	cmpopts.IgnoreUnexported(slack.UserProfileCustomFields{}),
}

func TestUserStore_Get(t *testing.T) {
	us := &userStore{dir: "testdata/user_read"}

	for _, tc := range []struct {
		id  string
		exp string
	}{
		{"UXXXX0001", `{"id":"UXXXX0001","name":"user01"}`},
		{"UXXXX0002", `{"id":"UXXXX0002","name":"user02"}`},
		{"UXXXX0003", `{"id":"UXXXX0003","name":"user03"}`},
		{"UXXXX0004", `{"id":"UXXXX0004","name":"user04"}`},
		{"UXXXX0005", `{"id":"UXXXX0005","name":"user05"}`},
	} {
		act, err := us.Get(tc.id)
		if err != nil {
			t.Fatalf("failed to get(%s): %s", tc.id, err)
		}
		exp := jsonToUser(t, tc.exp)
		testassert.Equal(t, exp, act, "id:"+tc.id, userCmpOpts...)
	}

	u, err := us.Get("CXXXX9999")
	if err == nil {
		t.Fatalf("should fail to get unknown ID: %+v", u)
	}
	if !strings.HasPrefix(err.Error(), "user not found, ") {
		t.Fatalf("unexpected error for getting unknown ID: %s", err)
	}
}

func TestUserStore_Write(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "user_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	us := &userStore{dir: dir}

	for i, s := range []string{
		`{"id":"U0001","name":"user01"}`,
		`{"id":"U0009","name":"user09"}`,
		`{"id":"U0005","name":"user05"}`,
		`{"id":"U0001","name":"user01a"}`,
	} {
		_, err := us.Upsert(*jsonToUser(t, s))
		if err != nil {
			t.Fatalf("upsert failed #%d: %s", i, err)
		}
	}
	err = us.Commit()
	if err != nil {
		t.Fatalf("commit failed: %s", err)
	}

	us2 := &userStore{dir: dir}
	err = us2.assureLoad()
	if err != nil {
		t.Fatalf("assureLoad failed: %s", err)
	}
	testassert.Equal(t, []store.User{
		*jsonToUser(t, `{"id":"U0001","name":"user01a"}`),
		*jsonToUser(t, `{"id":"U0005","name":"user05"}`),
		*jsonToUser(t, `{"id":"U0009","name":"user09"}`),
	}, us2.users, "wrote users.json", userCmpOpts...)
}

func TestUserStore_Upsert_NoID(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "user_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	us := &userStore{dir: dir}
	_, err = us.Upsert(*jsonToUser(t, `{"name":"foobar"}`))
	if err == nil {
		t.Fatal("upsert without ID should be failed")
	}
	if err.Error() != "empty ID is forbidden" {
		t.Fatalf("unexpected failure: %s", err)
	}
}

func TestUserStore_Commit_Empty(t *testing.T) {
	// 空のCommitは users.json を作らない
	dir, err := ioutil.TempDir("testdata", "user_write*")
	if err != nil {
		t.Fatalf("failed to TempDir: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	us := &userStore{dir: dir}

	err = us.Commit()
	if err != nil {
		t.Fatalf("unexpted failure: %s", err)
	}
	fi, err := os.Stat(us.path())
	if err == nil {
		t.Fatalf("users.json created unexpectedly: %s", fi.Name())
	}
	if !os.IsNotExist(err) {
		t.Fatalf("unexpected failure: %s", err)
	}
}
