# Roadmap

The roadmap is capability-driven rather than a promise of dates.

## Available in v0.1

- Stable Operation registry, progressive search, describe, and execution.
- Core Go operations for encoding, JSON, text, time, UUID, files, and URLs.
- Managed FFmpeg installation plus video trim/compress and audio extraction.
- Local directory/ZIP extension installation and process protocol v1.
- Agent Skill and structured CLI output.

## Delivered in v0.1.1–v0.1.3

- CLI input/output and extension recovery fixes.
- Structured JSON requests through the shared application service.
- Core PNG/JPEG inspection, resizing and conversion.
- CSV inspection and CSV/JSON conversion with explicit column selection.

## Delivered in v0.1.4

Batch A adds 97 Operations: 33 media, 35 pdfcpu, 7 Poppler and 22 ImageMagick. Batch A brought the catalog to 123 Operations; the existing three image Operations also support an explicit ImageMagick engine. Windows x64 acceptance uses pinned managed binaries. See [batch acceptance](batch-a-acceptance.md).

The first group within batch B adds 27 qsv table-cleaning Operations, bringing the source catalog to 150 at that stage. See [table contracts and acceptance](tables.md). Excelize remains deferred until its upstream parsing advisory has an acceptable fix. The Pandoc group adds 10 further Operations (160 total), with [document contracts and reproducible acceptance](documents.md). The 10-operation [archive group](archives.md) follows, taking the source catalog to 170. Both groups passed [local real-runtime acceptance](batch-b-acceptance.md) and have synchronized Web/README contracts.

## Candidate next work

- Spreadsheet cleaning and Excel workbooks (batch B).
- Document conversion and archive Operations (batch B).
- Signed remote extension registry and publisher identity.
- Hybrid lexical/vector search when catalog size justifies it.
- HTTP/job adapter built solely on `pkg/app`.

Protocol stability, package supply-chain security, and cross-platform behavior take precedence over operation count.

## How priorities are chosen

Roadmap order reflects user task coverage, reuse across agents, deterministic behavior, maintenance cost, supply-chain risk, and fit with the existing Operation model. A listed item is a direction, not a commitment to a version or date.

Concrete proposals should use the repository's capability issue template and include the proposed Operation contract. Accepted proposals may change as implementation reveals compatibility or security constraints. See [governance](../GOVERNANCE.md) for the decision process.

## 工具箱扩充提案（2026-09-07）

状态：A 批已实现并完成 Windows x64 真实引擎验收，纳入 v0.1.4；B 批的 qsv 表格、Pandoc 文档、7-Zip 压缩包组已实现并通过验收，纳入 v0.1.4；Excelize 暂缓，C 批仍待落地。未实现能力的 Operation ID 是建议命名，发布前需完成契约评审。开源项目的一手资料见 [工具箱调研](research/open-source-toolboxes.md)。以下优先级是针对 FinishBit 的产品判断，不代表用户量统计。

### 最新方向：以成熟依赖扩展整类能力

根据进一步讨论，扩充规模不再限于下文的 12 个轻依赖 Operation。最新优先级以本节为准；后文保留已有能力对照和具体契约边界作为候选池。允许引入高收益依赖：一个成熟引擎支持一组完整任务，FinishBit 负责将其转成统一、可发现、可验证的 Operation。

以下保留各批次的能力范围；A 批实际契约以 [音视频](media.md)、[PDF](pdf.md)、[图片增强](image-enhancements.md) 为准，B 批已实现契约见 [表格清洗](tables.md)、[文档](documents.md)和[压缩包](archives.md)，Excelize 及 C 批能力尚未实现。格式变体通常作为选项，不为每个格式单独增加 Operation。已有裁剪、压缩、提取音频、图片基础和 CSV 互转不重复计为新增。

