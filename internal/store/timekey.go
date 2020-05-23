package store

import (
	"time"
)

// TimeKey is key type for messages.
type TimeKey struct {
	Begin time.Time
	End   time.Time
}

var deafultLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	deafultLocation = loc
}

// TimeKeyYM creates a TimeKey of year/month.
func TimeKeyYM(year, month int) TimeKey {
	b := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, deafultLocation)
	e := b.AddDate(0, 1, 0)
	return TimeKey{Begin: b, End: e}
}

// TimeKeyYMD creates a TimeKey of year/month/day.
func TimeKeyYMD(year, month, day int) TimeKey {
	b := time.Date(year, time.Month(month), day, 0, 0, 0, 0, deafultLocation)
	e := b.AddDate(0, 0, 1)
	return TimeKey{Begin: b, End: e}
}
