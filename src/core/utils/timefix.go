package utils

import (
	"root/core/log"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// 时差修正 纳秒差值
	timefix int64
	// 服务器使用的标准时区
	location *time.Location = time.Local
)

const (
	STD_TIMEFORMAT         = "2006-01-02 15:04:05"
	STD_DAY_FORMAT         = "20060102"
	ZEROCLOCK              = "00:00:00"
	MILLISECONDS_OF_DAY    = 24 * MILLISECONDS_OF_HOUR   // 一天的毫秒数
	MILLISECONDS_OF_HOUR   = 60 * MILLISECONDS_OF_MINUTE //一小时的毫秒数
	MILLISECONDS_OF_MINUTE = 60 * MILLISECONDS_OF_SECOND //一分钟的毫秒数
	MILLISECONDS_OF_SECOND = 1000                        //一秒的毫秒数
)

func InitLocalTime(section string) {
	local, err := time.LoadLocation(section)
	if err != nil {
		return
	}
	location = local
}

//重置时间，传入utc标准时间UnixNano()
func ResetTime(t int64) {
	timefix = time.Now().UnixNano() - t
}

//当前本地时间 纳秒 已修正
func NanosTime() int64 {
	return time.Now().UnixNano()
}

//当前本地时间 秒 已修正
func SecondTimeSince1970() int64 {
	return time.Now().UnixNano() / int64(time.Second)
}

//获取毫秒
func MilliSecondTime() int64 {
	return NanosTime() / int64(time.Millisecond)
}

//获取毫秒
func MilliSecondTimeSince1970() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//时间转成毫秒
func TimeMillis(t time.Duration) int64 {
	return t.Nanoseconds() / int64(time.Millisecond)
}

//时间转成秒
func TimeSecond(t time.Duration) int64 {
	return t.Nanoseconds() / int64(time.Second)
}

//纳秒时间转成服务器时间(根据传入的时间的时区在返回的时间上进行时区纠正time.Time().In(tm.Location()))
func Nanos2Time(ns int64) time.Time {
	return time.Unix(ns/int64(time.Second), ns%int64(time.Second)).In(location)
}

//秒时间转成服务器时间(根据传入的时间的时区在返回的时间上进行时区纠正time.Time().In(tm.Location()))
func Second2Time(second int64) time.Time {
	return time.Unix(second, 0).In(location)
}

// 获得指定时间的凌晨时间
func Time2Midnight(tm time.Time) time.Time {
	year, month, day := tm.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, tm.Location())
}

//获取一个时间点的 x日后的凌晨时间
func NextMidnight(tm time.Time, day int) time.Time {
	midTime := Time2Midnight(tm)
	ns := midTime.UnixNano() + int64(day*MILLISECONDS_OF_DAY)*int64(time.Millisecond)
	return Nanos2Time(ns).In(tm.Location())
}

// 从一个时间戳获取下一个准点时间
func NextHour(tm time.Time) time.Time {
	year, month, day := tm.Date()
	hour, _, _ := tm.Clock()
	return time.Date(year, month, day, hour+1, 0, 0, 0, tm.Location())
}

// 从一个时间戳获取下一个准点时间
func NextHours(tm time.Time, n int) time.Time {
	year, month, day := tm.Date()
	hour, _, _ := tm.Clock()
	return time.Date(year, month, day, hour+n, 0, 0, 0, tm.Location())
}

// 判断两个时间是否是同一天（默认将两个时间转换成0时区的时间进行比较）
func OtherDay(t1, t2 time.Time) bool {
	year1, month1, day1 := t1.UTC().Date()
	year2, month2, day2 := t2.UTC().Date()
	return year1 == year2 && month1 == month2 && day1 == day2
}

// 判断两个时间是否是同一周
func OtherWeek(beforeTime, afterTime time.Time) bool {
	year1, week1 := beforeTime.UTC().ISOWeek()
	year2, week2 := afterTime.UTC().ISOWeek()
	return year1 == year2 && week1 == week2
}

// 判断两个时间是否是同一天(t1,t2为秒)
func OtherDayByUnix(t1, t2 int64) bool {
	div := (t1 - t2) * MILLISECONDS_OF_SECOND
	if div < 0 {
		div = -div
	}
	return div < MILLISECONDS_OF_DAY
}

// date format: "2006-01-02 13:04:00"
func S2UnixTime(value string, loc *time.Location) (*time.Time, error) {
	re := regexp.MustCompile(`([\d]+)-([\d]+)-([\d]+) ([\d]+):([\d]+):([\d]+)`)
	slices := re.FindStringSubmatch(value)
	if slices == nil || len(slices) != 7 {
		return nil, fmt.Errorf("time[%s] format error, expect format: 2006-01-02 13:04:00 \n", value)
	}
	year, _ := strconv.Atoi(slices[1])
	month, _ := strconv.Atoi(slices[2])
	day, _ := strconv.Atoi(slices[3])
	hour, _ := strconv.Atoi(slices[4])
	min, _ := strconv.Atoi(slices[5])
	sec, _ := strconv.Atoi(slices[6])
	tt := time.Date(year, time.Month(month), day, hour, min, sec, 0, loc)
	return &tt, nil
}

