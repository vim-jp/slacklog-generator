package slacklog

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Ts struct {
	IsNumber bool
	Value    string
}

var _ json.Marshaler = (*Ts)(nil)
var _ json.Unmarshaler = (*Ts)(nil)

func (ts *Ts) UnmarshalJSON(b []byte) error {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return nil
	}
	switch w := v.(type) {
	case string:
		ts.IsNumber = false
		ts.Value = w
	case float64:
		ts.IsNumber = true
		ts.Value = strconv.FormatFloat(w, 'g', -1, 64)
	default:
		return fmt.Errorf("unsupproted type for Ts: %t", v)
	}
	return nil
}

func (ts *Ts) MarshalJSON() ([]byte, error) {
	if ts.IsNumber {
		f, err := strconv.ParseFloat(ts.Value, 64)
		if err != nil {
			f = 0
		}
		return json.Marshal(f)
	}
	return json.Marshal(ts.Value)
}
