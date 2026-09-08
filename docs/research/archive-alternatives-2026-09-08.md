# 低星归档依赖的高星库与托管工具替代

核查日期：2026-09-08。只研究，不修改依赖或安装行为。结合当前 `go.mod`、`go mod graph`、`pkg/packagemanager/archive.go`、`nsis.go`、`registry.json` 以及一手上游源码和 release 资产清单。星数为当日 GitHub 页面显示的约数；API 遭遇限流，未把缺失值当成零。精确历史星数沿用 [同日健康报告](go-dependency-health-2026-09-08.md)。

## 判断

最值得实施验证的是 **官方 7-Zip 托管工具替换 Go sevenzip 整链**。更高星的 Go 归档包装库并没有消除底层依赖。XZ 先保留，等 pdfcpu 的非 Windows 安装包来源解决后再单独移除。不是自写压缩算法，也不是删除 indirect 行。

| 方案 | 热度和维护证据 | 对项目的实际效果 | 建议 |
|---|---|---|---|
| [mholt/archiver](https://github.com/mholt/archiver) | 约 4.4k 星；2024-11-19 已归档，明确弃用 | 不是活跃替代；旧 v4 的 7z 也是包装底层库 | 不换 |
| [mholt/archives](https://github.com/mholt/archives) | 441 星；前者的继续开发项目 | `7z.go` 直接 import bodgit/sevenzip；go.mod 还保留 xz、brotli、lz4 和小辅助模块，并增加 RAR、lzip 等格式库 | 不换；这是 API 整合，不是减依赖 |
| [官方 7-Zip](https://github.com/ip7z/7zip) | 约 3.9k 星；26.03 于 2026-09-03 发布 | 利用已有托管 7zip / 7zip-full，可让 sevenzip 和 9 个间接模块退出 Go 构建；增加引导工具及安装时进程依赖 | 第一优先级试验 |
| [libarchive / bsdtar](https://github.com/libarchive/libarchive) | 约 3.6k 星；[3.8.9](https://www.libarchive.org/) 于 2026-07-28 发布 | C 库与 CLI 支持 7z、xz；cgo 会增加本地构建工具链，CLI 需要可分发的二进制 | 次选，不比已有 7-Zip 省事 |

archives 的依赖穿透证据：[7z.go](https://raw.githubusercontent.com/mholt/archives/main/7z.go)、[go.mod](https://raw.githubusercontent.com/mholt/archives/main/go.mod)。表中的净收益不是已执行 tidy 后的结果，而是当前导入关系的推断。

## 为什么单个算法换高星库不能解决

当前低于 1k 星的 sevenzip、xz、brotli、lz4、ppmd、plumbing、windows、go4 形成同一主要安装链。除直接读取 tar.xz 外，FinishBit 没有自行选择 sevenzip 内部每个算法。上游 [register.go](https://github.com/bodgit/sevenzip/blob/v1.6.5/register.go) 在编译时直接导入和注册解码器；应用覆盖 RegisterDecompressor 不能让原 imports 消失。

- [google/brotli](https://github.com/google/brotli) 约 14.9k 星、[lz4/lz4](https://github.com/lz4/lz4) 约 12.1k 星，属于成熟的原生实现。但替换 sevenzip 的 Go 内部适配，需要上游修改或自己 fork，并增加 C 构建/动态库/外部进程，并非直接换 module。
- 已有 [klauspost/compress](https://github.com/klauspost/compress) 约 5.6k 星，主要覆盖其列出的 gzip、deflate、zstd、S2 等，不能把 Brotli、XZ、PPMd 等当成同一种格式直接交给它。sevenzip 自己已经用它，新增一次直接依赖没有减链收益。
- [dsnet/compress](https://github.com/dsnet/compress) 有 Brotli reader，但只有 419 星；功能表不含 XZ、7z，API 稳定性也有明确限制，不是更高关注度的整体替代。
- plumbing、windows、go4 的少量辅助代码可重写，但必须先 fork sevenzip 才能删原模块。不建议为了三个小模块接管上游维护。

## 官方 7-Zip package 路线

当前 7zip-full 解 Tesseract 的 NSIS，自己却由 Go sevenzip 解开；普通 7zip 也是 `.7z` 包。直接让它们互相安装会循环。

[官方固定版本下载页](https://7-zip.org/download.html) 提供 [26.03 裸 7zr.exe](https://github.com/ip7z/7zip/releases/download/26.03/7zr.exe)，是无需先解包的 Windows x86 控制台程序。建议增加独立 bootstrap 包：固定 URL、SHA-256、许可和版本，验证后仅用它解出完整 7-Zip，再由完整 7-Zip 处理业务包。不能用未固定的 latest，也不能执行安装器来代替解包。

可分阶段落地：

1. Windows x64：验证裸 7zr 解开锁定的 `7z2603-x64.exe`、`7z2603-extra.7z`；然后验证完整 7-Zip 解 ImageMagick 的 portable 包以及 Tesseract NSIS。bootstrap 下载本身不需要另一个归档工具。
2. Windows arm64：当前 ImageMagick 已有 arm64 产物，不能删除其安装能力。7zr 只有 x86 版本，必须在实际 arm64 环境核验仿真可执行性；可考虑它引导官方 arm64 完整安装包，并保留架构独立的 registry 描述。不能因 x64 成功就宣称 arm64 已支持。
3. 保留 ZIP/tar.gz 的标准库路径，第一阶段仍保留 Go xz。非 Windows 官方 7-Zip 以 tar.xz 分发，拿它替换 xz 会形成新的引导问题。

**预期 Go 模块 19 → 9**：移除 sevenzip 本体以及 brotli、plumbing、windows、golang-lru/v2、klauspost/compress、lz4/v4、afero、ppmd、go4 共 9 个间接模块。xz 仍用于 pdfcpu；x/text、xerrors 仍用于 gozxing。收益需实际改造后由 tidy、go list 和构建验证。运行时总依赖没有归零：新增一个 bootstrap package，并扩大已有 7-Zip package 的安装用途。

## 能否改成 ZIP，直接不需要解压工具

[ImageMagick 7.1.2-31 资产](https://github.com/ImageMagick/ImageMagick/releases/expanded_assets/7.1.2-31) 中 Windows portable 是 `.7z`，没有同等 portable ZIP；`Source code (zip)` 是源码，不能替换运行时。MSIX bundle 也不是已验证的 portable 等价包。

[pdfcpu 0.15.0 资产](https://github.com/pdfcpu/pdfcpu/releases/expanded_assets/v0.15.0) 中 Linux/macOS 运行时为 tar.xz，Windows 为 ZIP。要完全删除 xz，需上游增加 tar.gz/ZIP，或我们建立可审计的二次打包发布流程。后一种方案可把复杂解压移到发行流水线，但要维护源包校验、重打包哈希、原始许可、版本同步；不是只有改 URL。

[libarchive 3.8.9 release 资产](https://github.com/libarchive/libarchive/releases/expanded_assets/v3.8.9) 的 ZIP/tar.gz/tar.xz 是源码包，未发现官方同版本覆盖项目五个平台的现成 CLI 包。系统 tar/bsdtar 依赖系统版本和编译选项，不能默认各平台都有 XZ/7z/NSIS 支持，也不能保证锁定版本。若采用它，需额外维护二进制构建/分发，或显式接受系统前置条件。当前不优先。

## 实施难度和验证边界

估计官方 7-Zip 引导与替换为数天至一两周量级，主要是跨架构安装和安全回归，不是算法工作。以下是当前实现的等价性要求，并非本轮发现的可利用漏洞：

- 保持下载哈希校验、独立临时目录、安装成功后再提交结果及失败清理。
- 保持归档路径、绝对路径、Windows 特殊名称、大小写冲突、软/硬链接、条目数和大小约束。
- 当前 Go 路径逐条受限读取并由自己控制文件创建；CLI 一次性直接落盘的列表预检和事后 walk 不能完全替代运行期间约束。需要选择隔离执行或受控输出策略，并做恶意条目测试；只写一个 `7z x` 不够。
- 冷安装、并发安装、取消、损坏下载、修复安装、空缓存与离线已缓存运行都要覆盖；Windows arm64 的实际验证单列。

本轮未下载执行新解压器、未修改代码、未声称完成兼容验收；没有需要运行 Go 全量代码检查的实现变更。
