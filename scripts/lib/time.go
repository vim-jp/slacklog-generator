package slacklog

import (
	"strconv"
	"strings"
	"time"
)

func TsToDateTime(ts string) time.Time {
	t := strings.Split(ts, ".")
	if len(t) != 2 {
		return time.Time{}
	}
	sec, err := strconv.ParseInt(t[0], 10, 64)
	if err != nil {
		return time.Time{}
	}
	nsec, err := strconv.ParseInt(t[1], 10, 64)
	if err != nil {
		return time.Time{}
	}
	japan, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}
	}
	return time.Unix(sec, nsec).In(japan)
}
