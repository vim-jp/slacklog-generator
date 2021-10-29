package filestore

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// jsonReadFile reads a file and unmarshal its contents as JSON to `dst`
// destination object.
func jsonReadFile(name string, strict bool, dst interface{}) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	if strict {
		d.DisallowUnknownFields()
	}
	err = d.Decode(dst)
	if err != nil {
		return err
	}
	return nil
}

func jsonWriteFile(name string, src interface{}) error {
	dir := filepath.Dir(name)
	if dir != "." {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	e := json.NewEncoder(f)
	err = e.Encode(src)
	if err != nil {
		return err
	}
	return nil
}
