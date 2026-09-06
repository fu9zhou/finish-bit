package operation

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeCallContract(t *testing.T) {
	call, err := DecodeCall(strings.NewReader(`{"operation":"text.replace","inputs":["a\nb","a","--json"],"options":{"count":9007199254740993}}`))
	if err != nil {
		t.Fatal(err)
	}
	if call.Inputs[0] != "a\nb" || call.Options["count"] != json.Number("9007199254740993") {
		t.Fatalf("call=%#v", call)
	}
	n, err := IntOption(Request{Options: map[string]any{"count": json.Number("42")}}, "count", 0)
	if err != nil || n != 42 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	for _, raw := range []string{``, `null`, `[]`, `{}`, `{"operation":""}`, `{"operation":"json.format","unknown":true}`, `{"operation":"json.format","inputs":[1]}`, `{"operation":"json.format"} {}`, `{"operation":"json.format","options":[]}`, string([]byte{255}), strings.Repeat(" ", int(MaxCallBytes)+1)} {
		if _, err := DecodeCall(strings.NewReader(raw)); err == nil {
			t.Fatalf("accepted invalid request: %.100q", raw)
		}
	}
}

func TestStructuredIntegerPrecision(t *testing.T) {
	for _, raw := range []string{"2", "2.0", "2e0"} {
		n, err := IntOption(Request{Options: map[string]any{"n": json.Number(raw)}}, "n", 0)
		if err != nil || n != 2 {
			t.Fatalf("%s: %d %v", raw, n, err)
		}
	}
	for _, raw := range []string{"2.1", "1e1000000000", "9223372036854775808", "2.0000000000000000000000001"} {
		if _, err := IntOption(Request{Options: map[string]any{"n": json.Number(raw)}}, "n", 0); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
