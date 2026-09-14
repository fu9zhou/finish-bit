import { el } from "./dom.mjs";
import { t } from "./i18n.mjs";

export function formatBytes(bytes) {
  const value = Math.max(0, Number(bytes) || 0);
  const units = ["B", "KiB", "MiB", "GiB"];
  const index = Math.min(3, Math.floor(Math.log2(value || 1) / 10));
  return `${(value / 1024 ** index).toFixed(index ? 1 : 0)} ${units[index]}`;
}

// A percentage only describes the current download, never the whole install.
export function downloadDetails(progress) {
  const bytes = Math.max(0, progress.Bytes || 0);
  const total = progress.Total > 0 ? progress.Total : null;
  return {
    percent: total === null ? null : Math.min(100, Math.floor(bytes / total * 100)),
    size: total === null ? formatBytes(bytes) : `${formatBytes(bytes)} / ${formatBytes(total)}`,
  };
}

export function renderProgress(progress, running = true, historical = false) {
  const section = el("div", "install-progress");
  let stage = t("stage-" + progress.Stage);
  if (stage.startsWith("stage-")) stage = progress.Stage;
  if (historical) stage = t("lastProgress");
  const heading = el("div", "progress-heading");
  heading.append(el("span", "", [progress.Package, stage].filter(Boolean).join(" · ")));
  const downloading = progress.Stage === "downloading";
  const details = downloadDetails(progress);
  if (downloading && details.percent !== null)
    heading.append(el("strong", "", `${details.percent}%`));
  section.append(heading);
  if (running) {
    const bar = el("progress");
    bar.setAttribute("aria-label", stage);
    if (downloading && details.percent !== null) {
      bar.max = progress.Total;
      bar.value = Math.min(progress.Bytes, progress.Total);
    }
    section.append(bar);
  }
  const position = progress.SourceCount > 0 ? `${t("downloadSource")} ${progress.SourceIndex}/${progress.SourceCount}` : "";
  const description = [downloading ? details.size : "", position, progress.Source].filter(Boolean).join(" · ");
  if (description) section.append(el("p", "progress-detail", description));
  if (progress.Reason) section.append(el("p", "progress-detail", progress.Reason));
  if (progress.Stage === "source-failed") {
    const next = progress.NextSource
      ? `${t("switchingSource")} ${progress.SourceIndex + 1}/${progress.SourceCount}: ${progress.NextSource}`
      : t("noSourcesRemaining");
    section.append(el("p", "progress-detail", next));
  }
  return section;
}
