# 轻量开源工具依赖核对（2026-09-07）

本轮为功能对齐的依赖预研，没有安装依赖或修改功能代码。核对了当前 `go.mod`、`internal/builtin`、`internal/raster`、`internal/document`、`internal/tabular` 及包管理和路线图文档。下表的 Operation 名称均为建议，不代表已经实现。

“轻量候选”在这里指能直接作为 Go 库、无需额外解释器或模型运行时的候选；不等于已经测量二进制体积。主分支 `go.mod` 没有 `require` 只说明当前模块清单没有外部模块依赖，不代表零内存开销。实际接入应固定发布版本或 commit，记录依赖树、同条件构建的二进制增量及峰值内存。本文不提供未经测量的体积数字。

## 候选清单

| 候选及建议能力 | 许可证依据 | 运行时、依赖与重量判断 | 接入边界与建议 |
| --- | --- | --- | --- |
| [makiuchi-d/gozxing](https://github.com/makiuchi-d/gozxing)：`qrcode.generate/decode`；EAN、UPC、Code128 等条码读写 | [LICENSE](https://github.com/makiuchi-d/gozxing/blob/master/LICENSE) 同时包含移植作者 MIT 和原 ZXing Apache-2.0，不能只标 MIT | 官方明确为纯 Go；[go.mod](https://github.com/makiuchi-d/gozxing/blob/master/go.mod) 引用 `x/text`、`x/xerrors`。项目已有 `x/text`，最终版本由模块选择结果决定；无需 Java | 首选：一个库覆盖二维码生成和识别。只导入实际需要的格式子包。官方格式矩阵没有完成 PDF417、MaxiCode，不能承诺全条码。测试多码、旋转、低对比、非 UTF-8 内容及像素上限 |
| [skip2/go-qrcode](https://github.com/skip2/go-qrcode)：文本/链接生成 PNG，纠错级别和颜色 | [MIT](https://github.com/skip2/go-qrcode/blob/master/LICENSE) | [go.mod](https://github.com/skip2/go-qrcode/blob/master/go.mod) 无 `require`；Go 编码库，不需命令行运行时 | 仅生成场景备选；采用 gozxing 后通常无需重复引入。Wi-Fi、vCard 只是内容模板，按格式转义即可；保留二维码静区 |
| [rwcarlsen/goexif](https://github.com/rwcarlsen/goexif)：`image.exif`，拍摄时间、GPS、方向标签读取 | [BSD-2-Clause](https://github.com/rwcarlsen/goexif/blob/master/LICENSE) | Go EXIF/TIFF 解码源码；[解码实现](https://github.com/rwcarlsen/goexif/blob/master/exif/exif.go) 使用 Go 标准库和仓库内 TIFF 包，不要求 Perl/ExifTool | 作为只读候选。不是通用 EXIF 写入库，也不覆盖全部 IPTC/XMP/厂商私有信息。接入前补畸形 IFD、越界偏移、无标签和相机样本；ImageMagick 已有方向校正，不重复计新功能 |
| [lucasb-eyer/go-colorful](https://github.com/lucasb-eyer/go-colorful)：`color.convert/palette/gradient/distance`，RGB/HSL/HSV/Lab、色差、渐变配色 | [MIT](https://github.com/lucasb-eyer/go-colorful/blob/master/LICENSE) | Go 色彩算法；[go.mod](https://github.com/lucasb-eyer/go-colorful/blob/master/go.mod) 无 `require`，无需图片处理运行时 | 色值和配色工具可独立内置。HEX/RGB/HSL 基础转换也可自行实现；需要 Lab、感知色差和配色时引库更有价值。它不替代已有 ImageMagick 图片调色，也不是 ICC 色彩管理引擎 |
| [yaml/go-yaml](https://github.com/yaml/go-yaml)：`yaml.format/validate/to-json`、`json.to-yaml` | README 声明 MIT 与 Apache-2.0；当前 [LICENSE](https://github.com/yaml/go-yaml/blob/main/LICENSE) 为 Apache-2.0，分发应一并核对所选版本 NOTICE 和文件头 | 官方 YAML 组织维护的纯 Go 实现；[v4 主分支 go.mod](https://github.com/yaml/go-yaml/blob/main/go.mod) 无 `require` | 原 `go-yaml/yaml` 已被作者标为不维护；新功能优先评估 `go.yaml.in/yaml/v4` 的实际可用发行版。定义 YAML 到 JSON 的非字符串键、大整数、时间标签、多文档、重复键和别名策略；不承诺格式化后注释原样保留 |
| [pelletier/go-toml/v2](https://github.com/pelletier/go-toml/tree/v2)：`toml.format/validate/to-json`、`json.to-toml` | [v2 MIT](https://github.com/pelletier/go-toml/blob/v2/LICENSE)；不要混用 v1 的许可证/依赖结论 | [v2 go.mod](https://github.com/pelletier/go-toml/blob/v2/go.mod) 无 `require`；Go 解析器，无外部运行时 | 适合结构化转换。JSON 的 null、异构类型、顶层数组等到 TOML 需显式拒绝或约定表示；日期类型和注释不能笼统承诺无损往返 |
| [beevik/etree](https://github.com/beevik/etree)：`xml.format/validate/query`，节点查询修改 | [BSD-2-Clause](https://github.com/beevik/etree/blob/master/LICENSE) | [go.mod](https://github.com/beevik/etree/blob/master/go.mod) 无 `require`；Go XML 树库 | 简单格式化/校验先用标准库 `encoding/xml`。需要树操作和路径查询再引 etree；树解析应限制输入与深度，不能把路径语法标为完整 XPath/XSD。XML→JSON 必须明确属性、同名元素、混合文本及命名空间映射 |
| [yuin/goldmark](https://github.com/yuin/goldmark)：`markdown.to-html`、标题/链接/任务清单提取 | [MIT](https://github.com/yuin/goldmark/blob/master/LICENSE) | 官方声明纯 Go、只依赖标准库；[go.mod](https://github.com/yuin/goldmark/blob/master/go.mod) 无 `require` | 可为已有 Pandoc Markdown→HTML 增加无需下载运行时的 core 引擎；GFM 表格、删除线、任务列表有内置扩展。保留默认禁用原始 HTML 与危险链接，不默认启用 `WithUnsafe`。DOCX/PDF 不在此库能力内 |
| [robfig/cron/v3](https://github.com/robfig/cron)：`cron.validate/next`，输出未来触发时间 | [MIT](https://github.com/robfig/cron/blob/master/LICENSE) | [go.mod](https://github.com/robfig/cron/blob/master/go.mod) 无 `require`；Go 库，无外部 cron 守护进程 | 只使用解析和 `Schedule.Next` 即可，不需启动调度服务。明确五字段默认语法、可选秒、时区和 DST 行为；它不是 cron 中文解释器，也不等于 Quartz 全语法 |
| [6tail/lunar-go](https://github.com/6tail/lunar-go)：`calendar.lunar/solar`、生肖、节气、农历月份与节日查询 | [MIT](https://github.com/6tail/lunar-go/blob/master/LICENSE) | 官方声明无第三方依赖；[go.mod](https://github.com/6tail/lunar-go/blob/master/go.mod) 无 `require`。内置日历算法和数据，需实测链接增量 | 核心对齐公农历转换和传统节日即可。明确支持年份、闰月和日期边界。法定假期/调休数据需要按年维护，不把算法结果当作实时官方工作日公告；无需将宜忌、占卜等整套能力一起暴露 |

## 主清单补充候选：拼音、简繁、差异

| 候选 | 官方依据与边界 |
| --- | --- |
| [mozillazg/go-pinyin](https://github.com/mozillazg/go-pinyin) | MIT；[go.mod](https://github.com/mozillazg/go-pinyin/blob/master/go.mod) 没有 require。支持声调与多音字候选，不提供上下文自动消歧。内含拼音数据，接入需连同其引用的 pinyin-data 来源核实许可、测量嵌入数据增量，并保留无拼音字符。 |
| [longbridge/opencc](https://github.com/longbridge/opencc) | [Apache-2.0](https://github.com/longbridge/opencc/blob/main/LICENSE)，纯 Go 简繁转换。[go.mod](https://github.com/longbridge/opencc/blob/main/go.mod) 含字典 trie、go-diff 等依赖，不能标为零依赖；词典体积与上游词典许可也要单独核实。作为完整词组转换候选，尚未通过本项目轻量预算。 |
| [sergi/go-diff](https://github.com/sergi/go-diff) | [MIT](https://github.com/sergi/go-diff/blob/master/LICENSE)，提供文本 diff/match/patch。只做逐行差异时可自行实现；采用此库时限制输入、运行时间并明确 Unicode/行号输出约定。 |

## 可以不增加第三方库的补充项

这些是工程实现建议，并非已经完成：

- 开发文本：正则匹配/替换（Go RE2 语法，明确不支持反向引用和环视）、Unicode/HTML 转义、字符串大小写、行过滤、文本差异、进制转换、URL 分解、JWT 载荷解码（标明未验签）、IP/CIDR 计算。可利用 `regexp`、`unicode`、`html`、`strconv`、`net/url`、`encoding/base64`、`encoding/json`、`net/netip`。
- 日常计算：时间差、年龄、倒计时目标计算、比例百分比、长度/面积/质量/温度单位转换、随机抽取、分组、基础统计。使用 `time`、`math/big`、`math`、`crypto/rand` 与明确单位表。涉及价格、汇率、政策和实时天气的工具需要另列数据源，不伪装成本地恒定算法。
- 图片小工具：坐标取色、尺寸/像素单位计算、纯色图、九宫格切图、常用证件照画布尺寸；可在现有图片 provider 上组合。自动人像分割、美颜和智能抠图需要额外算法/模型，不能把白底透明化算作等价完成。

标准库行为以 [Go 官方包文档](https://pkg.go.dev/std) 为实现依据；具体接口仍需接入时核对。

## 已有外部引擎优先复用

现有状态依据本仓库 `docs/package-management.md` 与各 provider 注册源码；外部引擎不应作为“轻量内置库”宣传，也不因本次核对新增下载。

| 引擎 | 仓库现状及可复用方向 | 官方能力依据 / 限制 |
| --- | --- | --- |
| FFmpeg / ffprobe | 已托管并接入音视频 provider；复用转码、裁剪、压缩、抽帧、音轨、滤镜能力 | [FFmpeg CLI](https://ffmpeg.org/ffmpeg.html)、[滤镜](https://ffmpeg.org/ffmpeg-filters.html)。具体编解码器依实际托管构建，不能把音轨提取当语音转文字 |
| ImageMagick | 已接入裁剪、方向校正、压缩、拼图、水印、调色、GIF、ICO、清理元数据 | [选项](https://imagemagick.org/script/command-line-options.php)。新工具优先复用已有 Runner；WebP/HEIC 等格式按实际安装构建探测 |
| pdfcpu / Poppler | 已托管并接入 PDF 结构操作、渲染/读取 provider | [pdfcpu](https://pdfcpu.io/)、[Poppler](https://poppler.freedesktop.org/)。文本提取不覆盖扫描 OCR；结构编辑不等于 PDF→Word 精确保真 |
| Pandoc | 已接入 DOCX、Markdown、HTML、EPUB 等转换与资源提取 | [手册](https://pandoc.org/MANUAL.html)。goldmark 可补内置 Markdown HTML 引擎；复杂文档转换继续复用 Pandoc |
| qsv | 已接入 CSV 选择、排序、合并、统计、校验与转换 | [官方仓库](https://github.com/dathere/qsv)。日常表格清理优先复用；CSV 工作流不等于完整 Excel 工作簿布局编辑 |
| 7-Zip / Go 压缩库 | 已有 archive provider；`go.mod` 已引用 sevenzip 和 xz 及其压缩依赖 | [7-Zip](https://www.7-zip.org/)、[sevenzip](https://github.com/bodgit/sevenzip)。根据现有格式矩阵扩展，不再引同类压缩大依赖 |

## 尚未接入、不能记成已有能力的较重选项

- LibreOffice headless：当前只出现在 `docs/roadmap.md` 的 C 批次和既有研究文档中，包管理表与 provider 目录没有它。适合 Office→PDF/旧 Office 格式升级，但属于完整办公运行时，暂不纳入轻量内置方案。[官方启动参数](https://help.libreoffice.org/latest/en-US/text/shared/guide/start_parameters.html)
- Tesseract / OCRmyPDF：当前也是路线图项。OCR 需引擎与语言包，扫描 PDF 处理还需额外组件，应单独立项测量安装成本。[Tesseract 文档](https://tesseract-ocr.github.io/tessdoc/)、[OCRmyPDF 安装](https://ocrmypdf.readthedocs.io/en/latest/installation.html)
- ExifTool：路线图项，适合更广的 EXIF/IPTC/XMP 读写；需要 Perl 或打包运行时。若近期只要只读常见 EXIF，先验证 goexif；不要为少量标签读取立即增加通用外部引擎。[官方文档](https://exiftool.org/exiftool_pod.html)

## 建议接入顺序

1. 先完成标准库开发文本、颜色基础换算、日期差和单位换算；保持 `pkg/app` 用例与 `operation.Runner` 契约，CLI 只做参数适配。
2. 逐项引入 gozxing、YAML、TOML、cron；它们可覆盖清晰的功能缺口，每项单独测量链接增量。二维码生成库与 gozxing 二选一，优先后者统一读写。
3. 根据对标条目补 goldmark、go-colorful、lunar-go；etree 与 goexif 按节点查询/元数据实际需求决定，不因候选清单而全部安装。
4. 所有候选落地前固定版本、保留许可证、限制输入资源、提供正反样本；代码完成后运行仓库要求的 `go test ./...`、`go vet ./...`、`go build ./cmd/fnsh`。

## 固定版本与 API 实查补充

2026-09-07 在独立临时模块 `finishbit-dependency-audit` 中执行 `go list -m -json ...@latest`、固定版本 `go get`、`go doc` 和包构建，未修改项目根 `go.mod`。以下版本在 Go 1.25.13 下全部通过包构建；这里的最低 Go 版本来自各固定版本 `go.mod`，不是旧 Go 环境实测。

| 固定模块版本 | 声明 Go 版本 | 直接可用 API / 实际运行期外部模块 |
| --- | --- | --- |
| `github.com/makiuchi-d/gozxing@v0.1.1` | 1.17 | `qrcode.NewQRCodeWriter().Encode(text, gozxing.BarcodeFormat_QR_CODE, width, height, hints)` 返回 `*BitMatrix`，已实现 `image.Image`，直接 `png.Encode(w, matrix)`；没有 `ToImage` 方法。读取 `gozxing.NewBinaryBitmapFromImage(img)` → `qrcode.NewQRCodeReader().Decode(bitmap, hints)` → `result.GetText()`。运行期 `x/text`、`x/xerrors`；应 `go get github.com/makiuchi-d/gozxing/qrcode@v0.1.1` 或随后 tidy，以补齐子包依赖 |
| `github.com/mozillazg/go-pinyin@v0.21.0` | 1.11 | `pinyin.NewArgs()`、`Pinyin(text,args) [][]string`、`LazyPinyin(text,args) []string`；`Style` 可选 Normal、Tone、Tone2、Tone3、FirstLetter；`Heteronym` 返回备选读音；`Fallback func(rune,Args) []string` 应保留非汉字，默认会丢弃。运行期只用标准库 |
| `github.com/longbridgeapp/opencc@v0.3.13` | 1.16 | `opencc.New("s2t")`、`converter.Convert(text)`；嵌入词典/配置，无需外置文件。实际运行链 `liuzl/da` → `liuzl/cedar-go`。**仅 API 验证，不建议直接接入，见下面许可发现** |
| `go.yaml.in/yaml/v3@v3.0.5` | 1.16 | `yaml.NewDecoder(reader).Decode(&node)`、`yaml.Node`、`yaml.NewEncoder(writer)`、`SetIndent(2)`、`Encode(node)`；运行期只用标准库。必须再次 Decode 检查 EOF 才能拒绝多文档 |
| `github.com/yuin/goldmark@v1.8.6` | 1.22 | `goldmark.New(goldmark.WithExtensions(extension.GFM))`、`Convert([]byte,io.Writer)`；运行期只用标准库。默认不渲染原始 HTML，输出完整页面需自己用模板包装 |
| `github.com/sergi/go-diff@v1.4.0` | 1.13 | `diffmatchpatch.New()`、`DiffMain(a,b,true)`、`DiffCleanupSemantic`；行级用 `DiffLinesToRunes` → `DiffMainRunes(...,false)` → `DiffCharsToLines`。`DiffTimeout` 限制算法时间；另限输入长度。运行期只用标准库；go.mod 中 testify 等为测试相关，不能算成产品运行时 |

YAML 的 `go.yaml.in/yaml/v4@latest` 当前解析到 `v4.0.0-rc.6`，仍为预发布。前文建议评估 v4；本次需要稳定固定版时选择官方维护的 v3.0.5 安全修复线。后续再单独升级到 v4 正式版。[官方状态说明](https://github.com/yaml/go-yaml#project-status)、[v3.0.5 源码](https://github.com/yaml/go-yaml/tree/v3.0.5)

### 许可证与嵌入数据的实查发现

- OpenCC GitHub 仓库已改名为 `longbridge/opencc`，但 v0.3.13 `go.mod` 声明的模块仍是 `github.com/longbridgeapp/opencc`；使用新组织模块路径会失败。它自身为 Apache-2.0，词典来自 Apache-2.0 的上游 OpenCC；但运行依赖 `github.com/liuzl/cedar-go@v0.0.0-20170805034717-80a9c64b256d` 的 `LICENSE.md` 是 **GPL v2**，因此不能将整条链写为 Apache-2.0。建议本次使用上游 OpenCC 固定版本词典、自行实现 trie/最长词匹配，避免直接链接该 Go 包。尚未核对它嵌入词典具体对应的上游 commit，不把上游版本误标为已固定。[固定版 go.mod](https://github.com/longbridge/opencc/blob/v0.3.13/go.mod)、[嵌入实现](https://github.com/longbridge/opencc/blob/v0.3.13/opencc.go)、[实际 cedar 许可](https://github.com/liuzl/cedar-go/blob/80a9c64b256db37ac20aff007907c649afb714f1/LICENSE.md)、[上游 OpenCC 许可](https://github.com/BYVoid/OpenCC/blob/ver.1.1.7/LICENSE)
- go-pinyin v0.21.0 的代码为 MIT；CHANGELOG 指明嵌入 pinyin-data v0.15.0，该固定数据版也为 MIT。保留代码作者和数据来源声明，不能把多音字列表宣传成基于语境自动判读。[固定版 CHANGELOG](https://github.com/mozillazg/go-pinyin/blob/v0.21.0/CHANGELOG.md)、[数据许可](https://github.com/mozillazg/pinyin-data/blob/v0.15.0/LICENSE)
- go-diff 需要保留自身 MIT 与原 Google diff-match-patch 的 Apache-2.0；gozxing 同样需要自身 MIT 与 ZXing Apache-2.0。[go-diff 说明](https://github.com/sergi/go-diff/blob/v1.4.0/README.md#copyright-and-license)、[gozxing 许可](https://github.com/makiuchi-d/gozxing/blob/v0.1.1/LICENSE)
- YAML v3.0.5 的 LICENSE 为 Apache-2.0，部分扫描/发射器文件头为 MIT，并有 NOTICE；应一起保留。goldmark v1.8.6 为 MIT。[YAML 固定版](https://github.com/yaml/go-yaml/tree/v3.0.5)、[goldmark 固定版许可](https://github.com/yuin/goldmark/blob/v1.8.6/LICENSE)

固定下载命令（不包含因许可链问题暂缓的 OpenCC）：

```powershell
go get github.com/makiuchi-d/gozxing/qrcode@v0.1.1 github.com/mozillazg/go-pinyin@v0.21.0 go.yaml.in/yaml/v3@v3.0.5 github.com/yuin/goldmark@v1.8.6 github.com/sergi/go-diff@v1.4.0
```

## Windows OCR 包体实测与较重引擎边界

用户后续明确：新增依赖不超过 1 GB 可按需安装，超过再讨论。因此本轮仅在临时目录下载、解包并运行 OCR CLI 探测，没有执行系统安装器或修改注册表。以下是 2026-09-07 文件实际长度，MiB 按 1,048,576 字节计算。

### Tesseract：已有低于上限的实际可运行方案

| 构件 | 下载/文件字节数 | 固定来源与 SHA-256 |
| --- | ---: | --- |
| Tesseract 5.5.0 Windows x64 NSIS 包 | 21,381,872 | [官方 release 文件](https://github.com/tesseract-ocr/tesseract/releases/download/5.5.0/tesseract-ocr-w64-setup-5.5.0.20241111.exe)；`f3fc4236425b690c8be756f35793f77394ee004be0a6460a440c754d892f68bc` |
| tessdata_fast 简体中文 | 2,469,156 | [chi_sim](https://raw.githubusercontent.com/tesseract-ocr/tessdata_fast/87416418657359cb625c412a48b6e1d6d41c29bd/chi_sim.traineddata)；`a5fcb6f0db1e1d6d8522f39db4e848f05984669172e584e8d76b6b3141e1f730` |
| tessdata_fast 英文 | 4,113,088 | [eng](https://raw.githubusercontent.com/tesseract-ocr/tessdata_fast/87416418657359cb625c412a48b6e1d6d41c29bd/eng.traineddata)；`7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2` |

使用完整 7-Zip 的 NSIS 解包器解包后，**全部 140 个文件合计 75,658,744 字节**；加入上述中英模型后，**总计 82,240,988 字节（78.43 MiB）**。这个总数保留所有 DLL、训练工具、文档和安装器辅助文件，尚未裁剪，所以无需为满足上限冒险删 DLL。压缩包与两份模型的下载量合计 27,964,116 字节。SHA 为下载后实算；GitHub release API 此老资产没有发布 digest 字段，不把本地实算值称为厂商签名。

已通过直接运行 `tesseract.exe --version` 和 `--tessdata-dir ... --list-langs` 验证本机可运行，输出 `v5.5.0.20241111`、Leptonica 1.85.0，识别到 `chi_sim` 与 `eng`。本小节没有做识别准确率验收；provider 接入必须补真实中文、英文图像及异常文件测试。

包管理注意：普通 7za 不支持 NSIS，必须使用包含 NSIS 支持的完整 7z.exe + 7z.dll，或在受控构建流程生成固定散列的 ZIP。不能直接将 EXE 填入只支持 ZIP/TAR 的现有下载器。本次解包没有运行安装器，原包没有附带中英 traineddata，所以模型需要单独作为受校验构件管理。

最低识别调用建议 `tesseract input.png stdout --tessdata-dir <managed>/tessdata -l chi_sim+eng --oem 1 --psm 3`，TXT、TSV、hOCR 等输出按操作用途选择。`tessdata_fast` 仅支持 LSTM 引擎（OEM 1）；方向/文字系统自动检测还要单独的 `osd.traineddata`，官方仓库当前文件 10,562,727 字节（仅远端元数据核对，本轮未下载）；固定版也有繁体和竖排模型，按需求加，避免整仓库模型下载。[官方数据说明](https://tesseract-ocr.github.io/tessdoc/Data-Files.html)、[tessdata_fast Apache-2.0 许可](https://github.com/tesseract-ocr/tessdata_fast/blob/87416418657359cb625c412a48b6e1d6d41c29bd/LICENSE)

引擎自身 Apache-2.0，但上述 Windows 包还带 Leptonica、图像编解码、curl、ICU 等 DLL，最终托管包许可清单须覆盖这些组件，不能只复制 Tesseract 一份许可。Linux/macOS 宜采用发行版包或分别构建的受控归档；本轮仅验证 Windows x64 包，不宣称已经有跨平台可管理构件。

### LibreOffice：下载量可确认，安装量不能用下载量替代

[官方 26.2.6 Windows x64 MSI mirrorlist](https://download.documentfoundation.org/libreoffice/stable/26.2.6/win/x86_64/LibreOffice_26.2.6_Win_x86-64.msi.mirrorlist) 给出的文件长度是 **373,252,096 字节**，SHA-256 是 `f9877032fd908beb9c0ddf06df4af5c2e85f419c42e14876c4cce5aae5fb2660`。这只是安装包下载量。[官方系统要求](https://www.libreoffice.org/system-requirements/) 给出 Windows 最高需 1.5 GB 磁盘、macOS 800 MB、Linux 1.55 GB；这些是供应商要求，并非本机实测。未完成实际展开体积/精简组件验证前，不能承诺 Windows 安装在 1 GB 内。

Windows 官方主要分发 MSI，需要单独解决受控文件展开、组件目录映射、字体和独立 UserInstallation 配置；不要直接运行全局安装器。macOS 有 Intel/Apple silicon 平台，Linux 官方建议发行版包；本轮没有下载或安装其他平台构件。Office→PDF 通常不需要为 Base 数据库功能额外安装 Java。[启动参数](https://help.libreoffice.org/latest/en-US/text/shared/guide/start_parameters.html)

后续已在临时目录下载并实算验证上述 MSI SHA-256，然后通过 Windows Installer COM 的 `OpenDatabase(path, 0)` **只读查询** `File` 表，得到 **19,496 条文件记录、FileSize 合计 1,580,201,154 字节**。这是全量 MSI 文件载荷的声明长度，已经超过 1 GB，尚不含配置、字体替代和安装缓存；不是精简组件安装实测。没有执行 MSI 安装或展开全部文件。按用户上限，暂不安装全量 LibreOffice；若继续，可另行讨论精简功能组件验证或提高上限。

### AI 去背景：可行候选，尚未确认完整安装量

[rembg](https://github.com/danielgatis/rembg) 为 MIT，支持 CPU ONNX Runtime 与 `u2netp` 模型，需 Python 运行时及依赖。模型选择可参考 [u2netp session 实现](https://github.com/danielgatis/rembg/blob/main/rembg/sessions/u2netp.py) 和 [U-2-Net 上游](https://github.com/xuebinqin/U-2-Net)。这只是候选：Python、ONNX Runtime、图像科学计算依赖、模型应合计测量，模型许可与代码许可分别核对，不能用小模型文件推导总安装小于 1 GB。未下载未知大小的模型或 GPU/CUDA 包。

### Tesseract 安装包解包链：现有 Go 库可以引导完整 7-Zip

以下过程已在独立临时目录验证，未执行任何安装器：

1. `bodgit/sevenzip@v1.6.5` 直接 `OpenReader("tesseract-5.5.0.exe")` 返回 `sevenzip: not a valid 7-zip file`。该库读取 7z，不是 NSIS 通用解析器。7za 的格式表也没有 NSIS，直接解包失败；7zr 仅是精简 7z 工具，不应指望读取 NSIS。[sevenzip 官方仓库](https://github.com/bodgit/sevenzip)、[7-Zip 下载分类](https://www.7-zip.org/download.html)
2. [官方 7-Zip 26.03 x64 EXE](https://github.com/ip7z/7zip/releases/download/26.03/7z2603-x64.exe) 是 **7z 自解压容器**，实测 `sevenzip.OpenReader` 可读取，也能用 7za 解开。下载长度 **1,661,239 字节**，实算 SHA-256：`0859c524b8a63551848f0c246abddcb1d0b7b656b0fbfe879f8d85e61a9e6edd`。
3. 可使用已有 Go sevenzip 从该容器仅提取 `7z.exe`（577,536 字节）、`7z.dll`（1,906,688 字节）、`License.txt`（6,031 字节），并保持前两者同目录。这个引导过程无需系统 7-Zip、无需执行 `7z2603-x64.exe`。
4. 运行提取的完整 `7z.exe x <tesseract NSIS> -o<staging> -y` 即可。26.03 完整版重复解包后的文件实际总量仍为 75,658,744 字节，加中英模型仍为 82,240,988 字节，直接运行 CLI 成功。产品实现应先列清单、验证解包路径/数量/总量、设置超时并分阶段发布。

7-Zip 完整版 `License.txt` 对 `7z.dll` 标注主要 GNU LGPL、部分代码含 unRAR 限制；不能把整个完整引擎错误标为仅 LGPL 或 MIT。托管解包器仅用于 NSIS 不会自动移除其他代码的分发许可义务；随包保留完整原许可。NSIS 特定资产尚未发现同版本官方 ZIP，以上自解压容器引导方案已经可用。

OCR 托管目录最低布局为根目录 `tesseract.exe` 与其同包 DLL、`doc/LICENSE`、`doc/AUTHORS`，子目录 `tessdata/chi_sim.traineddata`、`tessdata/eng.traineddata` 和模型 LICENSE。本次保留了整包训练工具和辅助文件；不建议在未做 DLL 依赖审计前缩减。可用测试目录为 `C:/Users/wps/AppData/Local/Temp/finishbit-dependency-audit-aa014fd367d443678df41b81ec24b2bb/tesseract-2603`，官方完整解包器在同级 `7zip-full`。

## 避免 GPL 转换链的 OpenCC 最小词典准备

已从 OpenCC 上游固定版本 **ver.1.1.7** 下载以下原始文件；`git ls-remote` 核对标签指向 commit **e5d6c5f1b78e28a5797e7ad3ede3513314e544b7**。文件位于临时目录 `C:/Users/wps/AppData/Local/Temp/finishbit-dependency-audit-aa014fd367d443678df41b81ec24b2bb/opencc-data`，没有写入产品目录。七个文件合计 **1,088,411 字节**，四份词典为 **1,078,434 字节**。

| 文件 | SHA-256 |
| --- | --- |
| [LICENSE](https://raw.githubusercontent.com/BYVoid/OpenCC/e5d6c5f1b78e28a5797e7ad3ede3513314e544b7/LICENSE) | `b534e465949558eec2597b04f5092b5e161236a68dfbfd04d547592ac3964308` |
| [STCharacters.txt](https://raw.githubusercontent.com/BYVoid/OpenCC/e5d6c5f1b78e28a5797e7ad3ede3513314e544b7/data/dictionary/STCharacters.txt) | `9207708da9f2e2a248f39c457b2fccad26ec42e7efaf47a860e6900464f4cac5` |
| [STPhrases.txt](https://raw.githubusercontent.com/BYVoid/OpenCC/e5d6c5f1b78e28a5797e7ad3ede3513314e544b7/data/dictionary/STPhrases.txt) | `1411418f98dd7666a4ee673619654ed1e0518ec97953315cc10656c30c7015bb` |
| [TSCharacters.txt](https://raw.githubusercontent.com/BYVoid/OpenCC/e5d6c5f1b78e28a5797e7ad3ede3513314e544b7/data/dictionary/TSCharacters.txt) | `6b5a0a799bea2bb22c001f635eaa3fc2904310f0c08addbff275477a80ecf09a` |
| [TSPhrases.txt](https://raw.githubusercontent.com/BYVoid/OpenCC/e5d6c5f1b78e28a5797e7ad3ede3513314e544b7/data/dictionary/TSPhrases.txt) | `b2ef895dd4953b4bb77fc8ef8d26a2a9ca6d43a760ed9a1d767672cfafa6324f` |
| [s2t.json](https://raw.githubusercontent.com/BYVoid/OpenCC/e5d6c5f1b78e28a5797e7ad3ede3513314e544b7/data/config/s2t.json) | `710bb970e406d9c0a3dd33c609248686cf1c5578d064b7b7fc89e70fbb9ea75d` |
| [t2s.json](https://raw.githubusercontent.com/BYVoid/OpenCC/e5d6c5f1b78e28a5797e7ad3ede3513314e544b7/data/config/t2s.json) | `b818534194f27c2d95f01001edb0a5ec49b9050119892cb30a0504bb202cc07c` |

词典来自 Apache-2.0 上游本身，不复制 longbridge Go 实现或 GPL cedar 代码。原始文本按行 `源词<TAB>候选1 候选2 ...`，默认转换选择第一个候选。配置表明 s2t 需要 STPhrases 与 STCharacters、t2s 需要 TSPhrases 与 TSCharacters，词组优先于单字。自写算法可实现受限的最长词匹配，但不能把它宣传为已复刻 OpenCC 的完整 `mmseg` 分词；需要用词组、歧义字、非汉字、空白保留样本验证最小能力。暂不包含台湾/香港地区词汇专用词典。

## 无 CGO 的 AI 去背景最终候选核对

结论：本轮没有确认一个同时满足 **Windows 可直接托管、模型许可来源完整、全量下载/展开小于 1 GB、固定所有校验值、运行不隐式联网** 的现成方案。外部 Rust/C++ CLI 与主程序 `CGO_ENABLED=0` 可以共存；障碍是候选运行时的分发和版本控制，不是必须把 ONNX 链接到 Go。

| 候选 | 已确认 | 未通过的接入条件 |
| --- | --- | --- |
| [i-rocky/rembg v0.1.2](https://github.com/i-rocky/rembg/releases/tag/v0.1.2) | MIT Rust CLI；有 Windows x64 ZIP，并有 Linux/macOS CLI 归档。发布页 Windows ZIP 显示 3.59 MB，SHA-256 为 `6f6d10e066a163fd28bc6266017d676e23e3491d97c62f0b4fa45f72cae9dd2b`。动态载入 ONNX Runtime，可 CPU 推理，不执行 Python | [CLI 参数](https://github.com/i-rocky/rembg/blob/v0.1.2/rembg-rs/src/cli.rs) 没有本地模型、运行时、缓存目录或 offline 参数；[runtime.rs](https://github.com/i-rocky/rembg/blob/v0.1.2/rembg-rs/src/runtime.rs) 缺缓存时查询 PyPI 最新版本，存在缓存则挑可用的较高版本；[model.rs](https://github.com/i-rocky/rembg/blob/v0.1.2/rembg-rs/src/model.rs) 自动下载模型而且未配置 SHA/MD5。不能把固定 CLI ZIP 当作固定整套运行时。未完成完整下载/解包和推理测量 |
| [wyh2001/outline v0.1.2](https://github.com/wyh2001/outline/releases/tag/v0.1.2) | MIT；有带校验的 macOS ARM64、Linux x64/ARM64 发布归档；支持 ONNX Runtime，另有实验 RTen 路线 | 发布表没有 Windows 构件；官方 README 说明项目仍早期，RTen 模型兼容较窄。需要自建 Windows 发布与验证，而不是直接安装可复用包 |
| [WarRaft/rembg-rs](https://github.com/WarRaft/rembg-rs) | MIT；Rust ONNX Runtime + U2-Net 库/CLI | [发布页](https://github.com/WarRaft/rembg-rs/releases) 没有预编译 release，仍需独立构建交付、运行时和模型验证 |

模型来源要单列：[U-2-Net](https://github.com/xuebinqin/U-2-Net) 上游项目 LICENSE 是 Apache-2.0，i-rocky 的 `u2netp` 指向 danielgatis/rembg 的 `v0.0.0/u2netp.onnx` 转换模型资产。不能把包装器 MIT 自动视为所有模型权重的许可证；应记录固定权重 SHA、上游权重来源及转换依据。当前研究未下载未知尺寸的模型、未安装 Python/CUDA，也未执行去背景推理。

最小后续路线是维护独立 Rust 辅助 CLI：仅接收本地模型和固定 CPU ONNX Runtime DLL，默认禁用下载，输出 PNG 掩码/透明图，并将三者作为受校验包交付。这样保留 Go 的 `CGO_ENABLED=0`，但仍是待实现工作；本轮不把传统指定颜色透明化冒充 AI 抠图。

### 不 fork、预置缓存方案的精确核对

单靠设置子进程环境变量，在 Windows **不能**将上述 i-rocky CLI 缓存定向到 FinishBit 私有目录。已下载 crates.io 的 `directories 6.0.0` 与 `dirs-sys 0.5.0` 原源码核对：`ProjectDirs::from("rs", "rembg", "rembg-rs")` 经 Windows `project_dirs_from_path` 调用 `known_folder_local_app_data()`，后者使用 `SHGetKnownFolderPath(FOLDERID_LocalAppData)`；没有读取 `LOCALAPPDATA` 或 `XDG_CACHE_HOME` 的覆盖逻辑。[directories 固定版源码](https://docs.rs/crate/directories/6.0.0/source/src/win.rs)、[dirs-sys 固定版源码](https://docs.rs/crate/dirs-sys/0.5.0/source/src/lib.rs)

本机实际路径规则为：

```text
C:/Users/wps/AppData/Local/rembg/rembg-rs/cache/
  models/u2netp.onnx
  onnxruntime/onnxruntime/<version>/lib/onnxruntime.dll
```

`--device cpu` 可选择 CPU 包而不询问 GPU。缓存内有 DLL 时，`ensure_onnxruntime` 会在访问 PyPI 前直接返回；模型文件存在时也不会下载。可是目录属于全局 rembg 缓存，CLI 没有缓存根参数，发现多个版本会挑更高版本。缺少 DLL 时，源码先 `pypi::fetch_project`，之后才询问是否下载，因此“不加 -y、关闭 stdin”也不是离线保证；缺少模型时 `ensure_model` 直接允许下载且没有摘要校验。调用前检查散列只能确认检查时刻的文件状态，不能改变它后续缺失时联网和访问全局目录的行为。[固定版 runtime.rs](https://github.com/i-rocky/rembg/blob/v0.1.2/rembg-rs/src/runtime.rs)、[固定版 model.rs](https://github.com/i-rocky/rembg/blob/v0.1.2/rembg-rs/src/model.rs)

因此不采用修改全局 Known Folder、注入 API、建立全局缓存重定向等额外绕路。若坚持现成 CLI，本轮条件未满足，保持待实现；没有接入产品。

## Windows 桌面录制设备能力核对

2026-09-07 对本项目已有运行时 `D:/finish-bit/.finishbit-test-batch-a/packages/ffmpeg/6.1.1/ffmpeg.exe` 仅执行 `-version`、`-devices`、`-h demuxer=gdigrab`，**没有开始录屏或读取桌面帧**。版本输出是 `6.1.1-essentials_build-www.gyan.dev`，设备表包含 `D gdigrab`、`D dshow`，gdigrab 参数帮助确认 `draw_mouse`、`show_region`、`framerate`、`video_size`、`offset_x`、`offset_y`。

因此 `screen.record` 最小可实现为 Windows 有界时长的桌面/矩形区域录制，共用 FFmpeg provider；不增加依赖。建议明确 duration 必填且有上限、fps 范围、区域坐标和尺寸、鼠标开关、输出覆盖规则。示意参数（本轮未执行）：

```text
ffmpeg -nostdin -hide_banner -f gdigrab -framerate 15 -draw_mouse 1 -show_region 1 -video_size 1280x720 -offset_x 0 -offset_y 0 -i desktop -t 10 -an -c:v libx264 -pix_fmt yuv420p <staged-output.mp4>
```

设备只支持列出不证明实际捕获成功：多显示器负坐标、DPI 缩放、锁屏/远程会话、区域越界、编码器可用性仍需用户授权后的独立录制样本验收。输出尺寸应为偶数以适配 yuv420p。录制先落临时文件，成功后原子发布；超时/取消需正确终止进程并处理未完成 MP4。最小版不录麦克风或系统音频：DirectShow 的存在不意味着已经知道或验证可用音源。macOS、Linux 对应不同捕获设备与桌面权限，不从 Windows gdigrab 推导跨平台支持。[FFmpeg gdigrab 官方文档](https://ffmpeg.org/ffmpeg-devices.html#gdigrab)

## Python embeddable + 自有 ONNX 推理脚本的限定评估

这条架构在技术上可行：私有 Python、固定 CPU ONNX Runtime、NumPy、Pillow 和绝对路径模型，由 FinishBit 包管理完成下载校验，推理脚本没有网络逻辑。通过 `python312._pth` 只列嵌入 stdlib ZIP、包目录和脚本目录，可避开系统 Python/pip/用户 site；显式 `providers=["CPUExecutionProvider"]` 避免 CUDA。官方说明嵌入包面向应用集成，第三方包应作为应用随附构件，不把常规 pip 管理视为该分发的默认方案。[Python Windows 嵌入文档](https://docs.python.org/3.12/using/windows.html#the-embeddable-package)

但本次 **未达到可发布结论**：

- 已核对 CPython 3.12.10 x64 embeddable 与 ORT 1.22.1、NumPy 2.2.6、Pillow 11.3.0 的 `cp312-win_amd64` wheel 匹配，以及完整声明依赖。3.12.10 是最后官方 3.12 完整维护二进制，现已被安全修复版取代，因此仅作为兼容性预研组合，生产应重新选仍发布安全更新二进制的 Python 版本。[Python 3.12.10 官方说明](https://www.python.org/downloads/release/python-31210/)
- 这 11 个 wheel 的 PyPI 元数据长度合计 **39,875,205 字节**，Python ZIP HEAD 为 **11,133,606 字节**，合计 **51,008,811 字节（不含模型）**。这是下载长度，不是安装体积上界。二进制下载出现连接中断与超时，本轮停止重试；没有完成所有 ZIP 中央目录/展开长度测量，没有成功 import 或推理验收，因此不能声称“已验证总安装小于 1 GB”。
- 最终未确认 **u2netp ONNX 权重本身的明确可再分发许可链**。作者的 U-2-Net 仓库为 Apache-2.0，README 提供外部 `u2netp.pth` 权重；rembg 的模型 release 托管转换后的 `u2netp.onnx`，但是没有找到对该具体资产明确的独立许可/转换溯源声明。rembg README 又明确说权重许可独立于 MIT。不能仅把包装器或代码仓库 LICENSE 当作已核实的权重授权。[U-2-Net 权重入口](https://github.com/xuebinqin/U-2-Net#usage-for-salient-object-detection)、[rembg 模型发布](https://github.com/danielgatis/rembg/releases/tag/v0.0.0)、[权重许可提示](https://github.com/danielgatis/rembg#models)

所以本轮不接入 AI 权重、不下载 CUDA、不运行 rembg 自动下载逻辑，也不再扩散寻找更多候选。若后续提供明确允许再分发的权重与散列，可继续这条架构，先用 ZIP 表实际核算总展开长度，再做隔离 import 和自制图像推理。

完整声明依赖：ORT → coloredlogs、flatbuffers、numpy、packaging、protobuf、sympy；coloredlogs → humanfriendly；sympy → mpmath；humanfriendly 在 Windows/Python≥3.8 → pyreadline3。Pillow 的文档/测试 extras 不需要；NumPy 无额外 Python 运行依赖。原生 wheel 自带 DLL 及其第三方许可仍应保留，未检查完 wheel 内全部 DLL 许可前不称完整许可审计。[ORT 固定版 PyPI 元数据](https://pypi.org/pypi/onnxruntime/1.22.1/json)、[humanfriendly 条件依赖](https://pypi.org/pypi/humanfriendly/10.0/json)

CPython 3.12.10 ZIP 的固定来源为 [python-3.12.10-embed-amd64.zip](https://www.python.org/ftp/python/3.12.10/python-3.12.10-embed-amd64.zip)，SHA-256 `4acbed6dd1c744b0376e3b1cf57ce906f9dc9e95e68824584c8099a63025a3c3`，来自官方 [Sigstore bundle](https://www.python.org/ftp/python/3.12.10/python-3.12.10-embed-amd64.zip.sigstore) 的 messageDigest；本轮未完成 ZIP 本地摘要复核。

以下十行只说明推理接口，不是已运行/验收的产品脚本：模型必须由可信固定文件提供；黑图须避免除零，正式实现还应限制像素、时长、线程与输出覆盖。

```python
import sys, numpy as np, onnxruntime as ort
from PIL import Image
im = Image.open(sys.argv[1]).convert("RGB")
x = np.asarray(im.resize((320, 320)), dtype=np.float32)
x = x / max(float(x.max()), 1.0)
x = ((x - np.array([.485,.456,.406],np.float32)) / np.array([.229,.224,.225],np.float32)).transpose(2,0,1)[None]
s = ort.InferenceSession(sys.argv[2], providers=["CPUExecutionProvider"])
m = s.run(None, {s.get_inputs()[0].name:x})[0][0,0]
m = ((m-m.min()) / max(float(m.max()-m.min()), 1e-8) * 255).astype("uint8")
im.putalpha(Image.fromarray(m).resize(im.size, Image.Resampling.LANCZOS)); im.save(sys.argv[3], "PNG")
```

固定 wheel 清单（SHA-256 与长度来自对应 PyPI JSON，尚未全部下载复核）：

| 文件 | 字节 | SHA-256 |
| --- | ---: | --- |
| [onnxruntime-1.22.1-cp312-cp312-win_amd64.whl](https://files.pythonhosted.org/packages/5d/54/7139d463bb0a312890c9a5db87d7815d4a8cce9e6f5f28d04f0b55fcb160/onnxruntime-1.22.1-cp312-cp312-win_amd64.whl) | 12690910 | `6a64291d57ea966a245f749eb970f4fa05a64d26672e05a83fdb5db6b7d62f87` |
| [numpy-2.2.6-cp312-cp312-win_amd64.whl](https://files.pythonhosted.org/packages/36/fa/8c9210162ca1b88529ab76b41ba02d433fd54fecaf6feb70ef9f124683f1/numpy-2.2.6-cp312-cp312-win_amd64.whl) | 12614190 | `c1f9540be57940698ed329904db803cf7a402f3fc200bfe599334c9bd84a40b2` |
| [pillow-11.3.0-cp312-cp312-win_amd64.whl](https://files.pythonhosted.org/packages/8c/ce/e7dfc873bdd9828f3b6e5c2bbb74e47a98ec23cc5c74fc4e54462f0d9204/pillow-11.3.0-cp312-cp312-win_amd64.whl) | 6986324 | `a6444696fce635783440b7f7a9fc24b3ad10a9ea3f0ab66c5905be1c19ccf17d` |
| [coloredlogs-15.0.1-py2.py3-none-any.whl](https://files.pythonhosted.org/packages/a7/06/3d6badcf13db419e25b07041d9c7b4a2c331d3f4e7134445ec5df57714cd/coloredlogs-15.0.1-py2.py3-none-any.whl) | 46018 | `612ee75c546f53e92e70049c9dbfcc18c935a2b9a53b66085ce9ef6a6e5c0934` |
| [flatbuffers-25.2.10-py2.py3-none-any.whl](https://files.pythonhosted.org/packages/b8/25/155f9f080d5e4bc0082edfda032ea2bc2b8fab3f4d25d46c1e9dd22a1a89/flatbuffers-25.2.10-py2.py3-none-any.whl) | 30953 | `ebba5f4d5ea615af3f7fd70fc310636fbb2bbd1f566ac0a23d98dd412de50051` |
| [packaging-25.0-py3-none-any.whl](https://files.pythonhosted.org/packages/20/12/38679034af332785aac8774540895e234f4d07f7545804097de4b666afd8/packaging-25.0-py3-none-any.whl) | 66469 | `29572ef2b1f17581046b3a2227d5c611fb25ec70ca1ba8554b24b0e69331a484` |
| [sympy-1.14.0-py3-none-any.whl](https://files.pythonhosted.org/packages/a2/09/77d55d46fd61b4a135c444fc97158ef34a095e5681d0a6c10b75bf356191/sympy-1.14.0-py3-none-any.whl) | 6299353 | `e091cc3e99d2141a0ba2847328f5479b05d94a6635cb96148ccb3f34671bd8f5` |
| [humanfriendly-10.0-py2.py3-none-any.whl](https://files.pythonhosted.org/packages/f0/0f/310fb31e39e2d734ccaa2c0fb981ee41f7bd5056ce9bc29b2248bd569169/humanfriendly-10.0-py2.py3-none-any.whl) | 86794 | `1697e1a8a8f550fd43c2865cd84542fc175a61dcb779b6fee18cf6b6ccba1477` |
| [mpmath-1.3.0-py3-none-any.whl](https://files.pythonhosted.org/packages/43/e3/7d92a15f894aa0c9c4b49b8ee9ac9850d6e63b03c9c32c0367a13ae62209/mpmath-1.3.0-py3-none-any.whl) | 536198 | `a0b2b9fe80bbcd81a6647ff13108738cfb482d481d826cc0e02f5b35e5c88d2c` |
| [protobuf-5.29.5-cp310-abi3-win_amd64.whl](https://files.pythonhosted.org/packages/81/7f/73cefb093e1a2a7c3ffd839e6f9fcafb7a427d300c7f8aef9c64405d8ac6/protobuf-5.29.5-cp310-abi3-win_amd64.whl) | 434818 | `3f76e3a3675b4a4d867b52e4a5f5b78a2ef9565549d4037e06cf7b0942b1d3fc` |
| [pyreadline3-3.5.4-py3-none-any.whl](https://files.pythonhosted.org/packages/5a/dc/491b7661614ab97483abf2056be1deee4dc2490ecbf7bff9ab5cdbac86e1/pyreadline3-3.5.4-py3-none-any.whl) | 83178 | `eaf8e6cc3c49bcccf145fc6067ba8643d1df34d604a1ec0eccbf7a18e6d3fae6` |
