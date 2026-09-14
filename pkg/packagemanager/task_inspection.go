package packagemanager

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TaskInspection reports present evidence separately from a historical snapshot.
// Hashing is on demand, not part of the one-second task heartbeat poll.
type TaskInspection struct {
	TaskID              string    `json:"task_id"`
	Status              string    `json:"status"`
	Checked             time.Time `json:"checked"`
	Installation        string    `json:"installation"`
	InstalledVersion    string    `json:"installed_version,omitempty"`
	InstalledAt         time.Time `json:"installed_at,omitempty"`
	CompletionConfirmed bool      `json:"completion_confirmed"`
	Cache               string    `json:"cache"`
	ResidualDownloads   int       `json:"residual_downloads"`
	ResidualBytes       int64     `json:"residual_bytes"`
	TaskDownloadBytes   int64     `json:"task_download_bytes"`
	TaskDownloadFound   bool      `json:"task_download_found"`
	ResidualsChecked    bool      `json:"residuals_checked"`
	Errors              []string  `json:"errors,omitempty"`
}

func (m *Manager) InspectTask(id string) (TaskInspection, error) {
	decoded, err := hex.DecodeString(id)
	if err != nil || len(decoded) != 16 {
		return TaskInspection{}, os.ErrNotExist
	}
	path := filepath.Join(m.root, "tasks", id+".json")
	data, err := readTaskSnapshot(path)
	if err != nil {
		return TaskInspection{}, err
	}
	var task Task
	if json.Unmarshal(data, &task) != nil || task.ID != id {
		return TaskInspection{}, os.ErrNotExist
	}
	pkg, ok := m.registry.Find(task.Package)
	if !ok {
		return TaskInspection{}, os.ErrNotExist
	}
	result := TaskInspection{TaskID: id, Status: task.Status, Checked: time.Now().UTC(), Installation: "unchecked", Cache: "unchecked"}
	unlock, err := m.lockPackage(task.Package)
	if err != nil {
		result.Status = "unknown"
		if packageLockBusy(err) {
			result.Status = "recovering"
		} else {
			result.Errors = append(result.Errors, err.Error())
		}
		return result, nil
	}
	defer unlock()
	// Catch a completion racing the initial read before inspecting files.
	data, err = readTaskSnapshot(path)
	if err != nil {
		return TaskInspection{}, err
	}
	if json.Unmarshal(data, &task) != nil || task.ID != id || task.Package != pkg.Name {
		return TaskInspection{}, os.ErrNotExist
	}
	result.Status = task.Status
	if task.Status == "running" {
		if time.Since(task.Updated) <= taskLease {
			return result, nil
		}
		result.Status = "interrupted"
	}
	installed, err := m.Info(pkg.Name)
	if errors.Is(err, os.ErrNotExist) {
		result.Installation = "missing"
	} else if err != nil {
		result.Installation = "unverified"
		result.Errors = append(result.Errors, err.Error())
	} else {
		result.InstalledVersion, result.InstalledAt = installed.Version, installed.Installed
		if err := m.checkInstalled(pkg, installed); err != nil {
			result.Installation = "unverified"
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Installation = "verified"
			result.CompletionConfirmed = installed.TaskID == id
			if result.CompletionConfirmed {
				result.Status = "done"
			}
		}
	}
	if artifact, ok := pkg.Artifacts[PlatformKey()]; ok {
		cache := filepath.Join(m.root, "cache", strings.ToLower(artifact.SHA256))
		if info, err := os.Lstat(cache); errors.Is(err, os.ErrNotExist) {
			result.Cache = "missing"
		} else if err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else if !info.Mode().IsRegular() {
			result.Cache = "unverified"
		} else if err := verifyFile(cache, artifact.SHA256); err != nil {
			result.Cache = "unverified"
		} else {
			result.Cache = "verified"
		}
	}
	entries, err := os.ReadDir(filepath.Join(m.root, "packages", pkg.Name))
	result.ResidualsChecked = err == nil || errors.Is(err, os.ErrNotExist)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		result.Errors = append(result.Errors, err.Error())
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".install-") || strings.HasSuffix(entry.Name(), ".previous") {
			continue
		}
		info, err := os.Lstat(filepath.Join(m.root, "packages", pkg.Name, entry.Name(), "download"))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			result.ResidualsChecked = false
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		result.ResidualDownloads++
		result.ResidualBytes += info.Size()
		if strings.HasPrefix(entry.Name(), ".install-"+id+"-") {
			result.TaskDownloadFound = true
			result.TaskDownloadBytes += info.Size()
		}
	}
	result.Checked = time.Now().UTC()
	return result, nil
}
