package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Timestamp represents slack's timestamp.
type Timestamp string

// Time converts a Timestamp to time.Time.
func (ts Timestamp) Time() (time.Time, error) {
	ss := strings.SplitN(string(ts), ".", 2)
	sec, err := strconv.ParseInt(ss[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("syntax error in seconds part: %q", ts)
	}
	n := len(ss[1])
	if n > 9 {
		ss[1] = ss[1][:9]
	} else if n < 9 {
		ss[1] = ss[1] + strings.Repeat("0", 9-n)
	}
	nsec, err := strconv.ParseInt(ss[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("syntax error in nano seconds part: %q", ts)
	}
	return time.Unix(sec, nsec).In(defaultLocation), nil
}

// TimestampToTime converts slack's `Ts` string to time.Time.
func TimestampToTime(ts string) (time.Time, error) {
	return Timestamp(ts).Time()
}
