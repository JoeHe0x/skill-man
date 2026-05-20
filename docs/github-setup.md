# GitHub 仓库发现性设置

在 [skill-man 仓库 Settings](https://github.com/JoeHe0x/skill-man/settings) 里手动配置（无法通过 README 自动生效）。

## About（仓库简介）

**Description（复制到 About）：**

```
Keyboard-first TUI: browse, preview, bind Agent Skills & MCP across Cursor, Claude Code, Codex, Windsurf.
```

**Website（可选）：** 留空，或填你的博客 / 讨论帖链接。

**勾选：** Include in the home page（若你希望出现在个人主页 pinned 项目里，在 Profile 里 pin 本仓库）。

## Topics（标签，建议全选或按需删减）

在 About 右侧 **Topics** 添加：

```
cursor
claude-code
codex
windsurf
mcp
model-context-protocol
agent-skills
skills
bubbletea
tui
terminal
cli
golang
developer-tools
ai-coding
```

## 用 gh CLI 一次性设置（可选）

```bash
gh repo edit JoeHe0x/skill-man \
  --description "Keyboard-first TUI: browse, preview, bind Agent Skills & MCP across Cursor, Claude Code, Codex, Windsurf." \
  --add-topic cursor,claude-code,codex,windsurf,mcp,model-context-protocol,agent-skills,skills,bubbletea,tui,terminal,cli,golang,developer-tools,ai-coding
```

## Social preview

Settings → General → **Social preview**：上传一张 1280×640 图（可用 README 里的 demo GIF 截一帧，或终端截图 + 项目名）。

## Releases

每个 tag 写 3～5 行 Release notes，突出「一条命令安装」：

```bash
go install github.com/JoeHe0x/skill-man/cmd/skill-man@v0.1.0
```
