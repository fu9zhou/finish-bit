# 低热度 Go 依赖自写替代评估

核查日期：2026-09-08。本报告包含替代成本评估及同轮完成的农历替代结果。除农历外，其余依赖与运行时版本未修改。星数沿用同日 [Go 依赖健康报告](go-dependency-health-2026-09-08.md)，不是重新测量的下载量或安全评分。

## 结论

除农历的有界日期转换外，没有发现既能保留当前功能、又能直接低成本独立重写的 Go 依赖。拼音转换的算法小，但真正价值是约 789 KB 汉字读音数据；搬进仓库并不消除第三方数据维护责任。plumbing、windows、readerutil 本身确实小，却由 sevenzip 内部导入，不能只改我们的调用就移除。

最有减依赖价值的后续方向是调整安装包格式或解压引导，整体移除 sevenzip 链；这属于安装架构改造，不能视为几个工具函数的替换。二维码识别、压缩编解码、TOML/YAML/Markdown 解析均建议继续使用成熟实现。

## 已完成：移除 lunar-go

约 337 星的 lunar-go 原本提供农历、节气、节日、八字等大量能力，但项目只调用 `calendar.lunar` / `calendar.solar`，输入年限明确为 1900–2100。这个范围适合使用固定历表，不必自行重写天文计算。

- 自写 [日期转换](../../internal/localtools/lunar.go)，使用 202 个年份的月长和闰月记录（808 字节原始表），显式检查无效日期。表中额外的农历 1899 年用来覆盖公历 1900 年 1 月；农历 2100 年末转成公历 2101 年 1 月的既有行为保留。
- 数据通过原先固定的 lunar-go v1.4.6 计算产生，保留数据来源说明与 MIT 归因；没有复制其天文实现。这里移除的是编译依赖，不把历表数据声称为独立天文研究。[数据来源与编码](../../internal/localtools/data/lunar/README.md)
- 完整核对 73,414 个公历日期、73,412 个有效农历日期以及中文日期文本，与旧实现的预先生成 SHA-256 输出摘要一致；额外核对 HKO 的春节、闰日和 2033 年闰十一月实例，并测试无效日期与范围边界。持久测试不再导入原库。
- `go.mod` / `go.sum` 已移除 lunar-go；直接模块 9 → 8，间接仍为 11，合计 20 → 19，没有新增模块或本地替代 module。
- 同一 Windows amd64、Go 1.26.4，`go build -trimpath -ldflags='-s -w'` 实测：16,112,128 → 15,933,952 字节，减少 178,176 字节（174 KiB，约 1.1%）。此处是链接后二进制测量，不用源码体积代替收益。
- `go test ./...`、`go vet ./...`、`go build ./cmd/fnsh` 全部通过。本轮没有实际升级运行时，也没有将普通 Go 检查说成新版运行时兼容验收。

## 度量方法及限制

在当前 Windows 环境执行 `go list -deps -f '{{if .Module}}{{.Module.Path}}|{{.Dir}}|{{join .GoFiles ","}}{{end}}' ./cmd/fnsh`，按返回的生产 GoFiles 统计文件数、物理行数、字节数。包含注释、空行和生成表，不包括测试、未导入包、汇编或其他平台文件；Go 链接器还会移除未使用函数，因此这些数既不是最终二进制大小，也不是自写实现所需行数。

原始源码在本机 `C:/Users/wps/go/pkg/mod/`，版本来自核查时 go.mod。直接调用证据：[archive.go](../../pkg/packagemanager/archive.go)、[formats.go](../../internal/localtools/formats.go)、[images.go](../../internal/localtools/images.go)、[calendar.go](../../internal/localtools/calendar.go)。