//获取指定时间的下一个周末时间,自己处理好时区的问题,周一凌晨作为跨周，如果传入的时间没有超过周一凌晨，这返回周一凌晨的时间点，否则返回下一周的周一凌晨
func NextSundayMS(tm time.Time) time.Time {
	mt := NextMidnight(tm, 0)
	weekday := tm.Weekday()
	if weekday == time.Sunday {
		return Nanos2Time(mt.UnixNano()).In(tm.Location()).Add(24 * time.Hour)
	}
	return Nanos2Time(mt.UnixNano() + int64((7-int64(weekday))*MILLISECONDS_OF_DAY*int64(time.Millisecond))).In(tm.Location()).Add(24 * time.Hour)
}

// 返回当前的整点时间
func SharpClock(tm time.Time) time.Time {
	year, month, day := tm.Date()
	hour, _, _ := tm.Clock()
	return time.Date(year, month, day, hour, 0, 0, 0, tm.Location())
}

// 获取指定日期对应的零点时间对象
func GetTimeToMidnightTime(tm time.Time) time.Time {
	year, month, day := tm.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, tm.Location())
}

//func main() {
//	//	t1 := S2UnixTime("2015-12-31 00:00:00", Server_Location)
//	t1 := S2UnixTime("2016-12-31 00:00:00", Server_Location)
//	t2 := NextSundayMS(t1.UTC())
//	fmt.Println(t1.Format(TimeFormate))
//	fmt.Println(t2.UTC().Format(TimeFormate))
//}

//返回从整点到现在的差值
func NowToSharpClock(tm time.Time) time.Duration {
	return time.Duration(tm.UnixNano() - SharpClock(tm).UnixNano())
}

//检查是否跨周
func CheckCrossWeek(base time.Time, now time.Time) bool {
	year1, week1 := base.ISOWeek()
	year2, week2 := now.ISOWeek()
	if (year1 == year2 && week2-week1 > 0) || ((year1 < year2) && (year2-year1 > 1 || week2 > 1 || now.Unix() > NextSundayMS(base).Unix())) { //跨周
		return true
	}
	return false
}

//检查是否跨天
func CheckCrossDay(base time.Time, now time.Time) bool {
	year1, month1, day1 := base.Date()
	year2, month2, day2 := now.Date()
	if (year1 == year2 && ((month1 == month2 && day2-day1 >= 1) || (month1 < month2))) || (year1 < year2) { //跨天
		return true
	}
	return false
}

//检查是否跨月
func CheckCrossMonth(base time.Time, now time.Time) bool {
	year1, month1, _ := base.Date()
	year2, month2, _ := now.Date()
	if (year1 == year2 && month1 < month2) || (year1 < year2) { //跨天
		return true
	}
	return false
}

var days = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

//GetDayInMon 获取月份的天数
func GetDayInMon(year int, mon int) int {
	var day int
	if 2 == mon {
		if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
			day = 29
		} else {
			day = 28
		}
	} else {
		day = days[mon-1]
	}

	return day
}

//GetTimeFormatString 获取日期字符串; 参数单位: 秒
func GetTimeFormatString(stamp int64) string {
	t := time.Unix(stamp, 0)
	return t.Format(STD_TIMEFORMAT)
}

//  字符串转换成时间戳(毫秒)  字符串格式："2006-01-02 15:04:05"
func String2UnixStamp(strTime string) int64 {
	//loc, _ := time.LoadLocation("Local") //重要：获取时区
	t, err := time.ParseInLocation(STD_TIMEFORMAT, strTime, location)
	if err != nil {
		log.Errorf(err.Error())
		return 0
	}

	return t.UnixNano() / int64(time.Millisecond)
}

func DateString() string {
	return time.Now().Format(STD_TIMEFORMAT)
}

// 时间(秒)  帧间隔(秒)
func Second2sn(time float32) int32 {
	return int32(time / float32(0.05))
}

// 获得下一个时间点的时差 "15:04:05"
func GetStamp(date string) int64 {
	t := time.Now()
	slices := strings.Split(date, ":")
	year, month, day := t.Date()
	hour, _ := strconv.Atoi(slices[0])
	min, _ := strconv.Atoi(slices[1])
	sec, _ := strconv.Atoi(slices[2])
	tt := time.Date(year, month, day, hour, min, sec, 0, t.Location())
	timeStamp := tt.UnixNano() / 1e6
	now := MilliSecondTime()

	// 如果当前时间大于到期时间，切换到下一天该点的时间戳
	if now > timeStamp {
		timeStamp = time.Date(year, month, day+1, hour, min, sec, 0, t.Location()).UnixNano() / 1e6
	}

	return timeStamp
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			// 检查ip地址判断是否回环地址
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return "0.0.0.0"
}
