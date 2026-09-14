const messages = {
  toolTotal: ["{count} 项工具", "{count} tools"],
  tools: ["工具箱", "Toolbox"],
  packages: ["依赖管理", "Dependencies"],
  extensions: ["扩展管理", "Extensions"],
  health: ["环境检测", "Environment"],
  toolsTitle: ["工具箱", "Toolbox"],
  toolsSubtitle: [
    "搜索工具或浏览分类，在本地处理文件与数据。",
    "Search tools or browse categories to process files and data locally.",
  ],
  featured: ["常用工具", "Quick picks"],
  recent: ["最近打开", "Recently opened"],
  browseTools: ["按分类浏览", "Browse by category"],
  toolboxHome: ["首页", "Home"],
  backCategory: ["返回{category}", "Back to {category}"],
  categorySearch: ["在{category}中搜索…", "Search {category}…"],
  enterCategory: ["浏览工具", "Browse tools"],
  categoryUnavailable: ["未找到这个分类", "Category not found"],
  viewAll: ["查看全部", "View all"],
  clearFilters: ["清除筛选", "Clear filters"],
  searchResults: ["搜索结果", "Search results"],
  catalogUnavailable: ["暂时无法加载工具箱", "Toolbox unavailable"],
  reconnectHelp: ["请确认本地服务仍在运行。服务重启后，请打开终端中新生成的完整链接。", "Check that the local service is running. After restarting it, open the new complete link printed in the terminal."],
  staleCatalog: ["当前显示上次加载的目录，连接恢复后才能运行工具。", "Showing the last loaded catalog. Reconnect before running tools."],
  packagesTitle: ["依赖管理", "Dependencies"],
  packagesSubtitle: [
    "按需安装、修复或移除本地依赖包。",
    "Install, repair, or remove local dependencies on demand.",
  ],
  extensionsTitle: ["扩展管理", "Extensions"],
  extensionsSubtitle: [
    "从本地目录安装扩展，自动接入工具箱。",
    "Install extensions from local directories to expand your toolbox.",
  ],
  healthTitle: ["环境检测", "Environment"],
  healthSubtitle: [
    "检查运行环境、数据目录和依赖完整性。",
    "Inspect your runtime, data directory, and package integrity.",
  ],
  workspace: ["个人工作空间", "Personal workspace"],
  tagline: ["让琐碎的事，到此为止。", "A little less busywork."],
  localNote: ["本地运行", "Runs locally"],
  privacy: [
    "文件在本机处理，无需上传。安装依赖包时需要联网。",
    "Files stay on your computer, with no upload required. Package downloads need internet access.",
  ],
  connected: ["本地已连接", "Connected locally"],
  connecting: ["正在连接…", "Connecting…"],
  offline: ["连接不可用", "Connection unavailable"],
  refresh: ["刷新状态", "Refresh"],
  language: ["语言", "Language"],
  capabilities: ["项能力", "operations"],
  installedPackages: ["已安装依赖包", "installed packages"],
  local: ["本地", "Local"],
  processing: ["文件处理环境", "file processing"],
  search: ["搜索", "Search"],
  searchHint: [
    "搜索工具、能力或关键词…",
    "Search tools, operations, or keywords…",
  ],
  category: ["工具分类", "Category"],
  all: ["全部分类", "All categories"],
  empty: [
    "没有匹配的项目，试试其他关键词。",
    "No matches. Try another search.",
  ],
  loading: ["正在加载…", "Loading…"],
  installed: ["已安装", "Installed"],
  missing: ["需要依赖", "Needs dependencies"],
  builtin: ["内置能力", "Built in"],
  dependenciesReady: ["依赖已安装", "Dependencies installed"],
  notInstalled: ["未安装", "Not installed"],
  unsupported: ["当前平台不支持", "Unsupported platform"],
  install: ["安装", "Install"],
  repair: ["修复", "Repair"],
  remove: ["移除", "Remove"],
  removeConfirm: [
    "确定移除 {name}？依赖它的工具将需要重新安装。",
    "Remove {name}? Tools that depend on it will need it reinstalled.",
  ],
  extensionPath: ["扩展目录", "Extension directory"],
  installExtension: ["安装扩展", "Install extension"],
  noExtensions: [
    "尚无扩展。安装后，能力会自动出现在工具箱中。",
    "No extensions yet. Installed operations will appear in the toolbox.",
  ],
  check: ["开始检查", "Run checks"],
  checkReady: ["检查本地环境", "Check your local environment"],
  ok: ["正常", "OK"],
  attention: ["需要处理", "Needs attention"],
  back: ["返回工具箱", "Back to toolbox"],
  run: ["运行工具", "Run tool"],
  running: ["正在执行…", "Running…"],
  parameters: ["输入与配置", "Inputs & settings"],
  required: ["必填", "Required"],
  optional: ["可选", "Optional"],
  settingsHint: [
    "选择文件并填写配置，所有处理均在本地完成。",
    "Choose your files and settings. Processing runs on this computer.",
  ],
  details: [
    "查看完整契约与处理限制（原文）",
    "Full contract & processing limits",
  ],
  parameterDetails: ["参数说明（原文）", "Parameter reference"],
  browse: ["选择文件", "Choose file"],
  browseDirectory: ["选择目录", "Choose directory"],
  browsePath: ["选择文件或目录", "Choose file or folder"],
  saveLocation: ["选择保存位置", "Choose save location"],
  pathHint: [
    "可选择或粘贴本地绝对路径",
    "Choose or paste an absolute local path",
  ],
  textHint: ["输入文本内容", "Enter text"],
  lines: ["每行一项，按顺序处理。", "One item per line, processed in order."],
  default: ["使用默认值", "Use default"],
  pickerTitle: ["本地文件浏览器", "Local file browser"],
  closePicker: ["收起文件选择", "Close file browser"],
  go: ["前往", "Go"],
  up: ["上一级", "Parent folder"],
  home: ["用户目录", "Home"],
  currentFolder: ["使用此目录", "Use this folder"],
  folder: ["目录", "Folder"],
  file: ["文件", "File"],
  choose: ["选择", "Select"],
  filename: ["新文件名", "File name"],
  dirname: ["新目录名", "New folder name"],
  save: ["使用此位置", "Use this location"],
  invalidName: [
    "请输入名称，不要包含路径分隔符。",
    "Enter a name without path separators.",
  ],
  emptyFolder: ["此目录没有可选项目。", "No selectable items in this folder."],
  truncated: [
    "仅显示前 5000 项，请输入更具体的目录。",
    "Showing the first 5,000 entries. Navigate to a more specific folder.",
  ],
  result: ["任务结果", "Task result"],
  noResult: [
    "运行后，结果会显示在这里。",
    "Your result will appear here after running.",
  ],
  done: ["已完成", "Completed"],
  cancelled: ["已取消", "Cancelled"],
  unknown: ["暂时无法核实安装状态，正在自动重查。", "Cannot verify installation yet. Checking again automatically."],
  recovering: ["安装操作仍占用此包，正在等待进度恢复，请勿重复安装。", "Package work is still active. Waiting for progress to resume; do not start another install."],
  taskInterrupted: ["安装任务已停止，未收到完成确认。可重新安装；已安装的包可修复。", "The task stopped without a completion record. Retry installation, or repair an installed package."],
  lastProgress: ["最后记录（非实时进度）", "Last recorded progress (not live)"],
  lastUpdated: ["最后上报时间", "Last reported"],
  lastChecked: ["状态核实时间", "Status checked"],
  checkStatus: ["检查状态", "Check status"],
  taskChecking: ["核实中", "Checking"],
  taskStopped: ["已停止", "Stopped"],
  retryInstall: ["重新安装", "Retry install"],
  inspectHistory: ["核实历史", "Verify history"],
  historyChecking: ["正在核实文件…", "Verifying files…"],
  historyEvidence: ["历史核实结果", "Historical evidence"],
  taskCompleted: ["安装任务已完成", "Installation task completed"],
  needsRepair: ["需要修复", "Needs repair"],
  completedAt: ["完成时间", "Completed at"],
  verifiedCompletionBrief: ["已确认安装完成，文件完整。", "Installation completed and files verified."],
  taskDetails: ["详情", "Details"],
  expandHistory: ["展开", "Expand"],
  collapseHistory: ["收起", "Collapse"],
  showHistory: ["显示核实历史", "Show verification history"],
  hideHistory: ["隐藏核实历史", "Hide verification history"],
  stoppedBrief: ["未收到完成确认，可展开详情核实或重新安装。", "Completion was not confirmed. Check details or retry installation."],
  "short-origin-web": ["页面", "UI"],
  "short-origin-cli": ["命令行", "CLI"],
  version: ["安装版本", "Installed version"],
  installedAt: ["安装记录时间", "Installation recorded"],
  "evidence-installation-missing": ["当前版本没有安装记录。", "No installation record for the current version."],
  "evidence-installation-verified": ["当前安装文件已通过完整性校验。", "Current installation files passed integrity checks."],
  "evidence-installation-unverified": ["存在安装记录，但完整性未能确认。", "An installation record exists, but integrity could not be confirmed."],
  "evidence-installation-unchecked": ["安装状态暂未核实，可能仍有操作占用此包。", "Installation is not yet verified; package work may still be active."],
  "evidence-cache-verified": ["完整下载缓存可用。", "Verified download cache available."],
  "evidence-cache-missing": ["未找到当前版本的完整下载缓存。", "No complete download cache for the current version."],
  "evidence-cache-unverified": ["发现缓存，但校验不通过，不能作为完成证据。", "Cache found but verification failed; it is not completion evidence."],
  "evidence-cache-unchecked": ["下载缓存尚未核实。", "Download cache has not been verified."],
  historyConfirmed: ["安装记录与此任务 ID 一致，可确认该任务已完成安装。", "The installation record matches this task ID, confirming completion."],
  historyUnattributed: ["安装记录未关联此任务，不能据此判定这次历史任务成功。", "The installation is not linked to this task and cannot prove this historical run succeeded."],
  historyResidual: ["历史下载残留 {size}（{count} 个文件）。", "Residual downloads: {size} ({count} files)."],
  historyNoResidual: ["未发现残留下载文件；最后上报的字节数不代表文件仍可用。", "No residual downloads found; last-reported bytes do not imply reusable files remain."],
  historyTaskBytes: ["其中可归属此任务的实际下载文件：{size}（尚未确认完整）。", "Actual download files linked to this task: {size} (completeness unverified)."],
  evidenceChecked: ["文件核实时间", "Files verified at"],
  "origin-cli": ["来自命令行", "Started from CLI"],
  "origin-web": ["来自界面", "Started from UI"],
  "stage-starting": ["正在准备安装", "Preparing installation"],
  failed: ["执行失败", "Failed"],
  notFinished: ["任务未完成", "Task incomplete"],
  suggestion: ["建议", "Suggestion"],
  success: ["操作已完成。", "Operation completed."],
  interrupted: [
    "连接中断，任务状态未知。请刷新检查后再操作。",
    "Connection interrupted. Refresh and verify the task state before retrying.",
  ],
  serviceUnavailable: [
    "无法连接本地 UI 服务，任务状态未知。请确认终端中的 fnsh ui 仍在运行；若已退出，请重新启动并打开新的完整链接，检查安装状态后再重试。",
    "Cannot reach the local UI service; task state is unknown. Check that fnsh ui is running in your terminal. If it stopped, restart it, open the new complete link, and check installation status before retrying.",
  ],
  expired: [
    "会话已失效，请从终端重新打开完整 UI 链接。",
    "Session expired. Reopen the complete UI link from your terminal.",
  ],
  busy: [
    "另一个任务正在执行，请稍后重试。",
    "Another task is running. Please try again shortly.",
  ],
  notFound: ["未找到这个工具。", "This tool was not found."],
  footer: ["本地运行，数据由你掌握。", "Runs locally. Your data, your control."],
  literal: ["直接输入文本", "Literal text"],
  fileMode: ["从文件读取所有输入", "Read all inputs from files"],
  selectMode: ["选择处理方式", "Choose a mode"],
  "kind-file": ["选择已有的本地文件。", "Select an existing local file."],
  "kind-path": ["选择已有文件或目录。", "Select an existing file or folder."],
  "kind-directory": [
    "选择已有的本地目录。",
    "Select an existing local folder.",
  ],
  "kind-output-file": [
    "选择保存目录并填写文件名；运行时才会写入文件。",
    "Choose a folder and file name. The tool writes the file when run.",
  ],
  "kind-output-directory": [
    "选择父目录并填写新目录名；运行时才会创建目录。",
    "Choose a parent folder and new folder name. The tool creates it when run.",
  ],
  "kind-text-or-file": [
    "可直接输入文本，也可以选择本地文件。",
    "Enter literal text or choose a local file.",
  ],
  "stage-downloading": ["正在下载", "Downloading"],
  "stage-connecting": ["正在连接下载源", "Connecting to download source"],
  "stage-source-failed": ["当前下载源失败", "Download source failed"],
  "stage-source-retrying": ["正在重试当前下载源（仅一次）", "Retrying source once"],
  downloadSource: ["下载源", "Source"],
  switchingSource: ["正在切换到下载源", "Switching to source"],
  noSourcesRemaining: ["所有下载源均已尝试", "No download sources remaining"],
  "stage-ready": ["安装已就绪", "Installation ready"],
  "stage-verifying": ["正在校验", "Verifying"],
  "stage-extracting": ["正在解压", "Extracting"],
  "stage-cache-hit": ["使用已校验缓存", "Using verified cache"],
  "stage-already-ready": ["依赖已就绪", "Already ready"],
  "stage-installed": ["安装完成", "Installed"],
};

