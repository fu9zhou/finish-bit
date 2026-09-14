import {
  t,
  language,
  setLanguage,
  operationName,
  packageSummary,
} from "./i18n.mjs";
import { getPackageTasks, inspectPackageTask, startPackageInstall, getCatalog, initializeSession, performAction } from "./api.mjs";
import { parseRoute, toolURL, categoryURL } from "./router.mjs";
import { el, button, link } from "./dom.mjs";
import { createDraft, createToolPage } from "./forms.mjs";
import { packageTaskResults, packageTaskBusy } from "./tasks.mjs";
import { categories, categoryFor, findTools, categoryDescriptions } from "./catalog-view.mjs";
import { categoryIcon } from "./icons.mjs";
import { enhanceSelects } from "./select.mjs";
import { renderProgress, formatBytes } from "./progress.mjs";
import { createFilePicker } from "./file-picker.mjs";

initializeSession();
const $ = (id) => document.getElementById(id);
const state = {
  catalog: { operations: [], packages: [], extensions: [] },
  drafts: new Map(),
  results: new Map(),
  packageResults: new Map(),
  inspections: new Map(),
  taskDetails: new Map(),
  taskSignature: "",
  checks: [],
  busy: false,
  connected: false,
  loading: true,
  expired: false,
  recent: readRecent(),
  query: "",
  queries: new Map(),
  extensionPath: "",
  lastResult: null,
};
let route = parseRoute(location.hash);

