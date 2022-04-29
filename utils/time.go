package utils

import (
	"time"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
	DATE_FORMAT = "2006-01-02"
)

func Format() string {
	return time.Now().Format(TIME_FORMAT)
}

func String2TimeStamp(datetime string) int64 {
	loc, _ := time.LoadLocation("Local")                      //重要：获取时区
	tm, _ := time.ParseInLocation(TIME_FORMAT, datetime, loc) //使用模板在对应时区转化为time.time类型
	return tm.Unix()
}

func TimeStamp2String(timestamp int64) (datetime string) {
	datetime = time.Unix(timestamp, 0).Format(TIME_FORMAT)
	return
}
