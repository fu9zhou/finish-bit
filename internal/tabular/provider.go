// Package tabular exposes bounded local table workflows through managed qsv.
package tabular

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fu9zhou/finish-bit/internal/toolrun"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type spec struct {
	id, summary, alias, mode string
	inputs                   int
	options                  []operation.Parameter
}

func str(n, d, v string) operation.Parameter { return toolrun.Option(n, operation.TypeString, d, v) }
func integer(n, d string, v int) operation.Parameter {
	return toolrun.Option(n, operation.TypeInteger, d, v)
}
func boolean(n, d string, v bool) operation.Parameter {
	return toolrun.Option(n, operation.TypeBoolean, d, v)
}
func names(n, d string, required bool) operation.Parameter {
	return operation.Parameter{Name: n, Description: d, Type: operation.TypeStrings, Required: required}
}
func columns() operation.Parameter {
	return names("columns", "Exact column names in order; repeat option. Empty means all where optional", false)
}
func catalog() []spec {
	return []spec{
		{"csv.select", "Select and reorder CSV columns", "选择表格列", "csv", 1, []operation.Parameter{names("columns", "Exact column names in output order", true)}},
		{"csv.rename", "Replace the ordered CSV header row", "重命名表格列", "csv", 1, []operation.Parameter{names("names", "All replacement headers, matching input column count", true)}},
		{"csv.filter", "Filter CSV rows by selected cell contents", "筛选表格行", "csv", 1, []operation.Parameter{columns(), toolrun.Param("pattern", "Match text or Rust regular expression", true), str("mode", "contains, exact or regex", "contains"), boolean("invert", "Keep nonmatching rows", false), boolean("ignore-case", "Ignore case", false)}},
		{"csv.sort", "Sort CSV rows by selected keys", "表格排序", "csv", 1, []operation.Parameter{columns(), str("mode", "text, numeric or natural", "text"), boolean("reverse", "Descending order", false), boolean("ignore-case", "Ignore case in text sorting", false)}},
		{"csv.deduplicate", "Deduplicate CSV keys in sorted key order", "按键去重表格", "csv", 1, []operation.Parameter{columns(), str("keep", "Keep first or last input occurrence", "first"), boolean("ignore-case", "Ignore case in keys", false)}},
		{"csv.concat", "Combine CSV files by rows, aligned headers or columns", "合并表格", "csv", 1, []operation.Parameter{names("files", "Additional CSV files in order", true), str("mode", "rows, rowskey or columns", "rows"), boolean("pad", "Pad shorter column inputs", false)}},
		{"csv.join", "Join two CSV tables on explicit keys", "关联表格", "csv", 2, []operation.Parameter{names("left-keys", "Exact key names in first table", true), names("right-keys", "Exact key names in second table", true), str("kind", "inner, left, right, full, left-anti, left-semi, right-anti or right-semi", "inner"), boolean("ignore-case", "Ignore case in keys", false), boolean("nulls", "Match empty keys", false)}},
		{"csv.diff", "Compare tables with unique keys", "比较表格差异", "csv", 2, []operation.Parameter{names("keys", "Exact unique key names; same headers required", true)}},
		{"csv.frequency", "Count occurrences of selected column values", "表格频次统计", "csv", 1, []operation.Parameter{columns(), integer("limit", "Maximum values per column", 10), boolean("ignore-case", "Ignore case", false)}},
		{"csv.stats", "Compute column types, numeric and quality statistics", "表格统计摘要", "csv", 1, []operation.Parameter{columns(), boolean("extended", "Include median, quartiles and cardinality", true)}},
		{"csv.sample", "Sample a fixed number of CSV rows with a seed", "表格抽样", "csv", 1, []operation.Parameter{integer("count", "Number of data rows", 10), integer("seed", "Deterministic RNG seed", 42)}},
		{"csv.shuffle", "Shuffle CSV data rows with a seed", "打乱表格行", "csv", 1, []operation.Parameter{integer("seed", "Deterministic RNG seed", 42)}},
		{"csv.slice", "Select a contiguous range of CSV data rows", "截取表格行", "csv", 1, []operation.Parameter{integer("start", "Zero-based data row offset", 0), integer("count", "Number of rows", 100)}},
		{"csv.reverse", "Reverse CSV data row order", "反转表格行", "csv", 1, nil},
		{"csv.transpose", "Transpose all CSV cells including the header", "转置表格", "csv", 1, nil},
		{"csv.melt", "Reshape nonempty selected values into field/attribute/value rows", "表格宽转长", "csv", 1, []operation.Parameter{names("columns", "Exact value columns; other columns form pipe-joined identifiers", true)}},
		{"csv.explode", "Expand separated values in one column into rows", "展开表格多值列", "csv", 1, []operation.Parameter{toolrun.Param("column", "Exact column name", true), str("separator", "Literal nonempty value separator", ";")}},
		{"csv.fill", "Fill missing CSV cell values", "填充表格空值", "csv", 1, []operation.Parameter{names("columns", "Columns to fill", true), str("mode", "previous, first or value", "previous"), str("value", "Replacement for value mode", ""), boolean("backfill", "Fill leading blanks from the first value", false)}},
		{"csv.replace", "Replace selected CSV cell contents", "替换表格内容", "csv", 1, []operation.Parameter{columns(), toolrun.Param("pattern", "Literal text or Rust regular expression", true), str("replacement", "Replacement text; regex mode supports captures", ""), str("mode", "literal, exact or regex", "literal"), boolean("ignore-case", "Ignore case", false)}},
		{"csv.split", "Split a CSV into bounded row chunks", "拆分表格", "directory", 1, []operation.Parameter{integer("rows", "Data rows in each output file", 500)}},
		{"csv.partition", "Partition a CSV into files by one key", "按列分组拆表", "directory", 1, []operation.Parameter{toolrun.Param("column", "Column containing safe filename keys", true)}},
		{"csv.schema", "Infer a JSON Schema from CSV values", "推断表格规则", "json", 1, []operation.Parameter{integer("enum-limit", "Maximum distinct values for enum inference; 0 disables", 0)}},
		{"csv.validate", "Validate CSV structure or a local JSON Schema", "校验表格规则", "validate", 1, []operation.Parameter{str("schema", "Optional local JSON Schema; remote references are rejected", "")}},
		{"csv.format", "Rewrite CSV delimiters, quoting and line endings", "格式化表格文件", "csv", 1, []operation.Parameter{str("out-delimiter", "One output field separator", ","), boolean("crlf", "Use Windows line endings", false), boolean("quote-all", "Quote every value", false)}},
		{"csv.to-jsonl", "Convert CSV rows to JSON Lines with inferred types", "CSV 转 JSONL", "jsonl", 1, []operation.Parameter{boolean("trim", "Trim values before conversion", false), boolean("boolean", "Infer booleans", true)}},
		{"jsonl.to-csv", "Convert JSON Lines objects to CSV", "JSONL 转 CSV", "csv", 1, nil},
		{"csv.transform", "Apply a local text or number transformation to columns", "清洗表格单元格", "csv", 1, []operation.Parameter{names("columns", "Columns to transform", true), str("action", "trim, ltrim, rtrim, upper, lower, squeeze, titlecase, round or currencytonum", "trim"), integer("decimals", "Decimal places for round", 3)}},
	}
}

