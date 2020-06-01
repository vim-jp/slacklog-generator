package filestore

import (
	"errors"
	"sync"

	"github.com/vim-jp/slacklog-generator/internal/store"
)

type messages struct {
	rw   sync.RWMutex
	msgs []store.Message
	idx  map[string]int
}

func (msgs *messages) Upsert(m store.Message) (bool, error) {
	if m.ClientMsgID == "" {
		return false, errors.New("empty ID is forbidden")
	}

	m.Tidy()

	msgs.rw.Lock()
	defer msgs.rw.Unlock()
	x, ok := msgs.idx[m.ClientMsgID]
	if ok {
		msgs.msgs[x] = m
		return true, nil
	}
	msgs.idx[m.ClientMsgID] = len(msgs.msgs)
	msgs.msgs = append(msgs.msgs, m)
	return false, nil
}

type messageStore struct {
	dir string
}

// Begin starts a transaction for message.
func (ms *messageStore) Begin(channelID string) (store.MessageTx, error) {
	// TODO:
	return &messageTx{cid: channelID}, nil
}

type messageTx struct {
	cid string

	mu      sync.Mutex
	upserts map[store.TimeKey]*messages
}

var _ store.MessageTx = (*messageTx)(nil)

// Upsert updates or inserts a message in store.
func (mtx *messageTx) Upsert(m store.Message) (bool, error) {
	ts, err := m.TimestampTime()
	if err != nil {
		return false, err
	}
	tk := store.TimeKeyDate(ts)

	mtx.mu.Lock()
	msgs, ok := mtx.upserts[tk]
	if !ok {
		msgs = &messages{
			idx: map[string]int{},
		}
		mtx.upserts[tk] = msgs
	}
	mtx.mu.Unlock()

	return msgs.Upsert(m)
}

// Iterate iterates messages in a TimeKey.
func (mtx *messageTx) Iterate(key store.TimeKey, iter store.MessageIterator) error {
	// TODO:
	return nil
}

// Count counts messages in a TimeKey.
func (mtx *messageTx) Count(key store.TimeKey) (int, error) {
	// TODO:
	return 0, nil
}

// Commit persists changes in a transaction.
func (mtx *messageTx) Commit() error {
	// TODO:
	return nil
}

// Rollback discards changes in a transaction.
func (mtx *messageTx) Rollback() error {
	// TODO:
	return nil
}
