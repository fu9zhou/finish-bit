# 低星依赖的替代库与托管 package 再检索

后续实施：7-Zip 整链已移除，go.mod 声明依赖 19 → 9，详见[实施验收](sevenzip-migration-2026-09-08.md)。[YAML 实测评估](yaml-compatibility-2026-09-08.md) 后决定暂留。以下保留迁移前检索记录；需要自行分发的后续方案已按用户要求排除。

核查日期：2026-09-08。基于当前移除 lunar-go 后的 19 个 Go 模块。本轮只新增研究报告，没有再次修改实现、go.mod 或 runtime registry。星数为同日 GitHub 可见约数，不能单独代表使用规模或可靠性；分发资产与实际项目用途分别核对。

## 建议顺序

1. **减依赖收益最大：官方 7-Zip package 接管 7z 解包。** 有可固定版本的裸 7zr.exe，可解决安装器自身也要解压的引导问题。预计移除 10 个 Go 模块，19 → 9。必须先验证引导、Windows ARM64 和当前解压边界，不能直接改成一次 `7z x`。
2. **改动范围更集中：评估 goccy/go-yaml 替换现有 YAML v3。** 约 2.2k 星，纯 Go，v1.19.2 的 go.mod 没有其他模块依赖；不减少模块总数，但能换到独立维护的稳定解析器。需适配 AST 并验证当前四项 YAML 能力的严格语义。
3. **二维码暂留 gozxing；ZXing-C++ 作为后续 package 候选。** 后者较活跃，覆盖生成和识别，但官方 release 没有现成 CLI 二进制。我们需要额外维护构建分发，暂不属于低成本替换。
4. **XZ 暂留。** 非 Windows pdfcpu 仍需解 tar.xz；官方 7-Zip 的非 Windows 包也使用 tar.xz。完全删除需改发行产物来源或维护二次打包。

以上为源码与分发层面评估，尚未运行候选兼容实验；模块收益为当前依赖图推算，正式迁移须用 tidy/list 验证。托管 package 会把部分依赖移到运行时，并不使上游代码或更新责任消失。

## 覆盖所有低于 1k 星的当前模块

当前 4 个直接、8 个间接模块低于 1k，基线星数见[健康报告](go-dependency-health-2026-09-08.md)。

| 当前模块 | 约星数 | 替代方向 | 本轮判断 |
|---|---:|---|---|
| bodgit/sevenzip | 250 | 官方 7-Zip 约 3.9k，托管 CLI | 值得优先验证，整链移除收益最大 |
| ulikunitz/xz | 563 | 7-Zip / libarchive CLI；上游改 tar.gz/ZIP | 引导和发行格式尚未解决，暂留 |
| andybalholm/brotli | 733 | google/brotli 约 14.9k | sevenzip 内部依赖；单独替换要 fork 和原生适配，跟随整链处理 |
| pierrec/lz4/v4 | 970 | lz4/lz4 约 12.1k | 同上，原生高星实现不能直接替换 Go import |
| stangelandcl/ppmd | 0 | 7-Zip 内置解码 | 跟随 sevenzip 整链移除 |
| bodgit/plumbing | 7 | 标准库/少量辅助实现 | 上游 sevenzip 引用，单独自写不能让原模块消失 |
| bodgit/windows | 2 | FILETIME 转换可自写 | 同上，无需为它单独接管 sevenzip fork |
| go4.org | 332 | 有限 reader 辅助可自写 | 同上，整链处理 |
| makiuchi-d/gozxing | 668 | ZXing-C++ 约 1.9k，托管 CLI | 候选合适，二进制分发尚缺 |
| golang.org/x/xerrors | 279 | 随 gozxing 退出 | 不能直接删除上游仍引用的模块 |
| golang.org/x/text | 808 | 随 sevenzip 和 gozxing 两条路径一起退出 | Go 官方项目，不因 GitHub 镜像低星而单独替换 |
| go.yaml.in/yaml/v3 | 约 522 | goccy/go-yaml 约 2.2k；yq 约 15.9k package | 优先考虑纯 Go goccy，yq 作为接受运行时依赖的备选 |

