package localtools

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestLunarPublishedDates(t *testing.T) {
	// Independent month boundaries from the HKO 2024, 2033 and 2034 tables:
	// https://www.hko.gov.hk/en/gts/time/calendar/text/files/T2033e.txt
	// In particular, the two successive eleventh months in 2033 are intentional.
	for _, tc := range []struct {
		solar string
		lunar lunarDate
		text  string
	}{
		{"2024-02-09", lunarDate{2023, 12, 30, false}, "二〇二三年腊月三十"},
		{"2024-02-10", lunarDate{2024, 1, 1, false}, "二〇二四年正月初一"},
		{"2024-02-29", lunarDate{2024, 1, 20, false}, "二〇二四年正月二十"},
		{"2033-11-22", lunarDate{2033, 11, 1, false}, "二〇三三年冬月初一"},
		{"2033-12-21", lunarDate{2033, 11, 30, false}, "二〇三三年冬月三十"},
		{"2033-12-22", lunarDate{2033, 11, 1, true}, "二〇三三年闰冬月初一"},
		{"2034-01-19", lunarDate{2033, 11, 29, true}, "二〇三三年闰冬月廿九"},
		{"2034-01-20", lunarDate{2033, 12, 1, false}, "二〇三三年腊月初一"},
		{"2034-02-19", lunarDate{2034, 1, 1, false}, "二〇三四年正月初一"},
	} {
		t.Run(tc.solar, func(t *testing.T) {
			date, err := time.Parse("2006-01-02", tc.solar)
			if err != nil {
				t.Fatal(err)
			}
			got, err := lunarFromSolar(date)
			if err != nil || got != tc.lunar || got.text() != tc.text {
				t.Fatalf("got %+v (%s), %v; want %+v (%s)", got, got.text(), err, tc.lunar, tc.text)
			}
			back, err := solarFromLunar(tc.lunar)
			if err != nil || !back.Equal(date) {
				t.Fatalf("reverse: %s, %v", back, err)
			}
		})
	}
}

func TestLunarFullRangeCompatibility(t *testing.T) {
	// These digests were produced BEFORE the replacement by lunar-go v1.4.6,
	// across every supported day, including Chinese text. See data/lunar/README.md.
	// Keeping the reference output hash detects table drift without retaining
	// the removed library as a test dependency.
	h := sha256.New()
	count := 0
	for d := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC); d.Year() <= 2100; d = d.AddDate(0, 0, 1) {
		l, err := lunarFromSolar(d)
		if err != nil {
			t.Fatalf("%s: %v", d, err)
		}
		month := l.month
		if l.leap {
			month = -month
		}
		fmt.Fprintf(h, "%s %d %d %d %s\n", d.Format("2006-01-02"), l.year, month, l.day, l.text())
		count++
	}
	if count != 73414 || fmt.Sprintf("%x", h.Sum(nil)) != "d7065ef43ad2765a1a95f4c154c46cfbd79cd478c43a783571a55405a2b747bc" {
		t.Fatalf("solar compatibility: %d dates, digest %x", count, h.Sum(nil))
	}
	h.Reset()
	count = 0
	for year := 1900; year <= 2100; year++ {
		for month := 1; month <= 12; month++ {
			for _, leap := range []bool{false, true} {
				for day := 1; day <= 30; day++ {
					l := lunarDate{year, month, day, leap}
					d, err := solarFromLunar(l)
					if err != nil {
						continue
					}
					m := month
					if leap {
						m = -m
					}
					fmt.Fprintf(h, "%d %d %d %s\n", year, m, day, d.Format("2006-01-02"))
					count++
				}
			}
		}
	}
	if count != 73412 || fmt.Sprintf("%x", h.Sum(nil)) != "0b39f730429d71be47e4b818838b71b30853c0f7674bc5a7ee551298679c4af0" {
		t.Fatalf("lunar compatibility: %d dates, digest %x", count, h.Sum(nil))
	}
}

func TestLunarRejectsInvalidDates(t *testing.T) {
	for _, l := range []lunarDate{
		{1899, 1, 1, false}, {2101, 1, 1, false}, {2024, 0, 1, false},
		{2024, 13, 1, false}, {2024, 1, 0, false}, {2024, 1, 31, false},
		{2024, 1, 1, true}, {2033, 11, 30, true}, {2033, 7, 1, true},
		{2024, 1, 30, false},
	} {
		if _, err := solarFromLunar(l); err == nil {
			t.Errorf("accepted invalid lunar date %+v", l)
		}
	}
	for _, year := range []int{1899, 2101} {
		if _, err := lunarFromSolar(time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)); err == nil {
			t.Errorf("accepted Gregorian year %d", year)
		}
	}
	reg := operation.NewRegistry()
	if err := Register(reg); err != nil {
		t.Fatal(err)
	}
	c, _ := reg.Get("calendar.solar")
	_, err := c.Runner.Run(context.Background(), operation.Request{
		Inputs: []string{"2033", "11", "30"}, Options: map[string]any{"leap": true},
	})
	var opErr *operation.Error
	if !errors.As(err, &opErr) || opErr.Code != operation.CodeInvalidInput {
		t.Fatalf("expected structured invalid input: %v", err)
	}
}

func TestLunarOperationRangeBoundaries(t *testing.T) {
	got := call(t, "calendar.lunar", []string{"1900-01-01"}, nil).Data
	if got["year"] != 1899 || got["month"] != 12 || got["day"] != 1 || got["text"] != "一八九九年腊月初一" {
		t.Fatalf("Gregorian lower boundary: %v", got)
	}
	got = call(t, "calendar.lunar", []string{"2100-12-31"}, nil).Data
	if got["year"] != 2100 || got["month"] != 12 || got["day"] != 1 {
		t.Fatalf("Gregorian upper boundary: %v", got)
	}
	got = call(t, "calendar.solar", []string{"2100", "12", "29"}, nil).Data
	if got["date"] != "2101-01-28" {
		t.Fatalf("lunar upper boundary: %v", got)
	}
}
