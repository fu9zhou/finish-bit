package localtools

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func call(t *testing.T, id string, inputs []string, options map[string]any) operation.Result {
	t.Helper()
	r := operation.NewRegistry()
	if e := Register(r); e != nil {
		t.Fatal(e)
	}
	c, ok := r.Get(id)
	if !ok {
		t.Fatal("missing", id)
	}
	out, e := c.Runner.Run(context.Background(), operation.Request{Inputs: inputs, Options: options})
	if e != nil {
		t.Fatalf("%s: %v", id, e)
	}
	return out
}
func TestLocalTextAndCalculations(t *testing.T) {
	tests := []struct {
		id      string
		input   []string
		options map[string]any
		field   string
		want    any
	}{
		{"number.base", []string{"18446744073709551616"}, map[string]any{"to": 16}, "text", "10000000000000000"},
		{"unit.convert", []string{"1", "GiB", "MB"}, nil, "value", "1073.741824000000"},
		{"unit.convert", []string{"12", "in", "ft"}, nil, "exact", "1"},
		{"temperature.convert", []string{"32", "F", "C"}, nil, "value", "0.000000"},
		{"number.chinese", []string{"100010001.05"}, nil, "text", "壹亿零壹万零壹元零伍分"},
		{"number.chinese", []string{"10"}, map[string]any{"money": false}, "text", "十"},
		{"math.evaluate", []string{"sqrt(81)+pow(2,3)*2"}, nil, "value", 25.0},
		{"health.bmi", []string{"80", "200"}, nil, "bmi", 20.0},
		{"date.add", []string{"2024-01-31"}, map[string]any{"months": 1}, "date", "2024-02-29"},
		{"date.diff", []string{"2024-03-10", "2024-03-11"}, map[string]any{"timezone": "America/New_York"}, "elapsed_seconds", int64(23 * 3600)},
		{"date.business-days", []string{"2024-01-01", "2024-01-08"}, map[string]any{"holidays": []string{"2024-01-01"}, "working-dates": []string{"2024-01-06"}}, "business_days", 5},
		{"date.expiry", []string{"2024-01-31"}, map[string]any{"months": 1, "as-of": "2024-02-29"}, "expired", true},
		{"date.age", []string{"2000-02-29"}, map[string]any{"as-of": "2023-02-28"}, "years", 23},
		{"date.age", []string{"2000-02-29"}, map[string]any{"as-of": "2023-03-01"}, "next_birthday", "2024-02-29"},
		{"calendar.solar", []string{"2024", "1", "1"}, nil, "date", "2024-02-10"},
		{"calendar.lunar", []string{"2024-02-10"}, nil, "day", 1},
		{"text.map", []string{"ab", `{"a":"b","b":"x"}`}, nil, "text", "bx"},
		{"unicode.encode", []string{"中😀"}, nil, "text", `\u4e2d\ud83d\ude00`},
		{"unicode.decode", []string{`\u4e2d\ud83d\ude00`}, nil, "text", "中😀"},
		{"text.normalize", []string{"Ａ　B"}, map[string]any{"mode": "halfwidth"}, "text", "A B"},
		{"regex.replace", []string{"x=42", `(\d+)`, "[$1]"}, nil, "text", "x=[42]"},
		{"regex.validate", []string{"(?<=a)b"}, nil, "valid", false},
		{"text.pinyin", []string{"中国!"}, nil, "text", "zhong guo !"},
		{"text.chinese", []string{"头发发展，汉字。"}, nil, "text", "頭髮發展，漢字。"},
		{"text.chinese", []string{"漢字與頭髮"}, map[string]any{"direction": "t2s"}, "text", "汉字与头发"},
		{"cidr.inspect", []string{"192.168.1.12/24"}, nil, "last", "192.168.1.255"},
		{"cidr.inspect", []string{"::/0"}, nil, "addresses", "340282366920938463463374607431768211456"},
		{"yaml.to-json", []string{"value: 9007199254740993"}, nil, "text", "{\n  \"value\": 9007199254740993\n}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.id+"/"+tt.input[0], func(t *testing.T) {
			r := call(t, tt.id, tt.input, tt.options)
			if !reflect.DeepEqual(r.Data[tt.field], tt.want) {
				t.Fatalf("%s=%#v, want %#v", tt.field, r.Data[tt.field], tt.want)
			}
		})
	}
}
func TestDiffAndStructuredResults(t *testing.T) {
	r := call(t, "json.diff", []string{`{"x":null,"n":9007199254740993}`, `{"n":9007199254740992,"a/b":2}`}, nil)
	changes := r.Data["changes"].([]map[string]any)
	if len(changes) != 3 || changes[0]["path"] != "/a~1b" {
		t.Fatalf("%#v", changes)
	}
	r = call(t, "json.diff", []string{`{"n":1}`, `{"n":1.0}`}, nil)
	if r.Data["equal"] != true {
		t.Fatal(r.Data)
	}
	r = call(t, "text.diff", []string{"a\nb\nc", "a\nc"}, nil)
	lines := r.Data["lines"].([]map[string]any)
	if len(lines) != 3 || lines[1]["kind"] != "remove" {
		t.Fatal(lines)
	}
	r = call(t, "regex.match", []string{"中a1 a2", `a(\d)`}, map[string]any{"limit": 1})
	if r.Data["truncated"] != true || r.Data["matches"].([]map[string]any)[0]["start_byte"] != 3 {
		t.Fatal(r.Data)
	}
	r = call(t, "url.parse", []string{"https://example.com/a?x=1&x=2#part"}, nil)
	b, _ := json.Marshal(r.Data["query"])
	if string(b) != `{"x":["1","2"]}` {
		t.Fatal(string(b))
	}
	r = call(t, "finance.mortgage", []string{"1200", "0"}, map[string]any{"months": 12})
	s := r.Data["schedule"].([]map[string]any)
	if s[0]["payment"] != "100.00" || s[11]["balance"] != "0.00" {
		t.Fatal(r.Data)
	}
	r = call(t, "finance.contributions", []string{"10000", `{"pension":8,"medical":2}`}, nil)
	if r.Data["total"] != "1000.00" {
		t.Fatal(r.Data)
	}
}
func TestRandomAndAuthenticatedCrypto(t *testing.T) {
	r := call(t, "random.integer", []string{"-5", "4"}, map[string]any{"count": 10, "unique": true})
	seen := map[string]bool{}
	for _, n := range r.Data["values"].([]string) {
		if seen[n] {
			t.Fatal("duplicate", n)
		}
		seen[n] = true
	}
	r = call(t, "password.generate", nil, map[string]any{"length": 30, "count": 3, "alphabet": "中文密码"})
	for _, s := range r.Data["values"].([]string) {
		if len([]rune(s)) != 30 {
			t.Fatal(s)
		}
	}
	key := strings.Repeat("a1", 32)
	enc := call(t, "crypto.encrypt", []string{"private 中文"}, map[string]any{"key": key})
	ciphertext := enc.Data["text"].(string)
	plain := call(t, "crypto.decrypt", []string{ciphertext}, map[string]any{"key": key})
	if plain.Data["text"] != "private 中文" {
		t.Fatal(plain)
	}
	reg := operation.NewRegistry()
	_ = Register(reg)
	c, _ := reg.Get("crypto.decrypt")
	if _, e := c.Runner.Run(context.Background(), operation.Request{Inputs: []string{ciphertext}, Options: map[string]any{"key": strings.Repeat("b2", 32)}}); e == nil {
		t.Fatal("wrong key accepted")
	}
	hidden := call(t, "text.hide", []string{"cover", "secret 中文"}, nil)
	r = call(t, "text.reveal", []string{hidden.Data["text"].(string)}, nil)
	if r.Data["text"] != "secret 中文" {
		t.Fatal(r)
	}
}
func TestFileModePublicationAndCancellation(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "text.txt")
	if e := os.WriteFile(file, []byte("HELLO"), 0600); e != nil {
		t.Fatal(e)
	}
	r := call(t, "text.case", []string{file}, nil)
	if r.Data["text"] != strings.ToLower(file) {
		t.Fatal("implicit file read")
	}
	r = call(t, "text.case", []string{file}, map[string]any{"input-mode": "file"})
	if r.Data["text"] != "hello" {
		t.Fatal(r)
	}
	reg := operation.NewRegistry()
	_ = Register(reg)
	c, _ := reg.Get("text.case")
	_, e := c.Runner.Run(context.Background(), operation.Request{Inputs: []string{"replacement"}, Options: map[string]any{"output": file}})
	if e == nil {
		t.Fatal("overwrote existing file")
	}
	b, _ := os.ReadFile(file)
	if string(b) != "HELLO" {
		t.Fatal("destination changed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = c.Runner.Run(ctx, operation.Request{Inputs: []string{"X"}}); e == nil {
		t.Fatal("ignored cancellation")
	}
}
func TestRejectedInputs(t *testing.T) {
	cases := []struct {
		id string
		a  []string
		o  map[string]any
	}{
		{"unit.convert", []string{"1", "m", "kg"}, nil}, {"temperature.convert", []string{"-1", "K", "C"}, nil},
		{"number.chinese", []string{"1.001"}, nil}, {"math.evaluate", []string{"1/0"}, nil}, {"math.evaluate", []string{"os.Exit(0)"}, nil}, {"math.evaluate", []string{"sqrt(-1)"}, nil},
		{"yaml.to-json", []string{"a: 1\na: 2"}, nil}, {"yaml.to-json", []string{"x: &x [*x]"}, nil}, {"yaml.to-json", []string{"a: .nan"}, nil}, {"yaml.to-json", []string{"a: 1\n---\nb: 2"}, nil},
		{"xml.validate", []string{"<!DOCTYPE x><x/>"}, nil}, {"xml.validate", []string{"<x/><y/>"}, nil},
		{"random.integer", []string{"1", "2"}, map[string]any{"count": 3, "unique": true}},
		{"date.add", []string{"9999-12-31"}, map[string]any{"days": 1}},
		{"text.case", []string{"hello"}, map[string]any{"mode": "unknown"}},
		{"unicode.decode", []string{`\ud800`}, nil}, {"unicode.decode", []string{`\udc00`}, nil},
		{"regex.replace", []string{strings.Repeat("a", 10001), "a", "b"}, nil},
		{"calendar.solar", []string{"2024", "1", "1"}, map[string]any{"leap": true}},
		{"json.to-toml", []string{`{"a":null}`}, nil},
		{"json.to-toml", []string{`{"a":9223372036854775808}`}, nil},
	}
	reg := operation.NewRegistry()
	_ = Register(reg)
	for _, tt := range cases {
		c, _ := reg.Get(tt.id)
		if _, e := c.Runner.Run(context.Background(), operation.Request{Inputs: tt.a, Options: tt.o}); e == nil {
			t.Errorf("%s accepted %v", tt.id, tt.a)
		}
	}
}
func fixture(t *testing.T, dir string) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 65, 49))
	for y := 0; y < 49; y++ {
		for x := 0; x < 65; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x * 3), uint8(y * 4), 100, 255})
		}
	}
	p := filepath.Join(dir, "输入.png")
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	e = png.Encode(f, img)
	_ = f.Close()
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestImageAndQRWorkflows(t *testing.T) {
	dir := t.TempDir()
	input := fixture(t, dir)
	qr := filepath.Join(dir, "二维码.png")
	call(t, "qrcode.generate", []string{"https://example.com/中文"}, map[string]any{"output": qr})
	r := call(t, "qrcode.decode", []string{qr}, nil)
	if r.Data["text"] != "https://example.com/中文" {
		t.Fatal(r)
	}
	tiles := call(t, "image.split-grid", []string{input}, map[string]any{"output": filepath.Join(dir, "tiles")})
	if len(tiles.Outputs) != 9 {
		t.Fatal(tiles)
	}
	area := 0
	for _, p := range tiles.Outputs {
		img, e := loadImage(p)
		if e != nil {
			t.Fatal(e)
		}
		area += img.Bounds().Dx() * img.Bounds().Dy()
	}
	if area != 65*49 {
		t.Fatal("grid lost pixels", area)
	}
	hidden := filepath.Join(dir, "hidden.png")
	call(t, "image.hide", []string{input, "你好 hidden"}, map[string]any{"output": hidden})
	r = call(t, "image.reveal", []string{hidden}, nil)
	if r.Data["text"] != "你好 hidden" {
		t.Fatal(r)
	}
	padded := filepath.Join(dir, "padded.png")
	call(t, "image.minimum-bytes", []string{input}, map[string]any{"output": padded, "bytes": 10000})
	stat, _ := os.Stat(padded)
	if stat.Size() != 10000 {
		t.Fatal(stat.Size())
	}
	original, _ := loadImage(input)
	restored, e := loadImage(padded)
	if e != nil {
		t.Fatal(e)
	}
	if original.At(20, 20) != restored.At(20, 20) {
		t.Fatal("padding altered pixels")
	}
	pixelated := filepath.Join(dir, "pixelated.png")
	call(t, "image.pixelate", []string{input}, map[string]any{"output": pixelated, "block": 8})
	img, _ := loadImage(pixelated)
	if img.At(0, 0) != img.At(7, 7) {
		t.Fatal("block not uniform")
	}
	r = call(t, "image.ascii", []string{input}, map[string]any{"columns": 20})
	if len(strings.Split(strings.TrimSuffix(r.Data["text"].(string), "\n"), "\n")[0]) != 20 {
		t.Fatal(r)
	}
	r = call(t, "image.palette", []string{input}, nil)
	if len(r.Data["colors"].([]map[string]any)) != 8 {
		t.Fatal(r)
	}
	photo := filepath.Join(dir, "photo.png")
	call(t, "image.id-photo", []string{input}, map[string]any{"output": photo})
	img, _ = loadImage(photo)
	if img.Bounds().Dx() != 295 || img.Bounds().Dy() != 413 {
		t.Fatal(img.Bounds())
	}
	jpeg := filepath.Join(dir, "small.jpg")
	r = call(t, "image.target-size", []string{input}, map[string]any{"output": jpeg, "bytes": 1500})
	if r.Data["bytes"].(int) > 1500 {
		t.Fatal(r)
	}
}
func TestMarkdownDefaultBlocksActiveContent(t *testing.T) {
	r := call(t, "markdown.to-html", []string{"# Hello\n<script>alert(1)</script>\n[link](javascript:alert(1))"}, nil)
	s := r.Data["text"].(string)
	if strings.Contains(s, "<script>") || strings.Contains(s, `href="javascript:`) || !strings.Contains(s, "Content-Security-Policy") {
		t.Fatal(s)
	}
}
