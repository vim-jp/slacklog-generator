package filestore

import (
	"errors"
	"sort"
	"sync"

	"github.com/vim-jp/slacklog-generator/internal/store"
)

type messages struct {
	rw   sync.RWMutex
	msgs []store.Message
	idx  map[string]int

	sorted bool
	sparse bool
}

func (ms *messages) Merge(op *messages) {
	// lock and assure sorted both.
	ms.rw.Lock()
	ms.sort()
	op.rw.Lock()
	op.sort()
	op.rw.Unlock()
	op.rw.RLock()
	// merge sort.
	a, b := ms.msgs, op.msgs
	r := make([]store.Message, len(a)+len(b), 0)
	for len(a) > 0 && len(b) > 0 {
		if a[0].Before(b[0]) {
			r = append(r, a[0])
			a = a[1:]
		} else {
			r = append(r, b[0])
			b = b[1:]
		}
	}
	if len(a) > 0 {
		r = append(r, a...)
	}
	if len(b) > 0 {
		r = append(r, b...)
	}
	// build new index.
	idx := make(map[string]int)
	for i, m := range r {
		idx[m.ClientMsgID] = i
	}
	// update this messages and unlock.
	ms.msgs, ms.idx = r, idx
	op.rw.RUnlock()
	ms.rw.Unlock()
}

func (ms *messages) toDense() {
	if !ms.sparse || len(ms.msgs) == 0 {
		return
	}
	tail := len(ms.msgs) - 1
	for i := 0; i < tail; {
		if ms.msgs[i].Timestamp == "" {
			i++
			continue
		}
		for i < tail {
			if ms.msgs[tail].Timestamp == "" {
				break
			}
			tail--
		}
		if i == tail {
			tail = i - 1
			break
		}
		ms.msgs[i] = ms.msgs[tail]
		tail--
	}
	ms.msgs = ms.msgs[:tail+1]
	ms.sparse = false
}

func (ms *messages) sort() {
	ms.toDense()
	if ms.sorted || len(ms.msgs) <= 1 {
		return
	}
	sort.Slice(ms.msgs, func(i, j int) bool {
		return ms.msgs[i].Before(ms.msgs[j])
	})
	ms.sorted = true
}

func (ms *messages) Upsert(m store.Message) (bool, error) {
	if m.ClientMsgID == "" {
		return false, errors.New("empty ID is forbidden")
	}
	m.Tidy()
	ms.rw.Lock()
	defer ms.rw.Unlock()
	ms.sorted = false
	x, ok := ms.idx[m.ClientMsgID]
	if ok {
		ms.msgs[x] = m
		return true, nil
	}
	ms.idx[m.ClientMsgID] = len(ms.msgs)
	ms.msgs = append(ms.msgs, m)
	return false, nil
}

func (ms *messages) Delete(m store.Message) bool {
	ms.rw.Lock()
	defer ms.rw.Unlock()
	x, ok := ms.idx[m.ClientMsgID]
	if !ok {
		return false
	}
	ms.msgs[x] = store.Message{}
	delete(ms.idx, m.ClientMsgID)
	ms.sparse = true
	return true
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
	ms, ok := mtx.upserts[tk]
	if !ok {
		ms = &messages{
			idx: map[string]int{},
		}
		mtx.upserts[tk] = ms
	}
	mtx.mu.Unlock()

	return ms.Upsert(m)
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