const labels = Object.fromEntries(
  `input=输入内容|output=输出位置|entries=压缩包内的成员路径|entry=压缩包内的原路径|files=附加文件|format=文件格式|password-env=密码环境变量名|new-password-env=新密码环境变量名|drop-password=移除加密|overwrite=覆盖已有文件|level=压缩等级|name=名称|names=名称列表|input-mode=输入方式|path=路径|second=第二个文件|from=输入格式或单位|to=输出格式或单位|mode=处理模式|type=类型|language=语言|lang=语言标签|algorithm=哈希算法|axis=方向|preset=预设|position=位置|permissions=权限|angle=角度|engine=处理引擎|paper=纸张尺寸|wrap=自动换行|eol=换行符|track-changes=修订处理|highlight-style=代码高亮主题|style=样式|grid=格线样式|method=计算方式|direction=转换方向|unit=时间单位|keep=保留方式|kind=连接类型|action=处理方式|date=日期|year=年|month=月|day=日|days=天数|years=年数|months=月数|count=数量|first=起始页|last=结束页|dpi=分辨率 DPI|quality=质量|width=宽度|height=高度|size=尺寸|start=开始位置|end=结束位置|duration=持续时间|time=时间|interval=间隔|seconds=秒数|text=文本|old=查找文本|new=替换文本|replacement=替换内容|pattern=匹配表达式|expression=表达式|key=密钥|keys=键名列表|password=密码|owner-password=所有者密码|user-password=用户密码|font=字体文件|reference=参考文档|bibliography=参考文献文件|csl=引文样式文件|schema=校验规则文件|signature=签名图片|asset=素材文件|background=背景颜色|foreground=前景颜色|color=颜色|characters=字符集|delimiter=字段分隔符|out-delimiter=输出分隔符|separator=分隔符|columns=列名列表|column=列名|left-keys=左表键名|right-keys=右表键名|fields=字段列表|values=值列表|value=值|pages=页面范围|title=标题|author=作者|creator=创建者|subject=主题|phone=电话|email=邮箱|organization=组织|url=网址|voice=语音名称|timezone=时区|zones=时区列表|after=起始时间|as-of=参考日期|birth-date=出生日期|production-date=生产日期|holidays=节假日|working-dates=额外工作日|principal=本金|annual-rate=年利率|salary=工资|rates=缴费比例|contribution=每期缴存额|floor=下限|ceiling=上限|weight-kg=体重（千克）|height-cm=身高（厘米）|gain=增益|volume=音量|volume-mib=分卷大小（MiB）|audio-kbps=音频比特率|target-mb=目标大小（MB）|fps=帧率|crf=编码质量 CRF|lufs=目标响度|factor=倍率|rate=速率|threshold=阈值|tolerance=容差|opacity=透明度|brightness=亮度|saturation=饱和度|hue=色相|radius=半径|sigma=标准差|x=横坐标|y=纵坐标|left=左边距|right=右边距|top=上边距|bottom=下边距|margin=边距|gap=间距|rows=行数|loops=循环次数|delay=延迟|length=长度|limit=数量上限|min=最小值|max=最大值|seed=随机种子|precision=精度|decimals=小数位数|indent=缩进|index=索引|span=每组页数|scale-percent=缩放百分比|min-quality=最低质量|max-quality=最高质量|jpeg-quality=JPEG 质量|bytes=字节数|gib=容量（GiB）|alignment-kib=对齐大小（KiB）|parent-one=父方血型|parent-two=母方血型|before=原始内容|cover=载体文本|message=消息文本|mapping=字符映射|markdown=Markdown 内容|prefixes=前缀列表|suffixes=后缀列表|alphabet=允许的字符|instant=时间点|per-year=每年期数|periods=期数|enum-limit=枚举上限|layout=布局|reduction=缩减量|shift-headings=标题级别偏移|toc-depth=目录深度|audio=包含音频|backfill=向前填充|block=块大小|boolean=布尔值|crlf=使用 CRLF 换行|cursor=显示光标|descending=降序排列|extended=扩展格式|heteronym=多音字|ignore-case=忽略大小写|invert=反向选择|leap=闰月|literal=按字面量处理|money=金额格式|nulls=处理空值|number-sections=章节编号|null=空值|pad=补齐|quote-all=为所有字段加引号|repeat=重复|reverse=反转|standalone=生成独立文档|toc=生成目录|trim=去除首尾空白|unique=去重|unwrap=合并段落内换行|upscale=允许放大|url-safe=URL 安全编码`
    .split("|")
    .map((s) => s.split("=")),
);
const choiceLabels = {
  auto: "自动",
  core: "内置引擎",
  imagemagick: "ImageMagick",
  horizontal: "水平",
  vertical: "垂直",
  in: "淡入",
  out: "淡出",
  before: "之前",
  after: "之后",
  none: "不启用",
  all: "全部",
  print: "允许打印",
  strict: "严格",
  relaxed: "宽松",
  text: "文本",
  image: "图片",
  pdf: "PDF",
  literal: "字面文本",
  exact: "精确匹配",
  regex: "正则表达式",
  contains: "包含",
  upper: "大写",
  lower: "小写",
  title: "单词首字母大写",
  first: "第一项",
  last: "最后一项",
  rows: "按行",
  rowskey: "对齐表头后按行",
  columns: "按列",
  previous: "前一个值",
  value: "指定值",
  numeric: "数值",
  natural: "自然排序",
  accept: "接受修订",
  reject: "拒绝修订",
  preserve: "保留",
  s2t: "简体转繁体",
  t2s: "繁体转简体",
  plain: "无声调",
  tone: "带声调",
  number: "数字声调",
  initials: "首字母",
  space: "空白规范化",
  lines: "行规范化",
  fullwidth: "全角",
  halfwidth: "半角",
  square: "方格",
  cross: "田字格",
  rice: "米字格",
  eng: "英文",
  chi_sim: "简体中文",
  "chi_sim+eng": "中英文",
  seconds: "秒",
  milliseconds: "毫秒",
  annuity: "等额本息",
  "equal-principal": "等额本金",
  "one-inch": "一寸",
  "two-inch": "二寸",
  passport: "护照",
  custom: "自定义",
};
export function resolveLanguage(saved, systemLanguage) {
  if (["zh", "en"].includes(saved)) return saved;
  return (systemLanguage || "zh").toLowerCase().startsWith("zh") ? "zh" : "en";
}
let savedLanguage;
try {
  savedLanguage = localStorage.getItem("finishbit-language");
} catch {}
let locale = resolveLanguage(savedLanguage, globalThis.navigator?.language);
export function language() {
  return locale;
}
export function setLanguage(value) {
  locale = value === "en" ? "en" : "zh";
  try {
    localStorage.setItem("finishbit-language", locale);
  } catch {}
}
export function t(key, values = {}) {
  let s = messages[key]?.[locale === "zh" ? 0 : 1] || key;
  for (const [k, v] of Object.entries(values))
    s = s.replaceAll("{" + k + "}", String(v));
  return s;
}
export function operationName(op) {
  return locale === "zh"
    ? (op.aliases || []).find((x) => /[\u3400-\u9fff]/.test(x)) || op.summary
    : op.summary;
}
export function parameterName(p) {
  if (locale === "zh" && p.name === "level" && p.choices?.includes("L"))
    return "纠错等级";
  return locale === "zh"
    ? labels[p.name] || p.name
    : p.name.replaceAll("-", " ");
}
export function choiceName(value, p) {
  if (locale === "zh" && value === "plain")
    return p.name === "style" ? "无声调 · plain" : "纯文本 · plain";
  if (p.name === "input-mode")
    return value === "literal"
      ? t("literal")
      : value === "file"
        ? t("fileMode")
        : value;
  return locale === "zh" && choiceLabels[value]
    ? `${choiceLabels[value]} · ${value}`
    : value;
}

