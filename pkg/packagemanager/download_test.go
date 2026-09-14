package packagemanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDownloadHonorsCallerDeadlineWhileReceivingBytes(t *testing.T) {
	var mirrors atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mirror" {
			mirrors.Add(1)
			return
		}
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				_, _ = w.Write([]byte("data"))
				w.(http.Flusher).Flush()
			}
		}
	}))
	defer server.Close()
	m := New(t.TempDir(), Registry{})
	m.client.Transport = server.Client().Transport
	m.downloadIdleTimeout = time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err := m.download(ctx, Artifact{URL: server.URL + "/primary", Mirrors: []string{server.URL + "/mirror"}}, filepath.Join(t.TempDir(), "archive"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected caller deadline, got %v", err)
	}
	if mirrors.Load() != 0 {
		t.Fatal("tried a mirror after caller deadline")
	}
}

func TestDownloadHTTPPolicy(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		status                  int
		retryAfter              string
		recover                 bool
		wantPrimary, wantMirror int
	}{
		{"missing", 404, "", false, 1, 1},
		{"temporary", 503, "", true, 2, 0},
		{"exhausted", 502, "", false, 2, 1},
		{"rate-limit", 429, "1", true, 2, 0},
		{"long-rate-limit", 429, "120", false, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			content := []byte("verified")
			hash := sha256.Sum256(content)
			var primary, mirror atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/primary" {
					n := primary.Add(1)
					if !tc.recover || n == 1 {
						w.Header().Set("Retry-After", tc.retryAfter)
						w.WriteHeader(tc.status)
						return
					}
				} else {
					mirror.Add(1)
				}
				_, _ = w.Write(content)
			}))
			defer server.Close()
			m := New(t.TempDir(), Registry{})
			m.client = server.Client()
			var events []Progress
			ctx := WithProgress(context.Background(), func(p Progress) { events = append(events, p) })
			err := m.download(ctx, Artifact{URL: server.URL + "/primary", Mirrors: []string{server.URL + "/mirror"}, SHA256: hex.EncodeToString(hash[:])}, filepath.Join(t.TempDir(), "download"))
			if err != nil {
				t.Fatal(err)
			}
			if int(primary.Load()) != tc.wantPrimary || int(mirror.Load()) != tc.wantMirror {
				t.Fatalf("requests primary=%d mirror=%d", primary.Load(), mirror.Load())
			}
			for _, p := range events {
				if p.SourceCount != 2 || p.SourceIndex < 1 || p.SourceIndex > 2 {
					t.Fatalf("missing source position: %+v", p)
				}
				if p.Stage == "source-failed" && (p.Reason == "" || p.NextSource == "") {
					t.Fatalf("missing failure detail: %+v", p)
				}
			}
		})
	}
}

func TestDownloadStopsOnLocalFileFailure(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); _, _ = w.Write([]byte("data")) }))
	defer server.Close()
	m := New(t.TempDir(), Registry{})
	m.client = server.Client()
	// A missing parent makes opening the output fail after the HTTP request.
	err := m.download(context.Background(), Artifact{URL: server.URL, Mirrors: []string{server.URL + "/mirror"}}, filepath.Join(t.TempDir(), "absent", "download"))
	if err == nil || !strings.Contains(err.Error(), "local package file") || requests.Load() != 1 {
		t.Fatalf("err=%v requests=%d", err, requests.Load())
	}
}

func TestDownloadCancellationDuringRetry(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); w.WriteHeader(503) }))
	defer server.Close()
	m := New(t.TempDir(), Registry{})
	m.client = server.Client()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = WithProgress(ctx, func(p Progress) {
		if p.Stage == "source-retrying" {
			cancel()
		}
	})
	err := m.download(ctx, Artifact{URL: server.URL, Mirrors: []string{server.URL + "/mirror"}}, filepath.Join(t.TempDir(), "download"))
	if !errors.Is(err, context.Canceled) || requests.Load() != 1 {
		t.Fatalf("err=%v requests=%d", err, requests.Load())
	}
}