归档候选的源码、星数、发布与分发证据见[归档专项](archive-alternatives-2026-09-08.md)，二维码见[QR 专项](qr-alternatives-2026-09-08.md)。Go 官方 x/text 的仓库身份见 [golang/text](https://github.com/golang/text)。

## YAML：确实有可认真考虑的替代

| 候选 | 热度与维护证据 | 模块/部署代价 | 判断 |
|---|---|---|---|
| [goccy/go-yaml](https://github.com/goccy/go-yaml) | 约 2.2k；独立开发；稳定版 [v1.19.2](https://github.com/goccy/go-yaml/releases/tag/v1.19.2)，2026-01-08 | [固定版本 go.mod](https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/go.mod) 仅声明 module 和 Go 1.21，无 require；替换后一进一出 | 首选替代验证，继续纯 Go 内置运行 |
| [mikefarah/yq](https://github.com/mikefarah/yq) | 约 15.9k；[v4.53.6](https://github.com/mikefarah/yq/releases/tag/v4.53.6) 2026-08-20 发布 | 有独立 CLI，可移除本项目 YAML Go 模块，但增加下载、版本、进程与缓存管理 | 技术上可行；当前只有四项 YAML 能力，不是首选 |
| [go-yaml/yaml](https://github.com/go-yaml/yaml) | 约 7k，但已归档 | 回到旧上游 | 不因星多回退 |
| [kubernetes-sigs/yaml](https://github.com/kubernetes-sigs/yaml) | 355，未达到筛选线 | 包装 YAML 解析与 JSON 转换，并非独立的高星解析器 | 不推荐为此增加一层 |

现有 [yaml/go-yaml](https://github.com/yaml/go-yaml) 是 YAML 官方组织接管的维护仓库，不是毫无历史的新解析器。其 README 明确 v3 仅收安全修复，新特性与常规修复集中于 v4。因此保留当前库也有合理依据；goccy 是可比较的替代，并非仅凭星数已证明更可靠。

项目 [formats.go](../../internal/localtools/formats.go) 的四项能力为 yaml.to-json、json.to-yaml、yaml.format、yaml.validate。当前通过 yaml.Node 遍历实现：单文档、拒绝别名、唯一字符串键、128 层/100000 节点限制、JSON 数字精度和明确标量类型，并使用节点编码格式化。这决定了替换不能只改 import。

核对 goccy v1.19.2 的 [parser](https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/parser/parser.go) 与 [AST](https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/ast/ast.go)：提供文档列表、注释解析、映射/别名/数字等节点和原始 token，具备实现相同约束的入口；AST 类型不兼容现有 yaml.Node。验证要覆盖数字原文和大整数、显式标签、日期字符串、多行文本、重复键、别名、空/多文档、注释及格式化，以及极限嵌套。预计适配和回归约 1–3 天，属工程估计，不是兼容验收结论。

yq 的 [v4.53.6 实际资产清单](https://github.com/mikefarah/yq/releases/expanded_assets/v4.53.6) 已确认 Windows amd64/arm64、Linux amd64/arm64、macOS amd64/arm64 的独立二进制，并有压缩包及校验文件。它比 ZXing-C++ 更容易接入现有 package manager。但 [固定版本 go.mod](https://raw.githubusercontent.com/mikefarah/yq/v4.53.6/go.mod) 同时使用 goccy/go-yaml、go.yaml.in/yaml/v4 及多种格式库；这只是把解析依赖封装到 CLI，不是消除解析器。接入时固定表达式和输入通道，仍需证明上述严格语义及数值精度；不能假定普通 YAML→JSON 命令与现有契约相同。

## 哪些高星候选没有实际收益

- mholt/archiver 约 4.4k，但已归档；继任 archives 441 星，7z 实现仍直接引用 bodgit/sevenzip，还带入更多格式依赖。
- skip2/go-qrcode 约 3k，但只生成，主干停在 2020；boombuler/barcode 约 1.6k 也只生成，不能替代现有识别和生成后回读。
- GoCV 约 7.5k，但增加 CGo/OpenCV 构建及运行时，为目前 QR 三项能力成本过高。
- libarchive 约 3.6k，官方最新 release 的 ZIP 等资产是源码；不能直接作为跨平台便携 CLI 接入。

上述证据分别收录于两份专项报告，已检查底层 imports 和 release 资产，避免把包装库的星数或源码 ZIP 当成实际可部署替代。

## 可量化的阶段结果

| 阶段 | Go 模块预计数量 | 增加的维护职责 |
|---|---:|---|
| 当前 | 19 | 现有 managed packages |
| sevenzip 整链改官方 CLI，XZ 保留 | 9 | 裸引导工具、完整 7-Zip 安装依赖与解压约束 |
| 再将 YAML 改 goccy | 9 | YAML AST 适配；不新增 runtime |
| 再将 QR 完整改 ZXing-C++ CLI | 6 | 自建多平台 Reader/Writer；此时 x/text、xerrors 才都能退出 |
| 再解决 XZ 发行格式 | 5 | 上游替代产物或我们维护可审计的重打包发布 |

最后两阶段投入明显增加，不建议为数字好看一次推进。实际优先级是 7-Zip 的整链收益与 goccy 的集中适配；QR 和 XZ 留作后续有分发方案时再迁移。
