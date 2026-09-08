# 本地工具扩展

v0.1.6 在 170 个 Operation 基础上新增 **107 个，合计 277 个**。86 个新能力直接在 Go 二进制内运行，另外复用本地 PDF、图片、OCR、语音和录屏运行时。CLI 和后续 Web 适配器共用 `pkg/app` / `operation.Runner`；没有上传文件的后端、第三方处理 API 或常驻服务。

首次安装可选工具需要下载固定版本和校验散列；实际处理不需要网络。网站参照清单有 166 个工具入口，入口数不等于 Operation 数，基础子集也不等于网站全部效果。参见[逐项实现状态](research/qq-tool-implementation-2026-09-07.md)。

## 使用

```powershell
fnsh search "二维码" --json
fnsh describe qrcode.generate --json
fnsh run qrcode.generate "你好 FinishBit" --output qr.png
fnsh run regex.replace 'order=42' '(\d+)' '[$1]' --json
fnsh run text.chinese "简体中文" --direction s2t --json
fnsh run date.add 2024-01-31 --months 1 --json
fnsh run unit.convert 1 GiB MB --json
fnsh run markdown.to-html README.md --input-mode file --output preview.html
fnsh run image.split-grid photo.png --columns 3 --rows 3 --output tiles
fnsh run image.target-size photo.png --bytes 102400 --output small.jpg
fnsh run worksheet.handwriting "认真练字" --output worksheet.html
```

新增文字工具默认将位置参数视为**字面值**。`--input-mode file` 把全部位置参数当作 UTF-8 文件；`--input-mode stdin` 只从标准输入读取第一个参数，其余仍为字面值。图片、PDF、文档工具的位置参数按各自契约读取本地文件。输出默认不覆盖，只有显式 `--overwrite` 才替换；结果先写临时文件，再发布。

## 已增加的能力

| 范围 | Operation / 边界 |
| --- | --- |
| 编码与文字 | Unicode 代理对、HTML、十六进制、URL 解析、大小写、空白和全半角、倒序、字符映射、颜文字；零宽隐写带版本与 CRC32 |
| 开发工具 | RE2 正则校验/匹配/替换，逐行文本和结构 JSON diff，JWT 解码，IP/CIDR 计算，UA 规则解析，Hello World 模板 |
| 数据格式 | YAML/JSON、TOML/JSON、XML 校验与格式化、Markdown GFM 转本地 HTML；格式转换不保留注释和原始排版 |
| 中文 | 拼音、声调/首字母/多音候选，OpenCC 词组优先简繁映射；不承诺多音消歧或地区用词转换 |
| 数字 | 任意精度进制和单位换算、温度、人民币大写、白名单数学表达式、基础统计、颜色换算 |
| 日期 | 月末截断的日期加减、日期差、工作日（用户提供调休）、年龄、保质期、世界时间、农历互转、Cron 后续时间计算；不修改时钟或建立任务 |
| 随机与密码 | 密码学随机整数/抽签/密码、用户词库名字生成、密码模式检查、AES-256-GCM 文本加解密 |
| 公式工具 | 复利、房贷还款表、用户比例的五险一金、BMI 数值、简化 ABO 可能性、容量计算；无实时政策或个性化建议 |
| 图片 | QR 生成/配色/名片/解码，马赛克、字符画、网格切图、调色板、取色、纯色画布、证件照尺寸裁切、PNG 隐写/增大字节数、JPEG 目标体积、7 种普通滤镜 |
| PDF | 独立四边裁切、元数据、动态页码、纸张尺寸、签名图片；纯图 PDF、JPEG 页面压缩、长图、重排 DOCX、每页图片 PPTX、扫描 OCR |
| 文档与展示 | DOCX/PPTX/XLSX 包内图片和 ZIP 压缩、提取文字对比、图片 OCR 转 DOCX、可打印字帖、离线 LED 弹幕、可折叠 Markdown 大纲 |
| Windows | 已安装 System.Speech 声音导出 WAV；FFmpeg 显式录制桌面或矩形到无声 MP4 |

## OCR 与文档流程

```powershell
fnsh pkg add tesseract
fnsh run ocr.text scan.png --language chi_sim+eng --json
fnsh run ocr.words scan.png --language eng --json
fnsh run ocr.to-pdf scan.png --output searchable.pdf
fnsh run pdf.ocr source.pdf --first 1 --last 5 --dpi 150 --output recognized.pdf
fnsh run pdf.to-docx source.pdf --first 1 --last 5 --output text.docx
fnsh run pdf.to-pptx source.pdf --first 1 --last 5 --output slides.pptx
fnsh run pdf.compress-images source.pdf --first 1 --last 5 --dpi 100 --quality 70 --output compact.pdf
fnsh run document.compress slides.pptx --output compact.pptx
fnsh run speech.voices --json
fnsh run speech.synthesize "FinishBit" --output speech.wav
```

