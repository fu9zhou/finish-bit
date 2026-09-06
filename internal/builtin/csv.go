package builtin

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

const maxTableBytes = 8 << 20
const maxTableCells = 1_000_000

func registerCSV(registry *operation.Registry) error {
	input := []operation.Parameter{param("input", "UTF-8 text, file path, or - for stdin", true)}
	delimiter := option("delimiter", operation.TypeString, "One field separator character", ",")
	output := []operation.Parameter{option("output", operation.TypeString, "Write result to a file", ""), option("overwrite", operation.TypeBoolean, "Replace an existing destination", false)}
	return registerAll(registry,
		operation.Capability{Definition: operation.Definition{ID: "csv.info", Summary: "Inspect CSV headers and row count", Description: "Validate UTF-8 CSV with unique non-empty headers and consistent row widths.", Aliases: []string{"表格信息", "CSV 信息"}, Tags: []string{"csv", "table", "metadata"}, Inputs: input, Options: []operation.Parameter{delimiter}, Source: "core"}, Runner: operation.Func(runCSVInfo)},
		operation.Capability{Definition: operation.Definition{ID: "csv.to-json", Summary: "Convert CSV rows to JSON objects", Description: "Use the first row as field names and preserve all cell values as strings, including leading zeros.", Aliases: []string{"CSV 转 JSON", "表格转 JSON"}, Tags: []string{"csv", "json", "table", "convert"}, Inputs: input, Options: append([]operation.Parameter{delimiter}, output...), Source: "core"}, Runner: operation.Func(runCSVToJSON)},
		operation.Capability{Definition: operation.Definition{ID: "json.to-csv", Summary: "Convert a JSON object array to CSV", Description: "Write scalar fields as CSV. Infer sorted columns or select explicit ordered columns. Missing values and null become empty cells; nested values are rejected.", Aliases: []string{"JSON 转 CSV", "JSON 转表格"}, Tags: []string{"csv", "json", "table", "convert"}, Inputs: input, Options: append([]operation.Parameter{delimiter, option("columns", operation.TypeStrings, "Ordered columns; repeat the option for each column", nil)}, output...), Source: "core"}, Runner: operation.Func(runJSONToCSV)},
	)
}

func tableError(message string) error {
	return &operation.Error{Code: operation.CodeInvalidInput, Message: message}
}
func tableInput(request operation.Request) ([]byte, error) {
	if err := operation.RequireInputs(request, 1); err != nil {
		return nil, err
	}
	data, err := readInput(request.Inputs[0])
	if err != nil {
		return nil, err
	}
	if len(data) > maxTableBytes {
		return nil, tableError("table input exceeds 8 MiB")
	}
	if !utf8.Valid(data) {
		return nil, tableError("table input must be UTF-8")
	}
	return bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}), nil
}
func tableDelimiter(request operation.Request) (rune, error) {
	value, err := operation.StringOption(request, "delimiter", ",")
	if err != nil {
		return 0, err
	}
	if value == `\t` {
		value = "\t"
	}
	runes := []rune(value)
	if len(runes) != 1 || runes[0] == 0 || runes[0] == '"' || runes[0] == '\r' || runes[0] == '\n' || runes[0] == utf8.RuneError {
		return 0, tableError("delimiter must be one character other than quote, newline, NUL or replacement character")
	}
	return runes[0], nil
}
func validateColumns(columns []string) error {
	if len(columns) == 0 || len(columns) > 1000 {
		return tableError("table requires between 1 and 1000 columns")
	}
	seen := map[string]bool{}
	for _, column := range columns {
		if strings.TrimSpace(column) == "" || seen[column] {
			return tableError("column names must be non-empty and unique")
		}
		seen[column] = true
	}
	return nil
}
func readCSV(ctx context.Context, request operation.Request) ([][]string, error) {
	data, err := tableInput(request)
	if err != nil {
		return nil, err
	}
	delimiter, err := tableDelimiter(request)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	rows := [][]string{}
	cells := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, tableError(fmt.Sprintf("invalid CSV: %v", err))
		}
		if len(rows) == 0 {
			if err := validateColumns(row); err != nil {
				return nil, err
			}
		}
		cells += len(row)
		if cells > maxTableCells || len(rows) >= 100001 {
			return nil, tableError("table exceeds 100000 data rows or 1000000 cells")
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, tableError("CSV requires a header row")
	}
	return rows, nil
}

