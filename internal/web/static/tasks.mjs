export function packageTaskBusy(result) {
  return ["running", "recovering", "unknown"].includes(result?.status);
}

// Snapshots are newest first. Never let an older completion replace active work.
export function packageTaskResults(tasks) {
  const results = new Map();
  const ordered = [...tasks].sort((a,b) => {
    if (packageTaskBusy(a) !== packageTaskBusy(b)) return packageTaskBusy(a) ? -1 : 1;
    return String(b.started || "").localeCompare(String(a.started || ""));
  });
  for (const task of ordered) {
    if (results.has(task.package)) continue;
    results.set(task.package, {
      id: task.package, taskId: task.id, packageAction: true,
      status: task.status, origin: task.origin,
      updated: task.updated, checked: task.checked,
      progress: task.status === "running" ? task.progress : null,
      lastProgress: ["running", "done"].includes(task.status) ? null : task.progress,
      error: task.error ? { message: task.error } : null,
    });
  }
  return results;
}
