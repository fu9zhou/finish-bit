package cli

import (
	"context"
	"strings"
	"testing"
)

func TestLiteralArgumentsAndBinaryOutput(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"text", "replace", "--", "abc", "abc", "--json"}, "--json\n"},
		{[]string{"base64", "encode", "--", "--hello"}, "LS1oZWxsbw==\n"},
		{[]string{"base64", "decode", "/w=="}, string([]byte{255})},
		{[]string{"base64", "decode", "YQ=="}, "a"},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			c, out, errout := newTestCLI(t)
			if code := c.Run(context.Background(), tc.args); code != 0 || out.String() != tc.want {
				t.Fatalf("code=%d out=%q err=%s", code, out.String(), errout.String())
			}
		})
	}
	c, out, errout := newTestCLI(t)
	if code := c.Run(context.Background(), []string{"base64", "decode", "/w==", "--json"}); code != 0 || !strings.Contains(out.String(), `"base64": "/w=="`) {
		t.Fatalf("code=%d out=%s err=%s", code, out, errout)
	}
}

func TestOperationHelpAndOptionValue(t *testing.T) {
	c, out, errout := newTestCLI(t)
	if code := c.Run(context.Background(), []string{"json", "format", "--help"}); code != 0 || !strings.Contains(out.String(), "default: 2") {
		t.Fatalf("code=%d out=%s err=%s", code, out, errout)
	}
	enabled, args := c.takeJSONFlag([]string{"json", "format", "{}", "--output", "--json"})
	if enabled || len(args) != 5 {
		t.Fatalf("flag=%v args=%q", enabled, args)
	}
}
