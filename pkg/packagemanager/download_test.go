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
