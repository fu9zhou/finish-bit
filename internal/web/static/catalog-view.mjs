export const categories = [
  { id: "documents", zh: "文档", en: "Documents", prefixes: ["document", "ocr"] },
  { id: "pdf", zh: "PDF", en: "PDF", prefixes: ["pdf"] },
  { id: "media", zh: "音视频", en: "Audio & video", icon: "▷", prefixes: ["audio", "video", "media", "subtitle", "screen", "speech"] },
  { id: "images", zh: "图片与设计", en: "Images & design", icon: "◈", prefixes: ["image", "color", "qrcode"] },
  { id: "files", zh: "文件与压缩", en: "Files & archives", icon: "▱", prefixes: ["file", "archive", "disk"] },
  { id: "text", zh: "文本与写作", en: "Text & writing", icon: "T", prefixes: ["text", "markdown", "worksheet", "unicode"] },
  { id: "data", zh: "表格与数据", en: "Tables & data", icon: "⊞", prefixes: ["csv", "json", "jsonl", "yaml", "toml", "xml"] },
  { id: "developer", zh: "开发与网络", en: "Developer & network", icon: "{ }", prefixes: ["base64", "hex", "html", "url", "ip", "cidr", "jwt", "crypto", "hash", "password", "uuid", "regex", "cron", "code", "useragent"] },
  { id: "everyday", zh: "计算与生活", en: "Everyday utilities", icon: "＋", prefixes: [] },
];
export function categoryFor(op) {
  return categories.find(c => c.prefixes.includes(op.id.split(".")[0])) || categories.at(-1);
}
export function findTools(operations, query, category) {
  const terms = query.trim().toLocaleLowerCase().split(/\s+/).filter(Boolean);
  return operations.filter(op => {
    const group = categoryFor(op);
    const text = JSON.stringify([op.id, op.summary, op.aliases, op.tags, group.zh, group.en]).toLocaleLowerCase();
    return (!category || group.id === category) && terms.every(term => text.includes(term));
  });
}
export const categoryDescriptions = {
  documents: ["转换文档格式，提取内容与识别文字。", "Convert document formats, extract content and recognize text."],
  pdf: ["合并拆分文件，编辑页面与管理表单。", "Merge and split PDFs, edit pages and manage forms."],
  images: ["转换图片格式，调整画面与生成二维码。", "Convert image formats, edit images and generate QR codes."],
  media: ["转换剪辑音视频，处理字幕与识别语音。", "Convert and trim audio and video, process subtitles and recognize speech."],
  files: ["压缩解压文件，校验内容与查看磁盘。", "Compress and extract files, verify content and inspect disks."],
  text: ["整理比较文本，转换格式与制作字帖。", "Clean and compare text, convert formats and create worksheets."],
  data: ["清洗统计表格，转换格式与查询数据。", "Clean and summarize tables, convert formats and query data."],
  developer: ["编码解码内容，校验哈希与解析网络地址。", "Encode and decode content, verify hashes and parse network addresses."],
  everyday: ["计算日期时间，换算数值与处理日常计算。", "Calculate dates and times, convert values and solve everyday calculations."],
};
export const featuredDescriptions = {
  "image.convert": "转换图片格式，适配不同使用场景。",
  "document.convert": "转换文档格式，方便编辑、分享与归档。",
  "pdf.merge": "把多个 PDF 文件按顺序合并。",
  "archive.extract": "解压全部文件，或只提取选中的成员。",
  "json.format": "整理 JSON 缩进，让数据更容易阅读。",
  "qrcode.generate": "把文字或链接生成二维码图片。",
  "text.count": "统计文本字符、单词与行数。",
  "file.checksum": "计算文件校验值，检查内容是否一致。",
};