| 批次 | 能力包与依赖 | 可新增的任务清单 | 选择理由与边界 |
| --- | --- | --- | --- |
| A | 音视频：现有 FFmpeg，补 ffprobe | 媒体信息、视频转码、封装转换、视频拼接、画面裁剪、缩放、旋转翻转、变速、抽帧、封面、联系表、转 GIF、图片序列转视频、水印、静音、替换音轨、字幕提取/封装/烧录；音频裁剪、拼接、转码、混音、音量、响度标准化、降噪、静音检测/移除、淡入淡出、波形图 | 收益最高的现有依赖深化。托管 FFmpeg/ffprobe 均固定为 6.1.1，并已完成实测。滤镜/编码器/字体/libass 按实际构建验证；封装转换不能保证任意音视频流兼容。字幕提取不等于语音识别。[官方 CLI](https://ffmpeg.org/ffmpeg.html)、[滤镜](https://ffmpeg.org/ffmpeg-filters.html) |
| A | PDF 编辑：优先评估 pdfcpu CLI | 文档信息、结构校验、合并、拆分、提取页面、删除页面、重排页面、旋转、裁剪、加水印、盖章、图片转 PDF、优化、加密、持密码解密、附件列表/提取/添加、表单字段读取/填充 | 一次引入形成 PDF 编辑能力包。页面重排等可由选择与合并组合实现；优化不承诺固定压缩率；文字提取、渲染和 OCR 另行提供。[pdfcpu](https://github.com/pdfcpu/pdfcpu) |
| A | PDF 读取：Poppler | 按页提取文字、带文字坐标提取、页面转 PNG/JPEG、嵌入图片提取、字体清单、页面尺寸与元数据 | 补上 PDF 内容读取与预览；输出图片需指定 DPI 和页数上限；扫描件需 OCR。[项目源码](https://gitlab.freedesktop.org/poppler/poppler) |
| A | 图片增强：ImageMagick | 裁剪、旋转翻转、方向校正、格式扩展、质量压缩、水印、文字标注、横竖拼图、网格联系表、画布扩展、透明背景处理、灰度、模糊、锐化、颜色调整、差异图、GIF 生成/拆帧、图标生成 | 与 0.1.3 的基础 PNG/JPEG 操作衔接；更多格式是已有转换能力的扩展。WebP/AVIF/HEIC 等按安装构建探测，不根据项目总格式列表承诺全支持。[工具](https://imagemagick.org/convert/)、[选项](https://imagemagick.org/command-line-options/)、[格式](https://imagemagick.org/formats/) |
| B | 表格清洗：qsv | 选列、改列名、筛行、按列排序、按键去重、纵向合并、关联、分组统计、频次、数据质量摘要、抽样、拆分、转置、透视、JSONL 转换、规则校验 | 从“格式互转”升级为清洗分析能力包；按固定发行构建核对命令及可选特性。[命令目录](https://github.com/dathere/qsv) |
| B | Excel：Excelize Go 库 | 工作簿信息、工作表列表/新增/删除/复制、区域读写、查找替换、CSV 导入导出、样式与列宽、冻结窗格、数据验证、公式写入、图表、图片插入、报表模板填充 | 允许为整类能力引入 Go 依赖；同一 Runner 后封装。旧 .xls 不算原生支持，不保证任意公式重算及完整 Excel 排版兼容。[官方文档](https://xuri.me/excelize/en/) |
| B | 文档格式：Pandoc | Markdown→DOCX/HTML/EPUB/PPTX、DOCX→Markdown、HTML→Markdown、EPUB→Markdown、文档资源提取、目录和参考文献输出 | 共享 document.convert 契约并暴露格式矩阵；重点是内容结构转换，不能承诺像素级保留布局。PDF 输出需要额外排版引擎，不把 PDF 当通用输入。[官方手册](https://pandoc.org/MANUAL.html) |
| B | 压缩包：7-Zip | 列表、打包、解压、完整性测试、带密码打包/解压、分卷打包、按文件选择提取、更新压缩包 | 形成多格式能力包；按读写格式矩阵提供选项，RAR 支持解压不等于支持创建。[官方介绍](https://www.7-zip.org/) |
| C | OCR：Tesseract + OCRmyPDF | 图片文字识别、多语言识别、文字框与置信度、hOCR/TSV、扫描 PDF 加文字层、方向校正、纠偏、OCR 文本旁路输出、重做 OCR | Tesseract 面向图像识别，OCRmyPDF 负责 PDF 流程；后者还需按版本核实 Python、PDF 引擎及其他依赖。语言包独立管理，不承诺复杂表格恢复。[Tesseract](https://tesseract-ocr.github.io/tessdoc/)、[OCRmyPDF](https://ocrmypdf.readthedocs.io/en/latest/) |
| C | Office 导出：LibreOffice headless | Word/Excel/PowerPoint/ODF 转 PDF、旧 Office 格式升级、表格转 CSV、文档转 HTML、按导出参数生成 PDF | 重依赖按需安装，主要补办公文件渲染；转换需要字体，复杂版式实测，不能声称与 Microsoft Office 完全一致。[启动参数](https://help.libreoffice.org/latest/en-US/text/shared/guide/start_parameters.html) |
| C | 元数据：ExifTool | 批量读取元数据、EXIF/GPS 提取、选定标签修改/清除、元数据复制、拍摄时间调整、按标签生成重命名计划 | 形成照片和媒体资料整理包；写入支持按具体格式和标签判断，删除标签不等于消除所有敏感内容。[官方文档](https://exiftool.org/exiftool_pod.html) |
| C | 大数据文件：DuckDB | 本地文件 SQL、跨文件关联、分组聚合、窗口分析、CSV/JSON/Parquet 转换、分区导出、结构检查 | 待大表需求明确后引入；与 qsv 基础清洗存在重叠，不同时做两套相同公开命令。[数据读取](https://duckdb.org/docs/current/data/overview) |
| C | 配置处理：yq | YAML/JSON/XML 转换、TOML 读取转 JSON/YAML、路径读取/更新/删除、合并、多文档拆分/汇总、数组筛选、键排序 | 新依赖带来格式与结构操作族；不承诺 TOML 写回，XML 映射须固定规则。[运算符](https://mikefarah.gitbook.io/yq/operators)、[TOML 边界](https://mikefarah.gitbook.io/yq/usage/toml) |

A 批已实现并完成真实引擎验收：深化现有音视频，建立 PDF 编辑/读取和图片增强能力包。B 批先完成 qsv 表格清洗，后续补 Excel、文档互转和压缩包；C 按后续需求引入。这里的分批表示先后关系，不绑定版本，也不是以固定工具数量作为验收标准。

若要展示成用户工具箱，按“音视频、PDF、图片、表格、Excel、文档、压缩包、OCR、元数据、配置”组织；依赖名称留在安装和诊断信息中。相同能力只有一份公开契约，现有 core 路径仍可承担基础格式，增强 Provider 的选择规则需要显式设计。

引入方式：大型引擎沿用现有托管包思路；适合 Go 的库允许直接依赖或先采用独立 CLI。每个候选先核实目标平台发行资产、额外运行时和许可证，再固定版本及校验和；安装引擎不代表安装其所有可选组件。当前包管理只支持已定义的包/可执行文件，复杂 OCR/Office 依赖安装还需扩展，不能直接沿用一个静态 FFmpeg 文件的假设。

### 已有类别的深化与新类别的扩展

以下对照以 v0.1.3 的 26 个 Operation 为基线：23 个 core，3 个 FFmpeg；A 批后为 123 个，B 批表格、文档和压缩包组加入后当前源码为 170 个。已有 JSON、文本、编码、时间、UUID、文件、PNG/JPEG、CSV 和媒体的基础功能，详见 [当前目录](operations.md)。

0.1.3 已完成图片和表格能力的首批交付。后续增加裁剪、按列去重等属于深化，不代表这一版未完成目标；PDF、压缩包、证书等才属于新增类别。以下条目描述尚未覆盖的任务，不以数量判断现有版本是否合格。

- `json.query` 目前是点路径取值；缺少结构化比较、合并和 JSONL。
- CSV 目前能检查和互转；缺少筛选、按列去重、连接及数据质量报告。
- `file.info` 的 MIME 来自扩展名；尚不支持内容探测、重复文件分组或目录清单。
- `text.replace` 目前是字面替换；缺少正则提取、差异和脱敏。
- v0.1.3 图片仅有 PNG/JPEG 基础处理；A 批已补图片增强、EXIF 方向和 PDF。B 批已补充 Pandoc 文档及 7-Zip 压缩包能力；Excelize 相关 XLSX 能力仍暂缓。

### 早期候选池：高频、轻依赖的 12 个 Operation（不作为当前批次顺序）

这是最初基于 26 项目录提出的轻依赖候选池；当前执行顺序以上方 A/B/C 批次为准。它们原则上可用 Go 标准库实现，依然需要逐项验证输入边界和资源使用。先做前三项，尽早获得真实任务反馈，再推进同批其余能力。

| 顺序 | 候选 Operation | 解决的具体任务 | 首版范围与验收重点 |
| --- | --- | --- | --- |
| 1 | `json.diff` | 比较两个 API 响应或配置 | 输出 added/removed/changed、JSON Pointer 路径及新旧值；对象键顺序无关，数组按位置；保留大整数精度 |
| 2 | `csv.filter` | 从导出表筛选订单、用户、日志 | 显式列名、比较方式和值；默认字符串比较，数值模式显式选择；空值、缺列规则明确 |
| 3 | `text.extract` | 从文本按规则抽取字段 | Go 正则、命名分组、匹配位置和数量上限；声明不支持回溯引用和环视 |
| 4 | `csv.deduplicate` | 按业务主键去重 | 一个或多个键列，保留首条/末条；稳定顺序，返回删除行数 |
| 5 | `csv.select` | 抽列、重排列顺序 | 精确列名；缺列报错；保留身份证号、邮编等前导零 |
| 6 | `json.merge` | 合并基础配置与环境配置 | 显式冲突策略、数组替换规则；null 不暗中当成删除；不声称兼容尚未实现的补丁标准 |
| 7 | `jsonl.validate` | 检查日志或模型数据集 | 逐行解析，报告错误行号与有效行数；明确空行行为；限制单行大小及错误条数 |
| 8 | `jsonl.to-json` | 将 JSONL 转成 JSON 数组 | 明确总输入上限；不容许静默跳过坏行 |
| 9 | `json.to-jsonl` | 将 JSON 数组拆成逐行记录 | 只接受顶层数组；对象内的换行正确转义 |
| 10 | `url.parse` | 拆解链接、分析查询参数 | 返回协议、主机、端口、路径、重复参数；不发起网络请求 |
| 11 | `jwt.inspect` | 查看令牌头部、载荷和有效期字段 | 限三段 JWS 格式；显式返回 signature_verified=false；若判断过期，允许传入固定参考时间 |
| 12 | `file.detect` | 识别扩展名错误的文件 | 内容签名与扩展名分开返回；只保证已支持格式，未知类型保留 unknown |

### 第二批：扩大普通用户的任务覆盖

| 方向 | 候选 Operation | 典型使用场景 | 实现路径与复杂度判断 |
| --- | --- | --- | --- |
| PDF 基础 | `pdf.info`、`pdf.merge`、`pdf.split`、`pdf.extract-text` | 合并附件、拆分页码、读取电子文档正文 | 托管 Provider；先选型再承诺格式兼容；扫描件文字提取需另行 OCR |
| 压缩包 | `archive.list`、`archive.create`、`archive.extract` | 检查、打包、解压交付文件 | Core，首版 ZIP；路径越界、链接、解压大小及同名覆盖必须定义 |
| 表格互通 | `xlsx.info`、`xlsx.to-csv`、`csv.to-xlsx` | 接收 Excel 数据并接入现有 CSV 路径 | 独立扩展或评估新依赖；明确工作表选择、日期与公式缓存语义 |
| 配置互通 | `yaml.to-json`、`json.to-yaml`、`toml.to-json` | 配置检查和格式转换 | 需成熟解析库或扩展；重复键、日期、别名及非字符串键须明示；转换不承诺保留注释 |
| 图片实用 | `image.crop`、`image.rotate`、`image.contact-sheet` | 裁剪截图、旋转照片、一图预览多张素材 | Core 可先覆盖 PNG/JPEG；EXIF 方向规则先统一 |
| 二维码 | `qrcode.generate`、`qrcode.decode` | 把文本/链接制成图片或读取截图里的二维码 | 先扩展验证需求；成熟编码/识别库，新增依赖需评估 |
| 音视频 | `media.info`、`video.thumbnail`、`audio.convert`、`audio.normalize` | 查看媒体参数、提取封面、统一音频格式与响度 | 扩展现有 FFmpeg Provider；检查 ffprobe 的托管可用性，不默认系统已安装 |
| 开发辅助 | `text.diff`、`cron.next`、`certificate.inspect` | 比较文本、预览计划执行时间、查看本地证书 | Core 或轻依赖；Cron 必须指定方言及时区，文本 diff 限制规模 |

表格方向参考 [qsv 的命令目录](https://github.com/dathere/qsv)：它把筛选、选择、去重、连接、统计拆成可组合命令。更大规模的 CSV/JSON/Parquet 分析可另行评估 [DuckDB 的数据读取能力](https://duckdb.org/docs/current/data/overview)，暂不为简单 CSV 操作引入完整查询引擎。

### 第三批：更有 FinishBit 特色的任务能力

这些是产品提案，不声称参考项目具备完全相同的契约。

| 方向 | 候选 Operation | 为什么值得做 | 必须讲清的边界 |
| --- | --- | --- | --- |
| 有报告的脱敏 | `text.redact`、`json.redact`、`csv.redact` | 分享日志和样本前，按字段/规则替换并报告命中位置与数量 | 报告不回显原始敏感值；规则匹配不等于找出所有隐私；输出新文件 |
| 数据质量检查 | `csv.profile`、`csv.diff` | 给出空值、重复键、类型异常；按主键解释两份表哪里变了 | 区分全量和抽样；主键重复必须显式处理，不能静默丢行 |
| 重复文件报告 | `file.duplicates` | 整理下载目录和重复附件 | 按大小初筛、内容哈希分组；只报告；记录无法读取及扫描中发生变化的文件 |
| 交付清单 | `file.manifest`、`file.verify` | 发出一批文件后，可核对是否缺失或被修改 | 相对路径、大小、SHA-256；明确符号链接及平台路径规则 |
| 视频联系表 | `video.contact-sheet` | 把长视频按时间抽帧拼成带时间戳的总览，便于 Agent 定位片段 | 提供帧时间与拼图坐标；抽帧策略确定；不把抽帧描述成理解视频 |
| 文档内容提取 | `document.extract`、`ocr.extract` | 将 PDF/图片/办公文档内容转成可追溯的文本块 | 分别评估格式 Provider；保留页码和来源；OCR 输出声明模型/语言/版本 |
| 数据形态建议 | `data.inspect` | 输入未知内容，返回可能格式和适用的下一步 Operation | 候选建议和依据可检查；不自动执行解码或解压，不把猜测当确定结果 |

### 组合方式与实现约束

优先把单步做好，再提供示例请求串联：CSV 筛选 → 按键去重 → 脱敏 → 导出；视频信息 → 联系表 → 使用用户选定时间裁剪；PDF 提取 → 正则抽取 → JSONL 导出。工作流引擎暂缓，先验证是否确有超出已有 Agent 编排的需求。

新增能力遵循现有架构：复用用例放在 `pkg/app`，能力契约放在 `pkg/operation`，实现置于 `operation.Runner` 后；CLI 继续只做适配。多文件输入可先用已有字符串数组 options 表达，不暗中改变位置参数的数量校验。

所有新增 Operation 应可搜索、可 describe、可通过通用 run 调用；给出中英文搜索词、结构化结果和错误。落盘操作明确覆盖策略，批量任务报告逐项结果，严格区分成功、部分成功和失败。资源上限、取消执行、跨平台行为属于每项验收内容。

引入 PDF、OCR、Office 或数据引擎时独立评估 Provider 的体积、许可证、固定版本、校验和及跨平台安装，不直接把整个 Web 工具箱嵌入核心。核心保持精简，已为托管安装引入 ZIP/7z/xz 等解包依赖；继续采用新库时需记录用途与取舍。

优先暂缓：在线翻译/摘要等依赖模型的工具、账户集成、实时网络诊断、完整工作流编辑器，以及只有视觉交互价值的生成器。它们可以后续进入扩展生态，本轮先提升本地、结构化、可验证的任务覆盖。
