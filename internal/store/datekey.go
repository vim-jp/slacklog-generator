package store

import (
	"fmt"
	"time"
)

// DateKey represents a date.
type DateKey struct {
	Y uint16
	M uint8
	D uint8
}

// Time2DateKey converts time.Time to DateKey.
func Time2DateKey(v time.Time) DateKey {
	y, m, d := v.Date()
	return DateKey{Y: uint16(y), M: uint8(m), D: uint8(d)}
}

func (dk DateKey) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", dk.Y, dk.M, dk.D)
}

// Include checks a time `v` is in DateKey.
func (dk DateKey) Include(v time.Time) bool {
	return Time2DateKey(v) == dk
}