function localizeShell() {
  document.documentElement.lang = language() === "zh" ? "zh-CN" : "en";
  document
    .querySelectorAll("[data-i18n]")
    .forEach((e) => (e.textContent = t(e.dataset.i18n)));
  $("language").value = language();
  $("language").setAttribute("aria-label", t("language"));
  $("search").placeholder = t("searchHint");
  $("connection").classList.toggle("offline", !state.connected);
  $("refresh").disabled = state.busy;
  $("connection").textContent = t(state.connected ? "connected" : "offline");

}
function render() {
  localizeShell();
  const detail = route.page === "tool", categoryPage = route.page === "category";
  const currentOperation = detail ? state.catalog.operations.find(op => op.id === route.id) : null;
  const group = categoryPage ? categories.find(c => c.id === route.id) : currentOperation ? categoryFor(currentOperation) : null;
  const page = detail || categoryPage ? "tools" : route.page;
  document.querySelectorAll("[data-page]").forEach((a) => {
    a.classList.toggle("active", a.dataset.page === page);
    if (a.dataset.page === page) a.setAttribute("aria-current", "page");
    else a.removeAttribute("aria-current");
  });
  const breadcrumb = $("breadcrumb");
  breadcrumb.replaceChildren();
  if (page === "tools") {
    breadcrumb.append(link(t("tools"), "#tools"));
    if (group) {
      breadcrumb.append(el("span", "breadcrumb-divider", "/"), link(categoryLabel(group), categoryURL(group.id)));
      if (detail) breadcrumb.append(el("span", "breadcrumb-divider", "/"), el("span", "", operationName(currentOperation)));
    }
  } else breadcrumb.textContent = t(page);
  const back = $("page-back");
  back.replaceChildren();
  back.hidden = !detail && !categoryPage;
  if (!back.hidden) back.append(link("← " + (detail && group ? t("backCategory", {category:categoryLabel(group)}) : t("tools")), detail && group ? categoryURL(group.id) : "#tools", "back-link"));
  document.title = `${detail ? currentOperation ? operationName(currentOperation) : route.id : group ? categoryLabel(group) : t(page)} · FinishBit`;
  $("overview").hidden = detail || page === "not-found";
  $("toolbar").hidden =
    detail || !["tools", "packages", "extensions"].includes(page);
  $("title").textContent = group && categoryPage ? categoryLabel(group) : t(page + "Title");
  $("subtitle").textContent = group && categoryPage ? categoryDescription(group) : t(page + "Subtitle");
  $("search").placeholder = group && categoryPage ? t("categorySearch", {category:categoryLabel(group)}) : t("searchHint");
  document.querySelector("main").dataset.page = detail ? "tool" : page;
  $("page-count").hidden = !["extensions", "packages"].includes(page) && !(page === "tools" && !detail && !categoryPage);
  const installedCount = page === "packages" ? state.catalog.packages.filter(p => p.installed).length : state.catalog.extensions.length;
  $("page-count").textContent = ["extensions", "packages"].includes(page)
    ? `${t("installed")} ${state.connected ? installedCount : "—"}`
    : t("toolTotal", { count: state.connected ? state.catalog.operations.length : "—" });
  const content = $("content");
  content.replaceChildren();
  if (!state.catalog.operations.length && !state.connected) {
    $("toolbar").hidden = true;
    const status = el("div", "catalog-state");
    status.append(el("h2", "", t(state.loading ? "loading" : "catalogUnavailable")));
    if (!state.loading) status.append(el("p", "", t("reconnectHelp")), button(t("refresh"), "primary", load));
    content.append(status);
    $("global-result").hidden = true;
    enhanceSelects();
    return;
  }
  if (detail) {
    const op = state.catalog.operations.find((o) => o.id === route.id);
    if (!op) {
      content.append(
        el("div", "empty", state.connected ? t("notFound") : t("loading")),
      );
      enhanceSelects();
      return;
    }
    if (!state.drafts.has(op.id)) state.drafts.set(op.id, createDraft(op));
    content.append(
      createToolPage(op, {
        draft: state.drafts.get(op.id),
        packages: state.catalog.packages,
        busy: !state.connected || state.busy || (op.requirements || []).some(r => packageTaskBusy(state.packageResults.get(r.package))),
        onRun: (request) => action("run", op.id, request),
        onInstall: (name) => action("install", name),
        onChange: () => render(),
      }),
    );
    renderResult($("tool-result"), toolResult(op.id));
  } else if (categoryPage) {
    if (group) renderCategory(content, group);
    else content.append(el("h2", "", t("categoryUnavailable")), link(t("toolboxHome"), "#tools", "back-link"));
  } else if (page === "tools") renderTools(content);
  else if (page === "packages") renderPackages(content);
  else if (page === "extensions") renderExtensions(content);
  else if (page === "health") renderHealth(content);
  else
    content.append(
      el("div", "empty", t("notFound")),
      link(t("back"), "#tools", "back-link"),
    );
  const global = $("global-result");
  global.hidden = detail || page === "tools" || !state.lastResult || (page === "packages" && state.lastResult.packageAction);
  if (!global.hidden) renderResult(global, state.lastResult);
  enhanceSelects();
}
function matches(value) {
  return JSON.stringify(value)
    .toLowerCase()
    .includes(state.query.toLowerCase());
}
function categoryLabel(category) { return category[language()]; }
function readRecent() {
  try { const value = JSON.parse(localStorage.getItem("finishbit-recent-tools") || "[]"); return Array.isArray(value) ? value.filter(id => typeof id === "string").slice(0, 6) : []; } catch { return []; }
}
function rememberTool(id) {
  state.recent = [id, ...state.recent.filter(value => value !== id)].slice(0, 6);
  try { localStorage.setItem("finishbit-recent-tools", JSON.stringify(state.recent)); } catch {}
}
function categoryDescription(category) { return categoryDescriptions[category.id][language() === "zh" ? 0 : 1]; }
function renderToolCards(content, ops) {
  const grid = el("div", "grid");
  for (const op of ops) {
    const card = link("", toolURL(op.id), "tool-card");
    card.onclick = () => rememberTool(op.id);
    const top = el("div", "tile-top");
    const missing = (op.requirements || []).some(r => !state.catalog.packages.find(p => p.name === r.package)?.installed);
    top.append(el("h3", "", operationName(op)), el("span", "badge", t(missing ? "missing" : op.requirements?.length ? "dependenciesReady" : "builtin")));
    card.append(top);
    const foot = el("div", "card-bottom");
    foot.append(el("span", "", op.id), el("b", "", "→"));
    card.append(foot); grid.append(card);
  }
  content.append(grid);
}
function renderSearchResults(content, ops) {
  const heading = el("div", "section-heading");
  heading.append(el("h2", "", `${t("searchResults")} · ${ops.length}`));
  heading.append(button(t("clearFilters"), "text-button", () => {
    state.query = ""; state.queries.delete(queryKey(route)); $("search").value = ""; render(); $("search").focus();
  }));
  content.append(heading);
  if (ops.length) renderToolCards(content, ops);
  else content.append(el("p", "empty", t("empty")));
}
function renderCategory(content, category) {
  if (!state.connected) content.append(el("p", "stale-catalog", t("staleCatalog")));
  const ops = findTools(state.catalog.operations, state.query, category.id);
  if (state.query.trim()) { renderSearchResults(content, ops); return; }
  content.append(el("p", "section-label", `${ops.length} ${t("capabilities")}`));
  renderToolCards(content, ops);
}
function renderTools(content) {
  if (!state.connected) content.append(el("p", "stale-catalog", t("staleCatalog")));
  if (state.query.trim()) {
    renderSearchResults(content, findTools(state.catalog.operations, state.query, ""));
    return;
  }
  const recent = state.recent.map(id => state.catalog.operations.find(op => op.id === id)).filter(Boolean);
  if (recent.length) {
    const row = el("section", "recent-tools");
    row.append(el("h2", "", t("recent")));
    for (const op of recent) {
      const item = link(operationName(op), toolURL(op.id), "recent-tool");
      item.onclick = () => rememberTool(op.id); row.append(item);
    }
    content.append(row);
  }
  const grid = el("div", "category-grid");
  for (const category of categories) {
    const count = state.catalog.operations.filter(op => categoryFor(op).id === category.id).length;
    const card = link("", categoryURL(category.id), "category-card");
    card.dataset.category = category.id;
    const copy = el("div", "category-card-copy"), heading = el("div", "category-card-heading");
    const countLabel = el("span", "category-count", count);
    countLabel.setAttribute("aria-label", `${count} ${t("capabilities")}`);
    heading.append(el("h2", "", categoryLabel(category)), countLabel);
    copy.append(heading, el("p", "", categoryDescription(category)));
    card.append(categoryIcon(category.id), copy);
    grid.append(card);
  }
  content.append(grid);
}
function queryKey(value) { return value.page === "category" ? `category:${value.id}` : value.page; }
function renderPackages(content) {
  const packages = state.catalog.packages.filter((p) =>
    matches([p.name, p.description]),
  );
  for (const p of packages) {
    const result = state.packageResults.get(p.name);
    const taskBusy = packageTaskBusy(result) || state.inspections.get(result?.taskId)?.pending;
    const row = el("article", "package-row"),
      info = el("div", "package-info");
    info.append(
      el("h3", "", p.name),
      el("p", "", language() === "en" ? p.description : packageSummary(p.name)),
      el("code", "", `v${p.version} · ${p.license}`),
    );
    row.append(
      info,
      el(
        "span",
        "badge",
        t(
          taskBusy ? (result.status === "running" ? "running" : "taskChecking") :
          p.installed
            ? (p.needs_repair ? "needsRepair" : "installed")
            : result?.status === "interrupted" ? "taskStopped"
            : p.supported
              ? "notInstalled"
              : "unsupported",
        ),
      ),
    );
    const actions = el("div", "row-actions");
    if (p.supported) {
      const b = button(
        t(taskBusy ? (result.status === "running" ? "running" : "taskChecking") : p.installed ? "repair" : result?.status === "interrupted" ? "retryInstall" : "install"),
        p.installed ? "secondary" : "primary",
        () => action(p.installed ? "repair" : "install", p.name),
      );
      b.disabled = !state.connected || state.busy || taskBusy;
      actions.append(b);
    }
    if (p.installed) {
      const b = button(t("remove"), "danger", () => {
        if (confirm(t("removeConfirm", { name: p.name })))
          action("remove", p.name);
      });
      b.disabled = !state.connected || state.busy || taskBusy;
      actions.append(b);
    }
    row.append(actions);
    const activity = el("div", "package-activity");
    activity.id = "package-result-" + p.name;
    activity.hidden = !result;
    if (result) renderResult(activity, result);
    row.append(activity);
    content.append(row);
  }
  if (!packages.length) content.append(el("div", "empty", t("empty")));
}
function renderExtensions(content) {
  const form = el("form", "extension-form"),
    label = el("label", "", t("extensionPath")),
    input = el("input");
  input.required = true;
  input.value = state.extensionPath;
  input.placeholder = t("pathHint");
  input.oninput = () => (state.extensionPath = input.value);
  label.append(input);
  const picker = el("div"),
    choose = button(t("browseDirectory"), "secondary", () => {
      picker.replaceChildren(
        createFilePicker({
          kind: "directory",
          onSelect: (path) => {
            state.extensionPath = path;
            input.value = path;
            picker.replaceChildren();
          },
          onClose: () => picker.replaceChildren(),
        }),
      );
    });
  const submit = el("button", "primary", t("installExtension"));
  submit.type = "submit";
  submit.disabled = !state.connected || state.busy;
  choose.disabled = !state.connected || state.busy;
  form.append(label, choose, submit);
  form.onsubmit = (e) => {
    e.preventDefault();
    action("extension-install", input.value);
  };
  content.append(form, picker);
  const extensions = (state.catalog.extensions || []).filter((e) =>
    matches(e.name),
  );
  for (const ext of extensions) {
    const row = el("div", "package-row"),
      info = el("div", "package-info");
    info.append(
      el("h3", "", ext.name),
      el(
        "p",
        "",
        `v${ext.version} · ${(ext.operations || []).length} ${t("capabilities")}`,
      ),
    );
    const remove = button(t("remove"), "danger", () => {
      if (confirm(t("removeConfirm", { name: ext.name })))
        action("extension-remove", ext.name);
    });
    remove.disabled = !state.connected || state.busy;
    row.append(info, remove);
    content.append(row);
  }
  if (!extensions.length) content.append(el("div", "empty", t("noExtensions")));
}
function renderHealth(content) {
  const header = el("div", "health-heading");
  header.append(el("h2", "", t("checkReady")));
  const check = button(t("check"), "primary", () =>
    action("doctor", "environment"),
  );
  check.disabled = !state.connected || state.busy;
  header.append(check);
  content.append(header);
  for (const c of state.checks) {
    const row = el("article", "package-row"),
      info = el("div", "package-info");
    info.append(el("h3", "", c.name), el("p", "", c.message));
    row.append(
      info,
      el("span", c.ok ? "badge" : "badge danger", t(c.ok ? "ok" : "attention")),
    );
    content.append(row);
  }
}
function renderResult(target, result) {
  if (!target) return;
  if (result?.packageAction && result.taskId) {
    renderPackageActivity(target, result);
    return;
  }
  delete target.dataset.taskId;
  target.replaceChildren();
  target.append(el("h2", "", t("result")));
  if (!result) {
    target.append(el("p", "", t("noResult")));
    return;
  }
  target.append(el("code", "", result.id));
  const status = el("p", "result-status", t(result.status === "interrupted" ? "taskInterrupted" : result.status));
  status.setAttribute("role", "status");
  target.append(status);
  if (result.origin) target.append(el("p", "progress-detail", t("origin-" + result.origin)));
  if (result.progress) target.append(renderProgress(result.progress, result.status === "running" && state.connected, !state.connected));
  if (result.lastProgress) target.append(renderProgress(result.lastProgress, false, true));
  if (result.packageAction && result.taskId) {
    const timestamps = [["lastUpdated", result.updated], ["lastChecked", result.checked]]
      .filter(([, value]) => value && Number.isFinite(Date.parse(value)))
      .map(([key, value]) => `${t(key)}: ${new Date(value).toLocaleString(language() === "zh" ? "zh-CN" : "en-US")}`);
    if (timestamps.length) target.append(el("p", "progress-detail", timestamps.join(" · ")));
    const inspection = state.inspections.get(result.taskId);
    if (inspection) renderInspection(target, inspection);
  }
  const pre = el(
    "pre",
    "",
    result.error
      ? result.error.message +
          (result.error.suggestion
            ? "\n\n" + t("suggestion") + ": " + result.error.suggestion
            : "")
      : result.data === undefined
        ? ""
        : result.data === null
          ? t("success")
          : JSON.stringify(result.data, null, 2),
  );
  pre.tabIndex = 0;
  target.append(pre);
}
function renderPackageActivity(target, result) {
  // Keep disclosure controls mounted during polling so focus and clicks survive.
  if (target.dataset.taskId !== result.taskId) {
    target.replaceChildren();
    target.dataset.taskId = result.taskId;
    const headline = el("div", "task-headline");
    const status = el("span", "task-status");
    status.setAttribute("role", "status");
    const body = el("div", "task-detail-body");
    body.id = `${target.id}-history-${result.taskId}`;
    const check = button(t("showHistory"), "task-link task-check", () => {
      const open = !state.taskDetails.get(result.taskId);
      state.taskDetails.set(result.taskId, open);
      body.hidden = !open;
      check.textContent = t(open ? "hideHistory" : "showHistory");
      check.setAttribute("aria-expanded", String(open));
      const inspection = state.inspections.get(result.taskId);
      if (open && (!inspection || inspection.error)) inspectHistory(result.taskId);
    });
    check.setAttribute("aria-controls", body.id);
    const meta = el("div", "task-meta");
    meta.append(el("span", "task-times"), check);
    const controls = el("div", "task-controls");
    controls.append(el("strong", "task-percent"));
    headline.append(status, controls);
    target.append(headline, el("div", "task-live"), el("p", "task-notice"), meta, body);
  }
  const live = result.status === "running" && state.connected;
  let label = t(result.status === "done" ? "taskCompleted" : result.status === "interrupted" ? "taskStopped" : result.status);
  if (live && result.progress?.Stage) {
    const stage = t("stage-" + result.progress.Stage);
    if (!stage.startsWith("stage-")) label = stage;
  }
  if (!state.connected) label = t("taskChecking");
  const origin = result.origin ? t("short-origin-" + result.origin) : "";
  target.querySelector(".task-status").textContent = origin ? `${label}（${origin}）` : label;
  const progress = target.querySelector(".task-live");
  const percent = target.querySelector(".task-percent");
  percent.textContent = "";
  progress.replaceChildren();
  if (live && result.progress) {
    const view = renderProgress(result.progress, true);
    percent.textContent = view.querySelector(".progress-heading strong")?.textContent || "";
    view.querySelector(".progress-heading")?.remove();
    progress.append(view);
  }
  const notice = target.querySelector(".task-notice");
  notice.textContent = result.error?.message || (result.status === "interrupted" ? t("stoppedBrief") : "");
  notice.hidden = !notice.textContent;
  const body = target.querySelector(".task-detail-body");
  body.replaceChildren();
  if (!live && (result.lastProgress || result.progress)) body.append(renderProgress(result.lastProgress || result.progress, false, true));
  const timestamps = [[result.status === "done" ? "completedAt" : "lastUpdated", state.inspections.get(result.taskId)?.completion_confirmed ? state.inspections.get(result.taskId).installed_at : result.updated]]
    .filter(([, value]) => value && Number.isFinite(Date.parse(value)))
    .map(([key, value]) => `${t(key)}: ${new Date(value).toLocaleString(language() === "zh" ? "zh-CN" : "en-US")}`);
  target.querySelector(".task-times").textContent = timestamps.join(" · ");
  if (result.error?.suggestion) body.append(el("p", "progress-detail", result.error.suggestion));
  const inspection = state.inspections.get(result.taskId);
  if (inspection) renderInspection(body, inspection);
  const open = !!state.taskDetails.get(result.taskId);
  body.hidden = !open || !body.childElementCount;
  const check = target.querySelector(".task-check");
  check.textContent = t(open ? "hideHistory" : "showHistory");
  check.setAttribute("aria-expanded", String(open));
  check.disabled = !open && !inspection && (!state.connected || packageTaskBusy(result));
}
function renderInspection(target, inspection) {
  const section = el("div", "history-evidence");
  if (inspection.pending) section.append(el("p", "progress-detail", t("historyChecking")));
  else if (inspection.error) section.append(el("p", "progress-detail", inspection.error));
  else {
    section.append(el("p", "progress-detail", t(inspection.completion_confirmed ? "verifiedCompletionBrief" : "evidence-installation-" + inspection.installation)));
    section.append(el("p", "progress-detail", t("evidence-cache-" + inspection.cache)));
    if (inspection.residual_downloads) section.append(el("p", "progress-detail", t("historyResidual", { count:inspection.residual_downloads, size:formatBytes(inspection.residual_bytes) })));
    if (inspection.task_download_found) section.append(el("p", "progress-detail", t("historyTaskBytes", {size:formatBytes(inspection.task_download_bytes)})));
    for (const error of inspection.errors || []) section.append(el("p", "progress-detail", error));
    if (inspection.checked) section.append(el("p", "progress-detail", `${t("evidenceChecked")}: ${new Date(inspection.checked).toLocaleString(language() === "zh" ? "zh-CN" : "en-US")}`));
  }
  target.append(section);
}
async function inspectHistory(taskId) {
  if (state.inspections.get(taskId)?.pending || state.expired) return;
  state.inspections.set(taskId, {pending:true});
  render();
  try {
    state.inspections.set(taskId, await inspectPackageTask(taskId));
    await pollTasks();
    await load();
  } catch (e) {
    state.inspections.set(taskId, {error:e.message});
  }
  render();
}
async function load() {
  if (state.busy) return;
  state.loading = true;
  try {
    state.catalog = await getCatalog();
    state.catalog.extensions ||= [];
    state.connected = true;
    state.expired = false;
    $("notice").textContent = "";
  } catch (e) {
    state.connected = false;
    state.expired = e.status === 401;
    $("notice").textContent = e.message;
  }
  state.loading = false;
  render();
}
async function action(kind, id, request = {}) {
  if (state.busy || !state.connected) return;
  if (kind === "install" || kind === "repair") {
    state.busy = true;
    try {
      await startPackageInstall(id, kind === "repair");
      await pollTasks();
    } catch (e) { $("notice").textContent = e.message; }
    finally {
      state.busy = false;
      // Polling may already have consumed a fast task's terminal status while
      // busy suppressed catalog refresh. Reload installed state after release.
      await load();
    }
    return;
  }
  state.busy = true;
  const packageAction = ["install", "repair", "remove"].includes(kind);
  const result = { id, status: "running", packageAction };
  if (packageAction) state.packageResults.set(id, result);
  state.lastResult = result;
  const originatingTool = route.page === "tool" ? route.id : null;
  if (kind === "run") state.results.set(id, result);
  else if (originatingTool) state.results.set(originatingTool, result);
  render();
  const update = () => {
    if (route.page === "tool")
      renderResult($("tool-result"), toolResult(route.id));
    else if (route.page === "packages" && packageAction) {
      const target = $("package-result-" + id);
      if (target) { target.hidden = false; renderResult(target, result); }
    } else {
      $("global-result").hidden = false;
      renderResult($("global-result"), state.lastResult);
    }
  };
  try {
    await performAction({ action: kind, name: id, ...request }, (event) => {
      if (event.type === "progress") result.progress = event.data;
      if (event.type === "error") {
        result.status = "failed";
        result.error = event.data;
      }
      if (event.type === "result") {
        result.status = "done";
        result.data = event.data;
        result.progress = null;
        if (kind === "doctor") state.checks = event.data;
      }
      update();
    });
  } catch (e) {
    result.status = "notFinished";
    result.error = { message: e.message };
  } finally {
    state.busy = false;
    await load();
  }
}
$("language").onchange = () => {
  setLanguage($("language").value);
  render();
};
$("refresh").onclick = async () => {
  for (const [id, inspection] of state.inspections) if (!inspection.pending) state.inspections.delete(id);
  await pollTasks(); await load();
};
$("search").oninput = () => {
  state.query = $("search").value;
  state.queries.set(queryKey(route), state.query);
  render();
};
window.addEventListener("hashchange", () => {
  if (/^#[a-f0-9]{64}$/.test(location.hash)) { initializeSession(); state.expired = false; load(); }
  const next = parseRoute(location.hash);
  state.query = state.queries.get(queryKey(next)) || "";
  $("search").value = state.query;
  route = next;
  render();
  document.querySelector("main").scrollTo(0, 0);
  $("content").focus({ preventScroll: true });
});

function toolResult(id) {
  const op = state.catalog.operations.find(o => o.id === id);
  const dependency = (op?.requirements || []).map(r => state.packageResults.get(r.package)).find(packageTaskBusy);
  return dependency || state.results.get(id) || (op?.requirements || []).map(r => state.packageResults.get(r.package)).find(Boolean);
}
let pollingTasks = false;
async function pollTasks() {
  if (pollingTasks || state.expired) return;
  pollingTasks = true;
  try {
    const tasks = await getPackageTasks();
    const results = packageTaskResults(tasks);
    for (const result of results.values()) {
      const evidence = state.inspections.get(result.taskId);
      if (evidence?.completion_confirmed && evidence.status === "done" && !packageTaskBusy(result)) result.status = "done";
    }
    const signature = JSON.stringify([...results.values()].map(r => [r.taskId, r.status]));
    state.packageResults = results;
    if (signature !== state.taskSignature || !state.connected) {
      state.taskSignature = signature;
      if (!state.busy) await load();
      else render();
    } else if (route.page === "packages") {
      for (const [name, result] of results) {
        const target = $("package-result-" + name);
        if (target) { target.hidden = false; renderResult(target, result); }
      }
    } else if (route.page === "tool") {
      renderResult($("tool-result"), toolResult(route.id));
    }
  } catch (e) {
    state.connected = false;
    state.expired = e.status === 401;
    $("notice").textContent = e.message;
    render();
  } finally { pollingTasks = false; }
}
async function watchTasks() {
  await pollTasks();
  if (state.connected && route.page === "packages") {
    const task = [...state.packageResults.values()].find(result => result.status === "interrupted" && !state.inspections.has(result.taskId));
    if (task) await inspectHistory(task.taskId);
  }
  setTimeout(watchTasks, 1000);
}
render();
load().then(watchTasks);
