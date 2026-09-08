package localtools

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

var decimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]{1,3})?$`)

func decimal(s string) (*big.Rat, error) {
	if len(s) > 512 || !decimalPattern.MatchString(s) {
		return nil, invalid("expected a finite decimal with at most 512 characters and a three-digit exponent")
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, invalid("invalid decimal")
	}
	return r, nil
}
func finite(s string) (float64, error) {
	if len(s) > 128 {
		return 0, invalid("number too long")
	}
	f, e := strconv.ParseFloat(s, 64)
	if e != nil || math.IsInf(f, 0) || math.IsNaN(f) {
		return 0, invalid("expected finite number")
	}
	return f, nil
}
func numberSpecs() []spec {
	rows := []spec{
		{id: "number.base", summary: "Convert arbitrary precision integers between bases 2 and 36", alias: "进制转换", inputs: in("input"), options: []operation.Parameter{integer("from", "Input base", 10), integer("to", "Output base", 16)}},
		{id: "unit.convert", summary: "Convert length, area, volume, mass, duration and data units exactly", alias: "单位换算 字节数换算 长度转换", inputs: in("input", "from", "to"), options: []operation.Parameter{integer("precision", "Output decimal places", 12)}},
		{id: "temperature.convert", summary: "Convert Celsius, Fahrenheit and Kelvin", alias: "温度转换", inputs: in("input", "from", "to"), options: []operation.Parameter{integer("precision", "Output decimal places", 6)}},
		{id: "number.chinese", summary: "Write decimal numbers or CNY amounts in Chinese", alias: "数字人民币大写", inputs: in("input"), options: []operation.Parameter{boolean("money", "Use financial characters and yuan/jiao/fen", true)}},
		{id: "math.evaluate", summary: "Evaluate bounded arithmetic expressions and math functions locally", alias: "计算器", inputs: in("expression")},
		{id: "math.statistics", summary: "Calculate descriptive statistics for finite numbers", alias: "基础统计", inputs: in("input")},
		{id: "finance.mortgage", summary: "Calculate a fixed-rate amortization schedule", alias: "房贷计算", inputs: in("principal", "annual-rate"), options: []operation.Parameter{integer("months", "Number of monthly repayments", 360), str("method", "annuity or equal-principal", "annuity")}},
		{id: "finance.compound", summary: "Calculate compound growth with periodic contributions", alias: "投资收益计算", inputs: in("principal", "annual-rate"), options: []operation.Parameter{integer("periods", "Total periods", 12), integer("per-year", "Compounding periods per year", 12), str("contribution", "Contribution at end of each period", "0")}},
		{id: "finance.contributions", summary: "Calculate contributions from user-supplied bases, rates and limits", alias: "五险一金计算", inputs: in("salary", "rates"), options: []operation.Parameter{str("floor", "Minimum contribution base", "0"), str("ceiling", "Maximum base; 0 means no ceiling", "0")}},
		{id: "health.bmi", summary: "Calculate BMI from kilograms and centimetres", alias: "BMI计算", inputs: in("weight-kg", "height-cm")},
		{id: "biology.blood-types", summary: "Enumerate possible child ABO blood types under the simplified model", alias: "血型遗传规律", inputs: in("parent-one", "parent-two")},
		{id: "disk.capacity", summary: "Calculate disk capacity units without changing any disk", alias: "硬盘分区容量计算", inputs: in("gib"), options: []operation.Parameter{integer("alignment-kib", "Round byte count up to this alignment", 1024)}},
		{id: "color.convert", summary: "Convert HEX and RGB to RGB, HSL, HSV and simple CMYK", alias: "颜色转换", inputs: in("input")},
	}
	for i := range rows {
		id := rows[i].id
		rows[i].run = func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			return runNumber(ctx, id, v, a)
		}
	}
	return rows
}

type unit struct{ domain, ratio string }

var units = map[string]unit{
	"m": {"length", "1"}, "mm": {"length", "0.001"}, "cm": {"length", "0.01"}, "km": {"length", "1000"}, "in": {"length", "0.0254"}, "ft": {"length", "0.3048"}, "yd": {"length", "0.9144"}, "mi": {"length", "1609.344"}, "nmi": {"length", "1852"},
	"m2": {"area", "1"}, "cm2": {"area", "0.0001"}, "km2": {"area", "1000000"}, "ha": {"area", "10000"}, "acre": {"area", "4046.8564224"}, "ft2": {"area", "0.09290304"},
	"l": {"volume", "1"}, "ml": {"volume", "0.001"}, "m3": {"volume", "1000"}, "us-gal": {"volume", "3.785411784"}, "uk-gal": {"volume", "4.54609"},
	"kg": {"mass", "1"}, "g": {"mass", "0.001"}, "mg": {"mass", "0.000001"}, "t": {"mass", "1000"}, "lb": {"mass", "0.45359237"}, "oz": {"mass", "0.028349523125"}, "jin": {"mass", "0.5"},
	"s": {"duration", "1"}, "ms": {"duration", "0.001"}, "min": {"duration", "60"}, "h": {"duration", "3600"}, "d": {"duration", "86400"}, "week": {"duration", "604800"},
	"bit": {"data", "0.125"}, "B": {"data", "1"}, "KB": {"data", "1000"}, "MB": {"data", "1000000"}, "GB": {"data", "1000000000"}, "TB": {"data", "1000000000000"}, "KiB": {"data", "1024"}, "MiB": {"data", "1048576"}, "GiB": {"data", "1073741824"}, "TiB": {"data", "1099511627776"},
}

func runNumber(ctx context.Context, id string, v *toolrun.Values, a []string) (map[string]any, error) {
	switch id {
	case "number.base":
		from, to := v.Int("from", 10, 2, 36), v.Int("to", 16, 2, 36)
		if v.Err != nil {
			return nil, v.Err
		}
		if len(a[0]) > 65536 {
			return nil, invalid("integer exceeds 65536 digits")
		}
		n, ok := new(big.Int).SetString(a[0], from)
		if !ok {
			return nil, invalid("invalid integer for input base")
		}
		return value(n.Text(to)), nil
	case "unit.convert":
		x, e := decimal(a[0])
		if e != nil {
			return nil, e
		}
		u, ok := units[a[1]]
		w, ok2 := units[a[2]]
		if !ok || !ok2 || u.domain != w.domain {
			return nil, invalid("unknown or incompatible units; data unit names are case-sensitive (B/bit, KB/KiB)")
		}
		ur, _ := decimal(u.ratio)
		wr, _ := decimal(w.ratio)
		x.Mul(x, ur).Quo(x, wr)
		p := v.Int("precision", 12, 0, 30)
		if v.Err != nil {
			return nil, v.Err
		}
		return map[string]any{"value": x.FloatString(p), "exact": x.RatString(), "from": a[1], "to": a[2], "dimension": u.domain}, nil
	case "temperature.convert":
		x, e := decimal(a[0])
		if e != nil {
			return nil, e
		}
		from, to := strings.ToUpper(a[1]), strings.ToUpper(a[2])
		switch from {
		case "C":
			x.Add(x, big.NewRat(27315, 100))
		case "F":
			x.Sub(x, big.NewRat(32, 1)).Mul(x, big.NewRat(5, 9)).Add(x, big.NewRat(27315, 100))
		case "K":
		default:
			return nil, invalid("temperature units must be C, F or K")
		}
		if x.Sign() < 0 {
			return nil, invalid("temperature is below absolute zero")
		}
		switch to {
		case "C":
			x.Sub(x, big.NewRat(27315, 100))
		case "F":
			x.Sub(x, big.NewRat(27315, 100)).Mul(x, big.NewRat(9, 5)).Add(x, big.NewRat(32, 1))
		case "K":
		default:
			return nil, invalid("temperature units must be C, F or K")
		}
		p := v.Int("precision", 6, 0, 30)
		if v.Err != nil {
			return nil, v.Err
		}
		return map[string]any{"value": x.FloatString(p), "unit": to}, nil
	case "number.chinese":
		return chineseNumber(a[0], v.Bool("money", true))
	case "math.evaluate":
		if len(a[0]) > 4096 {
			return nil, invalid("expression exceeds 4096 bytes")
		}
		node, e := parser.ParseExpr(a[0])
		if e != nil {
			return nil, invalid("invalid arithmetic expression")
		}
		budget := 256
		f, e := evaluate(node, 0, &budget)
		if e != nil {
			return nil, e
		}
		return map[string]any{"value": f, "precision": "float64"}, nil
	case "math.statistics":
		parts := strings.FieldsFunc(a[0], func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t' || r == '\r' })
		if len(parts) == 0 || len(parts) > 100000 {
			return nil, invalid("provide 1 to 100000 numbers")
		}
		mean, m2, lo, hi, sum := 0.0, 0.0, math.Inf(1), math.Inf(-1), 0.0
		for i, p := range parts {
			if i%1024 == 0 {
				if e := ctx.Err(); e != nil {
					return nil, e
				}
			}
			x, e := finite(p)
			if e != nil {
				return nil, e
			}
			delta := x - mean
			mean += delta / float64(i+1)
			m2 += delta * (x - mean)
			lo = math.Min(lo, x)
			hi = math.Max(hi, x)
			sum += x
		}
		if math.IsInf(m2, 0) || math.IsNaN(m2) || math.IsInf(sum, 0) {
			return nil, invalid("statistical overflow")
		}
		return map[string]any{"count": len(parts), "sum": sum, "mean": mean, "min": lo, "max": hi, "population_variance": math.Max(0, m2/float64(len(parts))), "population_stddev": math.Sqrt(math.Max(0, m2/float64(len(parts))))}, nil
	case "finance.mortgage", "finance.compound":
		return finance(ctx, id, v, a)
	case "finance.contributions":
		salary, e := decimal(a[0])
		if e != nil {
			return nil, e
		}
		floor, e := decimal(v.String("floor", "0"))
		if e != nil {
			return nil, e
		}
		ceil, e := decimal(v.String("ceiling", "0"))
		if e != nil {
			return nil, e
		}
		if salary.Sign() < 0 || floor.Sign() < 0 || ceil.Sign() < 0 || (ceil.Sign() > 0 && floor.Cmp(ceil) > 0) {
			return nil, invalid("invalid salary or contribution limits")
		}
		base := new(big.Rat).Set(salary)
		if base.Cmp(floor) < 0 {
			base.Set(floor)
		}
		if ceil.Sign() > 0 && base.Cmp(ceil) > 0 {
			base.Set(ceil)
		}
		j, e := parseJSON(a[1])
		if e != nil {
			return nil, e
		}
		rates, ok := j.(map[string]any)
		if !ok || len(rates) == 0 || len(rates) > 50 {
			return nil, invalid("rates must be an object of 1 to 50 percentage rates")
		}
		total := new(big.Rat)
		items := map[string]string{}
		for k, raw := range rates {
			rate, e := decimal(fmt.Sprint(raw))
			if e != nil || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
				return nil, invalid("rates must be percentages from 0 to 100")
			}
			amount := new(big.Rat).Mul(base, rate)
			amount.Quo(amount, big.NewRat(100, 1))
			items[k] = amount.FloatString(2)
			total.Add(total, amount)
		}
		return map[string]any{"base": base.FloatString(2), "items": items, "total": total.FloatString(2), "policy": "user supplied rates and limits; no city policy database"}, nil
	case "health.bmi":
		w, e := finite(a[0])
		if e != nil {
			return nil, e
		}
		h, e := finite(a[1])
		if e != nil {
			return nil, e
		}
		if w <= 0 || w > 2000 || h <= 0 || h > 400 {
			return nil, invalid("weight must be (0,2000] kg and height (0,400] cm")
		}
		return map[string]any{"bmi": w / math.Pow(h/100, 2), "weight_kg": w, "height_cm": h}, nil
	case "biology.blood-types":
		genes := map[string][]string{"A": {"AA", "AO"}, "B": {"BB", "BO"}, "AB": {"AB"}, "O": {"OO"}}
		one, ok := genes[strings.ToUpper(a[0])]
		two, ok2 := genes[strings.ToUpper(a[1])]
		if !ok || !ok2 {
			return nil, invalid("blood type must be A, B, AB or O")
		}
		seen := map[string]bool{}
		for _, g := range one {
			for _, h := range two {
				for _, x := range g {
					for _, y := range h {
						t := "O"
						if x == 'A' || y == 'A' {
							t = "A"
						}
						if x == 'B' || y == 'B' {
							if t == "A" {
								t = "AB"
							} else {
								t = "B"
							}
						}
						seen[t] = true
					}
				}
			}
		}
		types := []string{}
		for _, t := range []string{"A", "B", "AB", "O"} {
			if seen[t] {
				types = append(types, t)
			}
		}
		return map[string]any{"possible": types, "model": "simplified ABO; excludes rare variants; not a paternity test"}, nil
	case "disk.capacity":
		n, e := decimal(a[0])
		if e != nil {
			return nil, e
		}
		align := v.Int("alignment-kib", 1024, 1, 1048576)
		if v.Err != nil {
			return nil, v.Err
		}
		if n.Sign() <= 0 {
			return nil, invalid("capacity must be positive")
		}
		bytes := new(big.Rat).Mul(n, big.NewRat(1<<30, 1))
		boundary := big.NewInt(int64(align) * 1024)
		q := new(big.Int).Quo(bytes.Num(), bytes.Denom())
		if new(big.Rat).SetInt(q).Cmp(bytes) < 0 {
			q.Add(q, big.NewInt(1))
		}
		q.Add(q, new(big.Int).Sub(boundary, big.NewInt(1))).Quo(q, boundary).Mul(q, boundary)
		return map[string]any{"bytes": q.String(), "mib": new(big.Rat).SetFrac(q, big.NewInt(1<<20)).FloatString(6), "decimal_gb": new(big.Rat).SetFrac(q, big.NewInt(1000000000)).FloatString(6), "alignment_kib": align, "disk_modified": false}, nil
	case "color.convert":
		return convertColor(a[0])
	}
	return nil, invalid("unknown numeric operation")
}
func evaluate(node ast.Expr, depth int, budget *int) (float64, error) {
	*budget--
	if depth > 32 || *budget < 0 {
		return 0, invalid("expression exceeds 32 levels or 256 nodes")
	}
	var f float64
	var e error
	switch n := node.(type) {
	case *ast.BasicLit:
		if n.Kind != token.INT && n.Kind != token.FLOAT {
			return 0, invalid("only numeric literals allowed")
		}
		f, e = finite(n.Value)
	case *ast.ParenExpr:
		f, e = evaluate(n.X, depth+1, budget)
	case *ast.Ident:
		switch n.Name {
		case "pi":
			f = math.Pi
		case "e":
			f = math.E
		default:
			e = invalid("unknown constant")
		}
	case *ast.UnaryExpr:
		f, e = evaluate(n.X, depth+1, budget)
		if n.Op == token.SUB {
			f = -f
		} else if n.Op != token.ADD {
			e = invalid("unsupported unary operator")
		}
	case *ast.BinaryExpr:
		var l, r float64
		l, e = evaluate(n.X, depth+1, budget)
		if e != nil {
			return 0, e
		}
		r, e = evaluate(n.Y, depth+1, budget)
		if e != nil {
			return 0, e
		}
		switch n.Op {
		case token.ADD:
			f = l + r
		case token.SUB:
			f = l - r
		case token.MUL:
			f = l * r
		case token.QUO:
			if r == 0 {
				return 0, invalid("division by zero")
			}
			f = l / r
		case token.REM:
			if r == 0 {
				return 0, invalid("modulo by zero")
			}
			f = math.Mod(l, r)
		default:
			return 0, invalid("unsupported operator; use pow(x,y) for exponentiation")
		}
	case *ast.CallExpr:
		name, ok := n.Fun.(*ast.Ident)
		if !ok || len(n.Args) < 1 || len(n.Args) > 2 {
			return 0, invalid("invalid function call")
		}
		args := []float64{}
		for _, a := range n.Args {
			x, e := evaluate(a, depth+1, budget)
			if e != nil {
				return 0, e
			}
			args = append(args, x)
		}
		unary := map[string]func(float64) float64{"sqrt": math.Sqrt, "abs": math.Abs, "sin": math.Sin, "cos": math.Cos, "tan": math.Tan, "log": math.Log, "log10": math.Log10, "exp": math.Exp, "floor": math.Floor, "ceil": math.Ceil, "round": math.Round}
		if fn, ok := unary[name.Name]; ok && len(args) == 1 {
			f = fn(args[0])
		} else if len(args) == 2 {
			switch name.Name {
			case "pow":
				f = math.Pow(args[0], args[1])
			case "min":
				f = math.Min(args[0], args[1])
			case "max":
				f = math.Max(args[0], args[1])
			default:
				e = invalid("unsupported function")
			}
		} else {
			e = invalid("unsupported function or argument count")
		}
	default:
		e = invalid("expression only permits arithmetic and named math functions")
	}
	if e != nil {
		return 0, e
	}
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return 0, invalid("non-finite arithmetic result")
	}
	return f, nil
}
func finance(ctx context.Context, id string, v *toolrun.Values, a []string) (map[string]any, error) {
	p, e := finite(a[0])
	if e != nil {
		return nil, e
	}
	annual, e := finite(a[1])
	if e != nil {
		return nil, e
	}
	if p < 0 || p > 1e15 || annual < 0 || annual > 1000 {
		return nil, invalid("principal must be [0,1e15], annual percentage [0,1000]")
	}
	if id == "finance.compound" {
		n := v.Int("periods", 12, 1, 12000)
		per := v.Int("per-year", 12, 1, 366)
		c, e := finite(v.String("contribution", "0"))
		if e != nil {
			return nil, e
		}
		if c < 0 || c > 1e15 {
			return nil, invalid("contribution must be [0,1e15]")
		}
		if v.Err != nil {
			return nil, v.Err
		}
		balance := p
		for i := 0; i < n; i++ {
			balance = balance*(1+annual/100/float64(per)) + c
			if math.IsInf(balance, 0) || balance > 1e100 {
				return nil, invalid("compound result overflows supported range")
			}
		}
		return map[string]any{"balance": fmt.Sprintf("%.2f", balance), "contributed": fmt.Sprintf("%.2f", p+c*float64(n)), "interest": fmt.Sprintf("%.2f", balance-p-c*float64(n)), "assumption": "constant supplied rate; contribution at period end"}, nil
	}
	n := v.Int("months", 360, 1, 1200)
	method := v.Enum("method", "annuity", "annuity", "equal-principal")
	if v.Err != nil {
		return nil, v.Err
	}
	rate := annual / 1200
	payment := p / float64(n)
	if rate > 0 {
		payment = p * rate / (-math.Expm1(-float64(n) * math.Log1p(rate)))
	}
	balance, total := p, 0.0
	schedule := []map[string]any{}
	for i := 1; i <= n; i++ {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		interest := balance * rate
		principal := payment - interest
		if method == "equal-principal" {
			principal = p / float64(n)
		}
		if i == n {
			principal = balance
		}
		balance = math.Max(0, balance-principal)
		total += interest
		schedule = append(schedule, map[string]any{"month": i, "principal": fmt.Sprintf("%.2f", principal), "interest": fmt.Sprintf("%.2f", interest), "payment": fmt.Sprintf("%.2f", principal+interest), "balance": fmt.Sprintf("%.2f", balance)})
	}
	return map[string]any{"schedule": schedule, "total_interest": fmt.Sprintf("%.2f", total), "total_payment": fmt.Sprintf("%.2f", p+total), "rounding": "unrounded internal calculation; displayed amounts rounded to cents"}, nil
}
func chineseNumber(s string, money bool) (map[string]any, error) {
	if !regexp.MustCompile(`^[+-]?[0-9]{1,16}(?:\.[0-9]{1,8})?$`).MatchString(s) {
		return nil, invalid("expected at most 16 integer digits and 8 decimal places")
	}
	negative := strings.HasPrefix(s, "-")
	s = strings.TrimLeft(s, "+-")
	parts := strings.Split(s, ".")
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if money && len(fraction) > 2 {
		return nil, invalid("money accepts at most two decimal places; round explicitly first")
	}
	n, _ := strconv.ParseUint(parts[0], 10, 64)
	digits := []rune("零一二三四五六七八九")
	small := []string{"", "十", "百", "千"}
	if money {
		digits = []rune("零壹贰叁肆伍陆柒捌玖")
		small = []string{"", "拾", "佰", "仟"}
	}
	group := func(x uint64) string {
		out := ""
		zero := false
		for power := 3; power >= 0; power-- {
			div := uint64(math.Pow10(power))
			d := x / div
			x %= div
			if d == 0 {
				if out != "" && x > 0 {
					zero = true
				}
				continue
			}
			if zero {
				out += string(digits[0])
				zero = false
			}
			out += string(digits[d]) + small[power]
		}
		return out
	}
	out := ""
	large := []string{"", "万", "亿", "兆"}
	zero := false
	for i := 3; i >= 0; i-- {
		div := uint64(math.Pow10(i * 4))
		g := n / div % 10000
		if g == 0 {
			if out != "" {
				zero = true
			}
			continue
		}
		if out != "" && (zero || g < 1000) {
			out += string(digits[0])
		}
		out += group(g) + large[i]
		zero = false
	}
	if out == "" {
		out = string(digits[0])
	}
	if !money && strings.HasPrefix(out, "一十") {
		out = strings.TrimPrefix(out, "一")
	}
	if money {
		out += "元"
		fraction = (fraction + "00")[:2]
		j, f := fraction[0]-'0', fraction[1]-'0'
		if j == 0 && f == 0 {
			out += "整"
		} else {
			if j > 0 {
				out += string(digits[j]) + "角"
			}
			if f > 0 {
				if j == 0 {
					out += "零"
				}
				out += string(digits[f]) + "分"
			}
		}
	} else if fraction != "" {
		out += "点"
		for _, r := range fraction {
			out += string(digits[r-'0'])
		}
	}
	if negative && (n > 0 || strings.Trim(fraction, "0") != "") {
		out = "负" + out
	}
	return value(out), nil
}
func convertColor(s string) (map[string]any, error) {
	var rgb [3]float64
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") {
		h := s[1:]
		if len(h) == 3 {
			h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		}
		if len(h) != 6 {
			return nil, invalid("HEX needs 3 or 6 digits")
		}
		n, e := strconv.ParseUint(h, 16, 24)
		if e != nil {
			return nil, invalid("invalid HEX color")
		}
		rgb = [3]float64{float64(n >> 16), float64(n >> 8 & 255), float64(n & 255)}
	} else {
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(s, "rgb("), ")"), ",")
		if len(parts) != 3 {
			return nil, invalid("color must be #RGB, #RRGGBB or r,g,b")
		}
		for i, p := range parts {
			n, e := finite(strings.TrimSpace(p))
			if e != nil || n < 0 || n > 255 || math.Trunc(n) != n {
				return nil, invalid("RGB channels must be integers in [0,255]")
			}
			rgb[i] = n
		}
	}
	r, g, b := rgb[0]/255, rgb[1]/255, rgb[2]/255
	hi, lo := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	d := hi - lo
	h, sat, sv := 0.0, 0.0, 0.0
	l := (hi + lo) / 2
	if d != 0 {
		switch hi {
		case r:
			h = math.Mod((g-b)/d, 6)
		case g:
			h = (b-r)/d + 2
		default:
			h = (r-g)/d + 4
		}
		h *= 60
		if h < 0 {
			h += 360
		}
		sat = d / (1 - math.Abs(2*l-1))
		sv = d / hi
	}
	k := 1 - hi
	c, m, y := 0.0, 0.0, 0.0
	if hi > 0 {
		c = (hi - r) / hi
		m = (hi - g) / hi
		y = (hi - b) / hi
	}
	return map[string]any{"hex": fmt.Sprintf("#%02X%02X%02X", int(rgb[0]), int(rgb[1]), int(rgb[2])), "rgb": rgb, "hsl": []float64{h, sat * 100, l * 100}, "hsv": []float64{h, sv * 100, hi * 100}, "cmyk": []float64{c * 100, m * 100, y * 100, k * 100}, "cmyk_model": "simple formula without ICC profile"}, nil
}
