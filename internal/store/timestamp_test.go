package store

import (
	"testing"
	"time"
)

func parseTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t.Fatalf("parseTime failed: %s", err)
	}
	return v
}

func TestTimestampToTime(t *testing.T) {
	for i, tc := range []struct {
		ts  string
		exp time.Time
	}{
		{"1583688494.253200", parseTime(t, "2020-03-09T02:28:14.253200000+09:00")},
		{"1583688494.2532001234", parseTime(t, "2020-03-09T02:28:14.253200123+09:00")},
		{"1583688494.253", parseTime(t, "2020-03-09T02:28:14.253000000+09:00")},
	} {
		act, err := TimestampToTime(tc.ts)
		if err != nil {
			t.Fatalf("TimestampToTime(%q) failed: %s", tc.ts, err)
		}
		if !act.Equal(tc.exp) {
			t.Fatalf("unexpected result for #%d: %+v %v", i, tc.exp, act)
		}
	}
}

func TestTimestampToTime_NG(t *testing.T) {
	for i, tc := range []struct {
		ts  string
		err string
	}{
		{"15836X8494.253200", `syntax error in seconds part: "15836X8494.253200"`},
		{"1583688494.253X00", `syntax error in nano seconds part: "1583688494.253X00"`},
	} {
		_, err := TimestampToTime(tc.ts)
		if err == nil {
			t.Fatalf("unexpected success #%d with %s", i, tc.err)
		}
		if act := err.Error(); act != tc.err {
			t.Fatalf("unexpected failure for #%d\nwant=%s\n got=%s", i, tc.err, act)
		}
	}
}
