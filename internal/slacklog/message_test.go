package slacklog

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMessageAttachment_Ts_JSON(t *testing.T) {
	for _, tc := range []struct {
		json string
		ma   MessageAttachment
	}{
		{`{"id":0,"ts":"1234.5678"}`, MessageAttachment{
			ID: 0,
			Ts: &Ts{false, "1234.5678"},
		}},
		{`{"id":0,"ts":1234.5678}`, MessageAttachment{
			ID: 0,
			Ts: &Ts{true, "1234.5678"},
		}},
		{`{"id":0}`, MessageAttachment{ID: 0, Ts: nil}},
	} {
		var act1 MessageAttachment
		err := json.Unmarshal([]byte(tc.json), &act1)
		if err != nil {
			t.Fatalf("unmarshal(%q) failed: %s", tc.json, err)
		}
		if diff := cmp.Diff(tc.ma, act1); diff != "" {
			t.Fatalf("unexpected unmarshal MessageAttachment: -want +got\n%s", diff)
		}

		b, err := json.Marshal(tc.ma)
		if err != nil {
			t.Fatalf("marshal(%+v) failed: %s", tc.ma, err)
		}
		act2 := string(b)
		if diff := cmp.Diff(tc.json, act2); diff != "" {
			t.Fatalf("unexpected marshal Ts: -want +got\n%s", diff)
		}
	}
}
