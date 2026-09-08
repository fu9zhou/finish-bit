package localtools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"github.com/pelletier/go-toml/v2"
	"github.com/robfig/cron/v3"
)

func extraSpecs() []spec {
	rows := []spec{
		{id: "cron.next", summary: "Validate a cron expression and calculate future occurrences without scheduling jobs", alias: "Cron下次执行时间", inputs: in("expression"), options: []operation.Parameter{str("after", "RFC3339 lower bound; empty means now", ""), str("timezone", "IANA timezone", "UTC"), boolean("seconds", "Use six fields including seconds", false), integer("count", "Number of future times", 10)}},
		{id: "calendar.lunar", summary: "Convert a Gregorian date to the Chinese lunar calendar", alias: "公历转农历", inputs: in("date")},
		{id: "calendar.solar", summary: "Convert a Chinese lunar date to a Gregorian date", alias: "农历转公历", inputs: in("year", "month", "day"), options: []operation.Parameter{boolean("leap", "Month is a leap lunar month", false)}},
		{id: "toml.to-json", summary: "Convert TOML to JSON with timestamps encoded as strings", alias: "TOML转JSON", inputs: in("input")},
		{id: "json.to-toml", summary: "Convert a JSON object to TOML and reject unrepresentable values", alias: "JSON转TOML", inputs: in("input")},
		{id: "toml.validate", summary: "Validate a TOML document", alias: "TOML校验", inputs: in("input")},
		{id: "toml.format", summary: "Re-encode TOML values with normalized layout", alias: "TOML格式化", inputs: in("input")},
	}
	for i := range rows {
		id := rows[i].id
		rows[i].run = func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			return runExtra(ctx, id, v, a)
		}
	}
	return rows
}
func runExtra(ctx context.Context, id string, v *toolrun.Values, a []string) (map[string]any, error) {
	switch id {
	case "cron.next":
		if len(a[0]) > 1024 {
			return nil, invalid("cron expression exceeds 1024 bytes")
		}
		zone := v.String("timezone", "UTC")
		loc, e := time.LoadLocation(zone)
		if e != nil {
			return nil, invalid("unknown IANA timezone")
		}
		after := time.Now().In(loc)
		if s := v.String("after", ""); s != "" {
			after, e = time.Parse(time.RFC3339, s)
			if e != nil {
				return nil, invalid("after must be RFC3339")
			}
			after = after.In(loc)
		}
		fields := cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor
		if v.Bool("seconds", false) {
			fields |= cron.Second
		}
		count := v.Int("count", 10, 1, 1000)
		if v.Err != nil {
			return nil, v.Err
		}
		schedule, e := cron.NewParser(fields).Parse(a[0])
		if e != nil {
			return nil, invalid("invalid cron: " + e.Error())
		}
		out := []string{}
		for i := 0; i < count; i++ {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			after = schedule.Next(after)
			if after.IsZero() {
				return nil, invalid("no future occurrence within the scheduler search horizon")
			}
			out = append(out, after.Format(time.RFC3339))
		}
		return map[string]any{"times": out, "timezone": zone, "jobs_scheduled": false}, nil
	case "calendar.lunar":
		t, e := parseDate(a[0], time.UTC)
		if e != nil {
			return nil, e
		}
		l, e := lunarFromSolar(t)
		if e != nil {
			return nil, e
		}
		return map[string]any{"year": l.year, "month": l.month, "day": l.day, "leap": l.leap, "text": l.text()}, nil
	case "calendar.solar":
		y, e := strconv.Atoi(a[0])
		if e != nil {
			return nil, invalid("invalid lunar year")
		}
		m, e := strconv.Atoi(a[1])
		if e != nil {
			return nil, invalid("invalid lunar month")
		}
		d, e := strconv.Atoi(a[2])
		if e != nil {
			return nil, invalid("invalid lunar day")
		}
		s, e := solarFromLunar(lunarDate{y, m, d, v.Bool("leap", false)})
		if e != nil {
			return nil, e
		}
		return map[string]any{"date": s.Format("2006-01-02")}, nil
	case "toml.to-json", "toml.validate", "toml.format":
		if len(a[0]) > 1<<20 {
			return nil, invalid("TOML input exceeds 1 MiB")
		}
		var data map[string]any
		if e := toml.Unmarshal([]byte(a[0]), &data); e != nil {
			return nil, invalid("invalid TOML: " + e.Error())
		}
		if id == "toml.validate" {
			return map[string]any{"valid": true}, nil
		}
		if id == "toml.format" {
			b, e := toml.Marshal(data)
			return value(string(b)), e
		}
		b, e := json.MarshalIndent(data, "", "  ")
		if e != nil {
			return nil, invalid("TOML includes values not representable in JSON: " + e.Error())
		}
		return value(string(b) + "\n"), nil
	case "json.to-toml":
		x, e := parseJSON(a[0])
		if e != nil {
			return nil, e
		}
		if _, ok := x.(map[string]any); !ok {
			return nil, invalid("TOML root must be an object")
		}
		budget := 100000
		converted, e := tomlValue(x, 0, &budget)
		if e != nil {
			return nil, e
		}
		b, e := toml.Marshal(converted)
		if e != nil {
			return nil, invalid("JSON values cannot be represented in TOML: " + e.Error())
		}
		return value(string(b)), nil
	}
	return nil, invalid("unknown additional utility")
}
func tomlValue(x any, depth int, budget *int) (any, error) {
	*budget--
	if depth > 128 || *budget < 0 {
		return nil, invalid("conversion exceeds nesting/node limit")
	}
	switch x.(type) {
	case nil:
		return nil, invalid("TOML cannot represent JSON null")
	}
	switch v := x.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, item := range v {
			c, e := tomlValue(item, depth+1, budget)
			if e != nil {
				return nil, e
			}
			out[k] = c
		}
		return out, nil
	case []any:
		out := []any{}
		for _, item := range v {
			c, e := tomlValue(item, depth+1, budget)
			if e != nil {
				return nil, e
			}
			out = append(out, c)
		}
		return out, nil
	case json.Number:
		if !strings.ContainsAny(string(v), ".eE") {
			i, e := strconv.ParseInt(string(v), 10, 64)
			if e != nil {
				return nil, invalid("TOML integer is outside signed 64-bit range")
			}
			return i, nil
		}
		f, e := finite(string(v))
		if e != nil {
			return nil, e
		}
		original, e := decimal(string(v))
		if e != nil {
			return nil, e
		}
		roundtrip, e := decimal(strconv.FormatFloat(f, 'g', -1, 64))
		if e != nil || roundtrip.Cmp(original) != 0 {
			return nil, invalid("TOML float conversion would lose decimal precision")
		}
		return f, nil
	case string, bool:
		return v, nil
	default:
		return nil, invalid(fmt.Sprintf("unsupported TOML value %T", x))
	}
}
