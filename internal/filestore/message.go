package filestore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

func (mm *messages) Merge(op *messages) {
	// lock and assure sorted both.
	mm.rw.Lock()
	mm.sort()
	op.rw.Lock()
	op.sort()
	op.rw.Unlock()
	op.rw.RLock()
	// merge sort.
	a, b := mm.msgs, op.msgs
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
	mm.msgs, mm.idx = r, idx
	op.rw.RUnlock()
	mm.rw.Unlock()
}

func (mm *messages) toDense() {
	if !mm.sparse || len(mm.msgs) == 0 {
		return
	}
	tail := len(mm.msgs) - 1
	for i := 0; i < tail; {
		if mm.msgs[i].Timestamp == "" {
			i++
			continue
		}
		for i < tail {
			if mm.msgs[tail].Timestamp == "" {
				break
			}
			tail--
		}
		if i == tail {
			tail = i - 1
			break
		}
		mm.msgs[i] = mm.msgs[tail]
		tail--
	}
	mm.msgs = mm.msgs[:tail+1]
	mm.sparse = false
}

func (mm *messages) sort() {
	mm.toDense()
	if mm.sorted || len(mm.msgs) <= 1 {
		return
	}
	sort.Slice(mm.msgs, func(i, j int) bool {
		return mm.msgs[i].Before(mm.msgs[j])
	})
	mm.sorted = true
}

func (mm *messages) Upsert(m store.Message) (bool, error) {
	if m.ClientMsgID == "" {
		return false, errors.New("empty ID is forbidden")
	}
	m.Tidy()
	mm.rw.Lock()
	defer mm.rw.Unlock()
	mm.sorted = false
	x, ok := mm.idx[m.ClientMsgID]
	if ok {
		mm.msgs[x] = m
		return true, nil
	}
	if mm.idx == nil {
		mm.idx = make(map[string]int)
	}
	mm.idx[m.ClientMsgID] = len(mm.msgs)
	mm.msgs = append(mm.msgs, m)
	return false, nil
}

func (mm *messages) Delete(m store.Message) bool {
	mm.rw.Lock()
	defer mm.rw.Unlock()
	x, ok := mm.idx[m.ClientMsgID]
	if !ok {
		return false
	}
	mm.msgs[x] = store.Message{}
	delete(mm.idx, m.ClientMsgID)
	mm.sparse = true
	return true
}

type messageStore struct {
	dir string
}

// Begin starts a transaction for message.
func (ms *messageStore) Begin(channelID string) (store.MessageTx, error) {
	return &messageTx{
		cid: channelID,
		dir: filepath.Join(ms.dir, channelID),
	}, nil
}

type messageTx struct {
	cid string
	dir string

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
	mm, ok := mtx.upserts[tk]
	if !ok {
		mm = &messages{
			idx: map[string]int{},
		}
		if mtx.upserts == nil {
			mtx.upserts = make(map[store.TimeKey]*messages)
		}
		mtx.upserts[tk] = mm
	}
	mtx.mu.Unlock()

	return mm.Upsert(m)
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
	mtx.mu.Lock()
	defer mtx.mu.Unlock()
	if len(mtx.upserts) == 0 {
		return nil
	}
	for tk, mm := range mtx.upserts {
		orig, err := mtx.readMsgsJSONL(tk)
		if err != nil {
			return err
		}
		mm.rw.RLock()
		orig.Merge(mm)
		mtx.writeMsgsJSONL(tk, orig)
		mm.rw.RUnlock()
	}
	mtx.upserts = nil
	return nil
}

// msgsFileName generate JSONL filename for `store.TimeKey.Begin`.
func (mtx *messageTx) msgsFileName(tk store.TimeKey) string {
	return filepath.Join(mtx.dir, tk.BeginDateString()+".jsonl")
}

// readMsgsJSONL reads messages from a JSONL file.
// JSONL is JSON Lines where defined at http://jsonlines.org/
func (mtx *messageTx) readMsgsJSONL(tk store.TimeKey) (*messages, error) {
	f, err := os.Open(mtx.msgsFileName(tk))
	if err != nil {
		if os.IsNotExist(err) {
			return new(messages), nil
		}
		return nil, err
	}
	defer f.Close()
	mm := new(messages)
	d := json.NewDecoder(f)
	for d.More() {
		var m store.Message
		err := d.Decode(&m)
		if err != nil {
			return nil, err
		}
		mm.Upsert(m)
	}
	return mm, nil
}

// writeMsgsJSONL writes messags as a JOSNL file.
// JSONL is JSON Lines where defined at http://jsonlines.org/
func (mtx *messageTx) writeMsgsJSONL(tk store.TimeKey, mm *messages) error {
	err := os.MkdirAll(mtx.dir, 0777)
	if err != nil {
		return err
	}
	f, err := os.Create(mtx.msgsFileName(tk))
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	for _, m := range mm.msgs {
		err := e.Encode(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// Rollback discards changes in a transaction.
func (mtx *messageTx) Rollback() error {
	mtx.mu.Lock()
	defer mtx.mu.Unlock()
	mtx.upserts = nil
	return nil
}
