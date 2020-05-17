package jsonwriter

import (
	"encoding/json"
	"os"
)

// WriteCloser writes objects as JSON array. It provides persistent layer for JSON value.
type WriteCloser interface {
	Write(interface{}) error
	Close() error
}

type fileWriter struct {
	name    string
	reverse bool

	// FIXME: いったん全部メモリにためるのであんまよくない
	buf []interface{}
}

func (fw *fileWriter) Write(v interface{}) error {
	// XXX: 排他してないのでgoroutineからは使えない
	fw.buf = append(fw.buf, v)
	return nil
}

func (fw *fileWriter) Close() error {
	// FIXME: ファイルの作成が Close まで遅延している。本来なら CreateFile のタ
	// イミングでやるのが好ましいが、いましばらく目を瞑る
	f, err := os.Create(fw.name)
	if err != nil {
		return err
	}
	defer f.Close()
	if fw.reverse {
		reverse(fw.buf)
		fw.reverse = false
	}
	err = json.NewEncoder(f).Encode(fw.buf)
	if err != nil {
		return err
	}
	fw.buf = nil
	return nil
}

// CreateFile creates a WriteCloser which implemented by file.
func CreateFile(name string, reverse bool) (WriteCloser, error) {
	return &fileWriter{name: name}, nil
}

func reverse(x []interface{}) {
	for i, j := 0, len(x)-1; i < j; {
		x[i], x[j] = x[j], x[i]
		i++
		j--
	}
}
