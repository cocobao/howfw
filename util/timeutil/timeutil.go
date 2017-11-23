package timeutil

import (
	"errors"
	"time"
)

type DateFormate string

const (
	DAY_PATTERN        DateFormate = "01月02日 15时04分"
	DAY_PATTERN_NORMAL DateFormate = "2006-01-02 15:04:05"
	DAY_PATTERN_ZONE   DateFormate = "2006-01-02T15:04:05-07:00"
)

func TimeToString(t time.Time) string {
	return t.Format(string(DAY_PATTERN_NORMAL))
}

func TimeSecToString(sec int64) string {
	return TimestampToString(sec, DAY_PATTERN_NORMAL)
}

func TimeToZoneStr(sec int64) string {
	return TimestampToString(sec, DAY_PATTERN_ZONE)
}

func TimeToStringZn(sec int64) string {
	return TimestampToString(sec, DAY_PATTERN)
}

func StringToTime(s string) (time.Time, error) {
	if s == "0000-00-00 00:00:00" {
		return time.Time{}, errors.New("StringToTime failed! cannot parse 0000-00-00 00:00:00 as time")
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	return t, err
}

func GetNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GetTimeZone() string {
	return time.Now().Format("-0700")
}

//时间戳转换成指定的时间格式
func TimestampToString(sec int64, formate DateFormate) string {
	return time.Unix(sec, 0).Format(string(formate))
}

//时间字符串转时间戳, 格式为"0000-00-00 00:00:00"
func StringToTimestamp(s string) int64 {
	t, err := StringToTime(s)
	if err != nil {
		return -1
	}
	return t.Unix()
}
