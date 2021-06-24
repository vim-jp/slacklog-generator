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

type channelStore struct {
	dir string

	rw       sync.RWMutex
	channels []store.Channel
	idxID    map[string]int
}

// Get gets a channel by ID.
func (cs *channelStore) Get(id string) (*store.Channel, error) {
	err := cs.assureLoad()
	if err != nil {
		return nil, err
	}
	cs.rw.RLock()
	defer cs.rw.RUnlock()

	x, ok := cs.idxID[id]
	if !ok {
		return nil, fmt.Errorf("channel not found, uknown id: id=%s", id)
	}
	if x < 0 || x >= len(cs.channels) {
		return nil, fmt.Errorf("channel index collapsed, ask developers: id=%s", id)
	}
	c := cs.channels[x]
	return &c, nil
}

// Iterate enumerates all channels by callback.
// 呼び出し時点でチャンネル一覧のコピーが本イテレート専用に作成される。
// コールバックが false を返すと store.ErrIterateAbort が返る
func (cs *channelStore) Iterate(iter store.ChannelIterator) error {
	err := cs.assureLoad()
	if err != nil {
		return err
	}
	cs.rw.RLock()
	channels := make([]store.Channel, len(cs.channels))
	copy(channels, cs.channels)
	// FIXME: cs.idxID に入ってないのは省くべきでは?
	cs.rw.RUnlock()
	for i := range channels {
		cont := iter.Iterate(&channels[i])
		if !cont {
			return store.ErrIterateAbort
		}
	}
	return nil
}

// Upsert updates or inserts a channel in store.
// This returns true as 1st parameter, when a channel inserted.
func (cs *channelStore) Upsert(c store.Channel) (bool, error) {
	if c.ID == "" {
		return false, errors.New("empty ID is forbidden")
	}

	err := cs.assureLoad()
	if err != nil {
		return false, err
	}
	cs.rw.Lock()
	defer cs.rw.Unlock()

	c.Tidy()

	x, ok := cs.idxID[c.ID]
	if ok {
		cs.channels[x] = c
		return true, nil
	}
	cs.idxID[c.ID] = len(cs.channels)
	cs.channels = append(cs.channels, c)
	return false, nil
}

// Commit saves channels to file:channels.json.
func (cs *channelStore) Commit() error {
	cs.rw.Lock()
	defer cs.rw.Unlock()
	if cs.channels == nil {
		log.Printf("[DEBUG] no channels to commit. not load yet?")
		return nil
	}

	ids := make([]string, 0, len(cs.idxID))
	for id := range cs.idxID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	ca := make([]store.Channel, len(ids))
	for i, id := range ids {
		ca[i] = cs.channels[cs.idxID[id]]
	}
	err := jsonWriteFile(cs.path(), ca)
	if err != nil {
		return err
	}

	cs.replaceChannels(ca)
	return nil
}

// path returns path for channels.json
func (cs *channelStore) path() string {
	return filepath.Join(cs.dir, "channels.json")
}

// assureLoad assure channels.json is loaded.
func (cs *channelStore) assureLoad() error {
	cs.rw.Lock()
	defer cs.rw.Unlock()
	if cs.channels != nil {
		return nil
	}
	var channels []store.Channel
	err := jsonReadFile(cs.path(), true, &channels)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	cs.replaceChannels(channels)
	return nil
}

func (cs *channelStore) replaceChannels(channels []store.Channel) {
	if len(channels) == 0 {
		cs.channels = []store.Channel{}
		cs.idxID = map[string]int{}
		return
	}
	idxID := make(map[string]int, len(channels))
	for i, c := range channels {
		idxID[c.ID] = i
	}
	cs.channels = channels
	cs.idxID = idxID
}
