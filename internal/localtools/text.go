package localtools

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"html"
	"io"
	"math/big"
	"net/netip"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func textSpecs() []spec {
	rows := []spec{
		{id: "url.parse", summary: "Parse URL components and repeated query parameters", alias: "URL解析", inputs: in("input")},
		{id: "unicode.encode", summary: "Encode Unicode as JSON Unicode escapes", alias: "Unicode编码", inputs: in("input")},
		{id: "unicode.decode", summary: "Decode JSON Unicode escapes", alias: "Unicode解码", inputs: in("input")},
		{id: "html.encode", summary: "Escape HTML text", alias: "HTML转义", inputs: in("input")},
		{id: "html.decode", summary: "Decode HTML entities", alias: "HTML反转义", inputs: in("input")},
		{id: "hex.encode", summary: "Encode UTF-8 text as hexadecimal bytes", alias: "十六进制编码", inputs: in("input")},
		{id: "hex.decode", summary: "Decode hexadecimal bytes to base64 and UTF-8 text when valid", alias: "十六进制解码", inputs: in("input")},
		{id: "text.case", summary: "Convert text letter case", alias: "大小写转换", inputs: in("input"), options: []operation.Parameter{str("mode", "upper, lower, title", "lower")}},
		{id: "text.normalize", summary: "Normalize whitespace, line endings and Unicode width", alias: "文本空白整理", inputs: in("input"), options: []operation.Parameter{str("mode", "space, lines, trim, fullwidth, halfwidth", "space")}},
		{id: "text.reverse", summary: "Reverse Unicode code points", alias: "文字倒序", inputs: in("input")},
		{id: "text.map", summary: "Apply a caller-supplied Unicode character mapping in one pass", alias: "火星文字符映射", inputs: in("input", "mapping")},
		{id: "text.emoticons", summary: "List a small original collection of text faces", alias: "文本颜艺 颜文字"},
		{id: "text.diff", summary: "Compare text lines with an exact bounded LCS", alias: "文本比较", inputs: in("before", "after")},
		{id: "json.diff", summary: "Compare JSON values with JSON Pointer paths and exact numbers", alias: "JSON差异", inputs: in("before", "after")},
		{id: "regex.validate", summary: "Validate a Go RE2 regular expression", alias: "正则校验", inputs: in("pattern")},
		{id: "regex.match", summary: "Find regular expression matches and capture byte offsets", alias: "正则匹配", inputs: in("input", "pattern"), options: []operation.Parameter{integer("limit", "Maximum returned matches", 1000)}},
		{id: "regex.replace", summary: "Replace Go RE2 regular expression matches", alias: "正则替换", inputs: in("input", "pattern", "replacement"), options: []operation.Parameter{boolean("literal", "Treat replacement literally instead of $ capture expansion", false)}},
		{id: "jwt.decode", summary: "Inspect JWT header and payload without verifying the signature", alias: "JWT解析", inputs: in("input")},
		{id: "ip.inspect", summary: "Inspect a local IPv4 or IPv6 address", alias: "IP地址解析", inputs: in("input")},
		{id: "cidr.inspect", summary: "Calculate network bounds and address count", alias: "子网计算", inputs: in("input")},
		{id: "useragent.parse", summary: "Parse common browser and operating system tokens locally", alias: "UA解析", inputs: in("input")},
		{id: "contact.extract", summary: "Extract phone, email and shipment-number candidates", alias: "快递信息提取", inputs: in("input")},
		{id: "text.hide", summary: "Encode a text payload using zero-width characters", alias: "文字隐写", inputs: in("cover", "message")},
		{id: "text.reveal", summary: "Decode a FinishBit zero-width text payload", alias: "文字隐写解码", inputs: in("input")},
		{id: "code.hello", summary: "Generate local Hello World source templates without executing code", alias: "开发语言HelloWorld", inputs: in("language")},
	}
	for i := range rows {
		id := rows[i].id
		rows[i].run = func(ctx context.Context, v *toolrun.Values, a []string) (map[string]any, error) {
			return runText(ctx, id, v, a)
		}
	}
	return rows
}
func runText(ctx context.Context, id string, v *toolrun.Values, a []string) (map[string]any, error) {
	x := ""
	if len(a) > 0 {
		x = a[0]
	}
	switch id {
	case "text.emoticons":
		return map[string]any{"values": []string{"(^_^)", "(T_T)", "(o_o)", "(-_-)", "(>_<)", "(=_=)", "(._.)", "(^o^)/", "(x_x)", "(0_0)"}}, nil
	case "text.map":
		var mapping map[string]string
		if e := json.Unmarshal([]byte(a[1]), &mapping); e != nil || len(mapping) > 10000 {
			return nil, invalid("mapping must be a JSON object with at most 10000 entries")
		}
		for k, v := range mapping {
			if utf8.RuneCountInString(k) != 1 || len(v) > 128 {
				return nil, invalid("map keys must be one code point and values at most 128 bytes")
			}
		}
		var b strings.Builder
		for _, r := range x {
			part, ok := mapping[string(r)]
			if !ok {
				part = string(r)
			}
			if b.Len()+len(part) > 8<<20 {
				return nil, invalid("mapped text exceeds 8 MiB")
			}
			b.WriteString(part)
		}
		return value(b.String()), nil
	case "url.parse":
		u, e := url.Parse(x)
		if e != nil {
			return nil, invalid("invalid URL: " + e.Error())
		}
		q, e := url.ParseQuery(u.RawQuery)
		if e != nil {
			return nil, invalid("invalid query: " + e.Error())
		}
		return map[string]any{"scheme": u.Scheme, "host": u.Host, "hostname": u.Hostname(), "port": u.Port(), "path": u.Path, "escaped_path": u.EscapedPath(), "query": q, "fragment": u.Fragment, "opaque": u.Opaque, "absolute": u.IsAbs()}, nil
	case "unicode.encode":
		var b strings.Builder
		for _, r := range x {
			if r > 0xffff {
				h, l := utf16.EncodeRune(r)
				fmt.Fprintf(&b, "\\u%04x\\u%04x", h, l)
			} else {
				fmt.Fprintf(&b, "\\u%04x", r)
			}
		}
		return value(b.String()), nil
	case "unicode.decode":
		if !validSurrogates(x) {
			return nil, invalid("Unicode escapes contain an unpaired surrogate")
		}
		var out string
		e := json.Unmarshal([]byte("\""+x+"\""), &out)
		if e != nil {
			return nil, invalid("expected a JSON string body with valid escapes")
		}
		return value(out), nil
	case "html.encode":
		return value(html.EscapeString(x)), nil
	case "html.decode":
		return value(html.UnescapeString(x)), nil
	case "hex.encode":
		return value(hex.EncodeToString([]byte(x))), nil
	case "hex.decode":
		b, e := hex.DecodeString(strings.TrimSpace(x))
		if e != nil {
			return nil, invalid("invalid hexadecimal bytes")
		}
		return byteResult(b), nil
	case "text.case":
		m := v.Enum("mode", "lower", "lower", "upper", "title")
		if m == "upper" {
			return value(strings.ToUpper(x)), nil
		}
		if m == "title" {
			var b strings.Builder
			start := true
			for _, r := range x {
				if start {
					r = unicode.ToTitle(r)
				}
				b.WriteRune(r)
				start = unicode.IsSpace(r)
			}
			return value(b.String()), nil
		}
		return value(strings.ToLower(x)), nil
	case "text.normalize":
		switch v.Enum("mode", "space", "space", "lines", "trim", "fullwidth", "halfwidth") {
		case "space":
			x = strings.Join(strings.Fields(x), " ")
		case "lines":
			x = strings.ReplaceAll(strings.ReplaceAll(x, "\r\n", "\n"), "\r", "\n")
		case "trim":
			x = strings.TrimSpace(x)
		case "fullwidth":
			x = strings.Map(func(r rune) rune {
				if r == ' ' {
					return 0x3000
				}
				if r >= 33 && r <= 126 {
					return r + 0xfee0
				}
				return r
			}, x)
		case "halfwidth":
			x = strings.Map(func(r rune) rune {
				if r == 0x3000 {
					return ' '
				}
				if r >= 0xff01 && r <= 0xff5e {
					return r - 0xfee0
				}
				return r
			}, x)
		}
		return value(x), nil
	case "text.reverse":
		r := []rune(x)
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return value(string(r)), nil
	case "text.diff":
		return lineDiff(ctx, x, a[1])
	case "json.diff":
		left, e := parseJSON(x)
		if e != nil {
			return nil, e
		}
		right, e := parseJSON(a[1])
		if e != nil {
			return nil, e
		}
		changes := []map[string]any{}
		e = jsonDiff(ctx, left, right, "", 0, &changes)
		return map[string]any{"equal": len(changes) == 0, "changes": changes}, e
	case "regex.validate", "regex.match", "regex.replace":
		pattern := x
		if id != "regex.validate" {
			pattern = a[1]
		}
		if len(pattern) > 8192 {
			return nil, invalid("regular expression exceeds 8192 bytes")
		}
		re, e := regexp.Compile(pattern)
		if id == "regex.validate" {
			if e != nil {
				return map[string]any{"valid": false, "error": e.Error(), "syntax": "Go RE2"}, nil
			}
			return map[string]any{"valid": true, "syntax": "Go RE2"}, nil
		}
		if e != nil {
			return nil, invalid(e.Error())
		}
		if re.NumSubexp() > 64 {
			return nil, invalid("regular expression exceeds 64 capture groups")
		}
		if id == "regex.replace" {
			literal := v.Bool("literal", false)
			return boundedReplace(ctx, re, x, a[2], literal)
		}
		limit := v.Int("limit", 1000, 1, 10000)
		if v.Err != nil {
			return nil, v.Err
		}
		indexes := re.FindAllStringSubmatchIndex(x, limit+1)
		truncated := len(indexes) > limit
		if truncated {
			indexes = indexes[:limit]
		}
		matches := []map[string]any{}
		resultBytes := 0
		for _, idx := range indexes {
			for j := 0; j < len(idx); j += 2 {
				resultBytes += idx[j+1] - idx[j]
			}
			if resultBytes > 8<<20 {
				return nil, invalid("matched text and captures exceed 8 MiB")
			}
			groups := []any{}
			for j := 2; j < len(idx); j += 2 {
				if idx[j] < 0 {
					groups = append(groups, nil)
				} else {
					groups = append(groups, x[idx[j]:idx[j+1]])
				}
			}
			matches = append(matches, map[string]any{"text": x[idx[0]:idx[1]], "start_byte": idx[0], "end_byte": idx[1], "groups": groups, "indexes": idx})
		}
		return map[string]any{"matches": matches, "group_names": re.SubexpNames()[1:], "truncated": truncated}, nil
	case "jwt.decode":
		parts := strings.Split(x, ".")
		if len(parts) != 3 {
			return nil, invalid("expected three JWT segments")
		}
		data := map[string]any{"verified": false, "signature": parts[2]}
		for i, k := range []string{"header", "payload"} {
			b, e := base64.RawURLEncoding.DecodeString(parts[i])
			if e != nil {
				return nil, invalid("invalid JWT base64url")
			}
			j, e := parseJSON(string(b))
			if e != nil {
				return nil, e
			}
			data[k] = j
		}
		return data, nil
	case "ip.inspect":
		p, e := netip.ParseAddr(x)
		if e != nil {
			return nil, invalid("invalid IP address")
		}
		return map[string]any{"address": p.String(), "ipv4": p.Is4(), "private": p.IsPrivate(), "loopback": p.IsLoopback(), "multicast": p.IsMulticast(), "unspecified": p.IsUnspecified(), "global_unicast": p.IsGlobalUnicast()}, nil
	case "cidr.inspect":
		p, e := netip.ParsePrefix(x)
		if e != nil {
			return nil, invalid("invalid CIDR")
		}
		p = p.Masked()
		bits := p.Addr().BitLen() - p.Bits()
		count := new(big.Int).Lsh(big.NewInt(1), uint(bits))
		first := p.Addr().AsSlice()
		last := new(big.Int).SetBytes(first)
		last.Add(last, new(big.Int).Sub(count, big.NewInt(1)))
		b := last.FillBytes(make([]byte, len(first)))
		ip, _ := netip.AddrFromSlice(b)
		return map[string]any{"network": p.String(), "first": p.Addr().String(), "last": ip.String(), "addresses": count.String()}, nil
	case "useragent.parse":
		if len(x) > 8192 {
			return nil, invalid("UA exceeds 8192 bytes")
		}
		browser, version := "unknown", ""
		for _, p := range []struct{ name, token string }{{"Edge", "Edg/"}, {"Opera", "OPR/"}, {"Chrome", "Chrome/"}, {"Firefox", "Firefox/"}, {"Safari", "Version/"}} {
			if i := strings.Index(x, p.token); i >= 0 {
				browser = p.name
				if fields := strings.Fields(x[i+len(p.token):]); len(fields) > 0 {
					version = fields[0]
				}
				break
			}
		}
		system := "unknown"
		for _, s := range []string{"Android", "iPhone", "iPad", "Windows", "Macintosh", "Linux"} {
			if strings.Contains(x, s) {
				system = s
				break
			}
		}
		return map[string]any{"browser": browser, "version": version, "system": system, "mobile": strings.Contains(x, "Mobile") || strings.Contains(x, "Android"), "raw": x, "coverage": "common tokens; no fingerprint or identity verification"}, nil
	case "contact.extract":
		out := map[string]any{"ambiguous": true}
		for k, p := range map[string]string{"emails": `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`, "phones": `(?:\+?86[ -]?)?1[3-9][0-9]{9}`, "shipment_candidates": `\b[A-Za-z]{0,4}[0-9]{10,20}\b`} {
			re := regexp.MustCompile(p)
			out[k] = re.FindAllString(x, 1000)
		}
		return out, nil
	case "text.hide":
		if len(a[1]) > 65536 {
			return nil, invalid("hidden message exceeds 64 KiB")
		}
		if strings.ContainsAny(x, "\u200b\u200c\u2063") {
			return nil, invalid("cover contains reserved zero-width characters")
		}
		var b strings.Builder
		b.WriteString(x)
		b.WriteRune('\u2063')
		for _, c := range []byte(fmt.Sprintf("FNSH1:%08x:%s", crc32.ChecksumIEEE([]byte(a[1])), a[1])) {
			for bit := 7; bit >= 0; bit-- {
				if c&(1<<bit) != 0 {
					b.WriteRune('\u200c')
				} else {
					b.WriteRune('\u200b')
				}
			}
		}
		b.WriteRune('\u2063')
		return value(b.String()), nil
	case "text.reveal":
		parts := strings.Split(x, "\u2063")
		if len(parts) != 3 {
			return nil, invalid("expected one FinishBit hidden payload")
		}
		r := []rune(parts[1])
		if len(r)%8 != 0 || len(r) > ((65536+15)*8) {
			return nil, invalid("invalid hidden payload length")
		}
		b := make([]byte, len(r)/8)
		for i, c := range r {
			if c != '\u200b' && c != '\u200c' {
				return nil, invalid("invalid hidden payload character")
			}
			if c == '\u200c' {
				b[i/8] |= 1 << (7 - i%8)
			}
		}
		if len(b) < 15 || !strings.HasPrefix(string(b), "FNSH1:") || b[14] != ':' {
			return nil, invalid("unknown hidden payload format")
		}
		checksum, err := strconv.ParseUint(string(b[6:14]), 16, 32)
		if err != nil || uint32(checksum) != crc32.ChecksumIEEE(b[15:]) || !utf8.Valid(b[15:]) {
			return nil, invalid("hidden payload checksum or UTF-8 validation failed")
		}
		return value(string(b[15:])), nil
	case "code.hello":
		templates := map[string]string{"go": "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"Hello, World!\") }\n", "python": "print(\"Hello, World!\")\n", "javascript": "console.log(\"Hello, World!\");\n", "typescript": "console.log(\"Hello, World!\");\n", "rust": "fn main() { println!(\"Hello, World!\"); }\n", "c": "#include <stdio.h>\nint main(void) { puts(\"Hello, World!\"); return 0; }\n", "cpp": "#include <iostream>\nint main() { std::cout << \"Hello, World!\\n\"; }\n", "java": "class Main { public static void main(String[] args) { System.out.println(\"Hello, World!\"); } }\n", "bash": "printf '%s\\n' 'Hello, World!'\n", "powershell": "Write-Output 'Hello, World!'\n", "ruby": "puts 'Hello, World!'\n", "php": "<?php echo \"Hello, World!\\n\";\n"}
		s, ok := templates[strings.ToLower(x)]
		if !ok {
			return nil, invalid("unsupported language; go/python/javascript/typescript/rust/c/cpp/java/bash/powershell/ruby/php")
		}
		return value(s), nil
	}
	return nil, invalid("unknown text operation")
}
func byteResult(b []byte) map[string]any {
	out := map[string]any{"base64": base64.StdEncoding.EncodeToString(b), "bytes": len(b)}
	if utf8.Valid(b) {
		out["text"] = string(b)
	}
	return out
}
func parseJSON(s string) (any, error) {
	d := json.NewDecoder(strings.NewReader(s))
	d.UseNumber()
	var x any
	if e := d.Decode(&x); e != nil {
		return nil, invalid("invalid JSON: " + e.Error())
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return nil, invalid("JSON must contain exactly one value")
	}
	return x, nil
}
func pointer(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1") }
func jsonDiff(ctx context.Context, a, b any, path string, depth int, out *[]map[string]any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 128 || len(*out) > 10000 {
		return invalid("JSON diff exceeds depth 128 or 10000 changes")
	}
	add := func(kind string, x, y any) {
		*out = append(*out, map[string]any{"kind": kind, "path": path, "before": x, "after": y})
	}
	if am, ok := a.(map[string]any); ok {
		if bm, ok := b.(map[string]any); ok {
			keys := map[string]bool{}
			for k := range am {
				keys[k] = true
			}
			for k := range bm {
				keys[k] = true
			}
			order := []string{}
			for k := range keys {
				order = append(order, k)
			}
			sort.Strings(order)
			for _, k := range order {
				av, ao := am[k]
				bv, bo := bm[k]
				p := path + "/" + pointer(k)
				if !ao || !bo {
					kind := "add"
					if !bo {
						kind = "remove"
					}
					*out = append(*out, map[string]any{"kind": kind, "path": p, "before": av, "after": bv})
					if len(*out) > 10000 {
						return invalid("JSON diff exceeds 10000 changes")
					}
				} else if e := jsonDiff(ctx, av, bv, p, depth+1, out); e != nil {
					return e
				}
			}
			return nil
		}
	}
	if aa, ok := a.([]any); ok {
		if ba, ok := b.([]any); ok {
			n := max(len(aa), len(ba))
			for i := 0; i < n; i++ {
				p := path + "/" + strconv.Itoa(i)
				if i >= len(aa) {
					*out = append(*out, map[string]any{"kind": "add", "path": p, "after": ba[i]})
				} else if i >= len(ba) {
					*out = append(*out, map[string]any{"kind": "remove", "path": p, "before": aa[i]})
				} else if e := jsonDiff(ctx, aa[i], ba[i], p, depth+1, out); e != nil {
					return e
				}
				if len(*out) > 10000 {
					return invalid("JSON diff exceeds 10000 changes")
				}
			}
			return nil
		}
	}
	if an, ok := a.(json.Number); ok {
		if bn, ok := b.(json.Number); ok {
			af, e := decimal(string(an))
			bf, f := decimal(string(bn))
			if e != nil || f != nil {
				return invalid("JSON numbers must be bounded decimals (512 characters, exponent within 999)")
			}
			if af.Cmp(bf) == 0 {
				return nil
			}
		}
	}
	ab, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	if string(ab) != string(bb) {
		add("replace", a, b)
	}
	return nil
}
func lineDiff(ctx context.Context, a, b string) (map[string]any, error) {
	left, right := strings.Split(a, "\n"), strings.Split(b, "\n")
	if len(left) > 2000 || len(right) > 2000 {
		return nil, invalid("text diff supports at most 2000 lines per input")
	}
	w := len(right) + 1
	dp := make([]uint16, (len(left)+1)*w)
	for i := len(left) - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for j := len(right) - 1; j >= 0; j-- {
			if left[i] == right[j] {
				dp[i*w+j] = dp[(i+1)*w+j+1] + 1
			} else {
				dp[i*w+j] = max(dp[(i+1)*w+j], dp[i*w+j+1])
			}
		}
	}
	edits := []map[string]any{}
	i, j := 0, 0
	for i < len(left) || j < len(right) {
		if i < len(left) && j < len(right) && left[i] == right[j] {
			edits = append(edits, map[string]any{"kind": "equal", "text": left[i], "before_line": i + 1, "after_line": j + 1})
			i++
			j++
		} else if j < len(right) && (i == len(left) || dp[i*w+j+1] >= dp[(i+1)*w+j]) {
			edits = append(edits, map[string]any{"kind": "add", "text": right[j], "after_line": j + 1})
			j++
		} else {
			edits = append(edits, map[string]any{"kind": "remove", "text": left[i], "before_line": i + 1})
			i++
		}
	}
	return map[string]any{"equal": a == b, "lines": edits}, nil
}

