import { el, button } from "./dom.mjs";
import {
  t,
  language,
  operationName,
  parameterName,
  parameterHelp,
  choiceName,
} from "./i18n.mjs";
import { createFilePicker } from "./file-picker.mjs";

export function createDraft(op) {
  const options = {};
  for (const p of op.options || [])
    if (p.default !== undefined) options[p.name] = structuredClone(p.default);
  return { inputs: (op.inputs || []).map((p) => p.default ?? ""), options };
}
export function requestFromDraft(op, draft) {
  const inputs = [...draft.inputs],
    options = {};
  while (
    inputs.length &&
    inputs.at(-1) === "" &&
    !op.inputs[inputs.length - 1].required
  )
    inputs.pop();
  for (const p of op.options || []) {
    const v = draft.options[p.name];
    if (v !== undefined && v !== "") options[p.name] = v;
  }
  return { inputs, options };
}

export function createToolPage(
  op,
  { draft, packages, busy, onRun, onInstall, onChange },
) {
  const root = el("div", "tool-page");
  const heading = el("div", "tool-heading");
  heading.append(
    el("span", "eyebrow", op.id),
    el("h1", "", operationName(op)),
    el("p", "", t("settingsHint")),
  );
  root.append(heading);
  const layout = el("div", "tool-layout"),
    form = el("form", "tool-form"),
    intro = el("div", "form-heading");
  intro.append(el("h2", "", t("parameters")));
  form.append(intro);
  const dependencies = el("div", "dependencies");
  for (const req of op.requirements || []) {
    const pkg = packages.find((p) => p.name === req.package),
      row = el("div", "dependency");
    row.append(
      el(
        "span",
        "",
        `${req.package} · ${t(pkg?.installed ? "installed" : "missing")}`,
      ),
    );
    if (!pkg?.installed) {
      const b = button(
        t(pkg?.supported ? "install" : "unsupported"),
        "secondary",
        () => onInstall(req.package),
      );
      b.disabled = busy || !pkg?.supported;
      row.append(b);
    }
    dependencies.append(row);
  }
  form.append(dependencies);
  function field(p, index, inputParam) {
    const value = () =>
      inputParam ? draft.inputs[index] : draft.options[p.name];
    const set = (value) => {
      if (inputParam) draft.inputs[index] = value;
      else draft.options[p.name] = value;
    };
    const container = el("div", "field"),
      label = el("label"),
      id = `parameter-${inputParam ? "in" : "opt"}-${index}`;
    label.htmlFor = id;
    label.append(
      el("span", "field-label", parameterName(p)),
      el("span", "field-required", t(p.required ? "required" : "optional")),
    );
    container.append(label);
    const choices = (p.choices || []).filter(
      (x) => p.name !== "input-mode" || x !== "stdin",
    );
    const modeFile = inputParam && draft.options["input-mode"] === "file";
    const kind = modeFile ? "file" : p.kind;
    let control;
    if (choices.length) {
      control = el("select");
      if (!p.required && p.default === undefined)
        control.add(new Option(t("default"), ""));
      choices.forEach((v) => control.add(new Option(choiceName(v, p), v)));
      if (
        value() !== undefined &&
        !choices.includes(String(value())) &&
        value() !== ""
      )
        control.add(new Option(String(value()), String(value())));
      control.value = value() ?? "";
    } else if (p.type === "boolean") {
      control = el("input", "switch");
      control.type = "checkbox";
      control.checked = value() ?? false;
      control.setAttribute("role", "switch");
    } else if (
      p.type === "strings" ||
      (inputParam && (!kind || kind === "text-or-file"))
    ) {
      control = el("textarea");
      control.rows = 4;
      control.value = Array.isArray(value())
        ? value().join("\n")
        : (value() ?? "");
    } else {
      control = el("input");
      control.type =
        p.type === "integer"
          ? "number"
          : (p.name.includes("password") && !p.name.endsWith("-env")) ||
              p.name === "key"
            ? "password"
            : "text";
      if (p.type === "integer") control.step = "1";
      control.value = value() ?? "";
    }
    control.id = id;
    control.name = p.name;
    control.required = !!p.required && p.type !== "boolean";
    control.disabled = busy;
    const update = () => {
      set(
        p.type === "boolean"
          ? control.checked
          : p.type === "integer" && control.value !== ""
            ? Number(control.value)
            : p.type === "strings"
              ? control.value
                ? control.value.split(/\r?\n/)
                : []
              : control.value,
      );
    };
    control.oninput = update;
    control.onchange = () => {
      update();
      if (p.name === "input-mode") onChange();
    };
    const inputRow = el("div", "field-control");
    inputRow.append(control);
    container.append(inputRow);
    if (kind) {
      control.placeholder = t("pathHint");
      const pickerSlot = el("div");
      const browse = button(
        t(
          kind.startsWith("output-")
            ? "saveLocation"
            : kind === "directory"
              ? "browseDirectory"
              : kind === "path"
                ? "browsePath"
                : "browse",
        ),
        "secondary",
        () => {
          if (pickerSlot.childElementCount) {
            pickerSlot.replaceChildren();
            return;
          }
          pickerSlot.append(
            createFilePicker({
              kind: kind === "text-or-file" ? "file" : kind,
              onClose: () => {
                pickerSlot.replaceChildren();
                browse.focus();
              },
              onSelect: (path) => {
                if (p.type === "strings") {
                  const values = value() || [];
                  set([...values, path]);
                  control.value = value().join("\n");
                } else {
                  set(path);
                  control.value = path;
                }
                pickerSlot.replaceChildren();
                control.focus();
              },
            }),
          );
        },
      );
      browse.disabled = busy;
      inputRow.append(browse);
      container.append(pickerSlot);
    }
    const help = parameterHelp({ ...p, kind });
    if (help) {
      const hint = el("small", "", help);
      hint.id = id + "-help";
      control.setAttribute("aria-describedby", hint.id);
      container.append(hint);
    }
    // English remains the canonical provider reference; never hide limits by translating heuristically.
    if (language() === "zh" && p.description) {
      const details = el("details", "parameter-reference");
      details.append(
        el("summary", "", t("parameterDetails")),
        el("p", "", p.description),
      );
      details.lastChild.lang = "en";
      container.append(details);
    }
    return container;
  }
  (op.inputs || []).forEach((p, i) => form.append(field(p, i, true)));
  (op.options || []).forEach((p, i) => form.append(field(p, i, false)));
  const actions = el("div", "form-actions"),
    run = el("button", "primary", t(busy ? "running" : "run") + " →");
  run.type = "submit";
  run.disabled = busy;
  actions.append(run);
  form.append(actions);
  form.onsubmit = (e) => {
    e.preventDefault();
    onRun(requestFromDraft(op, draft));
  };
  const side = el("aside", "tool-side"),
    result = el("section", "result-panel");
  result.id = "tool-result";
  result.setAttribute("aria-label", t("result"));
  side.append(result);
  const details = el("details", "contract-details");
  details.append(el("summary", "", t("details")), el("p", "", op.description));
  details.lastChild.lang = "en";
  side.append(details);
  layout.append(form, side);
  root.append(layout);
  return root;
}
