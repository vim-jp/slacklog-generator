package slacklog

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func TsToDateTime(ts string) time.Time {
	t := strings.Split(ts, ".")
	if len(t) != 2 {
		fmt.Fprintf(os.Stderr, "[warning] invalid timestamp: %s ...\n", ts)
		return time.Time{}
	}
	sec, err := strconv.ParseInt(t[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[warning] invalid timestamp: %s ...\n", ts)
		return time.Time{}
	}
	nsec, err := strconv.ParseInt(t[1], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[warning] invalid timestamp: %s ...\n", ts)
		return time.Time{}
	}
	japan, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[warning] invalid timestamp: %s ...\n", ts)
		return time.Time{}
	}
	return time.Unix(sec, nsec).In(japan)
}

// LevelOfDetailTime returns a label string which represents time.  The
// resolution of the string is determined by differece from base time in 4
// levels.
func LevelOfDetailTime(target, base time.Time) string {
	if target.Year() != base.Year() {
		return target.Format("2006年1月2日 15:04:05")
	}
	if target.Month() != base.Month() {
		return target.Format("1月2日 15:04:05")
	}
	if target.Day() != base.Day() {
		return target.Format("2日 15:04:05")
	}
	return target.Format("15:04:05")
}
