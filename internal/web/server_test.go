package web

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fu9zhou/finish-bit/pkg/app"
)

func TestLocalAPI(t *testing.T) {
	a, err := app.New(app.Config{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	h := newHandler(a, "127.0.0.1:8080", "test-secret")
	for _, test := range []struct {
		name, method, path, token, host, origin, body string
		status                                        int
		contains                                      string
	}{
		{name: "embedded UI", method: "GET", path: "/", status: 200, contains: "FinishBit"},
		{name: "embedded module", method: "GET", path: "/forms.mjs", status: 200, contains: "createToolPage"},
		{name: "embedded category icon", method: "GET", path: "/icons/media.svg", status: 200, contains: "viewBox=\"0 0 64 64\""},
		{name: "browse requires auth", method: "GET", path: "/api/browse", status: 401},
		{name: "tasks require auth", method: "GET", path: "/api/package-tasks", status: 401},
		{name: "inspection requires auth", method: "GET", path: "/api/package-tasks?inspect=missing", status: 401},
		{name: "inspection invalid ID", method: "GET", path: "/api/package-tasks?inspect=../secret", token: "test-secret", status: 404},
		{name: "tasks", method: "GET", path: "/api/package-tasks", token: "test-secret", status: 200, contains: "[]"},
		{name: "tasks cross origin", method: "POST", path: "/api/package-tasks", token: "test-secret", origin: "https://evil.example", status: 403},
		{name: "invalid task", method: "POST", path: "/api/package-tasks", token: "test-secret", body: `{"name":"missing"}`, status: 400},
		{name: "task trailing data", method: "POST", path: "/api/package-tasks", token: "test-secret", body: `{"name":"missing"} {}`, status: 400},
		{name: "browse", method: "GET", path: "/api/browse", token: "test-secret", status: 200, contains: "entries"},
		{name: "browse cross origin", method: "GET", path: "/api/browse", token: "test-secret", origin: "https://evil.example", status: 403},
		{name: "browse method", method: "POST", path: "/api/browse", token: "test-secret", status: 405},
		{name: "stdin option", method: "POST", path: "/api/action", token: "test-secret", body: `{"action":"run","name":"text.case","options":{"input-mode":"stdin"},"inputs":["ignored"]}`, status: 200, contains: "cannot read terminal stdin"},
		{name: "auth required", method: "GET", path: "/api/catalog", status: 401},
		{name: "catalog", method: "GET", path: "/api/catalog", token: "test-secret", status: 200, contains: "base64.encode"},
		{name: "uninstalled packages", method: "GET", path: "/api/catalog", token: "test-secret", status: 200, contains: "ffmpeg"},
		{name: "DNS rebinding", method: "GET", path: "/", host: "evil.example:8080", status: 403},
		{name: "cross origin", method: "POST", path: "/api/action", token: "test-secret", origin: "https://evil.example", status: 403},
		{name: "execute", method: "POST", path: "/api/action", token: "test-secret", body: `{"action":"run","name":"base64.encode","inputs":["hello"]}`, status: 200, contains: "aGVsbG8="},
		{name: "validation", method: "POST", path: "/api/action", token: "test-secret", body: `{"action":"run","name":"base64.encode"}`, status: 200, contains: "invalid_input"},
		{name: "stdin", method: "POST", path: "/api/action", token: "test-secret", body: `{"action":"run","name":"base64.encode","inputs":["-"]}`, status: 200, contains: "cannot read terminal stdin"},
		{name: "unknown package", method: "POST", path: "/api/action", token: "test-secret", body: `{"action":"install","name":"unknown-package"}`, status: 200, contains: "unknown package"},
		{name: "unknown action", method: "POST", path: "/api/action", token: "test-secret", body: `{"action":"shell"}`, status: 200, contains: "Unknown action"},
		{name: "malformed", method: "POST", path: "/api/action", token: "test-secret", body: `{`, status: 400},
		{name: "trailing", method: "POST", path: "/api/action", token: "test-secret", body: `{} {}`, status: 400},
		{name: "body limit", method: "POST", path: "/api/action", token: "test-secret", body: `{"name":"` + strings.Repeat("x", 2<<20) + `"}`, status: 400},
		{name: "method", method: "GET", path: "/api/action", token: "test-secret", status: 405},
		{name: "static traversal", method: "GET", path: "/server.go", status: 404},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "http://127.0.0.1:8080"+test.path, strings.NewReader(test.body))
			if test.host != "" {
				r.Host = test.host
			}
			if test.token != "" {
				r.Header.Set("Authorization", "Bearer "+test.token)
			}
			r.Header.Set("Origin", test.origin)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != test.status {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), test.contains) {
				t.Fatalf("missing %q in %s", test.contains, w.Body.String())
			}
			if w.Header().Get("Content-Security-Policy") == "" {
				t.Fatal("missing CSP")
			}
		})
	}
}

func TestServeStopsWithContext(t *testing.T) {
	a, err := app.New(app.Config{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, a, 0, false, io.Discard) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop")
	}
}
