export function el(tag, cls, text) {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
}
export function button(text, cls, action) {
  const b = el("button", cls, text);
  b.type = "button";
  b.onclick = action;
  return b;
}
export function link(text, href, cls = "") {
  const a = el("a", cls, text);
  a.href = href;
  return a;
}