func runCSVInfo(ctx context.Context, request operation.Request) (operation.Result, error) {
	rows, err := readCSV(ctx, request)
	if err != nil {
		return operation.Result{}, err
	}
	return operation.Result{Operation: "csv.info", Data: map[string]any{"columns": rows[0], "column_count": len(rows[0]), "row_count": len(rows) - 1}}, nil
}

// tableBuffer bounds serialized expansion, including repeated long field names.
type tableBuffer struct{ buffer bytes.Buffer }

func (b *tableBuffer) String() string { return b.buffer.String() }

func (b *tableBuffer) Write(data []byte) (int, error) {
	if int64(b.buffer.Len())+int64(len(data)) > maxInputBytes {
		return 0, tableError("table output exceeds 64 MiB")
	}
	return b.buffer.Write(data)
}

func runCSVToJSON(ctx context.Context, request operation.Request) (operation.Result, error) {
	rows, err := readCSV(ctx, request)
	if err != nil {
		return operation.Result{}, err
	}
	buffer := &tableBuffer{}
	buffer.Write([]byte("["))
	encoder := json.NewEncoder(buffer)
	for index, row := range rows[1:] {
		if err := ctx.Err(); err != nil {
			return operation.Result{}, err
		}
		if index > 0 {
			if _, err := buffer.Write([]byte(",")); err != nil {
				return operation.Result{}, err
			}
		}
		object := map[string]string{}
		for col, key := range rows[0] {
			object[key] = row[col]
		}
		if err := encoder.Encode(object); err != nil {
			return operation.Result{}, err
		}
	}
	if _, err := buffer.Write([]byte("]\n")); err != nil {
		return operation.Result{}, err
	}
	return finishArtifactText("csv.to-json", buffer.String(), request)
}

func runJSONToCSV(ctx context.Context, request operation.Request) (operation.Result, error) {
	data, err := tableInput(request)
	if err != nil {
		return operation.Result{}, err
	}
	delimiter, err := tableDelimiter(request)
	if err != nil {
		return operation.Result{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var rows []map[string]any
	if err := decoder.Decode(&rows); err != nil {
		return operation.Result{}, tableError("input must be a JSON array of objects")
	}
	if rows == nil {
		return operation.Result{}, tableError("input must be a JSON array of objects")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return operation.Result{}, tableError("input must contain exactly one JSON array")
	}
	if len(rows) > 100000 {
		return operation.Result{}, tableError("table exceeds 100000 rows")
	}
	all := map[string]bool{}
	cells := 0
	for _, row := range rows {
		if row == nil {
			return operation.Result{}, tableError("every JSON row must be an object")
		}
		cells += len(row)
		if cells > maxTableCells {
			return operation.Result{}, tableError("table exceeds 1000000 cells")
		}
		for key, value := range row {
			all[key] = true
			switch value.(type) {
			case nil, string, bool, json.Number:
			default:
				return operation.Result{}, tableError("nested arrays and objects cannot be converted to CSV")
			}
		}
	}
	columns := []string{}
	if raw, present := request.Options["columns"]; present {
		switch values := raw.(type) {
		case []string:
			columns = append(columns, values...)
		case []any:
			for _, value := range values {
				text, ok := value.(string)
				if !ok {
					return operation.Result{}, tableError("columns must be a string array")
				}
				columns = append(columns, text)
			}
		default:
			return operation.Result{}, tableError("columns must be a string array")
		}
	} else {
		for key := range all {
			columns = append(columns, key)
		}
		sort.Strings(columns)
	}
	if err := validateColumns(columns); err != nil {
		return operation.Result{}, err
	}
	if len(columns)*(len(rows)+1) > maxTableCells {
		return operation.Result{}, tableError("output table exceeds 1000000 cells")
	}
	buffer := &tableBuffer{}
	writer := csv.NewWriter(buffer)
	writer.Comma = delimiter
	if err := writer.Write(columns); err != nil {
		return operation.Result{}, err
	}
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return operation.Result{}, err
		}
		record := make([]string, len(columns))
		for index, key := range columns {
			switch value := row[key].(type) {
			case string:
				record[index] = value
			case json.Number:
				record[index] = string(value)
			case bool:
				record[index] = fmt.Sprint(value)
			}
		}
		if err := writer.Write(record); err != nil {
			return operation.Result{}, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return operation.Result{}, err
	}
	return finishArtifactText("json.to-csv", buffer.String(), request)
}