func TestDownloadAllSourcesFailAndCleanPartialFile(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("wrong")) }))
	defer server.Close()
	m := New(t.TempDir(), Registry{})
	m.client = server.Client()
	path := filepath.Join(t.TempDir(), "download")
	var last Progress
	ctx := WithProgress(context.Background(), func(p Progress) { last = p })
	err := m.download(ctx, Artifact{URL: server.URL, Mirrors: []string{server.URL + "/mirror"}, SHA256: strings.Repeat("0", 64)}, path)
	if err == nil || !strings.Contains(err.Error(), "source 1/2") || !strings.Contains(err.Error(), "source 2/2") {
		t.Fatalf("missing errors: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("failed bytes remain: %v", err)
	}
	if last.Stage != "source-failed" || last.NextSource != "" || last.Reason == "" {
		t.Fatalf("last event: %+v", last)
	}
}

func TestDownloadInterruptedTransferRestartsAtMirror(t *testing.T) {
	content := []byte("complete file")
	hash := sha256.Sum256(content)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/primary" {
			w.Header().Set("Content-Length", "100")
			_, _ = w.Write([]byte("partial"))
			return
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()
	m := New(t.TempDir(), Registry{})
	m.client = server.Client()
	path := filepath.Join(t.TempDir(), "download")
	if err := m.download(context.Background(), Artifact{URL: server.URL + "/primary", Mirrors: []string{server.URL + "/mirror"}, SHA256: hex.EncodeToString(hash[:])}, path); err != nil {
		t.Fatal(err)
	}
	if err := verifyFile(path, hex.EncodeToString(hash[:])); err != nil {
		t.Fatal(err)
	}
}

func TestDownloadKeepsMakingProgressAndFallsBackOnStall(t *testing.T) {
	for _, stall := range []bool{false, true} {
		name := "progress"
		if stall {
			name = "stall"
		}
		t.Run(name, func(t *testing.T) {
			content := []byte("0123456789")
			hash := sha256.Sum256(content)
			var mirrors atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/mirror" {
					mirrors.Add(1)
					_, _ = w.Write(content)
					return
				}
				w.WriteHeader(200)
				w.(http.Flusher).Flush()
				if stall {
					<-r.Context().Done()
					return
				}
				for _, b := range content {
					_, _ = w.Write([]byte{b})
					w.(http.Flusher).Flush()
					time.Sleep(30 * time.Millisecond)
				}
			}))
			defer server.Close()
			m := New(t.TempDir(), Registry{})
			m.client.Transport = server.Client().Transport
			m.downloadIdleTimeout = 150 * time.Millisecond
			path := filepath.Join(t.TempDir(), "download")
			artifact := Artifact{URL: server.URL + "/primary", Mirrors: []string{server.URL + "/mirror"}, SHA256: hex.EncodeToString(hash[:])}
			if err := m.download(context.Background(), artifact, path); err != nil {
				t.Fatal(err)
			}
			if got, _ := os.ReadFile(path); string(got) != string(content) {
				t.Fatalf("content %q", got)
			}
			want := int32(0)
			if stall {
				want = 1
			}
			if mirrors.Load() != want {
				t.Fatalf("mirror calls=%d", mirrors.Load())
			}
		})
	}
}

func TestDownloadRejectsWrongBytesBeforeTryingMirror(t *testing.T) {
	content := []byte("verified bytes")
	hash := sha256.Sum256(content)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bad" {
			_, _ = w.Write([]byte("wrong bytes"))
			return
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()
	m := New(t.TempDir(), Registry{})
	m.client = server.Client()
	var events []Progress
	ctx := WithProgress(context.Background(), func(event Progress) { events = append(events, event) })
	path := filepath.Join(t.TempDir(), "download")
	if err := m.download(ctx, Artifact{URL: server.URL + "/bad", Mirrors: []string{server.URL + "/good"}, SHA256: hex.EncodeToString(hash[:])}, path); err != nil {
		t.Fatal(err)
	}
	if err := verifyFile(path, hex.EncodeToString(hash[:])); err != nil {
		t.Fatal(err)
	}
	failed := false
	for _, e := range events {
		if e.Stage == "source-failed" {
			failed = true
		}
	}
	if !failed {
		t.Fatal("fallback not reported")
	}
}
