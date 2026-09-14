package packagemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func (m *Manager) download(ctx context.Context, artifact Artifact, destination string) error {
	var failures []error
	sources := append([]string{artifact.URL}, artifact.Mirrors...)
	for index, source := range sources {
		sourceCtx := context.WithValue(ctx, downloadSourceKey{}, Progress{SourceIndex: index + 1, SourceCount: len(sources)})
		var err error
		for attempt := 0; attempt < 2; attempt++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
				return err
			}
			err = m.downloadOnce(sourceCtx, source, destination)
			if err == nil {
				err = verifyFile(destination, artifact.SHA256)
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err == nil {
				return nil
			}
			var localError *os.PathError
			// Retrying the network cannot repair output or verification I/O.
			if errors.As(err, &localError) {
				return fmt.Errorf("local package file operation failed: %w", err)
			}
			var statusError *downloadHTTPError
			// Long Retry-After values skip this source, never retry it too early.
			if attempt != 0 || !errors.As(err, &statusError) || !statusError.retryable || statusError.delay > 2*time.Second {
				break
			}
			reportProgress(sourceCtx, Progress{Stage: "source-retrying", Source: sourceHost(source), Reason: err.Error()})
			timer := time.NewTimer(statusError.delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
		if cleanupErr := os.Remove(destination); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			return fmt.Errorf("remove failed download: %w", cleanupErr)
		}
		failures = append(failures, fmt.Errorf("source %d/%d (%s): %w", index+1, len(sources), sourceHost(source), err))
		next := ""
		if index+1 < len(sources) {
			next = sourceHost(sources[index+1])
		}
		reportProgress(sourceCtx, Progress{Stage: "source-failed", Source: sourceHost(source), Reason: err.Error(), NextSource: next})
	}
	return fmt.Errorf("all package download sources failed: %w", errors.Join(failures...))
}

type downloadHTTPError struct {
	status    string
	retryable bool
	delay     time.Duration
}

func (e *downloadHTTPError) Error() string { return "server returned " + e.status }

func retryDelay(value string) time.Duration {
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		if seconds > 2 {
			return 3 * time.Second
		}
		return max(time.Second, time.Duration(seconds)*time.Second)
	}
	if date, err := http.ParseTime(value); err == nil {
		return max(time.Second, time.Until(date))
	}
	return time.Second
}

func sourceHost(source string) string {
	u, err := url.Parse(source)
	if err != nil {
		return "source"
	}
	return u.Hostname()
}

func (m *Manager) downloadOnce(ctx context.Context, source, destination string) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	idle := m.downloadIdleTimeout
	if idle <= 0 {
		idle = 45 * time.Second
	}
	// Reset on actual bytes, rather than abandoning a healthy large download.
	timer := time.AfterFunc(idle, func() { cancel(fmt.Errorf("no download progress for %s", idle)) })
	defer timer.Stop()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "FinishBit/0.1 (+https://github.com/fu9zhou/finish-bit)")
	request.Header.Set("Accept", "application/octet-stream")
	host := sourceHost(source)
	reportProgress(ctx, Progress{Stage: "connecting", Source: host})
	response, err := m.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("download from %s stopped (%v): %w", host, context.Cause(ctx), err)
		}
		return fmt.Errorf("download from %s: %w", host, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		failure := &downloadHTTPError{status: response.Status, delay: retryDelay(response.Header.Get("Retry-After"))}
		switch response.StatusCode {
		case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			failure.retryable = true
		}
		return failure
	}
	if response.ContentLength > maxPackageBytes {
		return fmt.Errorf("package download exceeds size limit")
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create download: %w", err)
	}
	defer file.Close()
	event := Progress{Stage: "downloading", Source: host, Total: response.ContentLength}
	reportProgress(ctx, event)
	last := time.Now()
	buffer := make([]byte, 64*1024)
	for {
		n, readErr := response.Body.Read(buffer)
		if n > 0 {
			timer.Reset(idle)
			event.Bytes += int64(n)
			if event.Bytes > maxPackageBytes {
				return fmt.Errorf("package download exceeds size limit")
			}
			if _, err := file.Write(buffer[:n]); err != nil {
				return err
			}
			if time.Since(last) >= time.Second {
				reportProgress(ctx, event)
				last = time.Now()
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("download from %s stopped after %d bytes (%v): %w", host, event.Bytes, context.Cause(ctx), readErr)
			}
			return fmt.Errorf("download from %s interrupted after %d bytes: %w", host, event.Bytes, readErr)
		}
	}
	reportProgress(ctx, event)
	return file.Close()
}
