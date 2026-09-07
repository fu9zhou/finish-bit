# FinishBit

[![CI](https://github.com/fu9zhou/finish-bit/actions/workflows/ci.yml/badge.svg)](https://github.com/fu9zhou/finish-bit/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/fu9zhou/finish-bit?display_name=tag&sort=semver)](https://github.com/fu9zhou/finish-bit/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/fu9zhou/finish-bit.svg)](https://pkg.go.dev/github.com/fu9zhou/finish-bit)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](../LICENSE)

[English](../README.md) | 简体中文 | [完整文档](README.md)

![FinishBit — 多完成，少重写。](assets/finishbit-hero-zh-CN.png)

> **多完成，少重写。**

**用于 AI 任务的小型确定性能力单元。**

FinishBit 为 AI Agent 提供可复用的能力，覆盖 JSON 与文本转换、文件检查与校验、编码与时间工具，以及图片、PDF 和音视频处理。通过本地 CLI **`fnsh`**，Agent 可以发现能力、查看契约、执行操作并获得结构化结果。

**`fnsh` —— 搜索、复用、完成。**

```text
搜索 → 描述 → 执行 → 完成
```

## 核心理念

> **AI 会写，不代表每次都应该重写。**

常见任务不应每次都需要临时生成脚本。FinishBit 从一个简单原则出发：**用已有的确定性能力完成任务。** Agent 选择合适的能力并提供输入，由已有实现负责执行。没有合适的能力时，可以把新工具添加为可复用的 Operation。

**复用一个 Bit，完成一个任务。** Bit 表示一个小而专注的能力单元，用来完成任务中的一个步骤。在 FinishBit 的契约和文档中，这种能力统一称为 **Operation**：它具有可检索的标识符、强类型输入与选项，以及结构化结果和错误。

例如，`video.trim` 是 Agent 选择的 Operation，`ffmpeg` 是它的 Provider（能力提供方）。Operation 描述要完成的任务，Provider 提供具体实现。更大的任务可以使用多个 Operation。

确定性能力单元是设计目标：以行为明确的执行过程取代临时生成的任务逻辑。每个 Operation 的具体行为由其契约定义；UUID 生成器仍会产生新值，文件检查反映的则是文件的当前状态。

## 核心功能

| 领域 | 已提供的能力 |
| --- | --- |
| 能力发现 | 按自然语言意图搜索、查看强类型契约、以 JSON 列出完整能力目录 |
| JSON 与文本 | 格式化、压缩、校验和查询 JSON；统计、替换、排序和去重文本 |
| 文件与常用工具 | 查看文件信息、计算校验和与哈希、Base64/URL 编解码、时间转换、生成 UUID |
| 图片 | PNG/JPEG 基础处理；可选 ImageMagick 增强格式、编辑、拼图、透明处理、GIF 与图标 |
| 表格 | 0.1.3 的 CSV/JSON 基础能力，以及 27 项 qsv 清洗、关联、统计和校验能力 |
| PDF | 页面编辑、表单、附件、书签、文字提取与页面渲染，使用 pdfcpu 与 Poppler |
| 音视频处理 | 托管 FFmpeg/ffprobe 支持转码、拼接、字幕、预览及音频编辑，共 36 项 |
| 压缩包 | 10 项 [7-Zip 能力](archives.md)：8 种写入格式、加密、选择提取、修改、转换及分卷 |
| 文档 | 10 项 [Pandoc 文档能力](documents.md)：转换、合并、提取、拆分、模板与参考文献 |
| 工具箱扩充（v0.1.4） | 新增音视频、PDF、图片、表格、文档及压缩包共 144 项，目录达到 170 项；见[完整目录](operations.md)及[真实引擎验收](batch-b-acceptance.md) |
| 可读论文文本（v0.1.5） | `pdf.extract-text --unwrap` 将 PDF 视觉行还原为逻辑段落；`document.split` 默认不再插入硬折行 |
| 扩展能力 | 安装语言无关的本地扩展，通过协议 v1 发布新的 Operation |

核心 Operation 运行在单个 Go 二进制中，无须常驻服务；大型外部运行时只在需要时安装。`--json` 为 Agent 和自动化提供稳定的机器可读结果。

结构化调用可使用 `fnsh run --request request.json --json`，详见[CLI 文档](cli-reference.md)。新增图片与表格能力的格式、限制和示例见[使用说明](images-and-tables.md)。

新增能力已包含在 v0.1.4 中。完整真实引擎验收目前覆盖 Windows x64；Poppler、qsv、Pandoc 和 7-Zip 仅注册 Windows x64，ImageMagick 注册 Windows x64/arm64。按 `fnsh describe <id> --json` 查看依赖，用 `fnsh pkg add <package>` 按需安装。

[表格指南](tables.md)包含新增 27 项能力的契约及可重复执行的验收脚本。官网能力目录由应用注册表生成，CI 会检查是否与代码同步。

## 项目状态

FinishBit 当前沿 `v0.1.x` 系列迭代，当前版本为 `v0.1.5`，随后按 `v0.1.6` 等逐次递增。新增功能或依赖组不自动改变版本系列。核心运行时、`fnsh` CLI、托管音视频/PDF/图片/表格/文档/压缩包 Provider、本地扩展协议 v1 和 Agent Skill 已经可用。生产环境依赖首个稳定版本前的接口前，请阅读[兼容性政策](compatibility.md)。

## 为什么选择 FinishBit

- **任务导向：** 调用 `video.trim` 或 `json.format`，而不是要求 Agent 理解底层实现命令。
- **渐进式检索：** 默认只把最相关的少量能力放入上下文，确认后再读取完整契约。
- **确定性内核：** 轻量能力使用经过测试的 Go 实现，并提供结构化结果和错误。
- **依赖托管：** FFmpeg 等大型运行时按需下载到私有目录，来源、版本和 SHA-256 由 Runtime 固定。
- **语言无关扩展：** 任意可执行程序均可通过版本化本地 JSON 协议提供新 Operation。
- **统一应用层：** CLI 与未来 Web/API 共享 `pkg/app`，不会复制业务逻辑。

## 快速开始

从源码安装需要 Go 1.25.13 或更高版本；官方归档使用 Go 1.26.8 构建：

```bash
go install github.com/fu9zhou/finish-bit/cmd/fnsh@latest
```

搜索、确认并执行能力：

```bash
fnsh search "裁剪并压缩视频"
fnsh describe video.trim
fnsh json format data.json --indent 2
```

媒体 Operation 首次使用前需安装由 FinishBit 管理的 FFmpeg：

```bash
fnsh pkg add ffmpeg
fnsh video trim input.mp4 --start 10s --duration 20s -o clip.mp4
```

Agent 和自动化程序应使用 `--json`。成功 JSON 写入 stdout，结构化错误写入 stderr。

版本归档、安装脚本、PATH 配置、升级和卸载方法参见[安装指南](installation.md)。

## Agent Skill

仓库内的 `skills/finishbit` 指导兼容的 Agent 先发现 Operation，再执行操作，并优先使用结构化输出。将该目录复制或链接到 Agent 宿主支持的 Skill 目录即可使用。

Skill 是适配层，不是第二套运行时：所有能力仍通过 `fnsh` 和同一个应用服务执行。

## 扩展

一个扩展是包含 `finishbit-extension.json` 清单和一个可执行程序的目录或 ZIP 文件：

```bash
fnsh ext add ./my-extension
fnsh search "my capability"
```

请从[示例扩展](../examples/extensions/echo/README.md)开始，并阅读[扩展协议](extension-protocol.md)及其 [JSON Schema](../schemas/extension-v1.schema.json)。

扩展作为本地第三方程序，以当前用户权限执行。安装前请审查其源代码和来源。

## 文档

- [文档索引](README.md)
- [安装与升级](installation.md)
- [CLI 参考](cli-reference.md)
- [Operation 目录与契约](operations.md)
- [架构](architecture.md)
- [扩展协议](extension-protocol.md)
- [开发指南](development.md)
- [安全政策](../SECURITY.md)
- [路线图](roadmap.md)

## 贡献与支持

提议新能力或提交 Pull Request 前，请阅读[贡献指南](../CONTRIBUTING.md)。可复现的问题和范围明确的提案请提交到 [GitHub Issues](https://github.com/fu9zhou/finish-bit/issues)；支持方式及安全边界参见[支持指南](../SUPPORT.md)。

项目决策遵循[治理规则](../GOVERNANCE.md)，社区参与遵循[行为准则](../CODE_OF_CONDUCT.md)。

## 许可证

FinishBit 采用 [GNU AGPL v3.0](../LICENSE)。许可证允许商业使用，但分发修改版本以及符合 AGPL 条件的网络使用必须履行对应源码义务。本段仅为摘要，不构成法律意见，具体以许可证原文为准。

托管的第三方工具和用户安装的扩展保留各自许可证。
