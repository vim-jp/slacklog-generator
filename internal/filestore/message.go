package filestore

import (
	"github.com/vim-jp/slacklog-generator/internal/store"
)

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
}

var _ store.MessageTx = (*messageTx)(nil)

func (mtx *messageTx) Upsert(m store.Message) (bool, error) {
	// TODO:
	return false, nil
}

func (mtx *messageTx) Iterate(key store.TimeKey, iter store.MessageIterator) error {
	// TODO:
	return nil
}

func (mtx *messageTx) Count(key store.TimeKey) (int, error) {
	// TODO:
	return 0, nil
}

func (mtx *messageTx) Commit() error {
	// TODO:
	return nil
}

func (mtx *messageTx) Rollback() error {
	// TODO:
	return nil
}
