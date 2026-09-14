import { t } from "./i18n.mjs";
let token = "";
export function initializeSession() {
  const secret = location.hash.slice(1);
  if (/^[a-f0-9]{64}$/.test(secret)) {
    sessionStorage.setItem("finishbit-token", secret);
    history.replaceState(null, "", "#tools");
  }
  token = sessionStorage.getItem("finishbit-token") || "";
}
async function request(path, options = {}) {
  let response;
  try {
    response = await fetch(path, {
      ...options,
      headers: { Authorization: "Bearer " + token, ...options.headers },
    });
  } catch (cause) {
    throw new Error(t("serviceUnavailable"), { cause });
  }
  if (!response.ok) {
    if (response.status === 401) throw Object.assign(Error(t("expired")), { status: 401 });
    if (response.status === 409) throw Error(t("busy"));
    const raw = await response.text();
    try {
      throw Error(JSON.parse(raw).message || raw);
    } catch (e) {
      if (e instanceof SyntaxError) throw Error(raw);
      throw e;
    }
  }
  return response;
}
export const getCatalog = async () =>
  await (await request("/api/catalog")).json();
export const getPackageTasks = async () =>
  await (await request("/api/package-tasks")).json();
export const inspectPackageTask = async (id) =>
  await (await request("/api/package-tasks?inspect=" + encodeURIComponent(id))).json();
export const startPackageInstall = async (name, repair) =>
  await (await request("/api/package-tasks", {
    method: "POST", headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, repair }),
  })).json();
export const browseFiles = async (path) =>
  await (
    await request("/api/browse?path=" + encodeURIComponent(path || ""))
  ).json();
export async function performAction(payload, onEvent) {
  const response = await request("/api/action", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const reader = response.body.getReader(),
    decoder = new TextDecoder();
  let pending = "",
    finished = false;
  const consume = (line) => {
    if (!line.trim()) return;
    const event = JSON.parse(line);
    if (event.type === "result" || event.type === "error") finished = true;
    onEvent(event);
  };
  try {
    while (true) {
      let chunk;
      try { chunk = await reader.read(); }
      catch (cause) { throw new Error(t("interrupted"), { cause }); }
      const { done, value } = chunk;
      pending += decoder.decode(value, { stream: !done });
      let end;
      while ((end = pending.indexOf("\n")) >= 0) {
        consume(pending.slice(0, end));
        pending = pending.slice(end + 1);
      }
      if (done) {
        consume(pending);
        break;
      }
    }
  } finally {
    reader.releaseLock();
  }
  if (!finished) throw Error(t("interrupted"));
}
