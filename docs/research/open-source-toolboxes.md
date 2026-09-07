# 开源工具箱能力调研

调研日期：2026-09-07。仅使用项目 README、源码目录和官方文档；下列“借鉴方向”和优先级是针对 FinishBit 的产品判断，不代表上游项目承诺。“高频”指覆盖常见工作场景，未取得实际使用量统计。

## 项目对照

| 项目 | 已核实的能力 | 对 CLI / agent 的借鉴方向（建议） | 核实的许可信息 |
| --- | --- | --- | --- |
| [IT-Tools](https://github.com/CorentinTh/it-tools) | 工具注册源码包含 JSON/YAML/TOML 转换、JSON diff、JWT 解析、CIDR 计算、时间转换、二维码、Docker run 转 Compose 等。[工具清单源码](https://raw.githubusercontent.com/CorentinTh/it-tools/main/src/tools/index.ts) | 补齐本地数据转换；特色工具优先考虑 Docker 命令转换、结构化 diff、CIDR。 | 仓库 README 标注 GPLv3。 |
| [DevToys](https://github.com/DevToys-app/DevToys) | 包含格式化、编解码、哈希、JSONPath、正则、图片压缩和文本比较；支持根据剪贴板数据选择工具的 Smart Detection，并支持扩展。 | 输入识别后推荐可用 operation；错误说明和样例比增加大量独立小命令更有价值。 | 仓库标注 MIT。 |
| [CyberChef](https://github.com/gchq/CyberChef) | 编码、解压、字符集、哈希、X.509 解析；可保存组合 recipe，逐步执行、设置断点，并尝试识别嵌套编码。 | 声明式转换配方，以及可查看每一步中间结果的解释模式；编码识别输出候选及理由，避免静默猜测。 | README 标注 Apache-2.0。 |
| [Stirling-PDF](https://github.com/Stirling-Tools/Stirling-PDF) | 合并、拆分、签名、涂黑、转换、OCR、压缩；支持工作流及多数工具的 REST API。 | 面向办公交付：扫描件变可搜索 PDF、按页拆分合并、批量压缩。 | 当前是 open-core；根许可默认 MIT，但 engine、proprietary、部分前端目录有独立许可，不能概括为全仓库 MIT。[根许可证](https://raw.githubusercontent.com/Stirling-Tools/Stirling-PDF/main/LICENSE) |
| [FFmpeg](https://ffmpeg.org/ffmpeg.html) / [ffprobe](https://ffmpeg.org/ffprobe.html) | 转码、帧率与码率控制、缩放及滤镜管道；ffprobe 可输出容器、流和元数据的机器可读信息。 | 先探测再处理：统一输出视频信息、抽取音轨、缩小视频和生成预览的任务预设；复用成熟引擎。 | LGPL-2.1-or-later 为基础，启用部分可选组件会适用 GPL；具体取决于构建。[官方说明](https://ffmpeg.org/legal.html) |
| [qsv](https://github.com/dathere/qsv) | 表格筛选、排序、统计、校验、关联、转换；包含 CSV/JSONL 互转、按列分区、抽样和列值假名化。 | “拿到陌生表格→理解结构→清洗→导出”的连续流程；适合批量文件和 agent 的结构化摘要。 | README 标注 MIT，第三方组件另列 notices。 |
| [jq](https://github.com/jqlang/jq) | CLI 对 JSON 执行切片、筛选、映射和变换。 | JSON 查询和字段抽取应能接 stdin、输出稳定 JSON，并与其他 operation 串联。 | 程序 MIT；文档 CC BY 3.0，部分 decNumber 代码采用 ICU 许可。 |

## 与 FinishBit 的关系

0.1.3 已经增加 PNG/JPEG 信息读取、缩放、转换，以及 CSV 检查、CSV/JSON 互转。这些是已交付能力，不应重复列为待补充功能。JSON 点路径查询、文本行排序/去重、时间转换、音轨提取等也已经存在。

本次调研分为“已有类别的深化”和“新类别的扩展”。去重后的候选、范围和优先级统一维护在 [路线图](../roadmap.md)，当前可用命令以 [Operation 目录](../operations.md) 和源码注册为准。

## 形成特色的方法

1. 优先把“识别→处理→核验→交付”做成完整任务。例如表格清洗同时输出结果文件和行数、空值、重复数变化，而不只输出“成功”。
2. 纯文本和结构化数据能力尽量在本地运行；PDF/OCR/媒体使用可选引擎并明确依赖检测、版本、取消和超时。
3. 相同能力同时服务 CLI 和未来 Web：参数、结果、错误码放在共享契约，组合用例留在 `pkg/app`，执行通过 `operation.Runner`。
4. 不以工具数量为主要目标。基础转换扩大覆盖面，输入识别、可复用配方和可核验交付形成特色。

## 尚需验证

- 路线图已经按源码清单去重；具体实施批次仍是提案，需用真实任务反馈校准。
- 上游能力存在不等于可以直接移植源码；实际引入时需针对固定版本及所用目录核实许可。
- 复杂 PDF 的 OCR、签名及涂黑，媒体目标体积等能力都需要真实样本验证，不能仅按功能名称承诺质量。

## 按依赖扩展的能力事实（2026-09-07 补充）

以下是上游引擎可以支撑的候选操作，不表示 FinishBit 已接入。依赖应按能力簇引入：PDF 结构编辑、页面渲染和 OCR 分属不同引擎，不把单一库包装成“所有 PDF 处理”。

| 引擎 | 可以封装的具体能力 | 接入边界与官方依据 |
| --- | --- | --- |
| pdfcpu（Go 库 / CLI） | PDF 信息与校验、合并、拆分、选页与重排、插入/删除页、旋转/裁切、水印/印章、结构优化、附件管理、密码与权限设置、表单导出/填写、CSV/JSON 批量填表 | 适合成组增加 PDF 结构操作；优化不能承诺任意扫描件大幅缩小，页面结构编辑不等于 OCR 或恢复 Word 排版。[命令目录](https://pdfcpu.io/getting_started/usage/)、[批量填表](https://pdfcpu.io/form/form_multifill/) |
| Poppler（CLI 工具集） | PDF 转页面图片、PDF 文本抽取、PDF 转 HTML、PDF 转 PostScript、内嵌图片提取、字体清单、页数/尺寸/元数据检查、附件提取 | 提取现有文本不等于扫描件 OCR；渲染需要核实实际构建包含的功能和 Windows 分发方式。[官网](https://poppler.freedesktop.org/)、[维护者镜像的工具构建清单](https://raw.githubusercontent.com/tsdgeos/poppler_mirror/master/utils/CMakeLists.txt) |
| OCRmyPDF + Tesseract（外部引擎） | 扫描 PDF 增加可搜索文字层、多语言图片 OCR、导出纯文本、输出带文字坐标的 hOCR/TSV、页面方向纠正、扫描倾斜校正、按页 OCR、已有 OCR 重做、PDF/A 归档输出、扫描件优化 | 是 Python 应用与 OCR/渲染依赖组合；识别准确率依赖语言包、扫描质量和版式，不保证手写识别或表格语义恢复。17.x 可使用 pypdfium2 等路径，不能笼统写成始终强制依赖 Ghostscript。[介绍与限制](https://ocrmypdf.readthedocs.io/en/latest/introduction.html)、[操作示例](https://ocrmypdf.readthedocs.io/en/latest/cookbook.html)、[Tesseract CLI](https://tesseract-ocr.github.io/tessdoc/Command-Line-Usage.html) |
| LibreOffice（headless） | Word 转 PDF、Excel 转 PDF、PowerPoint 转 PDF、DOC 转 DOCX、XLS 转 XLSX、PPT 转 PPTX、ODF 与 Office 转换、文档转 TXT/HTML、表格转 CSV | 适合办公格式兼容与渲染；外部安装体积较大，需可写用户配置目录。复杂版式、字体替代、宏和公式兼容要用样本验证，不应承诺 PDF 无损还原为 Office。[命令行参数](https://help.libreoffice.org/latest/en-US/text/shared/guide/start_parameters.html)、[转换过滤器](https://help.libreoffice.org/latest/en-US/text/shared/guide/convertfilters.html) |
| Excelize（Go 库） | XLSX 工作簿检查、工作表/范围读取、CSV/JSON 导入导出、工作表增删改名、单元格批量填充、样式与数字格式、公式读写、图表生成、图片插入、数据验证、工作表保护、大表流式读写 | 读取单元格后转换 CSV/JSON 等需 FinishBit 封装；聚焦 OOXML，不能当作旧 XLS 解析器或 PDF 渲染器。当前官网版本 v2.11.0 要求 Go 1.25.0+，仓库声明 Go 1.25.13，满足最低版本要求。BSD-3-Clause。[介绍](https://xuri.me/excelize/zh-hans/)、[工作簿](https://xuri.me/excelize/en/workbook.html)、[工作表](https://xuri.me/excelize/en/sheet.html)、[单元格与公式](https://xuri.me/excelize/en/cell.html) |
| 7-Zip（CLI） | 创建 ZIP、创建 7z、TAR 打包、GZIP/XZ 压缩解压、解压 RAR、解包 ISO/磁盘镜像内文件、AES-256 加密归档、7z 自解压包 | 上游明确区分可打包格式与仅解包格式：RAR 是仅解包，不承诺创建 RAR；多数代码 LGPL，另有 BSD 和 unRAR 限制，固定分发版本再核实。[官方格式与许可说明](https://www.7-zip.org/) |
| ExifTool（CLI） | EXIF/IPTC/XMP 检查、GPS 查看/清理、作者版权修改、拍摄时间修正、元数据复制、按拍摄日期重命名/归档、缩略图提取、元数据批量导出、地理标记写入 | 各格式读/写能力不同；元数据清理不等于文件内容脱敏，也不应被包装成 PDF 涂黑。依赖 Perl 或打包运行时；README 说明可按 Perl Artistic 或 GPL 条款分发。[官方仓库与格式表](https://github.com/exiftool/exiftool)、[完整功能文档](https://raw.githubusercontent.com/exiftool/exiftool/master/html/index.html) |
