import { test } from "node:test";
import assert from "node:assert/strict";
import { parseRoute, toolURL, categoryURL } from "./static/router.mjs";
import { createDraft, requestFromDraft } from "./static/forms.mjs";
import { destinationPath } from "./static/file-picker.mjs";
import { getCatalog, performAction } from "./static/api.mjs";
import { downloadDetails } from "./static/progress.mjs";
import { packageTaskResults, packageTaskBusy } from "./static/tasks.mjs";
import { categoryFor, findTools } from "./static/catalog-view.mjs";
import {
  setLanguage,
  resolveLanguage,
  t,
  operationName,
  parameterName,
  choiceName,
} from "./static/i18n.mjs";

test("language restores a valid choice before falling back to system language", () => {
  assert.equal(resolveLanguage("en", "zh-CN"), "en");
  assert.equal(resolveLanguage("zh", "en-US"), "zh");
  assert.equal(resolveLanguage(null, "zh-TW"), "zh");
  assert.equal(resolveLanguage(undefined, "en-GB"), "en");
  assert.equal(resolveLanguage("invalid", "ZH-CN"), "zh");
  assert.equal(resolveLanguage(null, "fr-FR"), "en");
});

test("tool discovery searches localized categories and intersects multiple terms", () => {
  const tools = [{id:"pdf.merge",aliases:["合并 PDF"],summary:"Merge PDF"}, {id:"image.convert",aliases:["图片格式转换"],summary:"Convert image"}, {id:"new.feature",summary:"Extension"}];
  assert.equal(categoryFor(tools[2]).id, "everyday");
  assert.deepEqual(findTools(tools,"PDF 合并","pdf"), [tools[0]]);
  assert.deepEqual(findTools(tools,"PDF","documents"), []);
  assert.equal(categoryFor({id:"document.convert"}).id, "documents");
  assert.equal(categoryFor({id:"pdf.merge"}).id, "pdf");
  assert.deepEqual(findTools(tools,"图片",""), [tools[1]]);
  assert.deepEqual(findTools(tools,"PDF","images"), []);
  assert.deepEqual(findTools(tools,"   ",""), tools);
});

test("tool URLs can be refreshed and navigated independently", () => {
  assert.deepEqual(parseRoute(toolURL("archive.extract")), {
    page: "tool",
    id: "archive.extract",
  });
  assert.deepEqual(parseRoute("#packages"), { page: "packages" });
  assert.deepEqual(parseRoute("#tools/%ZZ"), { page: "not-found" });
  assert.deepEqual(parseRoute(categoryURL("media")), { page: "category", id: "media" });
  assert.deepEqual(parseRoute("#categories/%ZZ"), { page: "not-found" });
  assert.deepEqual(parseRoute("#categories/"), { page: "not-found" });
});
test("output selection preserves platform paths without creating them", () => {
  assert.equal(destinationPath("C:\\", "result.txt"), "C:\\result.txt");
  assert.equal(
    destinationPath("D:\\中文目录", "result.txt"),
    "D:\\中文目录\\result.txt",
  );
  assert.equal(destinationPath("/", "result.txt"), "/result.txt");
  assert.equal(destinationPath("/tmp", "new-directory"), "/tmp/new-directory");
});
test("typed drafts survive UI rerenders and omit unused optional values", () => {
  const op = {
    inputs: [{ name: "input", required: true }, { name: "extra" }],
    options: [
      { name: "format", default: "zip" },
      { name: "overwrite", type: "boolean", default: false },
      { name: "files", type: "strings", default: [] },
      { name: "output" },
    ],
  };
  const draft = createDraft(op);
  draft.inputs[0] = "D:\\文件\\input.zip";
  draft.options.files.push("D:\\other.txt");
  assert.deepEqual(requestFromDraft(op, draft), {
    inputs: ["D:\\文件\\input.zip"],
    options: { format: "zip", overwrite: false, files: ["D:\\other.txt"] },
  });
  assert.deepEqual(createDraft(op).options.files, []);
});
test("locale changes localize labels and preserve protocol values", () => {
  const op = { summary: "Extract archive", aliases: ["解压文件"] };
  setLanguage("zh");
  assert.equal(operationName(op), "解压文件");
  assert.equal(parameterName({ name: "output" }), "输出位置");
  assert.equal(t("install"), "安装");
  assert.equal(t("toolboxHome"), "首页");
  assert.equal(t("home"), "用户目录");
  assert.match(choiceName("horizontal", {}), /水平/);
  setLanguage("en");
  assert.equal(operationName(op), "Extract archive");
  assert.equal(t("install"), "Install");
  assert.equal(choiceName("horizontal", {}), "horizontal");
});

