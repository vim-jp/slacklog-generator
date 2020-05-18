package filestore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vim-jp/slacklog-generator/internal/store"
)

// FileStore is an implementation of Store on file system.
type FileStore struct {
	dir string

	cs *channelStore
	us *userStore
}

// New creates a FileStore.
func New(dir string) (*FileStore, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		if !fi.IsDir() {
			return nil, fmt.Errorf("path is used with not directory: %s", dir)
		}
	}
	return &FileStore{
		dir: dir,
		cs: &channelStore{
			dir: filepath.Join(dir, "slacklog_data"),
		},
		us: &userStore{
			dir: filepath.Join(dir, "slacklog_data"),
		},
	}, nil
}

// Channel returns a Channel by ID.
func (fs *FileStore) Channel(id string) (*store.Channel, error) {
	return fs.cs.Get(id)
}
