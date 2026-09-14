import { browseFiles } from "./api.mjs";
import { t } from "./i18n.mjs";
import { el, button } from "./dom.mjs";

// Inline picker uses real absolute paths. Selecting never creates an output.
export function destinationPath(parent, name) {
  return (
    parent.replace(/[\\/]$/, "") + (parent.includes("\\") ? "\\" : "/") + name
  );
}
export function createFilePicker({ kind, onSelect, onClose }) {
  const root = el("section", "file-picker");
  root.setAttribute("aria-label", t("pickerTitle"));
  const heading = el("div", "picker-heading");
  heading.append(
    el("strong", "", t("pickerTitle")),
    button(t("closePicker"), "secondary", onClose),
  );
  const toolbar = el("div", "picker-toolbar"),
    path = el("input");
  path.setAttribute("aria-label", t("pathHint"));
  const up = button("↑ " + t("up"), "secondary", () => load(listing?.parent));
  const go = button(t("go"), "secondary", () => load(path.value));
  path.onkeydown = (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      load(path.value);
    }
  };
  toolbar.append(up, path, go);
  const roots = el("div", "picker-roots"),
    status = el("p", "picker-status");
  status.setAttribute("role", "status");
  const entries = el("div", "file-list"),
    bottom = el("div", "picker-bottom");
  const output = kind.startsWith("output-"),
    directory = kind === "directory" || kind === "output-directory";
  let listing,
    sequence = 0;
  if (output) {
    const label = el("label", "filename"),
      name = el("input");
    label.append(el("span", "", t(directory ? "dirname" : "filename")), name);
    name.autocomplete = "off";
    bottom.append(
      label,
      button(t("save"), "primary", () => {
        if (!listing) return;
        const v = name.value.trim();
        if (!v || v === "." || v === ".." || /[\\/\x00]/.test(v)) {
          status.textContent = t("invalidName");
          name.focus();
          return;
        }
        onSelect(destinationPath(listing.path, v));
      }),
    );
  } else if (directory || kind === "path")
    bottom.append(
      button(
        t("currentFolder"),
        "primary",
        () => listing && onSelect(listing.path),
      ),
    );
  root.append(heading, toolbar, roots, status, entries, bottom);
  async function load(value) {
    const id = ++sequence;
    path.disabled = true;
    go.disabled = true;
    up.disabled = true;
    bottom.querySelectorAll("button").forEach((b) => (b.disabled = true));
    status.textContent = t("loading");
    entries.replaceChildren();
    listing = null;
    try {
      const data = await browseFiles(value);
      if (id !== sequence) return;
      listing = data;
      path.value = data.path;
      up.disabled = data.parent === data.path;
      roots.replaceChildren();
      data.roots.forEach((r) =>
        roots.append(
          button(r.name === "Home" ? t("home") : r.name, "secondary", () =>
            load(r.path),
          ),
        ),
      );
      for (const item of data.entries) {
        if (directory && !item.directory) continue;
        const row = button("", "file-entry", () => {
          if (item.directory) load(item.path);
          else if (output) bottom.querySelector("input").value = item.name;
          else onSelect(item.path);
        });
        row.append(
          el("span", "file-kind", t(item.directory ? "folder" : "file")),
          el("span", "file-name", item.name),
          el("span", "", item.directory ? "→" : t("choose")),
        );
        entries.append(row);
      }
      status.textContent = data.truncated
        ? t("truncated")
        : entries.childElementCount
          ? ""
          : t("emptyFolder");
    } catch (e) {
      if (id === sequence) status.textContent = e.message;
    } finally {
      if (id === sequence) {
        path.disabled = false;
        go.disabled = false;
        up.disabled = !listing || listing.parent === listing.path;
        bottom
          .querySelectorAll("button")
          .forEach((b) => (b.disabled = !listing));
      }
    }
  }
  load("");
  return root;
}
