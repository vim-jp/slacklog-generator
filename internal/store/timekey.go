package store

import (
	"time"
)

// TimeKey is key type for messages.
type TimeKey struct {
	Begin time.Time
	End   time.Time
}

// Include checks a time `v` is between `Begin` (inclusive) and `End`
// (exclusive)
func (tk TimeKey) Include(v time.Time) bool {
	if v.Before(tk.Begin) {
		return false
	}
	return tk.End.After(v)
}

var defaultLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	defaultLocation = loc
}

// TimeKeyYM creates a TimeKey of year/month.
func TimeKeyYM(year, month int) TimeKey {
	b := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, defaultLocation)
	e := b.AddDate(0, 1, 0)
	return TimeKey{Begin: b, End: e}
}

// TimeKeyYMD creates a TimeKey of year/month/day.
func TimeKeyYMD(year, month, day int) TimeKey {
	b := time.Date(year, time.Month(month), day, 0, 0, 0, 0, defaultLocation)
	e := b.AddDate(0, 0, 1)
	return TimeKey{Begin: b, End: e}
}

// TimeKeyDate creates a `TimeKey` of `ti.Date()`.
func TimeKeyDate(ti time.Time) TimeKey {
	y, m, d := ti.Date()
	return TimeKeyYMD(y, int(m), d)
}
