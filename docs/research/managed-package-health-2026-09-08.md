# 托管工具包维护状况核查（2026-09-08）

后续状态：此表保留升级前的盘点快照。同日 Tesseract 已升级到 5.5.3，见 [升级与验收记录](runtime-upgrade-evaluation-2026-09-08.md)。

本次只做依赖盘点与上游维护状况研究，没有升级、安装或修改功能代码。清单依据当前 `pkg/packagemanager/registry.json`，共 10 个 package，对应 8 个上游项目；FFmpeg/ffprobe、7zip/7zip-full 分别同源。Go 模块另见 [Go 依赖报告](go-dependency-health-2026-09-08.md)。本次覆盖应用工具包及 Go 模块，不包含 GitHub Actions、Go 工具链以及下载二进制内每个捆绑 DLL 的逐项审计；仓库未发现 npm/Python/Rust 包清单。

评价综合项目定位、公开社区规模、稳定版发布与当前固定版本。GitHub stars 为查询时页面显示的约数，只是关注度指标，不是下载量、实际用户数或质量保证；官网不在 GitHub 的项目不以某个镜像的星数替代。发布日期采用 GitHub 页面 HTML datetime 的 UTC 日期；例如 qsv 发布说明标题写 8 月 9 日，但页面发布时间为 UTC 8 月 8 日。没有执行漏洞扫描，不据此断言固定版本没有漏洞。

## 全部托管包

