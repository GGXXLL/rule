package dto

import (
	"encoding/json"
	"time"

	"github.com/spf13/cast"
)

// Payload provide common query
type Payload map[string]interface{}

func (p Payload) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (p Payload) Now() time.Time {
	return time.Now()
}

func (p Payload) Date(s string) time.Time {
	date, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		panic(err)
	}
	return date
}

func (p Payload) DaysAgo(s string) int {
	if s == "" {
		return 0
	}
	return int(time.Since(p.DateTime(s)).Hours() / 24)
}

func (p Payload) HoursAgo(s string) int {
	if s == "" {
		return 0
	}
	return int(time.Since(p.DateTime(s)).Hours())
}

func (p Payload) MinutesAgo(s string) int {
	if s == "" {
		return 0
	}
	return int(time.Since(p.DateTime(s)).Minutes())
}

func (p Payload) DateTime(s string) time.Time {
	date, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		panic(err)
	}
	return date
}

func (p Payload) IsBefore(s string) bool {
	var (
		t   time.Time
		err error
	)
	if len(s) == 10 {
		t, err = time.ParseInLocation("2006-01-02", s, time.Local)
	} else {
		t, err = time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	}
	if err != nil {
		panic(err)
	}
	return time.Now().Before(t)
}

func (p Payload) IsAfter(s string) bool {
	var (
		t   time.Time
		err error
	)
	if len(s) == 10 {
		t, err = time.ParseInLocation("2006-01-02", s, time.Local)
	} else {
		t, err = time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	}
	if err != nil {
		panic(err)
	}
	return time.Now().After(t)
}

func (p Payload) IsBetween(begin string, end string) bool {
	return p.IsAfter(begin) && p.IsBefore(end)
}

func (p Payload) IsWeekday(day int) bool {
	return time.Now().Weekday() == time.Weekday(day)
}

func (p Payload) IsWeekend() bool {
	if weekday := time.Now().Weekday(); weekday == 0 || weekday == 6 {
		return true
	}
	return false
}

func (p Payload) IsToday(s string) bool {
	return time.Now().Format("2006-01-02") == s
}

func (p Payload) IsHourRange(begin int, end int) bool {
	now := time.Now().Hour()
	return now >= begin && now <= end
}

func (p Payload) ToString(str interface{}) string {
	return cast.ToString(str)
}

func (p Payload) ToInt(int interface{}) int {
	return cast.ToInt(int)
}

type Data map[string]interface{}
