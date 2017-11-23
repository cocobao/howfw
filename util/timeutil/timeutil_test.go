package timeutil

import (
	"fmt"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	n := time.Now()
	s := n.Format("2006-01-02 15:04:05")
	if s1 := TimeToString(n); s != s1 {
		fmt.Println(s1)
		t.Errorf("time to string failed! %s", s)
	}
	s2 := "2016-12-04 15:44:00"
	if t1, _ := StringToTime(s2); t1.Format("2006-01-02 15:04:05") != s2 {
		t.Errorf("string to time failed! %s", s2)
	}

	s3 := "2016-aaaaaaaaa"
	if _, err := StringToTime(s3); err == nil {
		t.Errorf("应该转换失败，但没有! %s", s3)
	}

	s4 := "0000-00-00 00:00:"
	if _, err1 := StringToTime(s4); err1 == nil {
		t.Errorf("0000-00-00 00:00:00 应该转换失败，但没有! %s", s4)
	}
	fmt.Println("Now:" + GetNow())
	fmt.Println("Zone:" + GetTimeZone())
}
func TestTimestampToString(t *testing.T) {
	timestamp := int64(1485165212)
	fmt.Println(TimestampToString(timestamp, DAY_PATTERN))
}