| 模块 | 热度约数 | CLI 导入的生产 Go 文件 / 行 / 字节 | 当前使用面与替代判断 |
|---|---:|---:|---|
| sevenzip v1.6.5 | 250 | 26 / 4,214 / 86,221 | 7z 安装包及 7-Zip 自解压安装包读取；格式层加多算法依赖，自写高复杂度 |
| xz v0.5.16 | 563 | 39 / 7,098 / 170,745 | pdfcpu 非 Windows tar.xz；sevenzip 的 LZMA/LZMA2 也需要它；高复杂度 |
| brotli v1.2.2 | 733 | 81 / 261,239 / 2,545,656 | sevenzip 内部算法；大量生成字典表，不能按行数估工期；高复杂度 |
| lz4/v4 v4.1.27 | 970 | 17 / 2,857 / 72,011 | sevenzip 内部算法；块解码虽简短，帧、校验、截断、兼容边界仍非小工具 |
| ppmd v0.1.1 | 0 | 14 / 1,672 / 36,257 | sevenzip 内部概率模型解码；代码较小但算法复杂，不建议自写 |
| gozxing v0.1.1 | 668 | 65 / 8,806 / 249,406 | QR 生成、照片/图片识别、生成后回读校验；高复杂度 |
| go-pinyin v0.21.0 | 1.8k | 4 / 42,282 / 797,429 | 单字多读音、4 种风格；算法低至中复杂度，字典维护是主要成本 |
| go-toml/v2 v2.4.3 | 2.0k | 20 / 8,008 / 228,497 | 解析、校验、重编码、JSON 转换；完整输入格式解析，中高复杂度 |
| yaml/v3 v3.0.5 | 522 | 13 / 11,346 / 331,678 | Node 级解析、格式化、JSON 映射；高复杂度 |
| goldmark v1.8.6 | 5.0k | 50 / 12,802 / 449,976 | Markdown + GFM 转 HTML；输入语法完整，非简单字符串替换，高复杂度 |
| plumbing v1.3.0 | 7 | 8 / 299 / 6,643 | sevenzip 的 ReaderAt tee、关闭链、限长 reader、写入计数；单独算法小，但必须改上游 |
| windows v1.0.1 | 2 | 1 / 44 / 1,471 | sevenzip 的 FILETIME；单独算法小，但必须改上游 |
| go4.org 当前伪版本 | 332 | 6 / 400 / 9,786 | 只导入 readerutil，用于分卷 ReaderAt；单独实现中小，但必须改上游 |
| golang-lru/v2 v2.0.7 | 5.1k | 6 / 906 / 23,552 | sevenzip 的 AES 密钥缓存；超过阈值，附带检查；不能从主项目直接移除 |
| x/text v0.40.0 | 808（官方镜像） | 37 / 168,402 / 4,425,519 | QR ECI / 汉字等编码、afero 文本支持；生成映射表多，官方库不能按镜像星数判小众 |
| x/xerrors 当前伪版本 | 279（官方镜像） | 8 / 639 / 16,867 | gozxing 错误类型兼容；直接依赖时可考虑 errors/fmt，当前需修改 gozxing |

