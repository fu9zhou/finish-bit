export function parseRoute(hash) {
  const value = hash.replace(/^#/, "");
  if (["tools", "packages", "extensions", "health"].includes(value))
    return { page: value };
  if (value.startsWith("categories/")) {
    try {
      const id = decodeURIComponent(value.slice(11));
      return /^[a-z]+$/.test(id) ? { page: "category", id } : { page: "not-found" };
    } catch { return { page: "not-found" }; }
  }
  if (value.startsWith("tools/")) {
    try {
      return { page: "tool", id: decodeURIComponent(value.slice(6)) };
    } catch {}
  }
  return value === "" ? { page: "tools" } : { page: "not-found" };
}
export function toolURL(id) {
  return "#tools/" + encodeURIComponent(id);
}
export function categoryURL(id) { return "#categories/" + encodeURIComponent(id); }
export function navigate(hash) {
  if (location.hash !== hash) location.hash = hash;
}
