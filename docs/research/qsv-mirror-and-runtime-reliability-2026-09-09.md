# qsv 镜像与运行时可靠性修复

2026-09-09，Windows x64。本轮基于 v0.1.7 验收发现修复；未更换 qsv 版本、未自行重打包运行时。

## 镜像核查

官方来源：[dathere/qsv 22.0.1](https://github.com/dathere/qsv/releases/tag/22.0.1)。保持 Windows x64 MSVC 版本与原始归档，不重新打包，也不替换为 lite 版本。

| 备用来源 | 完整下载字节数 | 固定 SHA-256 校验 |
| --- | ---: | --- |
| [GH-Proxy](https://gh-proxy.com/en) | 319,028,376 | 通过 |
| [GHFast](https://ghfast.top/) | 319,028,376 | 通过 |
| [SourceForge qsv.mirror](https://sourceforge.net/projects/qsv.mirror/files/22.0.1/) | 319,028,376 | 通过 |

三个完整文件的 SHA-256 均为：

`9d95f7ae939b96124215ba9e34e18f067c5743f5961ce4caad32c164153290d7`

固定下载地址：

- `https://gh-proxy.com/https://github.com/dathere/qsv/releases/download/22.0.1/qsv-22.0.1-x86_64-pc-windows-msvc.zip`
- `https://ghfast.top/https://github.com/dathere/qsv/releases/download/22.0.1/qsv-22.0.1-x86_64-pc-windows-msvc.zip`
- `https://downloads.sourceforge.net/project/qsv.mirror/22.0.1/qsv-22.0.1-x86_64-pc-windows-msvc.zip`

本次分别从每个来源下载全部分段，检查 HTTP Content-Range、分段长度，按偏移合并后验证整个文件；未用已有官方缓存填补缺失字节。SourceForge 先前只收到 5,130,012 字节的限制，已由这次完整下载验证补足。过程包含并发、重试和断点恢复，不应把最后一次恢复的耗时当作完整下载测速。

运行验证使用各自新下载且通过摘要校验的归档，放入三个隔离安装目录的校验缓存，经 `fnsh pkg add qsv` 安装后运行 qsv 与表格集成测试。三个来源均成功安装，分别通过 32 项表格测试（27 个功能与 5 组异常/边界场景），且使用最终构建的 `fnsh doctor --json` 检查均通过。`go test ./...`、`go vet ./...`、`go build ./cmd/fnsh` 均通过。该流程验证镜像字节的安装和运行；网络分段下载由验收脚本完成，产品本身仍顺序下载，不具备跨命令断点续传。产品来源切换另有停滞、错误摘要和调用方取消的自动化测试。

注册顺序为 GitHub 官方 → GH-Proxy → GHFast → SourceForge。前两个为第三方 GitHub 代理，SourceForge 声明不隶属 qsv；各来源仍须通过相同固定摘要。其他探测节点中，ghproxy.net 只完成范围探测，Fastly/EdgeOne 与 SourceForge 其他直连节点出现超时或速度更低，未加入配置。镜像只提供备用传输路径，本次结果不保证日后的速度或可用性。

完整归档、各来源摘要和测试日志保存在本地 `.finishbit-test-fixes/mirror-full` 及 `.finishbit-test-fixes/*-table-tests.jsonl`。

## 修复与验证

- 安装记录全部文件摘要；doctor 校验 exe、DLL、OCR 模型及元数据，检查登记包和引导/解包辅助包。
- 旧安装缺少摘要时报告无法验证，提示 `pkg add` 从校验缓存或来源重装；不把现有可能损坏的文件作为可信基准。
- `pkg add` 从校验缓存恢复损坏安装；`pkg repair` 保持强制新下载语义。
- Windows 扩展卸载对访问拒绝、共享冲突、锁冲突有限重试；无关错误立即返回，持续失败保留最终错误。
- 下载无字节进展 45 秒才切换来源，移除会截断持续传输大文件的 30 分钟总时限；保留调用方取消/截止时间、30 秒响应头超时和 1 GiB 大小上限。错误摘要也可以尝试下一固定镜像。
- 人类输出增加 MiB、百分比、缓存命中和阶段提示；JSON 输出保持纯净。未实现跨进程断点续传。

回归覆盖：doctor 四种原有失败场景；缓存恢复和旧元数据迁移；Windows 实际短暂文件锁和重试上限；持续下载、停滞回退、错误摘要回退；进度事件与 JSON 隔离。真实依赖、CLI 故障注入及 Go test/vet/build 证据保存在本地 `.finishbit-test-fixes`。
