// Package web serves the embedded local UI and adapts HTTP to pkg/app.
package web

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fu9zhou/finish-bit/pkg/app"
	"github.com/fu9zhou/finish-bit/pkg/operation"
)

//go:embed static/*
var assets embed.FS

// Serve binds only to loopback. The fragment secret never appears in HTTP URLs.
func Serve(ctx context.Context, application *app.App, port int, open bool, output io.Writer) error {
	return serve(ctx, application, port, open, output, nil)
}

func serve(ctx context.Context, application *app.App, port int, open bool, output io.Writer, ready func(string, string) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if port < 0 || port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
	}
	listener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	defer listener.Close()
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return err
	}
	token := hex.EncodeToString(secret)
	host := listener.Addr().String()
	address := "http://" + host + "/#" + token
	handler := handlerWithContext(ctx, application, host, token)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/service" && r.Host == host && (r.Header.Get("Origin") == "" || r.Header.Get("Origin") == "http://"+host) && subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+token)) == 1 {
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusNoContent)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
				go cancel()
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}
		handler.ServeHTTP(w, r)
	}), ReadHeaderTimeout: 5 * time.Second, BaseContext: func(net.Listener) context.Context { return ctx }}
	if ready != nil {
		if err := ready(host, token); err != nil {
			return err
		}
	}
	fmt.Fprintf(output, "FinishBit UI: %s\nPress Ctrl+C to stop.\n", address)
	if open {
		if err := openBrowser(address); err != nil {
			fmt.Fprintf(output, "Open the URL above in your browser: %v\n", err)
		}
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = server.Close()
		case <-done:
		}
	}()
	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func openBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		command = exec.Command("open", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}

func newHandler(application *app.App, host, token string) http.Handler {
	return handlerWithContext(context.Background(), application, host, token)
}

func handlerWithContext(serviceContext context.Context, application *app.App, host, token string) http.Handler {
	files, _ := fs.Sub(assets, "static")
	static := http.FileServer(http.FS(files))
	var busy sync.Mutex
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		if r.Host != host || (r.Header.Get("Origin") != "" && r.Header.Get("Origin") != "http://"+host) {
			http.Error(w, "Forbidden origin", 403)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			info, e := fs.Stat(files, path)
			if !fs.ValidPath(path) || e != nil || info.IsDir() {
				http.NotFound(w, r)
				return
			}
			static.ServeHTTP(w, r)
			return
		}
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+token)) != 1 {
			http.Error(w, "Unauthorized", 401)
			return
		}
		if r.URL.Path == "/api/package-tasks" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if r.Method == "GET" {
				if id := r.URL.Query().Get("inspect"); id != "" {
					inspection, err := application.InspectPackageTask(id)
					if err != nil {
						http.Error(w, "Cannot inspect package task", http.StatusNotFound)
						return
					}
					_ = json.NewEncoder(w).Encode(inspection)
					return
				}
				tasks, err := application.PackageTasks()
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				_ = json.NewEncoder(w).Encode(tasks)
				return
			}
			if r.Method != "POST" {
				w.WriteHeader(405)
				return
			}
			var request struct {
				Name   string `json:"name"`
				Repair bool   `json:"repair"`
			}
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&request); err != nil {
				http.Error(w, "Invalid request", 400)
				return
			}
			var extra any
			if err := decoder.Decode(&extra); err != io.EOF {
				http.Error(w, "Expected one JSON request", 400)
				return
			}
			task, err := application.StartPackageInstall(serviceContext, request.Name, request.Repair)
			if err != nil {
				status := http.StatusInternalServerError
				if errors.Is(err, app.ErrPackageTaskRunning) {
					status = http.StatusConflict
				} else if operation.AsError(err).Code == operation.CodeInvalidInput {
					status = http.StatusBadRequest
				}
				http.Error(w, err.Error(), status)
				return
			}
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(task)
			return
		}
		if r.URL.Path == "/api/browse" {
			if r.Method != "GET" {
				w.WriteHeader(405)
				return
			}
			listing, err := application.BrowseFiles(r.URL.Query().Get("path"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if err != nil {
				w.WriteHeader(400)
				_ = json.NewEncoder(w).Encode(operation.AsError(err))
				return
			}
			_ = json.NewEncoder(w).Encode(listing)
			return
		}
		if r.URL.Path != "/api/catalog" && r.URL.Path != "/api/action" {
			http.NotFound(w, r)
			return
		}
		if !busy.TryLock() {
			http.Error(w, "另一个任务正在执行，请稍后重试", 409)
			return
		}
		defer busy.Unlock()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.URL.Path == "/api/catalog" {
			if r.Method != "GET" {
				w.WriteHeader(405)
				return
			}
			packages, err := application.PackageCatalog()
			if err != nil {
				w.WriteHeader(500)
				_ = json.NewEncoder(w).Encode(operation.AsError(err))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"operations": application.Capabilities(), "packages": packages, "extensions": application.Extensions(), "platform": runtime.GOOS + " / " + runtime.GOARCH})
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		var request struct {
			Action  string         `json:"action"`
			Name    string         `json:"name"`
			Inputs  []string       `json:"inputs"`
			Options map[string]any `json:"options"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
		decoder.DisallowUnknownFields()
		decoder.UseNumber()
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "Invalid request: "+err.Error(), 400)
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			http.Error(w, "Expected one JSON request", 400)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		var stream sync.Mutex
		emit := func(kind string, data any) {
			stream.Lock()
			defer stream.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"type": kind, "data": data})
			_ = http.NewResponseController(w).Flush()
		}
		ctx := app.WithPackageProgress(r.Context(), func(p app.PackageProgress) { emit("progress", p) })
		emit("started", request.Name)
		var result any
		var err error
		switch request.Action {
		case "run":
			if request.Options["input-mode"] == "stdin" {
				err = &operation.Error{Code: operation.CodeInvalidInput, Message: "The web UI cannot read terminal stdin; choose literal or file input"}
			}
			for _, input := range request.Inputs {
				if input == "-" {
					err = &operation.Error{Code: operation.CodeInvalidInput, Message: "The web UI cannot read terminal stdin; provide text or a local file path"}
					break
				}
			}
			if err == nil {
				result, err = application.Execute(ctx, request.Name, operation.Request{Inputs: request.Inputs, Options: request.Options})
			}
		case "install":
			result, err = application.InstallPackage(ctx, request.Name)
		case "repair":
			result, err = application.RepairPackage(ctx, request.Name)
		case "remove":
			err = application.RemovePackage(request.Name)
		case "extension-install":
			result, err = application.InstallExtension(request.Name)
		case "extension-remove":
			err = application.RemoveExtension(request.Name)
		case "doctor":
			result = application.Doctor()
		default:
			err = &operation.Error{Code: operation.CodeInvalidInput, Message: "Unknown action"}
		}
		if err != nil {
			emit("error", operation.AsError(err))
			return
		}
		emit("result", result)
	})
}
