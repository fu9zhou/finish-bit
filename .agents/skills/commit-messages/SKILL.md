---
name: commit-messages
description: Enforce Conventional Commits with an English type and module, a Chinese summary, and a detailed Chinese bullet-list body. Use whenever creating, amending, rewording, squashing, merging, reverting, or otherwise rewriting a Git commit.
---

# Commit Messages

Create every Git commit with a predictable machine-readable header and a detailed human-readable Chinese body.

## Format

Use a Conventional Commits header followed by a blank line and detailed Chinese bullets:

```text
<type>(<module>): <中文摘要>

- <具体改动一>
- <具体改动二>
```

Omit the entire module segment only when the change genuinely has no identifiable primary module:

```text
<type>: <中文摘要>

- <具体改动>
```

Never emit an empty `()`.

## Rules

1. Use an English lowercase Conventional Commits type: `feat`, `fix`, `refactor`, `test`, `docs`, `perf`, `build`, `ci`, `chore`, or `revert`.
2. Use the repository's English lowercase module identifier. Join multiple words with hyphens, such as `code-graph`, `agent-runner`, or `commit-rules`.
3. Write the summary in Chinese as a concise action statement with at least one Han character and no trailing period.
4. Add a Chinese bullet-list body after a blank line. Include at least one concrete bullet even for a small commit.
5. Describe verifiable behavior, implementation, validation, and impact. Replace vague statements such as “调整代码” or “更新内容” with concrete outcomes.
6. Keep one primary purpose per commit. Split unrelated changes.
7. Override Git-generated English merge, squash, or revert messages before committing.

## Examples

```text
feat(requirements): 支持图片辅助澄清客户需求

- 允许客户上传截图、草图和参考图片
- 将图片识别结果转化为待确认的业务选项
- 在需求包中记录客户确认后的多模态推断
```

```text
fix(release): 防止数据库迁移被重复执行

- 为迁移任务增加唯一执行标识
- 平台恢复后先查询迁移状态再决定是否重试
- 无法确认外部状态时转入人工接管
```

```text
docs(commit-rules): 统一仓库提交信息规范

- 规定 Conventional Commits 英文类型和英文模块
- 要求摘要及正文使用中文详细说明实际改动
- 增加提交前的暂存范围与格式检查
```

## Before committing

1. Inspect `git status --short` and the staged diff.
2. Confirm the header uses an allowed type, an English module when identifiable, and a Chinese summary.
3. Confirm the body contains specific Chinese bullets matching every material staged change.
4. Confirm the staged files share one primary purpose.
5. Commit with the validated header and body, then inspect the resulting log entry.
