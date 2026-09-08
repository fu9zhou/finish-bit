# Go 依赖热度与维护状态核查

后续状态：以下是移除依赖前的盘点快照。同日后续已用固定历表替代 lunar-go，当前为 8 个直接、11 个间接模块；实现与验证结果见 [减依赖评估](dependency-removal-evaluation-2026-09-08.md)。

核查日期：2026-09-08。范围是当前 `go.mod` 的 9 个直接、11 个间接模块；`go list -deps ./cmd/fnsh` 确认这 20 个模块均进入 CLI 的包依赖闭包。没有修改代码、依赖或锁定版本。本报告是维护健康检查，不是漏洞扫描或源码安全审计。

## 结论

整体可用，不能笼统称为“全部活跃”。go-toml、goldmark、xz、sevenzip 近期有版本更新；cron 很流行但多年缺少提交，gozxing 发布节奏明显落后于主干；sevenzip 的辅助依赖包含极小众项目，需要关注维护者集中度。YAML 已使用官方 YAML 组织维护的后继仓库，但我们锁定的 v3 属于仅接收安全修复的旧版本线。

Stars 只表示 GitHub 关注热度，不是安装量、质量或安全保证，Go 官方镜像和细分格式库尤其不能只看星数。以下星数取官方仓库页面显示值，`k` 为页面近似数。所有 20 个仓库页面均未显示归档提示；这不代表维护者承诺支持。GitHub API 本次受未认证速率限制，改用官方仓库网页和 Go 官方模块代理。