PDF 组合流程须显式选页，默认仅第一页，一次最多 50 页；具体依赖以 `describe` 为准。普通图像工具限制 32 MiB / 2500 万像素；文字一般限制每项 4 MiB，部分算法更小。正则最多 64 捕获组、10000 匹配，替换输出最多 8 MiB；文本 diff 每边最多 2000 行。JSON diff 用精确十进制比较，数字限 512 字符、指数绝对值不超过 999。

OCR 用 Tesseract 5.5.0 和固定版本 tessdata_fast 中英文模型，目标是印刷文字。提供文字、词坐标/置信度、hOCR 和可搜索 PDF；不承诺手写、发票表格或证件字段结构化识别。扫描 PDF 会重新栅格化，原有链接、表单、签名不保留。Word 输出为文字重排，PPT 输出为每页图片。压缩 PDF 不保证更小，返回实测大小；缩选页后的大小不能当作全文件压缩率。

文档减重保留非图片成员，遇到数字签名包拒绝处理，带 EXIF/ICC/分辨率元数据的图片保守保留。如果重打包不更小，输出原文件副本。普通图片变换会重新编码，通常不保留元数据，也不自动应用 EXIF 方向；可先用 `image.orient`。证件照只做尺寸/中心裁切与已有透明背景填色，无人脸检测或自动抠图。PNG/文字隐写的 CRC32 只检测损坏，保密用途用 `crypto.encrypt`。

Windows 语音使用用户已安装的系统声音，没有新增声音下载，也不会自动播放。`screen.record` 只在用户执行时开始，默认 10 秒、最多 300 秒，不录声音。其他系统返回明确的不支持错误。屏幕录制本轮只验证参数和 FFmpeg 支持，未自动录制用户桌面。

## 体积、许可和待定项

同一 Windows x64、Go 1.26.4、`-trimpath -ldflags '-s -w'` 实测：二进制由 9,542,144 增至 16,111,616 字节，约 **9.10 → 15.37 MiB，增加 6.27 MiB**。超过最初建议的 5 MiB 评审线，主要换来了二维码识别、中文词典、历法和内置时区等独立离线能力；5 MiB 并非用户硬限制。没有增加 Python/Node/Java 前置环境。用户对可选依赖给出的阈值是 1 GB。

Tesseract 安装器下载约 20.4 MiB，中英文模型约 6.3 MiB；完整 OCR 解包与模型约 **78.4 MiB**，另需小型 7-Zip 解包助手，明显低于 1 GB。管理器仅提取安装器，不运行系统安装程序；校验 SHA-256，限制条目数和合计展开体积。固定资源、许可与测量见[依赖研究](research/lightweight-tool-dependencies-2026-09-07.md)和[第三方声明](../THIRD_PARTY_NOTICES.md)。

LibreOffice 下载约 356 MiB，但 MSI 文件表展开约 **1.47 GiB**，超过用户阈值，等待是否允许例外后再接 Office 保真转 PDF。AI 抠图的私有 Python/ONNX 方案还未完成具体模型权重许可溯源及真实安装/推理验证，未接入产品。实时快递、搜索等在线服务，以及来源不清的生活词库不纳入本地同等覆盖。

## 验证

单元测试覆盖转换边界、精度、异常输入、随机无重复、AES 认证、隐写校验、二维码回扫、图像像素、ZIP 路径和输出保护。真实引擎测试包含中英文 OCR、词坐标、可搜索 PDF、全部新增 PDF 组合、动态页码、独立裁切、Word/PPTX、文档减重和 WAV 文件验证。

```powershell
go test ./...
go vet ./...
go build ./cmd/fnsh
node scripts/website-catalog.mjs --check
$env:FINISHBIT_TEST_LOCAL_HOME='D:\finish-bit\.finishbit-test-local-tools'
go test ./pkg/app ./internal/desktop -run 'TestLocalWorkflowsWithManagedTools|TestInstalledSpeechWithoutPlayback' -count=1 -v
```

真实引擎验收需先在该独立 `FINISHBIT_HOME` 安装测试所需的 pdfcpu、Poppler、Pandoc、ImageMagick、Tesseract；不会为测试录屏或播放声音。
