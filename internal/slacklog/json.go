package slacklog

import (
	"encoding/json"
	"os"
)

// ReadFileAsJSON reads a file and unmarshal its contents as JSON to `dst`
// destination object.
func ReadFileAsJSON(filename string, dst interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	err = d.Decode(dst)
	if err != nil {
		return err
	}
	return nil
}
