package slacklog

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTs_Unmarshal(t *testing.T) {
	for _, tc := range []struct {
		json string
		exp  Ts
	}{
		{`"1234.5678"`, Ts{false, "1234.5678"}},
		{`1234.5678`, Ts{true, "1234.5678"}},
	} {
		var act Ts
		err := json.Unmarshal([]byte(tc.json), &act)
		if err != nil {
			t.Fatalf("unmarshal(%q) failed: %s", tc.json, err)
		}
		if diff := cmp.Diff(tc.exp, act); diff != "" {
			t.Fatalf("unexpected unmarshal Ts: -want +got\n%s", diff)
		}
	}
}

func TestTs_Marshal(t *testing.T) {
	for _, tc := range []struct {
		ts  Ts
		exp string
	}{
		{Ts{false, "1234.5678"}, `"1234.5678"`},
		{Ts{true, "1234.5678"}, `1234.5678`},
	} {
		var act string
		b, err := json.Marshal(tc.ts)
		if err != nil {
			t.Fatalf("marshal(%+v) failed: %s", tc.ts, err)
		}
		act = string(b)
		if diff := cmp.Diff(tc.exp, act); diff != "" {
			t.Fatalf("unexpected marshal Ts: -want +got\n%s", diff)
		}
	}
}