test("local server connection errors explain recovery", async (context) => {
  setLanguage("en");
  context.mock.method(globalThis, "fetch", async () => { throw new TypeError("Failed to fetch"); });
  await assert.rejects(getCatalog(), /local UI service/);
});

test("expired sessions can be distinguished from empty catalogs", async (context) => {
  context.mock.method(globalThis, "fetch", async () => new Response("Unauthorized", {status:401}));
  await assert.rejects(getCatalog(), error => error.status === 401);
});

test("download progress is delivered before the action completes", async (context) => {
  let controller;
  const stream = new ReadableStream({ start(value) { controller = value; } });
  context.mock.method(globalThis, "fetch", async () => new Response(stream));
  const encoder = new TextEncoder();
  const events = [];
  let observed;
  const progress = new Promise(resolve => { observed = resolve; });
  const action = performAction({ action: "install", name: "example" }, event => {
    events.push(event);
    if (event.type === "progress") observed();
  });
  controller.enqueue(encoder.encode('{"type":"progress","data":{"Stage":"down'));
  controller.enqueue(encoder.encode('loading","Bytes":512,"Total":1024}}\n'));
  await progress;
  assert.equal(events[0].data.Bytes, 512);
  controller.enqueue(encoder.encode('{"type":"result","data":null}\n'));
  controller.close();
  await action;
  assert.equal(events.at(-1).type, "result");
});

test("known and unknown download sizes show only real progress", () => {
  assert.deepEqual(downloadDetails({ Bytes: 524288, Total: 1048576 }), {
    percent: 50, size: "512.0 KiB / 1.0 MiB",
  });
  assert.deepEqual(downloadDetails({ Bytes: 2048, Total: -1 }), {
    percent: null, size: "2.0 KiB",
  });
  assert.equal(downloadDetails({ Bytes: 0, Total: 100 }).percent, 0);
  assert.equal(downloadDetails({ Bytes: 110, Total: 100 }).percent, 100);
  // A mirror starts a fresh download; no previous-source bytes leak into it.
  assert.equal(downloadDetails({ Bytes: 0, Total: -1 }).percent, null);
});

test("restored tasks retain source and prefer active work over a rejected duplicate", () => {
  const tasks = [
    {id:"duplicate",package:"tool",status:"failed",origin:"web",started:"2026-09-10T12:00:01Z"},
    {id:"active",package:"tool",status:"running",origin:"cli",started:"2026-09-10T12:00:00Z",progress:{Bytes:512,Total:1024}},
  ];
  const result = packageTaskResults(tasks).get("tool");
  assert.equal(result.taskId,"active");
  assert.equal(result.origin,"cli");
  assert.equal(result.progress.Bytes,512);
  assert.equal(tasks[0].id,"duplicate");
});

test("stale task progress is historical and active recovery wins over old results", () => {
  const progress = {Stage:"downloading",Bytes:12,Total:100};
  const results = packageTaskResults([
    {id:"stopped",package:"stopped",status:"interrupted",updated:"2026-09-14T07:00:00Z",progress},
    {id:"newer",package:"tool",status:"failed",started:"2026-09-14T07:01:00Z"},
    {id:"owner",package:"tool",status:"recovering",started:"2026-09-14T07:00:00Z",progress},
  ]);
  assert.equal(results.get("stopped").progress, null);
  assert.equal(results.get("stopped").lastProgress.Bytes, 12);
  assert.equal(results.get("stopped").updated, "2026-09-14T07:00:00Z");
  assert.equal(results.get("tool").taskId, "owner");
  assert.equal(results.get("tool").progress, null);
  assert.equal(packageTaskBusy(results.get("tool")), true);
  assert.equal(packageTaskBusy(results.get("stopped")), false);
  assert.equal(packageTaskBusy({status:"unknown"}), true);
});

test("broken action streams report unknown state, not installation failure", async (context) => {
  setLanguage("en");
  context.mock.method(globalThis, "fetch", async () => new Response(new ReadableStream({
    start(controller) { controller.error(new TypeError("network error")); },
  })));
  await assert.rejects(performAction({}, () => {}), /Connection interrupted/);
});

test("download source errors retain the backend explanation", async (context) => {
  context.mock.method(globalThis, "fetch", async () => new Response(
    '{"type":"error","data":{"message":"all package download sources failed"}}\n',
  ));
  const events = [];
  await performAction({}, event => events.push(event));
  assert.equal(events[0].data.message, "all package download sources failed");
});
