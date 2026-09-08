# 7-Zip 托管解包落地与验收

日期：2026-09-08。用户确认继续复用已下载的 7-Zip，并排除需要我们构建、重打包或分发的方案。

## 现有能力与缺口

已有 `7zip`（7-Zip Extra）供归档操作使用，`7zip-full` 负责把 Tesseract NSIS 当作归档读取。缺口是这两个 package 自己分别从 `.7z` 和包含 7z 数据的安装器解出，此前必须先调用 Go sevenzip。

新增的 `7zip-bootstrap` 直接获取官方 26.03 原始 `7zr.exe`，无需任何解压器即可安装。它不是我们编译或托管的产物：

- [官方固定版本资产](https://github.com/ip7z/7zip/releases/download/26.03/7zr.exe)
- [官方资产清单及 SHA-256](https://github.com/ip7z/7zip/releases/expanded_assets/26.03)
- 固定 SHA-256：`ad4c82fadcbdf93c03b4fc440f300509c7d60c5c2f4d183e35d9d70d6957037d`，实际下载后匹配。
- 发布页显示约 589 KB；上游声明为 x86 控制台程序。

## 已实现

7z 安装解包优先复用当前版本、当前平台且可定位的 `7zip-full`、`7zip`、`7zip-bootstrap`，跳过正在安装/修复的包自身。没有现成读者时自动安装 raw bootstrap。Tesseract 的 NSIS 分支继续由完整 7-Zip 处理，不执行安装器。

实现位于 [sevenzip.go](../../pkg/packagemanager/sevenzip.go)：先读取有输出上限的技术列表，验证路径、重复/大小写冲突、父目录冲突、链接/特殊类型、加密、条目数与合计大小；再逐文件以字面量、区分大小写且不递归的选择器读取 stdout。Go 自己创建目标路径、使用排他写入、把数据流限制到声明大小，并验证进程成功和最终文件大小。因此新 7z 路径没有退化为“外部进程先随意落盘，再事后检查”。

逐文件读取可能重复解压 solid block，安装耗时比单次展开更多；好处是保持运行期间的写入上限。10 分钟的解包超时不包含此前的 bootstrap 下载。ZIP/tar.gz 继续标准库，tar.xz 继续 xz。

`go mod tidy` 后，**go.mod 中 require 列出的模块从 19 个变成 9 个**；`go list -deps ./cmd/fnsh` 的模块去重结果也只有这 9 个外部模块。移除：sevenzip、brotli、plumbing、windows、golang-lru/v2、klauspost/compress、lz4/v4、afero、ppmd、go4。XZ 仍直接使用；x/text 与 xerrors 仍由 gozxing 引用。

这里不是说 `go list -m all` 只有 9 行：它还会显示依赖模块图中的 x/mod、x/sync、x/tools，它们不在 fnsh 的实际包导入闭包内。不要混淆声明的模块、工具依赖图和编入程序的模块数量。

## 实际验证

Windows x64，独立临时 FINISHBIT_HOME，没有改动用户默认已安装工具：

1. 从空 package/cache 开始安装 `7zip-full`，自动下载 raw bootstrap 并成功解出 `7z.exe`、`7z.dll`、`License.txt`。
2. 复用已安装 full reader 安装 ImageMagick，使用已验证缓存，耗时约 11.42 秒。
3. 另一个没有预装任何 package 的 home，仅提供已验证下载缓存，自动引导安装 ImageMagick，耗时约 7.13 秒。这是不同读者/运行条件的观察值，不是严格性能对比，也不包含网络下载。
4. 安装普通 `7zip` 与 Tesseract 5.5.3 成功，覆盖 Extra 和 NSIS 调用链；随后 `pkg repair 7zip-full` 重新下载与激活成功，复用其他托管 reader 而不调用被修复的包自身。
5. 新旧解包产物逐文件 SHA-256 比较：完整 7-Zip 3 文件、ImageMagick 23 文件、Tesseract 与模型 141 文件，全部一致。
6. `go test ./...` 通过，同时启用 `FINISHBIT_TEST_7ZIP_HOME` 和 `FINISHBIT_TEST_LOCAL_HOME`，包括真实 stdout 解包以及中英文 OCR、PDF 和文档工作流；`go vet ./...`、`go build ./cmd/fnsh` 通过。

[新增测试](../../pkg/packagemanager/sevenzip_test.go) 覆盖列表校验、50000 条目上限、1 GiB 上限、已有 reader 复用和引导防循环；实际进程测试覆盖 solid archive、中文/空格/前导 @ 与 - 文件名、同名根文件和嵌套文件、空文件/目录、精确资源选择、缺失资源、禁止覆盖与取消。实际进程测试默认跳过，需要指定已安装工具 home，不在普通单元测试中下载运行时。

## 平台和剩余范围

- Windows arm64 的 ImageMagick 登记保留，bootstrap 使用同一官方 x86 原始文件，依赖 [Windows 自带 x86 仿真](https://learn.microsoft.com/en-us/windows/arm/apps-on-arm-x86-emulation)。本轮未获得 arm64 真机，不能声称完成本机运行验收。
- 不新增 ZXing-C++ 自建二进制，不维护 pdfcpu 重新打包。QR/XZ 保留。
- [goccy YAML 实验](yaml-compatibility-2026-09-08.md) 发现转换和格式化需要额外兼容层，且模块数量不变，本轮不迁移。