func Register(registry *operation.Registry, resolver toolrun.Resolver) error {
	for _, item := range catalog() {
		def := operation.Definition{ID: item.id, Summary: item.summary, Description: item.summary + ". Local UTF-8 tables through managed qsv 22.0.1. Exact column names, staged output and bounded input; see docs/tables.md.", Aliases: []string{item.alias}, Tags: []string{"csv", "table", "data"}, Source: "qsv", Options: append([]operation.Parameter{}, item.options...), Requirements: []operation.Requirement{{Package: "qsv"}}}
		for i := 0; i < item.inputs; i++ {
			name := "input"
			if i > 0 {
				name = "second"
			}
			def.Inputs = append(def.Inputs, toolrun.Param(name, "Local data file", true))
		}
		def.Options = append(def.Options, str("delimiter", "One input field separator", ","))
		if item.mode == "directory" {
			def.Options = append(def.Options, toolrun.Param("output", "New output directory", true))
		} else if item.mode != "validate" {
			def.Options = append(def.Options, toolrun.OutputOptions()...)
		}
		if err := registry.Register(operation.Capability{Definition: def, Runner: operation.Func(func(ctx context.Context, r operation.Request) (operation.Result, error) {
			if err := operation.ValidateRequest(def, r); err != nil {
				return operation.Result{}, err
			}
			return run(ctx, resolver, item, r)
		})}); err != nil {
			return err
		}
	}
	return nil
}

