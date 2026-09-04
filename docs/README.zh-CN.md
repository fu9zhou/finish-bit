# FinishBit

> 为完成 AI 任务而生的小型确定性能力单元。

FinishBit 是面向 AI Agent 的本地能力运行时。Agent 通过 `fnsh` 搜索、读取并执行已有 Operation，不再为常见任务反复生成 Python、Shell 或复杂的 FFmpeg 命令。

```text
搜索 → 描述 → 执行 → 完成
```

## 核心特点

- **任务导向：** 调用“裁剪视频”“格式化 JSON”，而不是要求 Agent 理解底层工具。
- **渐进式检索：** 默认只把最相关的少量能力放入上下文。
- **确定性内核：** 轻量能力使用经过测试的 Go 实现。
- **依赖托管：** FFmpeg 等大型依赖按需下载到 FinishBit 私有目录，版本和 SHA256 由 Runtime 固定。
- **社区扩展：** 任意语言均可通过版本化 JSON 进程协议提供新 Operation。
- **可扩展适配层：** CLI 和未来 Web/API 共用 `pkg/app`，不会复制业务逻辑。

## 快速开始

需要 Go 1.24 或更高版本：

```bash
go install github.com/fu9zhou/finish-bit/cmd/fnsh@latest

fnsh search "裁剪并压缩视频"
fnsh describe video.trim
fnsh pkg add ffmpeg
fnsh video trim input.mp4 --start 10s --duration 20s -o clip.mp4
```

Agent 和自动化程序应增加 `--json`，成功结果写入 stdout，结构化错误写入 stderr。

## 项目阶段

`v0.1.0` 是首个可运行版本。Operation ID 与 JSON 接口从第一版开始按兼容性原则维护，但 `v1.0.0` 前仍可能出现必要的协议调整。

详细设计参见：[架构](architecture.md)、[Operation](operations.md)、[扩展协议](extension-protocol.md)、[包管理](package-management.md)。

## 许可证

项目采用 [AGPL-3.0](../LICENSE)。允许商业使用和收费分发，但修改、衍生发布及符合 AGPL 条件的网络服务必须履行相应源码提供义务。
