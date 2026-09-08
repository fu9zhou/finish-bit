# QR 低星依赖替代检索

核查日期：2026-09-08。仅研究，没有替换实现、修改模块或安装 package。星数使用 GitHub 当日可见的约数；匿名 GitHub API 被限流，因此不提供伪精确统计。维护日期与分发资产单独检查，不能由星数推定。

## 结论

`gozxing` 存在更高星的候选，但没有找到“直接换掉、保持纯 Go、兼顾生成及识别、维护明显更好”的低成本替代。若接受新增托管 package，首选继续验证 **ZXing-C++ CLI**；难点是多平台二进制分发，而非 Go 调用代码。当前不建议仅为跨过 1k 星改用 GoCV，也不建议换成只生成的库。

## 项目实际边界

依据 [images.go](../../internal/localtools/images.go)：

- `qrcode.generate` 支持 UTF-8 文本、L/M/Q/H 纠错、四模块边距、64–4096 最小像素尺寸、自定义不透明前景/背景色；生成结果必须实际识别回读并逐字一致才保存。
- `qrcode.contact` 构造 vCard，再走生成与回读；不是独立编码引擎。
- `qrcode.decode` 从本地 PNG/JPEG/GIF 识别 QR；现有输入限制为 32 MiB、2500 万像素。
- 所以“仅生成”库不能单独替换 `gozxing`。只换编码器而保留解码器会增加直接依赖。

`go mod why -m` 本轮实际输出：`x/xerrors` 的引用路径是 `internal/localtools → gozxing → x/xerrors`；`x/text` 还有 `pkg/packagemanager → sevenzip → x/text/encoding/unicode`。去掉 gozxing 可移除它及 xerrors 两个模块，**不能单独连带删除 x/text**。最终应由 `go mod tidy` 和图检查确认，新候选也可能引入模块。

## 候选对比

