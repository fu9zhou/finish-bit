package localtools

import (
	"math/bits"
	"strconv"
	"strings"
	"time"
)

const firstLunarTableYear = 1899

// The preceding lunar year covers January 1900. The final lunar year ends
// in January 2101, as allowed by calendar.solar's lunar-year input contract.
var lunarEpoch = time.Date(1899, time.February, 10, 0, 0, 0, 0, time.UTC)

type lunarDate struct {
	year, month, day int
	leap             bool
}

type lunarYearBits uint32

func (y lunarYearBits) leapMonth() int { return int(y & 15) }

func (y lunarYearBits) monthCount() int {
	if y.leapMonth() != 0 {
		return 13
	}
	return 12
}

func (y lunarYearBits) monthDays(index int) int {
	return 29 + int((y>>(4+index))&1)
}

func (y lunarYearBits) days() int {
	return 29*y.monthCount() + bits.OnesCount32(uint32(y)>>4)
}

func lunarFromSolar(t time.Time) (lunarDate, error) {
	if t.Year() < 1900 || t.Year() > 2100 {
		return lunarDate{}, invalid("supported Gregorian years are 1900–2100")
	}
	days := int(civilDays(t) - civilDays(lunarEpoch))
	for offset, entry := range lunarYears {
		y := lunarYearBits(entry)
		if days >= y.days() {
			days -= y.days()
			continue
		}
		for index := 0; index < y.monthCount(); index++ {
			if days >= y.monthDays(index) {
				days -= y.monthDays(index)
				continue
			}
			month, leap := index+1, false
			if y.leapMonth() != 0 && index >= y.leapMonth() {
				month, leap = index, index == y.leapMonth()
			}
			return lunarDate{firstLunarTableYear + offset, month, days + 1, leap}, nil
		}
	}
	return lunarDate{}, invalid("date is outside the lunar table")
}

func solarFromLunar(l lunarDate) (time.Time, error) {
	if l.year < 1900 || l.year > 2100 || l.month < 1 || l.month > 12 || l.day < 1 || l.day > 30 {
		return time.Time{}, invalid("lunar year 1900–2100, month 1–12, day 1–30 required")
	}
	y := lunarYearBits(lunarYears[l.year-firstLunarTableYear])
	if l.leap && y.leapMonth() != l.month {
		return time.Time{}, invalid("invalid date or nonexistent lunar leap month")
	}
	index := l.month - 1
	if y.leapMonth() != 0 && (l.month > y.leapMonth() || l.leap) {
		index++
	}
	if l.day > y.monthDays(index) {
		return time.Time{}, invalid("nonexistent lunar date")
	}
	days := l.day - 1
	for _, entry := range lunarYears[:l.year-firstLunarTableYear] {
		days += lunarYearBits(entry).days()
	}
	for i := 0; i < index; i++ {
		days += y.monthDays(i)
	}
	return lunarEpoch.AddDate(0, 0, days), nil
}

func (l lunarDate) text() string {
	var b strings.Builder
	digits := []rune("〇一二三四五六七八九")
	for _, digit := range strconv.Itoa(l.year) {
		b.WriteRune(digits[digit-'0'])
	}
	b.WriteString("年")
	if l.leap {
		b.WriteString("闰")
	}
	months := [...]string{"正", "二", "三", "四", "五", "六", "七", "八", "九", "十", "冬", "腊"}
	b.WriteString(months[l.month-1])
	b.WriteString("月")
	switch l.day {
	case 10:
		b.WriteString("初十")
	case 20:
		b.WriteString("二十")
	case 30:
		b.WriteString("三十")
	default:
		prefix := [...]string{"初", "十", "廿"}
		b.WriteString(prefix[l.day/10])
		b.WriteRune(digits[l.day%10])
	}
	return b.String()
}
