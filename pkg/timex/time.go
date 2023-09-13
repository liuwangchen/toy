package timex

import (
	"time"

	"github.com/liuwangchen/toy/pkg/mathx"
)

type None struct{}

const (
	StandardTimeLayout       = "2006-01-02 15:04:05"
	ShortTimeLayout          = "2006-01-02 15:04"
	DateLayout               = "2006-01-02"
	DAY_SECS           int64 = 86400
)

// 获得当天0点
func DailyZero() int64 {
	now := time.Now()
	hour, minute, second := now.Hour(), now.Minute(), now.Second()
	return now.Unix() - int64((hour*3600)+(minute*60)+second)
}

// 获得下一天0点
func NextZero() int64 {
	return DailyZero() + DAY_SECS
}

// 当前整点时间戳
func CurrentHourUnix() int64 {
	now := time.Now()
	minute, second := now.Minute(), now.Second()
	return now.Unix() - int64((minute*60)+second)
}

func DailyZeroByTime(ts int64) int64 {
	t := time.Unix(ts, 0)
	hour, minute, second := t.Hour(), t.Minute(), t.Second()
	return ts - int64((hour*3600)+(minute*60)+second)
}

// 计算两个时间点之间相隔几天
func DaysBetweenTimes(ts1, ts2 int64) int64 {
	zero1 := DailyZeroByTime(ts1)
	zero2 := DailyZeroByTime(ts2)
	var sub int64
	if zero1 > zero2 {
		sub = zero1 - zero2
	} else {
		sub = zero2 - zero1
	}
	return sub / DAY_SECS
}

func UniqueWeek(tm time.Time) uint32 {
	year, week := tm.ISOWeek()
	return (uint32(year) << 16) + uint32(week)
}

// 获取指定时间当前周几零点(星期一作为一周的开始)
// @param day - 1~7
func WeekdayZeroByTime(ts int64, wd int) int64 {
	day := int64(time.Unix(ts, 0).Weekday())
	zero := DailyZeroByTime(ts)
	if day == 0 { // 星期日
		day = 7
	}
	return zero + (int64(wd)-day)*DAY_SECS
}

// 获取本周几零点(星期一作为一周的开始)
// @param day - 1~7
func WeekdayZero(wd int) int64 {
	day := int64(time.Now().Weekday())
	zero := DailyZero()
	if day == 0 { // 星期日
		day = 7
	}
	return zero + (int64(wd)-day)*DAY_SECS
}

func DoubleUint32ToUint64(p1, p2 uint32) uint64 {
	return uint64(p1)<<32 | uint64(p2)
}

func InSameDay(t1, t2 int64) bool {
	y1, m1, d1 := time.Unix(t1, 0).Date()
	y2, m2, d2 := time.Unix(t2, 0).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func GetMonthDays(year, month int) int {
	//	year := time.Now().Year()
	//	month := time.Now().Month()
	var days int

	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30
		} else {
			days = 31
		}
	} else {
		if (year%4) == 0 && (year%100) != 0 || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}

	return days
}

// DayZeroTime 每天的零点时间
// @param t1(s) time.Unix()
// @param cutTime(s) 一天的分割时间（如凌晨5点跨天，则cutTime=5*3600）
// @param timeZone(s) 时区偏移秒数
func DayZeroTime(t int64, cutTime int64, timeZone int64) int64 {
	return (t+timeZone-cutTime)/(3600*24)*(3600*24) - timeZone + cutTime
}

// WeekZeroTime 每天的零点时间
// @param t1(s) time.Unix()
// @param cutTime(s) 一天的分割时间（如凌晨5点跨天，则cutTime=5*3600）
// @param timeZone(s) 时区偏移秒数
func WeekZeroTime(t int64, cutTime int64, timeZone int64) int64 {
	var oneWeekSeconds int64 = 7 * 24 * 3600
	offset := 3*24*3600 + timeZone - cutTime // 1970.1.1是周4,周1，2，3转为下一周周四
	return (t+offset)/oneWeekSeconds*oneWeekSeconds - offset

}

// IsSameDay 判断t1/t2是否同一天
// @param t1(s) time.Unix()
// @param t2(s) time.Unix()
// @param cutTime(s) 一天的分割时间（如凌晨5点跨天，则cutTime=5*3600）
// @param timeZone(s) 时区偏移秒数
func IsSameDay(t1 int64, t2 int64, cutTime int64, timeZone int64) bool {
	ts1 := (t1 + timeZone - cutTime) / 24 / 3600
	ts2 := (t2 + timeZone - cutTime) / 24 / 3600
	return ts1 == ts2
}

// IsSameWeek 判断t1/t2是否同一周
// @param t1(s) time.Unix()
// @param t2(s) time.Unix()
// @param cutTime(s) 一天的分割时间（如凌晨5点跨天，则cutTime=5*3600）
// @param timeZone(s) 时区偏移秒数
func IsSameWeek(t1 int64, t2 int64, cutTime int64, timeZone int64) bool {
	// 1970/1/1为周四，所以加上3天（3*24*3600）
	ts1 := (t1+3*24*3600+timeZone-cutTime)/(24*3600*7) + 1
	ts2 := (t2+3*24*3600+timeZone-cutTime)/(24*3600*7) + 1
	return ts1 == ts2
}

// UnixToTimeString 把unix秒专成"2015-11-01 12:20:22"格式时间字符串
func UnixToTimeString(unix int64) string {
	return time.Unix(unix, 0).Format(StandardTimeLayout)
}

// GetWeekDayInRangeTime 获取这段时间内周几的天
func GetWeekDayInRangeTime(startTime, endTime time.Time, week time.Weekday) []time.Time {
	result := make([]time.Time, 0)
	for !startTime.After(endTime) {
		if startTime.Weekday() == week {
			result = append(result, startTime)
			startTime = startTime.Add(time.Hour * 24 * time.Duration(7))
			continue
		}
		w1 := startTime.Weekday() % 7
		if w1 == 0 {
			w1 = 7
		}

		w2 := week % 7
		if w2 == 0 {
			w2 = 7
		}

		// 计算startTime距离目标weekday差几天
		sub := mathx.AbsInt(int(w2 - w1))

		startTime = startTime.Add(time.Hour * 24 * time.Duration(sub))
	}
	return result
}