| 候选 | 可见热度、维护证据 | 能否覆盖当前用途 | 评估 |
| --- | --- | --- | --- |
| [makiuchi-d/gozxing](https://github.com/makiuchi-d/gozxing) | 668 星；[主干最后可见提交](https://github.com/makiuchi-d/gozxing/commits/master/) 2025-07-20 | 纯 Go，编码与识别均有；当前项目用 v0.1.1 | 更新低频，但不能据此说已停止维护；仍是当前完整且部署简单的方案 |
| [skip2/go-qrcode](https://github.com/skip2/go-qrcode) | 约 3k 星；[主干](https://github.com/skip2/go-qrcode/commits/master/)停在 2020-06-17 | 只生成，支持颜色及纠错 | 更高星却更久未更新；不能替换识别及回读，不推荐 |
| [boombuler/barcode](https://github.com/boombuler/barcode) | 约 1.6k 星；[主干](https://github.com/boombuler/barcode/commits/master/) 2025-07-25 有维护 | 创建多类条码，QR 只编码 | 可作为编码器候选，但仍缺解码器；目前没必要拆成两个引擎 |
| [liyue201/goqr](https://github.com/liyue201/goqr) | 88 星；2021-07-16 已归档 | 纯 Go 识别 | 更小且只读，排除 |
| [ZXing-C++](https://github.com/zxing-cpp/zxing-cpp) | 约 1.9k 星；[v3.1.1](https://github.com/zxing-cpp/zxing-cpp/releases/tag/v3.1.1) 发布于 2026-07-29，含 QR 检测回归修复 | 同时有 ZXingReader、ZXingWriter；可走命令行 package | 最值得继续验证，避免引 CGo 绑定；尚不具备立即迁移条件 |
| [ZBar](https://github.com/mchehab/zbar) | 约 1.4k 星；[最新可见 tag 0.23.93](https://github.com/mchehab/zbar/tags) 为 2024-01-09 | zbarimg 负责识别，仍要第二套生成器 | 可行备选，但增加生成/识别两个来源且发布频率较低 |
| [GoCV](https://github.com/hybridgroup/gocv) / OpenCV | GoCV 约 7.5k 星；[v0.43.0 tag](https://github.com/hybridgroup/gocv/tags) 为 2026-01-05 | 可接入计算机视觉算法；GoCV 使用 C 风格包装调用 OpenCV C++ | 会引入 CGo、OpenCV 原生构建/运行时，为三个 QR 能力代价过大，不推荐 |

## ZXing-C++ package 具体可行性

核对 v3.1.1 源码，而非只依据 README 名称：

- [ZXingReader.cpp](https://github.com/zxing-cpp/zxing-cpp/blob/v3.1.1/example/ZXingReader.cpp) 提供限定格式、单码停止、JSON 输出、从 stdin 读取图片等选项，可以作为可控的解码进程。
- [ZXingWriter.cpp](https://github.com/zxing-cpp/zxing-cpp/blob/v3.1.1/example/ZXingWriter.cpp) 提供生成器选项、缩放/目标尺寸、静区、旋转和反色，能生成 PNG/JPEG/SVG。CLI 未见任意 RGB 前景/背景选项；可在 Go 标准库里把黑白 PNG 重新着色，再交 Reader 回读。具体纠错参数、UTF-8/ECI 和最小尺寸语义仍需样例实测，不应默认完全等价。
- [v3.1.1 展开的 release 资产](https://github.com/zxing-cpp/zxing-cpp/releases/expanded_assets/v3.1.1)仅列 `test_samples.tar.gz`、源码 tar.gz/zip 及 GitHub 自动源码包；**没有 Windows/macOS/Linux CLI 成品**。源码 zip 不能当可执行安装包加入 registry。
- [Homebrew 公式](https://formulae.brew.sh/formula/zxing-cpp)有二进制 bottle，并明确包含 Reader/Writer，但 bottle 仍有 Homebrew 平台和依赖约束，不等于可直接作为五个平台统一的便携 ZIP。
- [上游构建文档](https://github.com/zxing-cpp/zxing-cpp)要求 CMake 与 C++20；提供 vcpkg、Conan、MinGW、Homebrew/Linux 分发线索。直接 Go 绑定会把这些原生构建要求带入 fnsh 构建，CLI 则把工作转移到独立 package 构建发布。优先后者。

工程估计：在二进制已经可用的前提下，适配、着色和兼容性验证约 1–3 天；完整多平台构建、固定版本资产、校验、依赖打包与安装回归还需独立投入，不能作为一两处 import 修改处理。此处是工程判断，不是上游承诺或性能结论。

## ZBar 的分发细节

[维护仓库 README](https://github.com/mchehab/zbar)说明 release 工作流会产出 Linux、macOS 和 Windows 32/64 位二进制，托管在 LinuxTV；`zbarimg` 使用 ImageMagick/GraphicsMagick 读图。项目已有 ImageMagick package 不代表其 DLL/ABI 与 zbarimg 自动兼容。该[二进制目录](https://linuxtv.org/downloads/zbar/binaries/)本轮返回反自动化页面，未核实具体版本、架构、下载体积、DLL 或校验值，故不能写成“已确认可直接下载接入”。

如果采用 ZBar + boombuler/barcode，旧 gozxing/xerrors 两模块可被一个新 Go 编码模块与一个新运行时 package 替代；还需测试 Unicode、二维码检测差异及颜色回读。因此并不是一个模块换一个模块的低风险简化。

## 迁移前验收范围

继续推进 ZXing-C++ 时先做隔离实验：中文/emoji/换行/vCard、四级纠错、低对比色拒绝、长文本与最小尺寸、JPEG/GIF 输入、旋转与透视样本、无 QR 图片、进程超时取消、JSON 文本转义。保留现有输入限制、原子保存和回读承诺。通过后再按项目边界将执行实现放在 operation.Runner 后，并由共享应用路径调用；最后才删除 Go 模块。
