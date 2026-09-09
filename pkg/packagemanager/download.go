package packagemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func (m *Manager) download(ctx context.Context, artifact Artifact, destination string) error {
	var failures []error
	for _, source := range append([]string{artifact.URL}, artifact.Mirrors...) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
			return err
		}
		err := m.downloadOnce(ctx, source, destination)
		if err == nil {
			err = verifyFile(destination, artifact.SHA256)
		}
		if err == nil {
			return nil
		}
		failures = append(failures, err)
		reportProgress(ctx, Progress{Stage: "source-failed", Source: sourceHost(source)})
	}
	return fmt.Errorf("all package download sources failed: %w", errors.Join(failures...))
}

func sourceHost(source string) string {
	u, err := url.Parse(source)
	if err != nil {
		return "source"
	}
	return u.Hostname()
}

func (m *Manager) downloadOnce(ctx context.Context, source, destination string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	idle := m.downloadIdleTimeout
	if idle <= 0 {
		idle = 45 * time.Second
	}
	// Reset on actual bytes, rather than abandoning a healthy large download.
	timer := time.AfterFunc(idle, cancel)
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
		return fmt.Errorf("download from %s: %w", host, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download from %s: server returned %s", host, response.Status)
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
			return fmt.Errorf("download from %s interrupted after %d bytes: %w", host, event.Bytes, readErr)
		}
	}
	reportProgress(ctx, event)
	return file.Close()
}
