package web

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestServiceLifecycleAuthentication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan serviceState, 1)
	done := make(chan error, 1)
	go func() {
		done <- serve(ctx, nil, 0, false, io.Discard, func(host, token string) error { ready <- serviceState{Host: host, Token: token}; return nil })
	}()
	var state serviceState
	select {
	case state = <-ready:
	case <-time.After(5 * time.Second):
		t.Fatal("not ready")
	}
	deadline, cancelDeadline := context.WithTimeout(ctx, 5*time.Second)
	defer cancelDeadline()
	for serviceRequest(deadline, state, http.MethodGet) != nil {
		if deadline.Err() != nil {
			t.Fatal(deadline.Err())
		}
		time.Sleep(10 * time.Millisecond)
	}
	wrong := state
	wrong.Token = strings.Repeat("0", 64)
	if serviceRequest(ctx, wrong, http.MethodDelete) == nil {
		t.Fatal("accepted invalid token")
	}
	if err := serviceRequest(ctx, state, http.MethodGet); err != nil {
		t.Fatal(err)
	}
	_ = serviceRequest(ctx, state, http.MethodDelete)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("did not stop")
	}
}

func TestServiceLockReleased(t *testing.T) {
	path := t.TempDir() + "/service.lock"
	release, err := lockServiceFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if other, err := lockServiceFile(path); err == nil {
		other()
		release()
		t.Fatal("duplicate lock")
	}
	release()
	release, err = lockServiceFile(path)
	if err != nil {
		t.Fatal(err)
	}
	release()
}