func escapeHTML(s string) string { return html.EscapeString(s) }

// Check escapes before encoding/json replaces malformed UTF-16 with U+FFFD.
func validSurrogates(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			continue
		}
		i++
		if i >= len(s) || s[i] != 'u' {
			continue
		}
		if i+4 >= len(s) {
			return false
		}
		n, err := strconv.ParseUint(s[i+1:i+5], 16, 16)
		if err != nil {
			return false
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return false
		}
		if n < 0xd800 || n > 0xdbff {
			continue
		}
		if i+6 >= len(s) || s[i+1:i+3] != `\u` {
			return false
		}
		low, err := strconv.ParseUint(s[i+3:i+7], 16, 16)
		if err != nil || low < 0xdc00 || low > 0xdfff {
			return false
		}
		i += 6
	}
	return true
}

func boundedReplace(ctx context.Context, re *regexp.Regexp, input, replacement string, literal bool) (map[string]any, error) {
	const budget = 8 << 20
	if len(replacement) > 65536 {
		return nil, invalid("replacement exceeds 65536 bytes")
	}
	indexes := re.FindAllStringSubmatchIndex(input, 10001)
	if len(indexes) > 10000 {
		return nil, invalid("replacement exceeds 10000 matches")
	}
	var out strings.Builder
	last := 0
	for _, idx := range indexes {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// Each dollar token can expand at most to the longest capture.
		longest := 0
		for j := 0; j < len(idx); j += 2 {
			longest = max(longest, idx[j+1]-idx[j])
		}
		if !literal && len(replacement)+strings.Count(replacement, "$")*longest > budget {
			return nil, invalid("capture expansion exceeds output budget")
		}
		part := replacement
		if !literal {
			part = string(re.ExpandString(nil, replacement, input, idx))
		}
		if out.Len()+idx[0]-last+len(part) > budget {
			return nil, invalid("replacement output exceeds 8 MiB")
		}
		out.WriteString(input[last:idx[0]])
		out.WriteString(part)
		last = idx[1]
	}
	if out.Len()+len(input)-last > budget {
		return nil, invalid("replacement output exceeds 8 MiB")
	}
	out.WriteString(input[last:])
	return value(out.String()), nil
}
