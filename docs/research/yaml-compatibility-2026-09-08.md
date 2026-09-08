# goccy/go-yaml 兼容评估（2026-09-08）

## 结论

goccy/go-yaml v1.19.2 可以作为纯 Go 候选，不需要我们构建或分发 CLI，也没有第三方 Go 模块依赖。但它不是现有 go.yaml.in/yaml/v3 的直接替代品。**本轮建议保留当前 YAML 依赖，不直接迁移。** 收益是社区规模更大，而非减少模块数量；适配复杂度超出了低成本依赖清理。

这里的结论基于实际实验：32 个定向 YAML 用例中，使用 AST 原始 token 加适配桥后的转换/拒绝行为匹配现有实现 29 个；其余 3 个涉及日期键、非特定标签和 directive 文档计数。格式化还有独立的注释及缩进问题，未完成生产级适配。5 个 JSON→YAML 样例证明，自定义数字编码可保留大数和长小数。上述样例不是完整 YAML 规范测试，也不是库质量排名。

版本依据：[官方 v1.19.2 release](https://github.com/goccy/go-yaml/releases/tag/v1.19.2)；[该版本 go.mod](https://github.com/goccy/go-yaml/blob/v1.19.2/go.mod) 只有 module 和 Go 1.21.0，没有 require。替换是一进一出，Go 模块数量净变化为 0。

## 当前项目契约

来源：[formats.go](../../internal/localtools/formats.go)，实际仓库位置为 internal/localtools/formats.go；既有回归样例在 internal/localtools/localtools_test.go。本评估复制现有 yamlValue、jsonYAML、decimal，仅把 invalid 错误包装简化为普通错误。

- yaml.to-json / yaml.validate / yaml.format：只接受一个文档；拒绝 alias，但允许没有被引用的 anchor；键必须解析为字符串且唯一；只接受指定标签；限制深度 128 和节点预算 100000。
- 转换从 scalar 文本生成 json.Number，不能经过 float64；timestamp 值保持原文字符串。
- yaml.format 使用原 AST，保留注释，输出两空格缩进。格式化前同样执行上述验证。
- json.to-yaml 使用 JSON UseNumber 的结果，对数字写显式类型语义，按键排序。
- 细节以代码为准：当前 012 实际转 JSON 为十进制 12，并非拒绝；0o12 和 0x10 被拒绝。迁移不能不经说明就把 012 改成 10。

## 实测差异

实验使用固定版本官方模块，对照现有代码，不修改主仓 go.mod。

| 输入/行为 | 当前实现 | goccy 现成 API | 可行适配 |
|---|---|---|---|
| 9007199254740993 | 精确数字 | 精确，采用 uint64 | 无需特殊修复，但不能推断任意数字都安全 |
| 18446744073709551616 | 精确 JSON 数字 | Unmarshal / YAMLToJSON 变字符串，AST 为 StringNode | 读取未加引号的 token，按当前解析规则补充数字判定 |
| 0.12345678901234567890123456789 | 保留全部位数 | YAMLToJSON 变 0.12345678901234568 | AST FloatNode.Token.Value 保留原文；不要读取 float64 Value |
| !!float 1e309 | 精确 json.Number | Unmarshal 变 +Inf；YAMLToJSON 输出 .inf | 显式标签加原 token，经当前 decimal 验证 |
| 普通 1e309 | 字符串 | 字符串 | 不能把所有数字样式 StringNode 一概改成数字 |
| TRUE、带引号 true、yes、null | 保留类型区别 | 样例一致 | 保留 token 引号类型及显式 tag |
| !!timestamp 2026-09-08 | 原文字符串 | YAMLToJSON 改成 2026-09-08T00:00:00Z | 用 tag 内容原文，不反序列化为 time.Time |
| 重复 a 或 a/带引号 a | 拒绝 | 默认 parser 拒绝 | 仍检查归一后的重复键值 |
| alias / 循环 alias | 拒绝 | 默认展开；循环样例变 null | AST 在反序列化前拒绝 AliasNode |
| 两个文档 | 拒绝 | Unmarshal / YAMLToJSON 只读第一个 | 检查实际文档数，考虑 directive 特例 |
| .nan、0x10、0o12 | 拒绝 | YAMLToJSON 接受；.nan 输出甚至不是合法 JSON | 复用当前数字和标签白名单 |
| 数字键 1: hi | 拒绝 | 默认转成字符串键 | 在 AST 检查原始键类型 |
| 空输入 / 显式空文档 | 前者拒绝，后者 null | 都可解析为空 body | 检查文档开始标记 |
| yaml.format 注释+四空格输入 | 保留注释并改两空格 | ast.File.String 保留注释和原缩进 | 尚需独立处理格式化 |
| 同一 AST 传给 Encoder、Indent(2) | 保留注释，两空格 | 实测丢注释且仍保留内部原缩进 | 不是加一个 Indent 选项即可解决 |

证据来源：[token 数字解析](https://github.com/goccy/go-yaml/blob/v1.19.2/token/token.go)、[AST 字段及 String 实现](https://github.com/goccy/go-yaml/blob/v1.19.2/ast/ast.go)、[Encoder 实现](https://github.com/goccy/go-yaml/blob/v1.19.2/encode.go)、[转换辅助 API](https://github.com/goccy/go-yaml/blob/v1.19.2/yaml.go)。表中具体输出来自附带实验，不将依赖库默认行为视为 FinishBit 必须继承的行为。

## JSON→YAML 的解决方向已验证

默认 goccy.Marshal 将 json.Number(9007199254740993) 编码成带引号字符串；直接 JSONToYAML 会舍入长小数、将超 uint64 整数写成字符串。因此这两个入口都不能直接使用。

官方 [CustomMarshaler 编码选项](https://github.com/goccy/go-yaml/blob/v1.19.2/option.go) 可以定制 json.Number。实验输出数字时附加 !!int 或 !!float，以下 4 个数字及 1 个易混淆字符串对象都可经候选 AST 桥恢复原值：

- 9007199254740993
- 18446744073709551616
- 0.12345678901234567890123456789
- 1e309
- 字符串 true / 2026-09-08 / 123 / null / yes 组成的对象

仅将 CustomMarshaler 返回值设为数字原文仍不够：1e309 不带 !!float 就会在下一次读取时变成字符串。生产实现还需保留现有 JSON 深度/节点预算，不把校验职责转交默认 Marshal。

## AST 适配原型的边界

原型用候选 parser/AST 解析，再桥接为旧 yaml.Node 数据结构以复用当前 validator。**旧 parser 没有参与候选路径；但原型仍引用旧 Node 类型，因此不是已移除旧依赖的生产实现。** 迁移时应改成内部值遍历，完全去掉旧模块。

32 个 YAML 样例中 29 个匹配现有“是否接受+输出 JSON 值”，含深度 129 和 100001 元素拒绝样例。剩余 3 个已定位：

1. 日期键 2026-09-08: hi：旧库把键识别为 timestamp，当前 profile 拒绝；goccy 识别为 string，简单桥错误接受。需补全键的隐式 timestamp 判定。
2. v: ! 123：旧实现解析成数字 123；goccy 产生非特定 TagNode，简单桥按未知 tag 拒绝。需要对非特定标签单独处理。
3. %YAML 1.1 后跟一个文档：goccy parser 将 directive 放在单独 Docs 项里，简单的 len(Docs)==1 会错误拒绝。需要按 Document/Directive 结构计数并校验 directive。

这些差异有源码级适配路径，并非候选库无法实现需求。不过浮点溢出边界、不同时间戳格式、显式标签、多文档终止符、literal/folded/chomping、复杂注释和指令组合还需回归。简单桥比原实现更严格计数键节点，恰好在 100000 附近的行为也需要统一。

yaml.format 尚未证明可以通过少量 API 配置同时实现两空格缩进、注释保留、标量类型不变。直接 decode→encode 会丢失词法信息，直接 AST.String 不归一缩进；改 token 列位置还要考虑块字符串内容和注释位置。需要专门实现并测试，不能宣布全兼容。

## 成本判断与建议

估算依据是现有 AST→值 / JSON→AST 逻辑，以及本次原型暴露的边界，不是已完成工期承诺：

- 转换、校验和 JSON 编码适配：估计 200–350 行内部访问器、标签及数字逻辑，工作量中等。
- 保留注释的两空格格式化：另一个需要验证的子问题，目前没有可靠的几行配置方案；不包含在上述估算中。
- 至少保留 32 项定向样例、5 项 JSON 往返，再补充 directive/tag/格式化回归；迁移后运行仓库要求的 test/vet/build。
- 不需要 CGo、CLI 安装、二进制构建或自行分发；但也不会减少依赖数量。

建议本轮优先完成 7-Zip 的依赖链清理，YAML 保留。若后续明确以替换 YAML 维护方为目的，应安排独立迁移：先完成 formatter 原型，再决定是否投入完整兼容层。不能只为跨过 1k star 阈值扩大自行维护的解析语义。

## 复现实验

附带目录：[yaml-compatibility-experiment](yaml-compatibility-experiment)。Go 代码以 .go.txt 存放，不参与项目构建。reference.go.txt 是当前函数快照，bridge.go.txt 是有意不完整的 AST 桥，main.go.txt 包含全部输入、默认 API 对照和格式化输出。

从仓库根在 PowerShell 运行：

    & ./docs/research/yaml-compatibility-experiment/run.ps1

脚本创建独立临时 module，执行 go mod tidy、go run .，输出摘要和完整 JSONL 路径；不修改主仓模块。固定依赖为 goccy v1.19.2 和旧库 v3.0.5，后者仅作参照和临时 AST 数据形状。首次运行需要 Go 模块网络下载。

本次实际 go run . 退出码 0；摘要保存在 [results-summary.json](yaml-compatibility-experiment/results-summary.json)。本轮仅新增研究文档和隔离实验材料，未迁移生产代码或主仓依赖。