| Package / 项目内用途 | 当前固定版本 | 流行程度依据 | 本次查到的稳定版及发布时间 | 判断 |
| --- | --- | --- | --- | --- |
| `ffmpeg`、`ffprobe`：音视频处理、媒体信息检查 | 6.1.1 | 成熟音视频基础设施；官网列出 Debian、Ubuntu、Fedora 等分发渠道，不使用第三方 ffmpeg-static 的星数衡量引擎 | 官方 9.0.1（2026-08-12），6.1 分支仍有 6.1.6（2026-06-20） | 上游活跃；我们的版本及分发渠道是首要更新关注项。[官方版本与渠道](https://ffmpeg.org/download.html) |
| `pdfcpu`：PDF 页面、结构、表单、附件处理 | 0.15.0 | [官方仓库](https://github.com/pdfcpu/pdfcpu)约 8.8k stars，垂直领域有一定社区规模 | [0.15.0](https://github.com/pdfcpu/pdfcpu/releases/tag/v0.15.0)，2026-08-11 | 活跃；与查到的最新稳定版一致 |
| `poppler`：PDF 文字、坐标、字体、图片提取及页面渲染 | 26.07.0-0 | [官方项目](https://poppler.freedesktop.org/)在 freedesktop GitLab，提供 C++/GLib/Qt 接口与多平台 CI；第三方 GitHub 仓库热度不能代表它 | 官方 26.09.0，2026-09-03；[Windows 分发 26.07.0-0](https://github.com/oschwartz10612/poppler-windows/releases/tag/v26.07.0-0)，2026-09-04 | 上游及分发均有近期活动，但打包版本落后上游 |
| `imagemagick`：图片转换、编辑、合成 | 7.1.2-31 | [官方仓库](https://github.com/ImageMagick/ImageMagick)约 17.4k stars，成熟图片工具 | [7.1.2-31](https://github.com/ImageMagick/ImageMagick/releases/tag/7.1.2-31)，2026-09-03 | 活跃；与查到的最新稳定版一致 |
| `qsv`：CSV 清理、连接、统计分析 | 22.0.1 | [官方仓库](https://github.com/dathere/qsv)约 3.8k stars，规模较小的专业数据工具 | [22.0.1](https://github.com/dathere/qsv/releases/tag/22.0.1)，2026-08-08 | 很活跃：发布页显示此后有 189 次主分支提交；与查到的最新稳定版一致。不能因规模较小判断为无人维护 |
| `pandoc`：结构化文档转换及 PDF 转 Word/PPT 工作流的文档生成 | 3.11 | [官方仓库](https://github.com/jgm/pandoc)约 46.2k stars，成熟文档转换工具 | [3.11](https://github.com/jgm/pandoc/releases/tag/3.11)，2026-08-29 | 活跃；与查到的最新稳定版一致 |
| `7zip`：压缩解压；`7zip-full`：提取 Tesseract NSIS 安装包 | 26.03 | [官网](https://7-zip.org/)提供成熟压缩工具；[官方 GitHub 仓库](https://github.com/ip7z/7zip)约 3.9k stars，GitHub 提交数不反映完整项目历史 | [26.03](https://github.com/ip7z/7zip/releases/tag/26.03)，2026-09-04 | 有近期发布；两个 package 是同一上游的不同发行形态，不是两个压缩引擎 |
| `tesseract`：离线 OCR，附带简体中文和英语 fast 模型 | 5.5.0-fast-8741641 | [官方仓库](https://github.com/tesseract-ocr/tesseract)约 76.4k stars，成熟 OCR 引擎 | [5.5.3](https://github.com/tesseract-ocr/tesseract/releases/tag/5.5.3)，2026-07-24 | 上游活跃，发布后仍有 60 次主分支提交；当前引擎落后三个补丁版本，应评估更新 |

## 二进制渠道与引擎要分开评价

1. **FFmpeg**：实际从 [eugeneware/ffmpeg-static](https://github.com/eugeneware/ffmpeg-static)下载，约 1.4k stars。其 [latest 发布](https://github.com/eugeneware/ffmpeg-static/releases/tag/b6.1.1)仍为 b6.1.1，页面时间 2025-11-14，发布后仅显示 1 次主分支提交。这说明跟随该渠道的最新版本仍会落后官方维护版本；不能只自动查询这个渠道的 latest。官方不直接提供二进制，其下载页列出了 gyan.dev、BtbN 等 Windows 渠道。是否迁移需要进一步核对各平台构建、功能、固定哈希和实际验收。
2. **Poppler**：实际渠道为 [oschwartz10612/poppler-windows](https://github.com/oschwartz10612/poppler-windows)，约 1.5k stars。维护者说明它重新打包 conda-forge 编译产物，不自行构建引擎。9 月 4 日的发布修正缺失 DLL，说明渠道仍有维护。我们用的是该渠道本次查到的 latest，但不是 Poppler 上游 latest。项目中没有引用 `innodatalabs/poppler`。
3. **Tesseract**：registry 的引擎下载 URL 位于上游 5.5.0 release；模型来自 `tesseract-ocr/tessdata_fast` 的固定提交 `87416418657359cb625c412a48b6e1d6d41c29bd`，包含 `eng` 和 `chi_sim`。模型与引擎版本应分别管理，不把模型提交当成引擎发布日期，也不以模型更新频率判断引擎活跃度。本次没有比较各模型版本识别质量。[模型仓库](https://github.com/tesseract-ocr/tessdata_fast)
4. **其余工具**：pdfcpu、ImageMagick、qsv、Pandoc、7-Zip 的当前 URL 均直接位于各自上游 GitHub release。仅凭下载托管在上游账号下，不推断所有捆绑组件都由同一维护者开发。

## 对 FinishBit 的建议

- 第一优先级：评估 FFmpeg/ffprobe 的维护版本与更新渠道。即使暂不跨大版本，也应检查 6.1.1 到 6.1.6 的修复与兼容性，而不是把分发仓库 latest 视为引擎最新状态。
- 第二优先级：评估 Tesseract 5.5.3 和较新 Poppler Windows 构建的可用性；只在对应构建与真实能力验收通过后更新 registry。
- 保留 pdfcpu、qsv、Pandoc、ImageMagick 和 7-Zip 的选型；本次没有发现仅凭热度或近期发布状况就应替换它们的依据。
- qsv 属于规模较小、功能演进快的项目。后续更新要核对实际使用命令的参数和输出，而不是只追版本号。
- 包平台覆盖是我们接入层的问题：当前 Poppler、qsv、Pandoc、Tesseract、7zip/7zip-full 只注册 Windows x64；ImageMagick 注册 Windows x64/arm64；FFmpeg/ffprobe/pdfcpu 注册 Windows x64、Linux x64/arm64、macOS x64/arm64。不能据此声称上游只支持 Windows。
- `docs/package-management.md` 当前表格没有列出 `tesseract` 和 `7zip-full`；本报告以 registry 实际清单为准，本轮没有改动该文档。

固定 SHA-256 解决下载字节完整性与可复现性，不保证下载版本足够新。这次核查的主要发现是个别版本/分发渠道滞后，而不是大面积依赖冷门无人维护的工具。
