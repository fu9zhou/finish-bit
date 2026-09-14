import { el } from "./dom.mjs";

let active = null;
let serial = 0;
const controls = new WeakMap();

function close() {
  if (!active) return;
  active.menu.remove();
  active.trigger.setAttribute("aria-expanded", "false");
  active.trigger.removeAttribute("aria-activedescendant");
  active = null;
}

function open(select, trigger) {
  close();
  const menu = el("div", "select-menu");
  menu.id = trigger.getAttribute("aria-controls");
  menu.setAttribute("role", "listbox");
  menu.setAttribute("aria-label", select.getAttribute("aria-label") || select.labels?.[0]?.textContent || "");
  const options = [...select.options].filter(option => !option.hidden);
  const items = options.map((option, index) => {
    const item = el("div", "select-option");
    item.id = menu.id + "-" + index;
    item.setAttribute("role", "option");
    item.setAttribute("aria-selected", String(option.selected));
    item.setAttribute("aria-disabled", String(option.disabled));
    item.append(el("span", "", option.text), el("span", "select-check", option.selected ? "✓" : ""));
    item.onpointerdown = event => event.preventDefault();
    item.onclick = () => { if (!option.disabled) choose(index); };
    menu.append(item);
    return item;
  });
  const choose = index => {
    const value = options[index]?.value;
    if (value === undefined || options[index].disabled) return;
    close();
    select.value = value;
    select.dispatchEvent(new Event("change", { bubbles: true }));
    enhanceSelects();
    // A dependent form or locale change may replace the original control.
    const replacement = document.getElementById(select.id);
    (controls.get(replacement)?.trigger || trigger).focus();
  };
  active = { select, trigger, menu, options, items, choose, index: Math.max(0, options.findIndex(o => o.selected)), prefix: "", typedAt: 0 };
  document.body.append(menu);
  const box = trigger.getBoundingClientRect();
  const width = Math.min(box.width, window.innerWidth - 24);
  menu.style.width = width + "px";
  menu.style.left = Math.max(12, Math.min(box.left, window.innerWidth - width - 12)) + "px";
  const below = window.innerHeight - box.bottom - 16;
  const above = box.top - 16;
  const upwards = below < Math.min(menu.scrollHeight, 280) && above > below;
  menu.style.maxHeight = Math.max(64, Math.min(320, upwards ? above : below)) + "px";
  if (upwards) menu.style.bottom = window.innerHeight - box.top + 6 + "px";
  else menu.style.top = box.bottom + 6 + "px";
  trigger.setAttribute("aria-expanded", "true");
  highlight(active.index);
}

function highlight(index) {
  if (!active) return;
  active.index = index;
  active.items.forEach((item, i) => item.classList.toggle("highlighted", i === index));
  const item = active.items[index];
  if (item) {
    active.trigger.setAttribute("aria-activedescendant", item.id);
    item.scrollIntoView({ block: "nearest" });
  }
}

function keydown(event, select, trigger) {
  const key = event.key;
  if (key === "Tab" || key === "Escape") { close(); return; }
  if (!["ArrowDown", "ArrowUp", "Home", "End", "Enter", " "].includes(key) && (key.length !== 1 || event.ctrlKey || event.metaKey || event.altKey)) return;
  event.preventDefault();
  if (!active || active.trigger !== trigger) {
    open(select, trigger);
    if (["Enter", " ", "ArrowDown", "ArrowUp"].includes(key)) return;
  }
  if (key === "Enter" || key === " ") { active.choose(active.index); return; }
  const enabled = active.options.map((option, index) => option.disabled ? -1 : index).filter(index => index >= 0);
  if (!enabled.length) return;
  if (key === "Home") highlight(enabled[0]);
  else if (key === "End") highlight(enabled.at(-1));
  else if (key === "ArrowDown" || key === "ArrowUp") {
    const next = enabled.indexOf(active.index) + (key === "ArrowDown" ? 1 : -1);
    highlight(enabled[Math.max(0, Math.min(enabled.length - 1, next))]);
  } else {
    const now = Date.now();
    active.prefix = (now - active.typedAt > 700 ? "" : active.prefix) + key.toLocaleLowerCase();
    active.typedAt = now;
    const index = enabled.find(i => active.options[i].text.toLocaleLowerCase().startsWith(active.prefix));
    if (index !== undefined) highlight(index);
  }
}

// Preserve native values, change handlers and form validation; replace only presentation.
export function enhanceSelects() {
  if (active && (!active.select.isConnected || active.select.hidden)) close();
  for (const select of document.querySelectorAll("select")) {
    let control = controls.get(select);
    if (!control) {
      const trigger = el("button", "select-trigger");
      trigger.type = "button";
      trigger.setAttribute("role", "combobox");
      trigger.setAttribute("aria-haspopup", "listbox");
      trigger.setAttribute("aria-expanded", "false");
      trigger.setAttribute("aria-controls", "select-menu-" + ++serial);
      trigger.onclick = () => active?.trigger === trigger ? close() : open(select, trigger);
      trigger.onkeydown = event => keydown(event, select, trigger);
      select.classList.add("select-native");
      select.tabIndex = -1;
      select.setAttribute("aria-hidden", "true");
      select.onfocus = () => trigger.focus();
      for (const label of select.labels || []) label.addEventListener("click", event => { event.preventDefault(); trigger.focus(); });
      select.addEventListener("invalid", event => { event.preventDefault(); trigger.focus(); trigger.setAttribute("aria-invalid", "true"); });
      select.after(trigger);
      control = { trigger };
      controls.set(select, control);
    }
    const { trigger } = control;
    trigger.hidden = select.hidden;
    trigger.disabled = select.disabled;
    trigger.classList.toggle("select-language", select.id === "language");
    trigger.classList.toggle("select-category", select.id === "category");
    const text = select.selectedOptions[0]?.text || "";
    const chevron = el("span", "select-chevron");
    chevron.setAttribute("aria-hidden", "true");
    trigger.replaceChildren(el("span", "select-value", text), chevron);
    const label = select.getAttribute("aria-label") || [...(select.labels?.[0]?.children || [])].map(child => child.textContent.trim()).join(" ") || select.labels?.[0]?.textContent?.trim();
    trigger.setAttribute("aria-label", label ? `${label}: ${text}` : text);
    trigger.setAttribute("aria-required", String(select.required));
    if (select.validity.valid) trigger.removeAttribute("aria-invalid");
    if (select.hasAttribute("aria-describedby")) trigger.setAttribute("aria-describedby", select.getAttribute("aria-describedby"));
  }
}

if (typeof document !== "undefined") {
  document.addEventListener("pointerdown", event => {
    if (active && !active.menu.contains(event.target) && !active.trigger.contains(event.target)) close();
  });
  document.addEventListener("focusin", event => { if (active && event.target !== active.trigger && !active.menu.contains(event.target)) close(); });
  document.addEventListener("scroll", event => { if (active && event.target !== active.menu) close(); }, true);
  window.addEventListener("resize", close);
}
