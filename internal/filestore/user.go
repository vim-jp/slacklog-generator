package filestore

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/vim-jp/slacklog-generator/internal/store"
)

type userStore struct {
	dir string

	rw    sync.RWMutex
	users []store.User
	idxID map[string]int
}

// Get gets a user by ID.
func (us *userStore) Get(id string) (*store.User, error) {
	err := us.assureLoad()
	if err != nil {
		return nil, err
	}
	us.rw.RLock()
	defer us.rw.RUnlock()

	x, ok := us.idxID[id]
	if !ok {
		return nil, fmt.Errorf("user not found, uknown id: id=%s", id)
	}
	if x < 0 || x >= len(us.users) {
		return nil, fmt.Errorf("user index collapsed, ask developers: id=%s", id)
	}
	u := us.users[x]
	return &u, nil
}

// Upsert updates or inserts a user in store.
// This returns true as 1st parameter, when a channel inserted.
func (us *userStore) Upsert(u store.User) (bool, error) {
	if u.ID == "" {
		return false, errors.New("empty ID is forbidden")
	}

	err := us.assureLoad()
	if err != nil {
		return false, err
	}
	us.rw.Lock()
	defer us.rw.Unlock()

	u.Tidy()

	x, ok := us.idxID[u.ID]
	if ok {
		us.users[x] = u
		return true, nil
	}
	us.idxID[u.ID] = len(us.users)
	us.users = append(us.users, u)
	return false, nil
}

// Commit saves users to file:users.json.
func (us *userStore) Commit() error {
	us.rw.Lock()
	defer us.rw.Unlock()
	if us.users == nil {
		log.Printf("[DEBUG] no users to commit. not load yet?")
		return nil
	}

	ids := make([]string, 0, len(us.idxID))
	for id := range us.idxID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	ua := make([]store.User, len(ids))
	for i, id := range ids {
		ua[i] = us.users[us.idxID[id]]
	}
	err := jsonWriteFile(us.path(), ua)
	if err != nil {
		return err
	}

	us.replaceUsers(ua)
	return nil
}

// path returns path for users.json
func (us *userStore) path() string {
	return filepath.Join(us.dir, "users.json")
}

// assureLoad assure users.json is loaded.
func (us *userStore) assureLoad() error {
	us.rw.Lock()
	defer us.rw.Unlock()
	if us.users != nil {
		return nil
	}
	var users []store.User
	err := jsonReadFile(us.path(), true, &users)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	us.replaceUsers(users)
	return nil
}

func (us *userStore) replaceUsers(users []store.User) {
	if len(users) == 0 {
		us.users = []store.User{}
		us.idxID = map[string]int{}
		return
	}
	idxID := make(map[string]int, len(users))
	for i, u := range users {
		idxID[u.ID] = i
	}
	us.users = users
	us.idxID = idxID
}
