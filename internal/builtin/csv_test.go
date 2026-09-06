package builtin

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCSVConversionsPreserveValues(t *testing.T) {
	input := "\ufeffid,name,note\r\n001,张三,\"a,b\"\r\n002,Alice,\"line1\nline2\"\r\n"
	result := runTestOperation(t, "csv.to-json", operation.Request{Inputs: []string{input}})
	var rows []map[string]string
	if err := json.Unmarshal([]byte(result.Data["text"].(string)), &rows); err != nil {
		t.Fatal(err)
	}
	if rows[0]["id"] != "001" || rows[0]["name"] != "张三" || rows[1]["note"] != "line1\nline2" {
		t.Fatal(rows)
	}
	converted := runTestOperation(t, "json.to-csv", operation.Request{Inputs: []string{result.Data["text"].(string)}, Options: map[string]any{"columns": []string{"id", "name", "note"}}})
	actual, err := csv.NewReader(strings.NewReader(converted.Data["text"].(string))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, [][]string{{"id", "name", "note"}, {"001", "张三", "a,b"}, {"002", "Alice", "line1\nline2"}}) {
		t.Fatal(actual)
	}
	info := runTestOperation(t, "csv.info", operation.Request{Inputs: []string{"id;name\n1;A"}, Options: map[string]any{"delimiter": ";"}})
	if info.Data["row_count"] != 1 || info.Data["column_count"] != 2 {
		t.Fatal(info)
	}
}

func TestJSONToCSVScalarsAndColumnOrder(t *testing.T) {
	result := runTestOperation(t, "json.to-csv", operation.Request{Inputs: []string{`[{"z":true,"a":9007199254740993,"b":null},{"a":"001"}]`}})
	if result.Data["text"] != "a,b,z\n9007199254740993,,true\n001,,\n" {
		t.Fatal(result)
	}
	empty := runTestOperation(t, "json.to-csv", operation.Request{Inputs: []string{`[]`}, Options: map[string]any{"columns": []string{"id"}}})
	if empty.Data["text"] != "id\n" {
		t.Fatal(empty)
	}
	header := runTestOperation(t, "csv.to-json", operation.Request{Inputs: []string{"id\n"}})
	if header.Data["text"] != "[]\n" {
		t.Fatal(header)
	}
}

func TestTableRejectsAmbiguityAndBadData(t *testing.T) {
	for _, tc := range []struct{ id, input string }{
		{"csv.to-json", "a,a\n1,2"}, {"csv.info", ",b\n1,2"}, {"csv.info", "a,b\n1"}, {"csv.info", "a\n\"broken"}, {"csv.info", ""},
		{"json.to-csv", `[null]`}, {"json.to-csv", `null`}, {"json.to-csv", `[{"x":{}}]`}, {"json.to-csv", `[{"x":1}] {}`}, {"json.to-csv", `[]`},
	} {
		cap, _ := testRegistry(t).Get(tc.id)
		if _, err := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{tc.input}}); err == nil {
			t.Fatalf("accepted %s %q", tc.id, tc.input)
		}
	}
	cap, _ := testRegistry(t).Get("csv.info")
	if _, err := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{"a\n1"}, Options: map[string]any{"delimiter": "\n"}}); err == nil {
		t.Fatal("invalid delimiter accepted")
	}
	if _, err := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{strings.Repeat("x", maxTableBytes+1)}}); err == nil {
		t.Fatal("oversized table accepted")
	}
}

func TestTableOutputProtection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "table.json")
	os.WriteFile(path, []byte("keep"), 0600)
	request := operation.Request{Inputs: []string{"a\n1"}, Options: map[string]any{"output": path}}
	if _, err := runCSVToJSON(context.Background(), request); err == nil {
		t.Fatal("overwrote existing destination")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "keep" {
		t.Fatal("destination changed")
	}
	request.Options["overwrite"] = true
	if _, err := runCSVToJSON(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	var decoded []map[string]string
	if err := json.Unmarshal(data, &decoded); err != nil || decoded[0]["a"] != "1" {
		t.Fatalf("data=%s err=%v", data, err)
	}
}

func TestCSVRejectsSerializedExpansion(t *testing.T) {
	input := strings.Repeat("h", 1024) + "\n" + strings.Repeat("v\n", 70000)
	if _, err := runCSVToJSON(context.Background(), operation.Request{Inputs: []string{input}}); err == nil {
		t.Fatal("repeated long headers expanded beyond the output limit")
	}
}

func TestCSVTabDelimiterEscape(t *testing.T) {
	result := runTestOperation(t, "csv.info", operation.Request{Inputs: []string{"id\tname\n1\tA"}, Options: map[string]any{"delimiter": `\t`}})
	if result.Data["column_count"] != 2 {
		t.Fatal(result)
	}
}