const maxInputBytes = 16 << 20
const maxRows = 100000
const maxCells = 1000000

type table struct{ rows [][]string }

func readTable(data []byte, separator rune) (table, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = separator
	var rows [][]string
	cells := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return table{}, toolrun.Invalid("invalid CSV: " + err.Error())
		}
		cells += len(row)
		if len(rows) > maxRows || cells > maxCells {
			return table{}, toolrun.Invalid("table exceeds 100000 rows or 1000000 cells")
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return table{}, toolrun.Invalid("CSV needs a header row")
	}
	seen := map[string]bool{}
	for _, h := range rows[0] {
		if h == "" || seen[h] {
			return table{}, toolrun.Invalid("CSV headers must be nonempty and unique")
		}
		seen[h] = true
	}
	return table{rows}, nil
}
func separator(value string) (rune, error) {
	r := []rune(value)
	if len(r) != 1 || r[0] == '"' || r[0] == '\r' || r[0] == '\n' || r[0] == 0 || r[0] > 127 {
		return 0, toolrun.Invalid("delimiter must be one ASCII non-quote field separator")
	}
	return r[0], nil
}
func selection(t table, values []string) (string, error) {
	if len(values) == 0 {
		return "1-", nil
	}
	result := []string{}
	seen := map[string]bool{}
	for _, name := range values {
		if seen[name] {
			return "", toolrun.Invalid("duplicate selected column: " + name)
		}
		seen[name] = true
		found := false
		for i, h := range t.rows[0] {
			if h == name {
				result = append(result, strconv.Itoa(i+1))
				found = true
				break
			}
		}
		if !found {
			return "", toolrun.Invalid("unknown column: " + name)
		}
	}
	return strings.Join(result, ","), nil
}
func localBytes(path string) ([]byte, error) {
	path, err := toolrun.LocalFile(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxInputBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxInputBytes || !utf8.Valid(b) {
		return nil, toolrun.Invalid("input must be UTF-8 and at most 16 MiB")
	}
	return bytes.TrimPrefix(b, []byte{0xef, 0xbb, 0xbf}), nil
}

func run(ctx context.Context, resolver toolrun.Resolver, item spec, r operation.Request) (operation.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return operation.Result{}, err
	}
	v := &toolrun.Values{Request: r}
	delim := v.String("delimiter", ",")
	sep, err := separator(delim)
	if err != nil {
		return operation.Result{}, err
	}
	work, err := os.MkdirTemp("", "finishbit-table-")
	if err != nil {
		return operation.Result{}, err
	}
	defer os.RemoveAll(work)
	paths := append([]string{}, r.Inputs...)
	if item.id == "csv.concat" {
		paths = append(paths, v.Strings("files")...)
	}
	if len(paths) > 32 {
		return operation.Result{}, toolrun.Invalid("at most 32 input tables")
	}
	var tables []table
	var inputs []string
	total := 0
	for i, path := range paths {
		b, e := localBytes(path)
		if e != nil {
			return operation.Result{}, e
		}
		total += len(b)
		if total > 32<<20 {
			return operation.Result{}, toolrun.Invalid("combined inputs exceed 32 MiB")
		}
		t := table{}
		if item.id != "jsonl.to-csv" && item.id != "csv.validate" {
			t, e = readTable(b, sep)
			if e != nil {
				return operation.Result{}, e
			}
		}
		if item.id == "jsonl.to-csv" {
			if e = checkJSONL(b); e != nil {
				return operation.Result{}, e
			}
		}
		file := fmt.Sprintf("input-%03d.csv", i)
		if item.id == "jsonl.to-csv" {
			file = "input.jsonl"
		}
		if e = os.WriteFile(filepath.Join(work, file), b, 0600); e != nil {
			return operation.Result{}, e
		}
		tables = append(tables, t)
		inputs = append(inputs, file)
	}
	first := inputs[0]
	var prepare func() error
	args := []string{}
	pos := []string{first}
	columns := func(option string, index int) string {
		s, e := selection(tables[index], v.Strings(option))
		v.Check(e == nil, fmt.Sprint(e))
		return s
	}
	flag := func(name string, fallback bool, qsv string) {
		if v.Bool(name, fallback) {
			args = append(args, qsv)
		}
	}
	number := func(name string, fallback, low, high int) string {
		return strconv.Itoa(v.Int(name, fallback, low, high))
	}
	switch item.id {
	case "csv.select":
		args = []string{"select"}
		pos = []string{columns("columns", 0), first}
	case "csv.rename":
		headers := v.Strings("names")
		v.Check(len(headers) == len(tables[0].rows[0]), "replacement header count must match input")
		seen := map[string]bool{}
		for _, h := range headers {
			v.Check(h != "" && !seen[h], "replacement headers must be nonempty and unique")
			seen[h] = true
		}
		var b bytes.Buffer
		w := csv.NewWriter(&b)
		_ = w.Write(headers)
		w.Flush()
		args = []string{"rename"}
		pos = []string{strings.TrimSuffix(b.String(), "\n"), first}
	case "csv.filter", "csv.replace":
		command := "search"
		fallback := "contains"
		if item.id == "csv.replace" {
			command = "replace"
			fallback = "literal"
		}
		args = []string{command, "--select", columns("columns", 0), "--not-one", "--unicode"}
		mode := v.Enum("mode", fallback, fallback, "exact", "regex")
		if mode == fallback {
			args = append(args, "--literal")
		} else if mode == "exact" {
			args = append(args, "--exact")
		}
		flag("ignore-case", false, "--ignore-case")
		pattern := v.String("pattern", "")
		v.Check(len(pattern) <= 4096, "pattern exceeds 4096 bytes")
		pos = []string{pattern}
		if item.id == "csv.replace" {
			replacement := v.String("replacement", "")
			if mode != "regex" {
				replacement = strings.ReplaceAll(replacement, "$", "$$")
			}
			pos = append(pos, replacement)
		} else {
			flag("invert", false, "--invert-match")
		}
		pos = append(pos, first)
	case "csv.sort":
		args = []string{"sort", "--select", columns("columns", 0)}
		mode := v.Enum("mode", "text", "text", "numeric", "natural")
		if mode != "text" {
			args = append(args, "--"+mode)
		}
		flag("reverse", false, "--reverse")
		flag("ignore-case", false, "--ignore-case")
	case "csv.deduplicate":
		args = []string{"dedup", "--select", columns("columns", 0), "--quiet"}
		flag("ignore-case", false, "--ignore-case")
		if v.Enum("keep", "first", "first", "last") == "first" {
			sorting := []string{"sort", "--select", columns("columns", 0), "--delimiter", delim}
			if v.Bool("ignore-case", false) {
				sorting = append(sorting, "--ignore-case")
			}
			sorting = append(sorting, "--", first)
			prepare = func() error {
				b, e := toolrun.RunWithEnv(ctx, resolver, "qsv", "qsv", work, qsvEnv(work), sorting...)
				if e != nil {
					return e
				}
				return os.WriteFile(filepath.Join(work, "sorted.csv"), b, 0600)
			}
			args = append(args, "--sorted")
			pos = []string{"sorted.csv"}
			delim = ","
		}
	case "csv.concat":
		mode := v.Enum("mode", "rows", "rows", "rowskey", "columns")
		args = []string{"cat", mode}
		pos = inputs
		if v.Bool("pad", false) {
			v.Check(mode == "columns", "pad requires columns mode")
			args = append(args, "--pad")
		}
		if mode == "rows" {
			for _, t := range tables[1:] {
				v.Check(equal(t.rows[0], tables[0].rows[0]), "row concatenation requires identical ordered headers")
			}
		}
	case "csv.join":
		left := columns("left-keys", 0)
		right := columns("right-keys", 1)
		v.Check(len(v.Strings("left-keys")) == len(v.Strings("right-keys")), "join key counts differ")
		args = []string{"join"}
		kind := v.Enum("kind", "inner", "inner", "left", "right", "full", "left-anti", "left-semi", "right-anti", "right-semi")
		if kind != "inner" {
			args = append(args, "--"+kind)
		}
		flag("ignore-case", false, "--ignore-case")
		flag("nulls", false, "--nulls")
		pos = []string{left, first, right, inputs[1]}
	case "csv.diff":
		v.Check(equal(tables[0].rows[0], tables[1].rows[0]), "diff requires identical ordered headers")
		sel := columns("keys", 0)
		indexes := strings.Split(sel, ",")
		for i, s := range indexes {
			n, _ := strconv.Atoi(s)
			indexes[i] = strconv.Itoa(n - 1)
		}
		args = []string{"diff", "--key", strings.Join(indexes, ","), "--sort-columns", strings.Join(indexes, ",")}
		pos = inputs
		for _, t := range tables {
			seen := map[string]bool{}
			for _, row := range t.rows[1:] {
				var key []string
				for _, s := range indexes {
					n, _ := strconv.Atoi(s)
					if n < 0 || n >= len(row) {
						continue
					}
					key = append(key, row[n])
				}
				b, _ := json.Marshal(key)
				v.Check(!seen[string(b)], "diff keys must uniquely identify rows")
				seen[string(b)] = true
			}
		}
	case "csv.frequency":
		args = []string{"frequency", "--select", columns("columns", 0), "--limit", number("limit", 10, 1, 1000), "--no-trim", "--no-other", "--unq-limit", "0"}
		flag("ignore-case", false, "--ignore-case")
	case "csv.stats":
		args = []string{"stats", "--select", columns("columns", 0), "--cache-threshold", "0"}
		if v.Bool("extended", true) {
			args = append(args, "--median", "--quartiles", "--cardinality")
		}
	case "csv.sample":
		count := v.Int("count", 10, 1, maxRows)
		v.Check(count <= len(tables[0].rows)-1, "sample count exceeds input rows")
		args = []string{"sample", "--seed", number("seed", 42, 0, 2147483647)}
		pos = []string{strconv.Itoa(count), first}
	case "csv.shuffle":
		args = []string{"sort", "--random", "--seed", number("seed", 42, 0, 2147483647)}
	case "csv.slice":
		args = []string{"slice", "--start", number("start", 0, 0, maxRows), "--len", number("count", 100, 1, maxRows)}
	case "csv.reverse":
		args = []string{"reverse"}
	case "csv.transpose":
		args = []string{"transpose"}
	case "csv.melt":
		_ = columns("columns", 0)
		values := map[string]bool{}
		for _, name := range v.Strings("columns") {
			values[name] = true
		}
		var identifiers []string
		for _, name := range tables[0].rows[0] {
			if !values[name] {
				identifiers = append(identifiers, name)
			}
		}
		v.Check(len(identifiers) > 0, "melt needs at least one identifier column")
		selected, e := selection(tables[0], identifiers)
		v.Check(e == nil, fmt.Sprint(e))
		args = []string{"transpose", "--long", selected}
	case "csv.explode":
		sel, e := selection(tables[0], []string{v.String("column", "")})
		v.Check(e == nil, fmt.Sprint(e))
		value := v.String("separator", ";")
		v.Check(value != "" && len(value) <= 64, "separator must contain 1 to 64 bytes")
		args = []string{"explode"}
		pos = []string{sel, value, first}
	case "csv.fill":
		args = []string{"fill"}
		mode := v.Enum("mode", "previous", "previous", "first", "value")
		if mode == "first" {
			args = append(args, "--first")
		} else if mode == "value" {
			args = append(args, "--default", v.String("value", ""))
		}
		flag("backfill", false, "--backfill")
		pos = []string{columns("columns", 0), first}
	case "csv.split":
		size := v.Int("rows", 500, 1, maxRows)
		if size > 0 {
			v.Check((len(tables[0].rows)-1+size-1)/size <= 500, "split exceeds 500 files")
		}
		args = []string{"split", "--size", strconv.Itoa(size), "--pad", "6", "--quiet"}
		pos = []string{"OUTPUT_DIR", first}
	case "csv.partition":
		sel, e := selection(tables[0], []string{v.String("column", "")})
		v.Check(e == nil, fmt.Sprint(e))
		if e == nil {
			index, _ := strconv.Atoi(sel)
			seen := map[string]string{}
			for _, row := range tables[0].rows[1:] {
				key := row[index-1]
				v.Check(safePartition.MatchString(key), "partition values must be 1-80 letters, numbers, underscores or hyphens")
				fold := strings.ToLower(key)
				if previous, exists := seen[fold]; exists {
					v.Check(previous == key, "partition keys collide on case-insensitive filesystems")
				}
				seen[fold] = key
			}
			v.Check(len(seen) <= 500, "partition exceeds 500 files")
		}
		args = []string{"partition", "--filename", "part-{}.csv"}
		pos = []string{sel, "OUTPUT_DIR", first}
	case "csv.schema":
		args = []string{"schema", "--stdout", "--enum-threshold", number("enum-limit", 0, 0, 1000)}
	case "csv.validate":
		args = []string{"validate", "--fail-fast"}
		schema := v.String("schema", "")
		if schema != "" {
			b, e := localBytes(schema)
			if e != nil {
				return operation.Result{}, e
			}
			if e = localSchema(b); e != nil {
				return operation.Result{}, e
			}
			if e = os.WriteFile(filepath.Join(work, "rules.json"), b, 0600); e != nil {
				return operation.Result{}, e
			}
			pos = append(pos, "rules.json")
		} else {
			args = append(args, "--json")
		}
	case "csv.format":
		out := v.String("out-delimiter", ",")
		_, e := separator(out)
		v.Check(e == nil, fmt.Sprint(e))
		args = []string{"fmt", "--out-delimiter", out}
		flag("crlf", false, "--crlf")
		flag("quote-all", false, "--quote-always")
	case "csv.to-jsonl":
		args = []string{"tojsonl", "--quiet"}
		flag("trim", false, "--trim")
		if !v.Bool("boolean", true) {
			args = append(args, "--no-boolean")
		}
	case "jsonl.to-csv":
		args = []string{"jsonl"}
	case "csv.transform":
		action := v.Enum("action", "trim", "trim", "ltrim", "rtrim", "upper", "lower", "squeeze", "titlecase", "round", "currencytonum")
		args = []string{"apply", "operations", action}
		if action == "round" {
			args = append(args, "--formatstr", number("decimals", 3, 0, 15))
		}
		pos = []string{columns("columns", 0), first}
	}
	if v.Err != nil {
		return operation.Result{}, v.Err
	}
	if prepare != nil {
		if err := prepare(); err != nil {
			return operation.Result{}, err
		}
	}
	args = append(args, "--delimiter", delim, "--")
	args = append(args, pos...)
	call := func() ([]byte, error) {
		return toolrun.RunWithEnv(ctx, resolver, "qsv", "qsv", work, qsvEnv(work), args...)
	}
	if item.mode == "directory" {
		outputs, e := toolrun.Directory(ctx, v.String("output", ""), func(destination string) error {
			for i, a := range args {
				if a == "OUTPUT_DIR" {
					args[i] = destination
				}
			}
			_, e := call()
			if e != nil {
				return e
			}
			entries, e := os.ReadDir(destination)
			if e != nil {
				return e
			}
			if len(entries) > 500 {
				return toolrun.Invalid("too many output files")
			}
			for _, entry := range entries {
				if !entry.Type().IsRegular() {
					return toolrun.Invalid("unexpected nested table output")
				}
				if _, e := localBytes(filepath.Join(destination, entry.Name())); e != nil {
					return e
				}
			}
			return nil
		})
		return operation.Result{Outputs: outputs}, e
	}
	var data map[string]any
	generate := func(destination string) error {
		b, e := call()
		if e != nil {
			return e
		}
		if item.mode == "validate" {
			data = map[string]any{"valid": true}
			if len(bytes.TrimSpace(b)) > 0 && json.Valid(b) {
				var report any
				_ = json.Unmarshal(b, &report)
				data["report"] = report
			}
			return nil
		}
		if item.mode == "json" && !json.Valid(b) {
			return fmt.Errorf("qsv produced invalid JSON")
		}
		if item.mode == "jsonl" {
			if e = checkJSONL(b); e != nil {
				return e
			}
		}
		if len(b) == 0 {
			return fmt.Errorf("qsv produced empty output")
		}
		if item.id == "csv.join" || item.id == "csv.concat" && v.String("mode", "rows") == "columns" {
			b, e = uniqueOutputHeaders(b)
			if e != nil {
				return e
			}
		}
		return os.WriteFile(destination, b, 0600)
	}
	if item.mode == "validate" {
		err = generate("")
	} else {
		err = toolrun.Artifact(ctx, v.String("output", ""), v.Bool("overwrite", false), generate)
	}
	result := operation.Result{Data: data}
	if err == nil && item.mode != "validate" {
		result.Outputs = []string{v.String("output", "")}
	}
	return result, err
}

