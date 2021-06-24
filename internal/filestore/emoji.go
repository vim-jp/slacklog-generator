package filestore

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/vim-jp/slacklog-generator/internal/store"
)

type emojiItem struct {
	store.Emoji

	AliasTo string `json:"alias_to,omitempty"`
	Path    string `json:"path,omitempty"`
}

const aliasPrefix = "alias:"

func newEmojiItem(src store.Emoji) emojiItem {
	dst := emojiItem{Emoji: src}
	if strings.HasPrefix(src.URL, aliasPrefix) {
		dst.AliasTo = src.URL[len(aliasPrefix):]
	} else {
		d := fmt.Sprintf("%x", md5.Sum([]byte(src.URL)))
		ext := path.Ext(src.URL)
		dst.Path = path.Join(d[len(d)-2:], d+"."+ext)
	}
	return dst
}

type emojiStore struct {
	dir string

	rw      sync.RWMutex
	emojis  []emojiItem
	idxName map[string]int
}

// Get gets an emoji by name.
func (es *emojiStore) Get(name string) (*store.Emoji, error) {
	err := es.assureLoad()
	if err != nil {
		return nil, err
	}
	es.rw.RLock()
	defer es.rw.RUnlock()

	x, ok := es.idxName[name]
	if !ok {
		return nil, fmt.Errorf("emoji not found, uknown name: name=%s", name)
	}
	if x < 0 || x >= len(es.emojis) {
		return nil, fmt.Errorf("emoji index collapsed, ask developers: name=%s", name)
	}
	e := es.emojis[x]
	return &e.Emoji, nil
}

// Upsert updates or inserts a emoji in store.
// This returns true as 1st parameter, when a emoji inserted.
func (es *emojiStore) Upsert(e store.Emoji) (bool, error) {
	if e.Name == "" {
		return false, errors.New("empty name is forbidden")
	}

	err := es.assureLoad()
	if err != nil {
		return false, err
	}
	es.rw.Lock()
	defer es.rw.Unlock()

	e.Tidy()

	x, ok := es.idxName[e.Name]
	if ok {
		es.emojis[x] = newEmojiItem(e)
		return false, nil
	}
	es.idxName[e.Name] = len(es.emojis)
	es.emojis = append(es.emojis, newEmojiItem(e))
	return true, nil
}

// Commit saves emojis to file:emojis.json.
func (es *emojiStore) Commit() error {
	es.rw.Lock()
	defer es.rw.Unlock()
	if es.emojis == nil {
		log.Printf("[DEBUG] no emojis to commit. not load yet?")
		return nil
	}

	names := make([]string, 0, len(es.idxName))
	for name := range es.idxName {
		names = append(names, name)
	}
	sort.Strings(names)
	ca := make([]emojiItem, len(names))
	for i, name := range names {
		ca[i] = es.emojis[es.idxName[name]]
	}
	err := jsonWriteFile(es.path(), ca)
	if err != nil {
		return err
	}

	es.replaceEmojis(ca)
	return nil
}

// path returns path for emojis.json
func (es *emojiStore) path() string {
	return filepath.Join(es.dir, "emojis.json")
}

// assureLoad assure emojis.json is loaded.
func (es *emojiStore) assureLoad() error {
	es.rw.Lock()
	defer es.rw.Unlock()
	if es.emojis != nil {
		return nil
	}
	var emojis []emojiItem
	err := jsonReadFile(es.path(), true, &emojis)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	es.replaceEmojis(emojis)
	return nil
}

func (es *emojiStore) replaceEmojis(emojis []emojiItem) {
	if len(emojis) == 0 {
		es.emojis = []emojiItem{}
		es.idxName = map[string]int{}
		return
	}
	idxName := make(map[string]int, len(emojis))
	for i, e := range emojis {
		idxName[e.Name] = i
	}
	es.emojis = emojis
	es.idxName = idxName
}
