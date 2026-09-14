package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fu9zhou/finish-bit/pkg/app"
	"github.com/fu9zhou/finish-bit/pkg/packagemanager"
)

type serviceState struct {
	Host  string `json:"host"`
	Token string `json:"token"`
	PID   int    `json:"pid"`
}

func (s serviceState) url() string { return "http://" + s.Host + "/#" + s.Token }

// serviceRequest verifies the token with the actual listener, never a PID alone.
func serviceRequest(ctx context.Context, s serviceState, method string) error {
	host, _, err := net.SplitHostPort(s.Host)
	if err != nil || host != "127.0.0.1" || len(s.Token) != 64 {
		return fmt.Errorf("invalid UI service state")
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://"+s.Host+"/api/service", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.Token)
	client := &http.Client{Timeout: time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("UI service returned %s", resp.Status)
	}
	return nil
}

func readService(path string) (serviceState, error) {
	var state serviceState
	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}

func waitService(ctx context.Context, path string) (serviceState, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		s, err := readService(path)
		if err == nil && serviceRequest(ctx, s, http.MethodGet) == nil {
			return s, nil
		}
		select {
		case <-ctx.Done():
			return serviceState{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

// Manage owns only local HTTP service lifecycle. Capability execution stays in app.
func Manage(ctx context.Context, application *app.App, action string, port int, open bool, output io.Writer) error {
	root, err := packagemanager.DefaultRoot()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, "ui")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "service.json")
	if action == "run" {
		unlock, err := lockServiceFile(filepath.Join(dir, "running.lock"))
		if err != nil {
			return fmt.Errorf("UI service already running: %w", err)
		}
		defer unlock()
		defer os.Remove(path)
		return serve(ctx, application, port, false, output, func(host, token string) error {
			data, err := json.Marshal(serviceState{Host: host, Token: token, PID: os.Getpid()})
			if err != nil {
				return err
			}
			return os.WriteFile(path, data, 0600)
		})
	}
	// Serialize start/stop commands; the daemon uses a separate lifetime lock.
	var unlock func()
	deadline := time.Now().Add(15 * time.Second)
	for {
		unlock, err = lockServiceFile(filepath.Join(dir, "control.lock"))
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("another UI command is busy: %w", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	defer unlock()
	state, readErr := readService(path)
	running := readErr == nil && serviceRequest(ctx, state, http.MethodGet) == nil
	switch action {
	case "status":
		if !running {
			fmt.Fprintln(output, "FinishBit UI: stopped")
			return nil
		}
	case "stop":
		if !running {
			fmt.Fprintln(output, "FinishBit UI: stopped")
			return nil
		}
		// Closing the server can close this response too; verify release of its lifetime lock.
		_ = serviceRequest(ctx, state, http.MethodDelete)
		deadline := time.NewTimer(10 * time.Second)
		defer deadline.Stop()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			release, e := lockServiceFile(filepath.Join(dir, "running.lock"))
			if e == nil {
				release()
				fmt.Fprintln(output, "FinishBit UI: stopped")
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-deadline.C:
				return fmt.Errorf("UI service did not stop in time")
			case <-ticker.C:
			}
		}
	case "start":
		if running && port != 0 {
			_, existing, _ := net.SplitHostPort(state.Host)
			if existing != strconv.Itoa(port) {
				return fmt.Errorf("UI already runs on port %s; use fnsh ui stop before changing ports", existing)
			}
		}
		if !running {
			executable, err := os.Executable()
			if err != nil {
				return err
			}
			log, err := os.OpenFile(filepath.Join(dir, "service.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}
			command := exec.Command(executable, "ui", "run", "--port", strconv.Itoa(port), "--no-open")
			command.Stdout, command.Stderr = log, log
			detachService(command)
			err = command.Start()
			_ = log.Close()
			if err != nil {
				return err
			}
			_ = command.Process.Release()
			ready, cancel := context.WithTimeout(ctx, 15*time.Second)
			state, err = waitService(ready, path)
			cancel()
			if err != nil {
				return fmt.Errorf("UI did not become ready; see %s: %w", filepath.Join(dir, "service.log"), err)
			}
		}
	default:
		return errors.New("unknown UI service action")
	}
	fmt.Fprintf(output, "FinishBit UI: running\nPID: %d\nURL: %s\nStop: fnsh ui stop\n", state.PID, state.url())
	if action == "start" && open {
		if err := openBrowser(state.url()); err != nil {
			fmt.Fprintf(output, "Open the URL above in your browser: %v\n", err)
		}
	}
	return nil
}
