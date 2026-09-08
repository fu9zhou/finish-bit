# 运行时与 YAML 升级可行性评估

后续实施状态（2026-09-08）：Tesseract 已按本报告升级为 `5.5.3-fast-8741641`；FFmpeg、Poppler 与 YAML 仍按下述条件暂缓。下文“仅评估”描述的是最初调查，实施证据见文末。

核查日期：2026-09-08。范围：FFmpeg/ffprobe、Tesseract、Poppler、go.yaml.in/yaml。仅评估，未改 registry/go.mod，未下载或执行候选二进制。页面列出资产不等于下载校验和运行验收通过。

## 结论

| 项目 | 建议 | 具体条件或阻碍 |
| --- | --- | --- |
| Tesseract 5.5.0 → 5.5.3 | 推荐优先升级，工作量较小 | 官方已有 Windows x64 安装包；仍须验证 NSIS 解包后的 DLL、tessdata/configs 和 OCR 四种输出 |
| FFmpeg/ffprobe 6.1.1 | 推荐更换停滞的分发渠道；暂缓全平台统一改版本 | 6.1.6 是低风险版本方向，但未核实齐备的五平台预编译包；BtbN 有较新 Windows/Linux 包，缺 macOS；需要固定构建、完整编解码器与平台矩阵 |
| Poppler 26.07 → 26.09 | 暂缓直接升级 | 上游源码已发布，但现用 Windows 分发 latest 仍是 26.07.0-0，没有核实到同渠道 26.09 ZIP |
| YAML v3 → v4 | 暂缓生产迁移，保留 v3.0.5 | 可取得的 v4 最新标签仍是 v4.0.0-rc.6；v3 仍收安全修复。API 改动预计小，但没有项目必须立即升级的新功能需求 |

## FFmpeg / ffprobe

### 当前约束与候选资产