表中版本源码及算法注册可从 [sevenzip go.mod](https://github.com/bodgit/sevenzip/blob/v1.6.5/go.mod)、[sevenzip register.go](https://github.com/bodgit/sevenzip/blob/v1.6.5/register.go)、[gozxing QR 目录](https://github.com/makiuchi-d/gozxing/tree/v0.1.1/qrcode)、[gozxing ECI](https://github.com/makiuchi-d/gozxing/blob/v0.1.1/common/character_set_eci.go) 核对。统计全模块会夸大实际使用面，例如 gozxing 全模块约 20,453 行，但 CLI 导入的 QR 相关包为 8,806 行。

## 看起来职责单一，实际要接管什么

**拼音：保留依赖。** `text.pinyin` 支持 plain、tone、number、initials，还支持 heteronym=true、转换后读音去重、每个非汉字 rune 回退、separator 及 syllables 结果。不是只取首读音，也没有上下文消歧。上游 `pinyin_dict.go` 为 788,900 字节，算法文件 `pinyin.go` 仅 7,325 字节。固定字典配合自写风格转换是可行的，工程估计为数天含全字典差分、鼻音/ü/多音字边界验证；但仍需保留数据来源与 MIT 许可，并承担新字和读音修订。不能将搬表称作独立自研。本轮不实施。[上游源码](https://github.com/mozillazg/go-pinyin/blob/v0.21.0/pinyin.go)、[字典](https://github.com/mozillazg/go-pinyin/blob/v0.21.0/pinyin_dict.go)、[许可](https://github.com/mozillazg/go-pinyin/blob/v0.21.0/LICENSE)。

**QR：生成不等于识别。** 项目要求编码级别 L/M/Q/H、UTF-8、quiet zone、尺寸、配色、vCard，并用同一识别器回读；外部图片还需要二值化、定位、几何采样、格式识别、Reed–Solomon 纠错和文本编码处理。只写二维码矩阵生成无法替代当前能力。DENSO 的版本说明列出 1–40 的符号规模及容量差异；成本判断基于当前 QR encoder/decoder/detector 源码，估计完整替代为数周以上且需真实图像语料，不是承诺工期。[DENSO 官方说明](https://www.qrcode.com/en/about/version.html)、[上游 QR 源码](https://github.com/makiuchi-d/gozxing/tree/v0.1.1/qrcode)。

**TOML/YAML：调用少不代表语法少。** TOML 当前库支持 1.1：要处理引号/多行串、dotted key、数组表、时间类型、整数精度、重定义拒绝及合法重编码。YAML 即使业务拒绝 alias 和非字符串键，仍先依赖完整 scanner/parser 构建 Node；块字符串、flow style、标量解析和缩进歧义都不能省略。两者缩成 JSON 子集都会改变当前功能。自写合规实现估计数周以上，应保留库并单独评估 YAML 版本线。[TOML 规范](https://toml.io/en/v1.1.0)、[go-toml README](https://github.com/pelletier/go-toml/tree/v2.4.3)、[YAML 规范](https://yaml.org/spec/1.2.2/)。

**Markdown：保留 goldmark。** 当前启用 GFM，需要块和行内语法优先级、嵌套列表、代码块、引用、强调分隔符、链接、表格等兼容；还要保持现有默认 HTML/危险链接处理语义。短小正则实现会在正常文档上退化。替代成本为数周以上及规范测试，不因只有一次 Convert 调用而变小。[CommonMark 规范及测试设计](https://spec.commonmark.org/0.31.2/)、[GFM 规范](https://github.github.com/gfm/)、[goldmark 源码](https://github.com/yuin/goldmark/tree/v1.8.6)。

**压缩：保留实现，考虑减少格式需求。** XZ 涉及流、块、索引、校验及 filter chain，Brotli 有 Huffman/LZ77、上下文和预定义字典，LZ4 的块格式较小但当前支持含帧与校验，PPMd 涉及概率模型。对固定安装包可以减少支持算法，却需要先证明锁定产物及未来更新的兼容范围；不是“当前一次没见到该算法”就能安全删除。已有 sevenzip 在 init 中直接导入并注册算法，覆盖 RegisterDecompressor 不会删除编译依赖；裁剪必须 fork 上游。[XZ 规范入口](https://tukaani.org/xz/format.html)、[Brotli RFC 7932](https://www.rfc-editor.org/info/rfc7932/)、[LZ4 帧规范](https://github.com/lz4/lz4/blob/dev/doc/lz4_Frame_format.md)、[sevenzip 注册代码](https://github.com/bodgit/sevenzip/blob/v1.6.5/register.go)。

## 小辅助模块能否单独去掉

技术上 FILETIME 换算和简单 I/O 包装可用几十至几百行实现。工程上 imports 位于 sevenzip 内部：主项目新增同类函数不会替换它们，删除 go.mod 的 indirect 行后 tidy 会恢复。要么上游接受修改，要么维护 fork/replace；在本仓库伪装同 module path 的本地兼容包也是维护 fork，不是无成本消除依赖。

go4 只用 readerutil 的 SizeReaderAt、NewMultiReaderAt，支持 sevenzip 分卷随机读取；不能因为当前 registry 无分卷就假定上游该逻辑不编译。plumbing 涉及正确传播 Read/Close 错误、CRC tee 和长度限制；windows 是跨平台 FILETIME 类型。单独重写本身不复杂，但为了这些模块接管 7z 的发布同步不划算。LRU 和 xerrors 同理，前者在 sevenzip 内，后者在 gozxing 内。[sevenzip reader.go](https://github.com/bodgit/sevenzip/blob/v1.6.5/reader.go)、[struct.go](https://github.com/bodgit/sevenzip/blob/v1.6.5/struct.go)、[types.go](https://github.com/bodgit/sevenzip/blob/v1.6.5/types.go)、[windows 源码](https://github.com/bodgit/windows/blob/v1.0.1/filetime.go)。

## 整体移除 sevenzip 的可行路线

当前 [registry.json](../../pkg/packagemanager/registry.json) 的 7z 输入有 ImageMagick Windows x64/arm64、7zip extra、7zip-full 安装器；pdfcpu 的四个非 Windows 产物是 tar.xz。Tesseract 的 NSIS 已调用托管的 7zip-full，但 7zip-full 自己仍靠 Go sevenzip 解开。因此直接将所有 7z 解压改成“先安装 7zip”会产生引导循环。

1. **优先争取 ZIP/tar.gz 产物。** 标准库即可读取，最容易维持逐条安全检查。但目前上游对应包不是这些格式；若自行重打包，要建立来源校验、许可保留、发布与更新流程，依赖转移到发行流水线。不能只把 URL 后缀改掉。ImageMagick、7zip 自身、pdfcpu 要分别解决。
2. **Windows 使用官方裸 7zr.exe 引导。** 官方下载页确有无需先解包的 x86 `7zr.exe`，可固定版本/URL/hash，先验证再调用，以解开 7zip extra 或完整版本；不执行安装器。需要实际验证该二进制对锁定包及 ARM64 仿真环境的兼容。Linux/macOS 官方 CLI 仍以 tar.xz 分发，不能顺带声称 xz 也可删除。[官方 7-Zip 下载格式](https://7-zip.org/download.html)。
3. **保留提取安全边界。** 目前 Go 路径在写入前逐项检查路径、链接/特殊文件、大小写冲突、50,000 项/1 GiB 总量，并使用 O_EXCL 和受限 reader。现有 NSIS 路径有列表预检、超时、事后核对，可以复用设计，不能直接复制成通用 7z 解压器便宣称等价：外部解压器在落盘时的路径/重解析点/覆盖语义，以及运行期间的真实写盘上限，需要单独验证或隔离；事后发现越界不能撤销根目录之外的写入。见 [archive.go](../../pkg/packagemanager/archive.go)、[nsis.go](../../pkg/packagemanager/nsis.go)。

去掉 sevenzip 后理论上可随 tidy 消失的是 brotli、plumbing、windows、LRU、compress、lz4、afero、ppmd、go4 共 9 个间接模块。xz 仍由 tar.xz 直接需要；x/text 和 xerrors 仍由 QR 引入。此处是依据 import 图的预测，须实施后用 go list/tidy 验证，不是已完成结果。

该路线有明显收益，但估计是数天至一两周的引导、跨平台和恶意归档回归工作，不是本轮低复杂度自写范围。所有工期仅为源码审阅后的量级估计，不含持续数据/格式兼容维护。

## 本轮建议

有清晰有限输入范围的农历替代已完成；其余暂保留。下一次减依赖优先做 sevenzip 整链引导设计，优先改变托管产物而非重写压缩算法。拼音保留上游数据和库，不为降低 go.mod 数量搬一整份字典。运行时及 YAML 升级建议见 [升级评估](runtime-upgrade-evaluation-2026-09-08.md)。
