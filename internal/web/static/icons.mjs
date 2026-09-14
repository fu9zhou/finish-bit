import { el } from "./dom.mjs";
import { categories } from "./catalog-view.mjs";

// Hand-drawn SVG assets share a 64-unit grid and a 2.2-unit rounded stroke.
// Visible headings provide the accessible name, so illustrations are decorative.
export function categoryIcon(id, className = "category-icon") {
  const image = el("img", className);
  image.src = "/icons/" + (categories.some(category => category.id === id) ? id : "everyday") + ".svg";
  image.alt = "";
  image.setAttribute("aria-hidden", "true");
  image.width = 64;
  image.height = 64;
  return image;
}