// Joined tables must remain valid inputs for subsequent name-based operations.
func uniqueOutputHeaders(b []byte) ([]byte, error) {
	reader := csv.NewReader(bytes.NewReader(b))
	rows, e := reader.ReadAll()
	if e != nil {
		return nil, e
	}
	if len(rows) == 0 {
		return b, nil
	}
	reserved := map[string]bool{}
	used := map[string]bool{}
	for _, h := range rows[0] {
		reserved[h] = true
	}
	for i, h := range rows[0] {
		if used[h] {
			for n := 2; ; n++ {
				name := fmt.Sprintf("%s_%d", h, n)
				if !reserved[name] && !used[name] {
					rows[0][i] = name
					break
				}
			}
		}
		used[rows[0][i]] = true
	}
	var output bytes.Buffer
	w := csv.NewWriter(&output)
	e = w.WriteAll(rows)
	return output.Bytes(), e
}
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var safePartition = regexp.MustCompile(`^[\pL\pN_-]{1,80}$`)

func checkJSONL(b []byte) error {
	count := 0
	var keys map[string]bool
	cells := 0
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var object map[string]json.RawMessage
		if e := json.Unmarshal(line, &object); e != nil || object == nil {
			return toolrun.Invalid("JSONL requires one JSON object per nonempty line")
		}
		if keys == nil {
			keys = map[string]bool{}
			for key := range object {
				if key == "" {
					return toolrun.Invalid("JSONL keys must be nonempty")
				}
				keys[key] = true
			}
		}
		if len(object) != len(keys) {
			return toolrun.Invalid("JSONL rows must have identical keys")
		}
		for key, val := range object {
			if !keys[key] {
				return toolrun.Invalid("JSONL rows must have identical keys")
			}
			v := bytes.TrimSpace(val)
			if len(v) > 0 && (v[0] == '{' || v[0] == '[') {
				return toolrun.Invalid("JSONL cells must be scalar values")
			}
		}
		count++
		cells += len(object)
		if count > maxRows || cells > maxCells {
			return toolrun.Invalid("JSONL exceeds row or cell limits")
		}
	}
	if count == 0 || len(keys) == 0 {
		return toolrun.Invalid("JSONL input is empty")
	}
	return nil
}
func localSchema(b []byte) error {
	var value any
	if e := json.Unmarshal(b, &value); e != nil {
		return toolrun.Invalid("invalid JSON Schema")
	}
	var walk func(any) error
	walk = func(v any) error {
		switch x := v.(type) {
		case map[string]any:
			for key, val := range x {
				if key == "dynamicEnum" {
					return toolrun.Invalid("dynamicEnum lookups are not supported")
				}
				if key == "$ref" || key == "$dynamicRef" || key == "$recursiveRef" {
					s, ok := val.(string)
					if !ok || !strings.HasPrefix(s, "#") {
						return toolrun.Invalid("only local fragment schema references are supported")
					}
				}
				if e := walk(val); e != nil {
					return e
				}
			}
		case []any:
			for _, item := range x {
				if e := walk(item); e != nil {
					return e
				}
			}
		}
		return nil
	}
	return walk(value)
}
func qsvEnv(work string) []string {
	env := []string{}
	for _, entry := range os.Environ() {
		key := strings.ToUpper(strings.SplitN(entry, "=", 2)[0])
		if strings.HasPrefix(key, "QSV_") || key == "RAYON_NUM_THREADS" {
			continue
		}
		env = append(env, entry)
	}
	return append(env, "QSV_MAX_JOBS=2", "RAYON_NUM_THREADS=2", "QSV_CACHE_DIR="+work)
}