“最新版本日期”统一使用 Go proxy `@latest` 的 `Time`（版本对应提交时间），不是 GitHub Release 公告发布时间；只对当前 module 路径比较，不跨 major 版本。这也避免把“多年没打 tag”等同于“多年没提交”。伪版本行标明提交日期。Go proxy 对版本查询的定义见 [Go Modules Reference](https://go.dev/ref/mod#version-queries)。

## 9 个直接依赖

用途依据本地 `internal/localtools/calendar.go`、`formats.go`、`images.go` 与 `pkg/packagemanager/archive.go` 的 import 和调用。

| 模块 / 当前版本 | 项目用途 | GitHub stars | 最新同路径版本及日期 | 判断 |
|---|---|---:|---|---|
| [6tail/lunar-go](https://github.com/6tail/lunar-go) v1.4.6 | 公历与农历转换 | 337 | [v1.4.6 · 2025-11-05](https://proxy.golang.org/github.com/6tail/lunar-go/@latest) | 细分领域小众，近一年有更新，可保留 |
| [bodgit/sevenzip](https://github.com/bodgit/sevenzip) v1.6.5 | 解压托管工具的 7z 安装包 | 250 | [v1.6.5 · 2026-07-10](https://proxy.golang.org/github.com/bodgit/sevenzip/@latest) | 小众但近期有发布；辅助依赖较多，值得集中管理 |
| [makiuchi-d/gozxing](https://github.com/makiuchi-d/gozxing) v0.1.1 | 二维码生成与识别 | 668 | [v0.1.1 · 2021-10-05](https://proxy.golang.org/github.com/makiuchi-d/gozxing/@latest) | 需要关注发布滞后，不能判为彻底停维护 |
| [mozillazg/go-pinyin](https://github.com/mozillazg/go-pinyin) v0.21.0 | 汉字转拼音 | 1.8k | [v0.21.0 · 2025-07-19](https://proxy.golang.org/github.com/mozillazg/go-pinyin/@latest) | 细分领域有一定热度，更新频率不高，可保留 |
| [pelletier/go-toml/v2](https://github.com/pelletier/go-toml) v2.4.3 | TOML 校验、格式化及 JSON 转换 | 2.0k | [v2.4.3 · 2026-07-05](https://proxy.golang.org/github.com/pelletier/go-toml/v2/@latest) | 热度与近期发布均可接受 |
| [robfig/cron/v3](https://github.com/robfig/cron) v3.0.1 | cron.next 表达式解析及未来时间计算 | 14.2k | [v3.0.1 · 2020-01-04](https://proxy.golang.org/github.com/robfig/cron/v3/@latest) | 很流行，但明显缺乏近期开发活动；保留并观察 |
| [ulikunitz/xz](https://github.com/ulikunitz/xz) v0.5.16 | 解压安装包中的 XZ 数据 | 563 | [v0.5.16 · 2026-07-20](https://proxy.golang.org/github.com/ulikunitz/xz/@latest) | Go 专用格式库，近期有更新，可保留 |
| [yuin/goldmark](https://github.com/yuin/goldmark) v1.8.6 | Markdown / GFM 转换 | 5.0k | [v1.8.6 · 2026-09-03](https://proxy.golang.org/github.com/yuin/goldmark/@latest) | 热度较高，更新活跃 |
| [go.yaml.in/yaml/v3](https://github.com/yaml/go-yaml) v3.0.5 | YAML 解析与序列化 | 约 521–522 | [v3.0.5 · 2026-07-26](https://proxy.golang.org/go.yaml.in/yaml/v3/@latest) | 上游项目活跃，但 v3 只接受安全修复；后续评估 v4 |

9 个直接依赖的当前版本均与本次对应 module 路径的 `@latest` 相同，但这不意味着包含主干上的所有改动，也不意味着是最新 major 版本。

## 11 个间接依赖

`go mod graph` 与 sevenzip 源码确认：除 xerrors 由 gozxing 引入外，其余 10 个都出现在 sevenzip 的依赖声明中；x/text 同时也被 gozxing 和 afero 引用。sevenzip 用 go4.org 的 readerutil、afero 文件抽象、plumbing I/O 辅助、windows 元数据、LRU 密钥缓存及多种解压器。来源：[sevenzip v1.6.5 go.mod](https://github.com/bodgit/sevenzip/blob/v1.6.5/go.mod)、[reader.go](https://github.com/bodgit/sevenzip/blob/v1.6.5/reader.go)、[gozxing v0.1.1 go.mod](https://github.com/makiuchi-d/gozxing/blob/v0.1.1/go.mod)。

| 模块 / 当前版本 | 用途 | Stars | 最新同路径版本及日期 | 判断 |
|---|---|---:|---|---|
| [andybalholm/brotli](https://github.com/andybalholm/brotli) v1.2.2 | Brotli 解压 | 733 | [v1.2.3 · 2026-08-27](https://proxy.golang.org/github.com/andybalholm/brotli/@latest) | 近期更新，当前落后一个补丁版本 |
| [bodgit/plumbing](https://github.com/bodgit/plumbing) v1.3.0 | I/O 工具 | 7 | [v1.3.0 · 2022-11-18](https://proxy.golang.org/github.com/bodgit/plumbing/@latest) | 极小众、低频；2024-05-03 最后可见提交是 CI 更新 |
| [bodgit/windows](https://github.com/bodgit/windows) v1.0.1 | Windows 格式元数据 | 2 | [v1.0.1 · 2022-09-14](https://proxy.golang.org/github.com/bodgit/windows/@latest) | 极小众、低频；2024-05-03 最后可见提交是 CI 更新 |
| [hashicorp/golang-lru/v2](https://github.com/hashicorp/golang-lru) v2.0.7 | 缓存 | 5.1k | [v2.0.7 · 2023-09-21](https://proxy.golang.org/github.com/hashicorp/golang-lru/v2/@latest) | 有较大热度，成熟小型组件，低发布频率不能单独判停维护 |
| [klauspost/compress](https://github.com/klauspost/compress) v1.19.0 | 多格式解压，包括 Zstd | 5.6k | [v1.20.0 · 2026-09-02](https://proxy.golang.org/github.com/klauspost/compress/@latest) | 活跃，有新 minor 版本 |
| [pierrec/lz4/v4](https://github.com/pierrec/lz4) v4.1.27 | LZ4 解压 | 970 | [v4.1.29 · 2026-08-18](https://proxy.golang.org/github.com/pierrec/lz4/v4/@latest) | 近期更新，当前落后两个补丁版本 |
| [spf13/afero](https://github.com/spf13/afero) v1.15.0 | 文件系统抽象 | 6.7k | [v1.15.0 · 2025-09-08](https://proxy.golang.org/github.com/spf13/afero/@latest) | 热度较高，约一年没有新版本，保留观察即可 |
| [stangelandcl/ppmd](https://github.com/stangelandcl/ppmd) v0.1.1 | 7z 的 PPMd 解压 | 0 | [v0.1.1 · 2026-05-23](https://proxy.golang.org/github.com/stangelandcl/ppmd/@latest) | 极小众、近期有版本；应关注维护者集中度和格式样本覆盖 |
| [go4.org](https://github.com/go4org/go4) 20260112-a5071408f32f | readerutil | 332 | [同当前伪版本 · 提交 2026-01-12](https://proxy.golang.org/go4.org/@latest) | 低热度工具集合，年内有提交；不要将其全部模块依赖都误计为 CLI 实际导入 |
| [golang.org/x/text](https://github.com/golang/text) v0.40.0 | 编码、Unicode 文本转换 | 808（镜像） | [v0.41.0 · 2026-08-11](https://proxy.golang.org/golang.org/x/text/@latest) | Go 官方扩展包，近期更新；镜像星数不是生态使用规模 |
| [golang.org/x/xerrors](https://github.com/golang/xerrors) 20200804-5ec99f83aff1 | gozxing 的错误包装兼容层 | 279（镜像） | [20240903-7835f813f4da · 提交 2024-09-03](https://proxy.golang.org/golang.org/x/xerrors/@latest) | 官方历史过渡包，当前伪版本较旧；随 gozxing 一起评估 |

## 优先关注的细节

1. **cron：流行与活跃度明显分离。** [主干提交历史](https://github.com/robfig/cron/commits/master/)最后可见提交是 2021-01-06，最新稳定 tag 是 2020 年。不能称它仍在积极维护，也不能仅凭无提交断言已正式放弃。FinishBit 只使用表达式解析与 `Next` 计算，没有启动任务调度器，现有使用面很窄。
2. **gozxing：tag 旧，但主干并未完全停止。** [提交历史](https://github.com/makiuchi-d/gozxing/commits/master/)显示 2025-07-20 合入矩阵并发访问修复，2023 年也有性能改动；`@latest` 仍是 2021 年的 v0.1.1。我们当前代码每次新建 bitmap 并读取，不共享 bitmap；本核查不将该上游并发修复认定为项目已触发的 bug。应跟踪正式发布与真实识别效果，不要机械升级到任意主干提交。
3. **YAML：不是随意换了一个低星 fork。** [官方 README 的 Project Status / Version Intentions](https://github.com/yaml/go-yaml#project-status)说明，原作者 2025 年 4 月标记旧仓库不再维护后，由 YAML 官方组织接手。v1/v2/v3 为冻结旧版本，只接收安全修复；新功能与常规修复在 v4。我们已接到正确的维护方，接下来是评估版本线，而不是因为新仓库星少回退旧仓库。
4. **sevenzip：真正小众的部分在传递依赖。** plumbing、windows 和 ppmd 热度极低；[plumbing 提交历史](https://github.com/bodgit/plumbing/commits/main/)与 [windows 提交历史](https://github.com/bodgit/windows/commits/main/)显示最后可见更新是 2024 年 CI 维护。它们是窄职责组件，而 sevenzip 自身 2026 年还在发布，因此不能简单判整条链失效。建议将这条链作为一个单元跟进，关注安装包格式覆盖、上游维护响应和升级回归。
5. **xerrors：旧不等于来历不明。** [官方 README](https://github.com/golang/xerrors)明确它是 Go 1.13 新错误机制的过渡包。当前由 gozxing 引入，不应为了清理日期直接在主模块强行删除；随着上游去掉兼容层再减少依赖更合理。

## 建议

当前没有仅凭流行度就必须替换的直接依赖。把 cron、gozxing 的发布/维护情况及 sevenzip 的极小众子依赖列入观察清单；把 YAML v4 列为可独立验证的升级事项。另有 brotli、compress、lz4、x/text、xerrors 新版本可评估，但“存在新版本”不等于“当前有漏洞”，本次没有执行升级或漏洞扫描。

本报告只新增研究文档，没有代码变更，因此未运行代码变更要求的 test/vet/build；执行了只读的模块图、CLI 包依赖及源码调用核对。没有用未获取到的 GitHub API 字段推算任何星数或日期。