const help = {
  entries:
    "填写压缩包内部的相对路径，每行一项。这里不是本地磁盘路径；留空时的行为请查看该工具的完整契约。",
  entry: "填写压缩包内现有成员的相对路径。",
  "password-env":
    "填写保存密码的环境变量名称，不要填写密码本身；留空表示不使用密码。",
  "new-password-env": "填写保存新密码的环境变量名称；留空复用输入密码。",
  pages: "页面从 1 开始，支持范围和 odd/even；留空通常表示全部页面。",
  "input-mode": "“文件”方式会从文件中读取所有输入；界面不支持终端标准输入。",
  overwrite: "允许替换已有输出。运行前请核对输出路径。",
  format: "选择处理格式。自动模式会根据文件扩展名判断。",
  "volume-mib": "每个分卷的最大大小，单位为 MiB。",
  level: "压缩等级为 0–9；数值越高，通常压缩越慢。",
  timezone: "填写 IANA 时区名称，例如 Asia/Shanghai 或 UTC。",
  delimiter: "填写一个字段分隔字符，例如逗号。",
  "owner-password": "用于管理 PDF 权限的所有者密码。",
  "user-password": "打开加密 PDF 时使用的密码。",
};
export function parameterHelp(p) {
  if (locale === "en") return p.description;
  if (p.name === "level" && p.choices?.includes("L"))
    return "选择二维码纠错等级 L、M、Q 或 H。";
  if (p.name === "format" && !p.choices?.length) return "";
  if (help[p.name]) return help[p.name];
  if (p.kind) return t("kind-" + p.kind);
  if (p.choices?.length) return t("selectMode");
  if (p.type === "strings") return t("lines");
  return "";
}

export function packageSummary(name) {
  return (
    {
      ffmpeg: "本地音视频转换与处理",
      ffprobe: "音视频信息检查",
      pdfcpu: "PDF 页面、表单及附件处理",
      poppler: "PDF 文字提取与页面渲染",
      imagemagick: "图片转换、编辑与合成",
      qsv: "CSV 表格清洗、关联与统计",
      pandoc: "文档与参考文献格式转换",
      "7zip-bootstrap": "7z 压缩包引导解压工具",
      "7zip-full": "完整的压缩包读取与依赖解压工具",
      tesseract: "离线中英文文字识别",
      "7zip": "压缩包创建与解压",
    }[name] || name
  );
}
