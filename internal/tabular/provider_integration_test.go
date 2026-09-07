package tabular

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type runtimeResolver string

func (p runtimeResolver) Executable(string, string) (string, error) { return string(p), nil }

func TestTablesWithManagedQSV(t *testing.T) {
	binary := os.Getenv("FINISHBIT_TEST_QSV")
	if binary == "" {
		t.Skip("set FINISHBIT_TEST_QSV to the managed qsv executable")
	}
	root := filepath.Join(t.TempDir(), "表格 ' [data]")
	if e := os.MkdirAll(root, 0755); e != nil {
		t.Fatal(e)
	}
	write := func(name, body string) string {
		p := filepath.Join(root, name)
		if e := os.WriteFile(p, []byte(body), 0600); e != nil {
			t.Fatal(e)
		}
		return p
	}
	input := write("input.csv", "id,name,team,amount,tags,note\n001,Alice,A,10,x;y,\n002,Bob,B,20,z,ready\n003,陈明,A,30,x,\n004,Dana,B,20,y,done\n")
	right := write("right.csv", "id,city\n001,北京\n003,上海\n005,杭州\n")
	changed := write("changed.csv", "id,name,team,amount,tags,note\n001,Alice,A,11,x;y,\n002,Bob,B,20,z,ready\n003,陈明,A,30,x,\n004,Dana,B,20,y,done\n")
	duplicates := write("duplicates.csv", "id,value\n02,first\n01,earlier\n02,later\n")
	jsonl := write("input.jsonl", "{\"id\":\"001\",\"name\":\"陈明\"}\n{\"id\":\"002\",\"name\":\"Alice\"}\n")
	schema := write("schema.json", `{"type":"object","properties":{"amount":{"type":"integer","minimum":0}},"required":["amount"]}`)
	registry := operation.NewRegistry()
	if e := Register(registry, runtimeResolver(binary)); e != nil {
		t.Fatal(e)
	}
	run := func(t *testing.T, id string, inputs []string, opts map[string]any) operation.Result {
		t.Helper()
		cap, ok := registry.Get(id)
		if !ok {
			t.Fatal("missing", id)
		}
		result, e := cap.Runner.Run(context.Background(), operation.Request{Inputs: inputs, Options: opts})
		if e != nil {
			t.Fatalf("%s: %+v", id, e)
		}
		return result
	}
	read := func(t *testing.T, path string) [][]string {
		t.Helper()
		f, e := os.Open(path)
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		records, e := csv.NewReader(f).ReadAll()
		if e != nil {
			t.Fatal(e)
		}
		return records
	}
	cases := []struct {
		id     string
		inputs []string
		opts   map[string]any
		rows   int
	}{
		{"csv.select", nil, map[string]any{"columns": []string{"id", "amount"}}, 5},
		{"csv.rename", nil, map[string]any{"names": []string{"编号", "姓名", "组", "金额", "标签", "备注"}}, 5},
		{"csv.filter", nil, map[string]any{"columns": []string{"team"}, "pattern": "A", "mode": "exact"}, 3},
		{"csv.sort", nil, map[string]any{"columns": []string{"amount"}, "mode": "numeric", "reverse": true}, 5},
		{"csv.deduplicate", []string{duplicates}, map[string]any{"columns": []string{"id"}}, 3},
		{"csv.concat", nil, map[string]any{"files": []string{input}}, 9},
		{"csv.join", []string{input, right}, map[string]any{"left-keys": []string{"id"}, "right-keys": []string{"id"}}, 3},
		{"csv.diff", []string{input, changed}, map[string]any{"keys": []string{"id"}}, 3},
		{"csv.frequency", nil, map[string]any{"columns": []string{"team"}, "limit": 10}, 3},
		{"csv.stats", nil, map[string]any{"columns": []string{"amount"}}, 2},
		{"csv.sample", nil, map[string]any{"count": 2, "seed": 7}, 3},
		{"csv.shuffle", nil, map[string]any{"seed": 7}, 5},
		{"csv.slice", nil, map[string]any{"start": 1, "count": 2}, 3},
		{"csv.reverse", nil, nil, 5},
		{"csv.transpose", nil, nil, 6},
		{"csv.melt", nil, map[string]any{"columns": []string{"amount", "note"}}, 7},
		{"csv.explode", nil, map[string]any{"column": "tags", "separator": ";"}, 6},
		{"csv.fill", nil, map[string]any{"columns": []string{"note"}, "mode": "value", "value": "missing"}, 5},
		{"csv.replace", nil, map[string]any{"columns": []string{"name"}, "pattern": "Alice", "replacement": "$cost"}, 5},
		{"csv.split", nil, map[string]any{"rows": 2}, 0},
		{"csv.partition", nil, map[string]any{"column": "team"}, 0},
		{"csv.schema", nil, nil, 0},
		{"csv.validate", nil, map[string]any{"schema": schema}, 0},
		{"csv.format", nil, map[string]any{"quote-all": true, "crlf": true}, 5},
		{"csv.to-jsonl", nil, nil, 0},
		{"jsonl.to-csv", []string{jsonl}, nil, 3},
		{"csv.transform", nil, map[string]any{"columns": []string{"name"}, "action": "upper"}, 5},
	}
	covered := map[string]bool{}
	for _, test := range cases {
		covered[test.id] = true
		t.Run(test.id, func(t *testing.T) {
			opts := map[string]any{}
			for k, v := range test.opts {
				opts[k] = v
			}
			out := filepath.Join(root, test.id+".csv")
			opts["output"] = out
			if test.id == "csv.validate" {
				delete(opts, "output")
			}
			inputs := test.inputs
			if inputs == nil {
				inputs = []string{input}
			}
			result := run(t, test.id, inputs, opts)
			if test.id == "csv.validate" {
				if result.Data["valid"] != true {
					t.Fatal("validation result")
				}
				return
			}
			if test.id == "csv.split" || test.id == "csv.partition" {
				if len(result.Outputs) != 2 {
					t.Fatalf("expected two files: %v", result.Outputs)
				}
				total := 0
				for _, p := range result.Outputs {
					total += len(read(t, p)) - 1
				}
				if total != 4 {
					t.Fatal("split lost rows")
				}
				return
			}
			data, e := os.ReadFile(out)
			if e != nil || len(data) == 0 {
				t.Fatal("no output", e)
			}
			if test.id == "csv.schema" {
				var schema map[string]any
				if e = json.Unmarshal(data, &schema); e != nil || schema["properties"] == nil {
					t.Fatal("bad schema", e)
				}
				return
			}
			if test.id == "csv.to-jsonl" {
				lines := strings.Split(strings.TrimSpace(string(data)), "\n")
				if len(lines) != 4 {
					t.Fatal("JSONL lost rows")
				}
				var row map[string]any
				if e = json.Unmarshal([]byte(lines[0]), &row); e != nil || row["id"] != "001" {
					t.Fatal("leading zero lost", row, e)
				}
				return
			}
			records := read(t, out)
			if len(records) != test.rows {
				t.Fatalf("rows=%d expected=%d: %v", len(records), test.rows, records)
			}
			switch test.id {
			case "csv.select":
				if !equal(records[0], []string{"id", "amount"}) || records[1][0] != "001" {
					t.Fatal(records)
				}
			case "csv.rename":
				if records[0][0] != "编号" {
					t.Fatal(records)
				}
			case "csv.filter":
				if records[2][0] != "003" {
					t.Fatal(records)
				}
			case "csv.sort":
				if records[1][3] != "30" {
					t.Fatal(records)
				}
			case "csv.deduplicate":
				if records[2][1] != "first" {
					t.Fatal("wrong retained record", records)
				}
			case "csv.join":
				if records[1][len(records[1])-1] != "北京" {
					t.Fatal(records)
				}
			case "csv.diff":
				if records[1][0] == records[2][0] {
					t.Fatal("missing change markers", records)
				}
			case "csv.frequency":
				found := false
				for i, h := range records[0] {
					if h == "count" {
						found = true
						if records[1][i] != "2" {
							t.Fatal(records)
						}
					}
				}
				if !found {
					t.Fatal("missing counts", records)
				}
			case "csv.stats":
				found := false
				for i, h := range records[0] {
					if h == "mean" {
						found = true
						if records[1][i] != "20" {
							t.Fatal(records)
						}
					}
				}
				if !found {
					t.Fatal("missing mean", records)
				}
			case "csv.slice":
				if records[1][0] != "002" {
					t.Fatal(records)
				}
			case "csv.reverse":
				if records[1][0] != "004" {
					t.Fatal(records)
				}
			case "csv.transpose":
				if records[0][1] != "001" {
					t.Fatal(records)
				}
			case "csv.fill":
				if records[1][5] != "missing" || records[2][5] != "ready" {
					t.Fatal(records)
				}
			case "csv.replace":
				if records[1][1] != "$cost" {
					t.Fatal("literal replacement changed", records)
				}
			case "csv.transform":
				if records[1][1] != "ALICE" {
					t.Fatal(records)
				}
			}
		})
	}
	for _, s := range catalog() {
		if !covered[s.id] {
			t.Error("untested operation", s.id)
		}
	}
	t.Run("failure-and-overwrite", func(t *testing.T) {
		out := write("keep.csv", "ORIGINAL")
		cap, _ := registry.Get("csv.select")
		for _, opts := range []map[string]any{{"output": out, "columns": []string{"id"}}, {"output": out, "columns": []string{"absent"}, "overwrite": true}} {
			if _, e := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: opts}); e == nil {
				t.Fatal("expected refusal")
			}
			b, _ := os.ReadFile(out)
			if string(b) != "ORIGINAL" {
				t.Fatal("damaged existing output")
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e := cap.Runner.Run(ctx, operation.Request{Inputs: []string{input}, Options: map[string]any{"output": out, "columns": []string{"id"}, "overwrite": true}}); e == nil {
			t.Fatal("cancel succeeded")
		}
		run(t, "csv.select", []string{input}, map[string]any{"output": out, "columns": []string{"id"}, "overwrite": true})
		if read(t, out)[1][0] != "001" {
			t.Fatal("overwrite failed")
		}
		cap, _ = registry.Get("csv.split")
		if _, e := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{"rows": 0, "output": filepath.Join(root, "invalid")}}); e == nil {
			t.Fatal("zero chunk size accepted")
		}
	})
	t.Run("join-modes-and-column-composition", func(t *testing.T) {
		for kind, count := range map[string]int{"left": 5, "right": 4, "full": 6, "left-anti": 3, "left-semi": 3, "right-anti": 2, "right-semi": 3} {
			out := filepath.Join(root, "join-"+kind+".csv")
			run(t, "csv.join", []string{input, right}, map[string]any{"left-keys": []string{"id"}, "right-keys": []string{"id"}, "kind": kind, "output": out})
			records := read(t, out)
			if len(records) != count {
				t.Fatalf("%s: %v", kind, records)
			}
			b, _ := os.ReadFile(out)
			if _, e := readTable(b, ','); e != nil {
				t.Fatal("joined result cannot be reused", e)
			}
		}
		out := filepath.Join(root, "columns.csv")
		run(t, "csv.concat", []string{input}, map[string]any{"files": []string{right}, "mode": "columns", "pad": true, "output": out})
		records := read(t, out)
		if len(records) != 5 || records[0][6] != "id_2" {
			t.Fatal(records)
		}
		run(t, "csv.concat", []string{input}, map[string]any{"files": []string{right}, "mode": "rowskey", "output": out, "overwrite": true})
		if len(read(t, out)) != 8 {
			t.Fatal("union concatenation lost rows")
		}
	})
	t.Run("delimiter-and-exact-column-names", func(t *testing.T) {
		input := write("quoted.tsv", "\"a,b\"\t编号\nvalue\t001\n")
		out := filepath.Join(root, "quoted-out.csv")
		run(t, "csv.select", []string{input}, map[string]any{"columns": []string{"编号", "a,b"}, "delimiter": "\t", "output": out})
		records := read(t, out)
		if records[1][0] != "001" || records[1][1] != "value" {
			t.Fatal(records)
		}
		run(t, "csv.deduplicate", []string{duplicates}, map[string]any{"columns": []string{"id"}, "keep": "last", "output": out, "overwrite": true})
		if read(t, out)[2][1] != "later" {
			t.Fatal("last occurrence was not retained")
		}
	})
	t.Run("bad-data-and-schema", func(t *testing.T) {
		cap, _ := registry.Get("csv.validate")
		bad := write("ragged.csv", "a,b\n1\n")
		if _, e := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{bad}}); e == nil {
			t.Fatal("ragged CSV valid")
		}
		invalid := write("invalid.csv", "amount\n-1\n")
		if _, e := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{invalid}, Options: map[string]any{"schema": schema}}); e == nil {
			t.Fatal("invalid row accepted")
		}
		remote := write("remote.json", `{"$ref":"https://example.com/schema"}`)
		if _, e := cap.Runner.Run(context.Background(), operation.Request{Inputs: []string{input}, Options: map[string]any{"schema": remote}}); e == nil {
			t.Fatal("remote schema accepted")
		}
	})
	t.Run("no-matches-and-seed", func(t *testing.T) {
		out := filepath.Join(root, "empty.csv")
		run(t, "csv.filter", []string{input}, map[string]any{"pattern": "NEVER MATCH", "output": out})
		if len(read(t, out)) != 1 {
			t.Fatal("no-match result")
		}
		var previous string
		for i := 0; i < 2; i++ {
			run(t, "csv.sample", []string{input}, map[string]any{"count": 2, "seed": 17, "output": out, "overwrite": true})
			b, _ := os.ReadFile(out)
			if i == 1 && previous != string(b) {
				t.Fatal("seed not reproducible")
			}
			previous = string(b)
		}
	})
}