`pkg/packagemanager/registry.json` 将两个包分别固定为 6.1.1，覆盖 win32-x64、linux-x64、linux-arm64、darwin-x64、darwin-arm64，全部来自 eugeneware 的单文件 gzip。其 [latest release](https://github.com/eugeneware/ffmpeg-static/releases/latest) 本次仍跳转到 b6.1.1，继续跟这个渠道无法获得新版。

[FFmpeg 官方下载页](https://ffmpeg.org/download.html) 明确只提供源码；6.1.6 是 2026-06-20 发布的 6.1 分支维护版本。**版本方向上优先考虑 6.1.6，但本次没有核实到可直接替换上述五平台的 6.1.6 二进制集合**。不能把源码存在写成“升级只需换 URL”。

| 当前平台 | 本次核实到的替代分发 | 状态 |
| --- | --- | --- |
| Windows x64 | [Gyan release](https://www.gyan.dev/ffmpeg/builds/) 列出 9.0.1 essentials ZIP/7z、对应 SHA-256 和 8.1.2 历史包；包含 ffmpeg 与 ffprobe | 有资产，跨大版本兼容性未运行验证 |
| Linux x64 | [BtbN 资产页](https://github.com/BtbN/FFmpeg-Builds/releases/expanded_assets/latest) 列出 `ffmpeg-n8.1-latest-linux64-gpl-8.1.tar.xz` 和 n9.0 对应包 | 有资产；分支构建不能未经执行就标成精确 8.1.2/9.0.1 |
| Linux arm64 | 同页列出 `ffmpeg-n8.1-latest-linuxarm64-gpl-8.1.tar.xz` 和 n9.0 对应包 | 有资产；同上 |
| macOS x64 | [Evermeet](https://evermeet.cx/ffmpeg/) 列出 9.0.1 ffmpeg/ffprobe，ZIP/7z 和 GPG 签名 | 有资产；要求 macOS 10.13+，完整运行与路径未验证 |
| macOS arm64 | Evermeet 明确不提供原生 Apple Silicon 构建，BtbN 不构建 macOS | 上述候选渠道未覆盖；不能拿 Intel 包冒充原生 arm64 |

[BtbN 的保留策略](https://github.com/BtbN/FFmpeg-Builds#release-retention-policy) 是日包保留 14 天、月末包两年，`latest` 会漂移。生产 registry 应使用固定构建并解决长期镜像保留，不能把 `latest` URL 与固定 SHA 拼起来，否则下一次更新便校验失败。BtbN 的 Linux 构建最低环境为 glibc 2.28、Linux 4.18；更换前应明确我们的平台基线。不同平台的构建号也需准确体现在 registry 元数据中。

### 与项目的兼容边界

`internal/ffmpeg/provider.go` 和 `catalog.go` 调用 CLI，没有链接 libav ABI，因此 ABI 升级不是直接阻碍。实际约束是 `libx264`、`libvpx-vp9`、`libmp3lame`、`libopus`、AAC、`subtitles`/libass、`drawtext`、lavfi 与静音过滤器；`internal/desktop/provider.go` 还需要 Windows `gdigrab`。BtbN 明确 LGPL 变体缺少 libx264/libx265，不能选它；应选 GPL static。[BtbN 变体说明](https://github.com/BtbN/FFmpeg-Builds#targets-variants-and-addins)

Gyan 声明其主构建为 GPLv3，项目当前登记 GPL-3.0-or-later；换包后仍须随实际构建保存许可证与组件信息，不能用构建脚本仓库的 MIT 许可证替代二进制许可。Gyan essentials 公布的组件表包含项目上述外部库，仍需执行 `-encoders`、`-filters` 和 `-devices` 确认最终产物。[Gyan 构建说明](https://www.gyan.dev/ffmpeg/builds/)

跨大版本应验证：字幕字体发现、滤镜选项/默认值、音视频时间戳、双遍压缩体积、ffprobe JSON 的 streams/format/chapters，以及静音检测的 `lavfi.silence_*` frame tags。项目解析 JSON 较宽松，但上游增加字段会改变对外返回数据，不能只比较进程退出码。

实施验收：固定同一构建的 ffmpeg/ffprobe；下载比对哈希、检查包内路径后再登记。设置 `FINISHBIT_TEST_FFMPEG`、`FINISHBIT_TEST_FFPROBE` 运行 `internal/ffmpeg/extended_integration_test.go` 的真实工具测试，并补充 Windows 桌面录制检查。五平台都须执行，普通 `go test` 会在没设置环境变量时跳过主要真实运行时测试。

推荐将分发改造作为单独变更：先解决五平台固定版本/长期资产，再确定 6.1.6 自行构建，还是统一迁移到可获取的较新维护线。自行构建 FFmpeg 是维护编译流水线，并非自行重写媒体引擎，投入明显大于改 registry。

## Tesseract

[官方 5.5.3 资产页](https://github.com/tesseract-ocr/tesseract/releases/expanded_assets/5.5.3) 已列出 Windows x64 的 `tesseract-ocr-w64-setup-5.5.3.20260724.exe`（25.3 MB），发布页 SHA-256 为 `bee9e3434bd94fd65387d9be28cd467a41f61b1275383b55b0f59a1331270ae4`。这是页面元数据，本次未下载复算。没有在该发布页看到其他平台二进制；当前项目也仅登记 Windows x64。

[5.5.3 发布说明](https://github.com/tesseract-ocr/tesseract/releases/tag/5.5.3) 包含模型反序列化内存安全、LSTM 整数溢出、缺少 LSTM 时崩溃等修复，也修改了 Windows 安装器。升级有实际收益。项目模型固定为官方 tessdata_fast 提交 `87416418657359cb625c412a48b6e1d6d41c29bd` 的 eng/chi_sim，不接受任意用户模型，不能据发布说明直接推断当前应用可被这些模型漏洞利用。

`internal/ocr/provider.go` 使用 `--oem 1`、`--psm`、`--dpi`、显式 tessdata 目录与 txt/tsv/pdf/hocr；不涉及 C API。预计无需调整 Runner 接口。`parseTSV` 要求 12 列及 level 表头；真实版本必须验收词框、置信度和中文识别，不能假定输出逐字不变。

具体升级预计是 registry 的版本、source、URL、SHA 和描述更新，保留既有模型及其许可证；不应同时换模型以免难以区分效果变化。现有 `pkg/packagemanager/nsis.go` 以托管 7zip-full 列目录、检查路径、解包，不执行安装器。须先确认新版 NSIS 能被此流程解包，DLL/configs 完整且路径仍对应 `tesseract.exe`。引擎登记 Apache-2.0 加捆绑组件许可；新包第三方组件清单仍需随下载核对，未做无变化保证。

验收：全新目录安装、哈希错误失败、repair、`--version`、`--list-langs`、中文/英文图片的四个 OCR 能力，以及 PDF OCR/提取文本/扫描图片转 Word。项目已有 `pkg/app/local_workflows_test.go`（`FINISHBIT_TEST_LOCAL_HOME`）可复用；必须使用新安装目录，避免误测旧运行时。该项建议优先落地。

## Poppler

[上游 26.09.0 发布说明](https://poppler.freedesktop.org/) 已有源码包，包含异常 PDF 崩溃修复、pdftohtml 加速及 pdftotext 新的可选 `-urls`；构建新增 harfbuzz 字体子集要求。但 [Windows 分发 latest](https://github.com/oschwartz10612/poppler-windows/releases/latest) 本次仍是 v26.07.0-0，且最新提交说明正是修复缺 DLL。**不能直接将版本与路径字符串替换成 26.09**。

当前依赖边界在 `internal/pdf/reading.go` 的 CLI；预计现有参数变动较小，此判断来自发布说明，不是回归通过结论。下一版须同步替换 registry 所有 `poppler-26.07.0/Library/bin/*` 路径，校验完整 DLL/编码数据，保留 GPL 与捆绑组件许可证。不要仅拷贝 exe。

验收：`FINISHBIT_TEST_HOME` 指向新 runtime 运行 `internal/pdf/provider_integration_test.go`，覆盖中文文本、bbox、PNG/JPEG、内嵌图片、字体表、HTML/PS；再测 OCR、图片 PDF、长图等 `pkg/app` 组合能力。关注 bbox 数值、文字顺序与渲染输出，避免要求跨版本像素完全相等。建议待 Windows 分发有固定 26.09 包后做；现在自行搭 conda/编译打包的投入大于本次常规升级收益。

## YAML v4

需要修正“上游在做 v4，所以现在该升级”的推论：[官方标签](https://github.com/yaml/go-yaml/tags) 和本次读取的 [Go module proxy 列表](https://proxy.golang.org/go.yaml.in/yaml/v4/@v/list) 仅有 v4.0.0-rc.1 至 rc.6，没有正式 v4.0.0。rc.6 的 [模块元数据](https://proxy.golang.org/go.yaml.in/yaml/v4/@v/v4.0.0-rc.6.info) 时间为 2026-06-17。[官方 README](https://github.com/yaml/go-yaml) 说明 v3 继续收安全修复，常规开发在 v4；因此暂留已使用的 v3.0.5 是合理选择。

`internal/localtools/formats.go` 是唯一使用点，负责 YAML/JSON 转换、格式化、校验，使用 Node AST、Decoder/Encoder、SetIndent。核对 [rc.6 源码](https://github.com/yaml/go-yaml/blob/v4.0.0-rc.6/yaml.go) 后这些入口仍在；[go.mod](https://proxy.golang.org/go.yaml.in/yaml/v4/@v/v4.0.0-rc.6.mod) 仅要求 Go 1.18，当前 Go 1.25.13 足够。预计迁移规模小，正式版后再做行为回归；本次未编译候选版本，不能承诺只改 import 即完全兼容。

项目有意拒绝 aliases、多文档、非字符串键、重复键、非有限数，并限制深度 128/节点 100000；保存 JSON 大整数精度。迁移应保留这些条件，再检查 YAML 1.1/1.2 的 yes/on、八进制、时间戳 tag、空文档、注释与多行字符串格式化、Encoder 缩进。已有 `internal/localtools/localtools_test.go` 包含部分反例，应补充差异行为，而非复制新库实现写测试。

YAML 解析器虽在本仓库星数不足 1k，实际是老 go-yaml 由 YAML 官方组织接手的延续；不是新写的小玩具。只用四个入口也需要完整词法、缩进、引号、转义和标量语义。推荐保留库，不自写 YAML 子集来悄悄缩减支持范围。上游声明 MIT/Apache-2.0 双许可，迁移时继续保留适用许可证。

## 验证范围

本评估完成仓库调用点/测试入口检查、官方发布/源码说明读取、候选资产页面与 Go proxy 元数据核实。没有下载候选 runtime，没有执行升级回归，也没有变更依赖。未来任何实际代码或 registry 变更按仓库要求运行 `go test ./...`、`go vet ./...`、`go build ./cmd/fnsh`，并启用上述真实工具测试；三项普通 Go 检查通过不能替代二进制兼容验收。

## Tesseract 实施与验收（2026-09-08）

用户确认按评估实施后，已将 `pkg/packagemanager/registry.json` 的引擎固定为 `5.5.3-fast-8741641`，更新 source、安装包 URL 和 SHA-256。保留原来的 eng/chi_sim 模型提交、三个补充资源哈希和 NSIS 解包路径，不执行安装器。FFmpeg、Poppler、YAML 没有升级。

### 实测产物

| 项目 | 结果 |
| --- | --- |
| 安装包 | `tesseract-ocr-w64-setup-5.5.3.20260724.exe`，26,573,224 字节（约 25.3 MiB） |
| 下载后复算 SHA-256 | `bee9e3434bd94fd65387d9be28cd467a41f61b1275383b55b0f59a1331270ae4`，与发布值一致 |
| 实际引擎 | `tesseract v5.5.3.20260724`，Leptonica 1.87.0 |
| 模型 | `--list-langs` 返回 `chi_sim`、`eng`；两个文件的 SHA-256 与原 registry 固定值一致 |
| 展开大小 | payload 共 141 个文件，108,267,960 字节（约 103.3 MiB），包括模型，未含另行安装的 7-Zip 助手及下载缓存 |
| 资源 | `doc/LICENSE`（Apache-2.0）、`doc/AUTHORS`、模型 LICENSE 及 `tessdata/configs/{txt,tsv,pdf,hocr}` 均保留 |

验收使用独立目录 `C:/Users/wps/AppData/Local/Temp/finishbit-tesseract-553-740cfee6810c44f3b66350100ee6d1fd/home`。该目录没有旧版 Tesseract，先完成全新安装，再执行 `pkg repair tesseract` 强制重新下载、校验和激活，最后重复 `pkg add tesseract` 确认识别已安装版本。7zip-full、ImageMagick、pdfcpu、Pandoc、Poppler 在同一目录从既有缓存重新校验安装；未修改用户默认运行时目录。

### 回归结果

- `TestLocalWorkflowsWithManagedTools` 启用上述新目录后通过（测试主体约 45 秒）。覆盖英文和中文 OCR、词坐标/置信度、英文 hOCR、中英文可搜索 PDF，以及 PDF OCR、扫描 PDF 提取文字、扫描图片转 Word 与其他 PDF/文档组合流程。
- 增加词框必须在 1200×240 测试图片内、尺寸大于零、置信度在 0–100 的断言；增加中文 PDF 文字层检查。首次检查发现提取结果在相邻中文字符间带换行，改为忽略提取空白后比较实际文字，未要求不属于契约的固定换行布局。
- `go test ./...`、`go vet ./...`、`go build ./cmd/fnsh` 全部通过。已有校验和拒绝、repair 失败保留当前包等包管理器单元测试随全套测试运行。
- `node scripts/website-catalog.mjs --check` 验证 277 个能力的目录未漂移；`node --check website/app.js` 与 `git diff --check` 通过。
- 同步更新 package-management、local-tools、第三方声明、网站的体积说明及 Unreleased 变更记录。运行时适配器接口无需修改。

已有安装的使用者需要使用更新后的 CLI 执行 `fnsh pkg add tesseract`（或 `fnsh pkg repair tesseract`）。本次验证了注册版本和独立安装，不代表已有发布包或所有用户安装目录已经自动更新。
