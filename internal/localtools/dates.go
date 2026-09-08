package localtools

import (
	"context"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func timeSpecs() []spec {
	rows := []spec{
		{id: "date.add", summary: "Add calendar years, months and days with clamped month ends", alias: "日期加减", inputs: in("date"), options: []operation.Parameter{integer("years", "Calendar years to add", 0), integer("months", "Calendar months to add", 0), integer("days", "Calendar days to add", 0)}},
		{id: "date.diff", summary: "Calculate calendar-day and elapsed-time differences", alias: "日期间隔", inputs: in("start", "end")},
		{id: "date.business-days", summary: "Count business days with explicit holidays and working-date overrides", alias: "工作日计算", inputs: in("start", "end"), options: []operation.Parameter{list("holidays", "Non-working YYYY-MM-DD dates"), list("working-dates", "Working YYYY-MM-DD dates, overriding holidays/weekends")}},
		{id: "date.expiry", summary: "Calculate expiry dates from calendar months or days", alias: "保质期计算", inputs: in("production-date"), options: []operation.Parameter{integer("months", "Shelf-life calendar months", 0), integer("days", "Shelf-life days", 0), str("as-of", "Reference date; empty means today", "")}},
		{id: "date.age", summary: "Calculate completed calendar years and next birthday", alias: "年龄计算", inputs: in("birth-date"), options: []operation.Parameter{str("as-of", "Reference date; empty means today", "")}},
		{id: "time.world", summary: "Show one instant in multiple IANA timezones", alias: "世界时间", inputs: []operation.Parameter{{Name: "instant", Type: operation.TypeString, Description: "RFC3339 instant; omit for now"}}, options: []operation.Parameter{list("zones", "IANA timezone names; default UTC, Asia/Shanghai, America/New_York, Europe/London")}},
	}
	for i := range rows {
		id := rows[i].id
		rows[i].options = append(rows[i].options, str("timezone", "IANA timezone for date inputs", "UTC"))
		rows[i].run = func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			return runDate(ctx, id, v, a)
		}
	}
	return rows
}
func parseDate(s string, loc *time.Location) (time.Time, error) {
	t, e := time.ParseInLocation("2006-01-02", s, loc)
	if e != nil || t.Year() < 1 || t.Year() > 9999 {
		return time.Time{}, invalid("date must be YYYY-MM-DD within years 1–9999")
	}
	return t, nil
}
func civilDays(t time.Time) int64 {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix() / 86400
}
func clampAdd(t time.Time, years, months, days int) (time.Time, error) {
	y := t.Year() + years
	m := int(t.Month()) - 1 + months
	y += m / 12
	m %= 12
	if m < 0 {
		m += 12
		y--
	}
	if y < 1 || y > 9999 {
		return time.Time{}, invalid("resulting date outside years 1–9999")
	}
	last := time.Date(y, time.Month(m+2), 0, 0, 0, 0, 0, t.Location()).Day()
	out := time.Date(y, time.Month(m+1), min(t.Day(), last), 0, 0, 0, 0, t.Location()).AddDate(0, 0, days)
	if out.Year() < 1 || out.Year() > 9999 {
		return time.Time{}, invalid("resulting date outside years 1–9999")
	}
	return out, nil
}
func runDate(ctx context.Context, id string, v *toolrun.Values, a []string) (map[string]any, error) {
	loc, e := time.LoadLocation(v.String("timezone", "UTC"))
	if e != nil {
		return nil, invalid("unknown IANA timezone")
	}
	if id == "time.world" {
		instant := time.Now()
		if len(a) > 0 && a[0] != "" {
			instant, e = time.Parse(time.RFC3339, a[0])
			if e != nil {
				return nil, invalid("instant must be RFC3339")
			}
		}
		zones := v.Strings("zones")
		if len(zones) == 0 {
			zones = []string{"UTC", "Asia/Shanghai", "America/New_York", "Europe/London"}
		}
		if len(zones) > 100 {
			return nil, invalid("at most 100 zones")
		}
		rows := []map[string]any{}
		for _, z := range zones {
			l, e := time.LoadLocation(z)
			if e != nil {
				return nil, invalid("unknown IANA timezone: " + z)
			}
			t := instant.In(l)
			name, offset := t.Zone()
			rows = append(rows, map[string]any{"timezone": z, "time": t.Format(time.RFC3339), "abbreviation": name, "offset_seconds": offset})
		}
		return map[string]any{"zones": rows, "unix": instant.Unix(), "clock_changed": false}, nil
	}
	start, e := parseDate(a[0], loc)
	if e != nil {
		return nil, e
	}
	switch id {
	case "date.add":
		years := v.Int("years", 0, -9998, 9998)
		months := v.Int("months", 0, -120000, 120000)
		days := v.Int("days", 0, -3660000, 3660000)
		if v.Err != nil {
			return nil, v.Err
		}
		end, e := clampAdd(start, years, months, days)
		if e != nil {
			return nil, e
		}
		return map[string]any{"date": end.Format("2006-01-02"), "timezone": loc.String(), "month_end": "clamped"}, nil
	case "date.diff", "date.business-days":
		end, e := parseDate(a[1], loc)
		if e != nil {
			return nil, e
		}
		days := civilDays(end) - civilDays(start)
		if id == "date.diff" {
			return map[string]any{"days": days, "elapsed_seconds": end.Unix() - start.Unix(), "timezone": loc.String()}, nil
		}
		if days < 0 || days > 36600 {
			return nil, invalid("business-day range must be forward and at most 36600 days")
		}
		holidays, working := map[string]bool{}, map[string]bool{}
		for _, entry := range []struct {
			key string
			out map[string]bool
		}{{"holidays", holidays}, {"working-dates", working}} {
			list := v.Strings(entry.key)
			if len(list) > 10000 {
				return nil, invalid("date overrides exceed 10000")
			}
			for _, s := range list {
				if _, e := parseDate(s, loc); e != nil {
					return nil, e
				}
				entry.out[s] = true
			}
		}
		count := 0
		for t := start; t.Before(end); t = t.AddDate(0, 0, 1) {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			key := t.Format("2006-01-02")
			if working[key] || (!holidays[key] && t.Weekday() != time.Saturday && t.Weekday() != time.Sunday) {
				count++
			}
		}
		return map[string]any{"business_days": count, "calendar_days": days, "interval": "start inclusive, end exclusive", "policy": "weekends plus caller-supplied overrides"}, nil
	case "date.expiry", "date.age":
		asOf := v.String("as-of", "")
		if asOf == "" {
			asOf = time.Now().In(loc).Format("2006-01-02")
		}
		ref, e := parseDate(asOf, loc)
		if e != nil {
			return nil, e
		}
		if id == "date.expiry" {
			months := v.Int("months", 0, 0, 120000)
			days := v.Int("days", 0, 0, 3660000)
			if v.Err != nil {
				return nil, v.Err
			}
			if months == 0 && days == 0 {
				return nil, invalid("provide positive months or days")
			}
			end, e := clampAdd(start, 0, months, days)
			if e != nil {
				return nil, e
			}
			return map[string]any{"expires_on": end.Format("2006-01-02"), "days_remaining": civilDays(end) - civilDays(ref), "expired": !ref.Before(end), "interval": "production date inclusive; expires_on exclusive"}, nil
		}
		if ref.Before(start) {
			return nil, invalid("birth date is after reference date")
		}
		years := ref.Year() - start.Year()
		birthday, e := clampAdd(start, years, 0, 0)
		if e != nil {
			return nil, e
		}
		if ref.Before(birthday) {
			years--
		}
		next, e := clampAdd(start, ref.Year()-start.Year(), 0, 0)
		if e != nil {
			return nil, e
		}
		if next.Before(ref) {
			next, e = clampAdd(start, ref.Year()-start.Year()+1, 0, 0)
			if e != nil {
				return nil, e
			}
		}
		return map[string]any{"years": years, "days": civilDays(ref) - civilDays(start), "next_birthday": next.Format("2006-01-02"), "leap_day_policy": "February 28 in non-leap years"}, nil
	}
	return nil, invalid(fmt.Sprintf("unknown date operation %s", strings.TrimSpace(id)))
}
