<div align="center">

```
████ █  █ ███ █    ▓      ▓   ▒  ▒▒  ▒  ▒
█    █ █   █  ▓    ▓      ▓▓ ▒▒ ▒  ▒ ▒▒ ▒
███  ██    █  ▓    ▓      ▓ ▒ ▒ ▒▒▒▒ ▒ ▒▒
   █ █ █   █  ▓    ▓      ▓   ▒ ▒  ▒ ▒  ▒
████ █  █ █▓▓ ▓▓▓▓ ▓▓▓▓   ▒   ▒ ▒  ▒ ▒  ░
```

# skill-man

**Stop grepping five `mcp.json` files.** One keyboard-first TUI to browse Skills, preview MCP, and bind across Cursor, Claude Code, Codex & Windsurf.

[![CI](https://github.com/JoeHe0x/skill-man/actions/workflows/ci.yml/badge.svg)](https://github.com/JoeHe0x/skill-man/actions/workflows/ci.yml)
[![GitHub stars](https://img.shields.io/github/stars/JoeHe0x/skill-man?style=social)](https://github.com/JoeHe0x/skill-man/stargazers)
[![npm](https://img.shields.io/npm/v/@joehe0x/skill-man?style=flat&logo=npm)](https://www.npmjs.com/package/@joehe0x/skill-man)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Bubble Tea](https://img.shields.io/badge/Powered%20by-Bubble%20Tea-ff87d7)](https://github.com/charmbracelet/bubbletea)

[English](#why-skill-man) · [中文](#中文简介) · [Features](#features) · [Quick start](#quick-start) · [Keybindings](#keybindings)

```bash
npm install -g @joehe0x/skill-man && cd your-project && skill-man
```

**[⭐ Star on GitHub](https://github.com/JoeHe0x/skill-man)** if this saves you from juggling ten config files.

![skill-man demo](docs/demo.gif)

</div>

---

## Why skill-man?

You use more than one AI coding agent. Each one hides **Skills** and **MCP servers** in different folders (`mcp.json`, `config.toml`, `.claude.json`, …). You grep, edit blind, rescan mentally, and still wonder which tool actually picked up the change.

**skill-man** is a keyboard-first terminal workbench — split list + live preview, mutations on **real files on disk**:

## How is this different?

| Tool | Focus | skill-man |
|------|--------|-----------|
| [vercel-labs/skills](https://github.com/vercel-labs/skills) CLI | Install / update / remove skills | **Complements it** — same `SKILL.md` model; adds MCP + cross-agent bind in a TUI |
| Config-sync repos (e.g. one source → many agents) | Declarative sync, often YAML-driven | **Interactive** — see what’s on disk, preview, toggle, bind with confirmations |
| Generic JSON editors | Edit one file at a time | **Agent-aware** — scans 70+ layout conventions, filters by agent, Skills **and** MCP in one app |

**Who is it for?** Developers who daily-switch between Cursor, Claude Code, Codex, or Windsurf and want one place to manage skills *and* MCP without leaving the terminal.

**skill-man** puts everything in one split-pane UI:

| Instead of… | Use skill-man to… |
|-------------|-------------------|
| `find` + `cat` + editor hopping | Browse, search, and preview in one screen |
| Guessing which agent sees which skill | Filter by agent and **bind** across toolchains |
| Hand-editing `mcp.json` / `config.toml` | **Toggle**, **bind**, and **remove** with confirmations |
| Fragile one-off shell scripts | **Rescan** disk and see live counts in the header |

Built with [Charm](https://charm.sh/) — [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), [Glamour](https://github.com/charmbracelet/glamour).

---

## Features

### Skills tab

- **Scan** project & global skill dirs across **70+ agent** layout conventions
- **List / find / filter** by agent
- **Live preview** of `SKILL.md` (Markdown)
- **Inspect** skill file trees · **install** · **init** templates · **update**
- **Bind / unbind** to agents (symlinks) · **enable / disable** · **remove** (confirmed)

### MCP tab

- **Discover** real MCP entries from JSON & TOML (not a placeholder list)
- **Skills ↔ MCP** via `Tab` / `Shift+Tab`
- **Preview** stdio vs URL transport and raw config
- **Toggle** enable/disable · **bind** into another agent’s config · **remove** (confirmed)

### UX

- Split **list + preview** (stacks on narrow terminals)
- Branded header: ASCII logo + live **overview** in a bordered banner
- Status bar: scope, agents, skill/MCP counts, readiness
- Mouse-friendly scrolling where supported

---

## Quick start

### Requirements

- **Node.js 18+** (recommended) or **Go 1.26+** (from source / fallback)
- True-color terminal recommended (iTerm2, WezTerm, Kitty, Windows Terminal, …)

### Install

**npm (recommended — no Go toolchain required):**

```bash
npm install -g @joehe0x/skill-man
# or from this repo before publish:
npm install -g .
```

Installs a prebuilt binary from [GitHub Releases](https://github.com/JoeHe0x/skill-man/releases). If the asset is not published yet, the installer builds from local source (when installing from a clone) or falls back to `go install` when Go is on your `PATH`.

> **Maintainers:** push tag `v*` to trigger [GoReleaser](.github/workflows/release.yml), then `npm publish` from the repo root.

**From GitHub (Go):**

```bash
go install github.com/JoeHe0x/skill-man/cmd/skill-man@v0.1.0
```

**From source:**

```bash
git clone https://github.com/JoeHe0x/skill-man.git
cd skill-man
make install   # → $GOPATH/bin/skill-man
```

**Or run without installing:**

```bash
make dev       # go run ./cmd/skill-man
```

### Run

```bash
cd your-project
skill-man
```

Uses your **current working directory** as the project root and scans project + user-level configs.

---

## Re-record demo

To refresh `docs/demo.gif` (Enter · X · MCP tab — short tape, see `docs/demo/demo.tape`):

```bash
go install github.com/charmbracelet/vhs@latest
brew install ffmpeg ttyd
make demo
```

See [docs/demo/README.md](docs/demo/README.md).

---

## Keybindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch **Skills** / **MCP** |
| `↑` `↓` / `Ctrl+K` `Ctrl+J` | Move selection |
| `Enter` | Inspect skill tree or refresh MCP preview |
| `X` | Toggle enable / disable |
| `B` | Bind to agents (`Enter` to apply) |
| `Del` | Remove (confirmation) |
| `Ctrl+P` | Command palette (fuzzy search actions) |
| `Ctrl+F` / `/` | Filter list (inline fuzzy) |
| `Ctrl+A` | Cycle agent filter |
| `Ctrl+R` | Rescan disk |
| `Ctrl+L` | Focus list |
| `Ctrl+U` | Update skill(s) |
| `Ctrl+D` | Install skill (prompt) |
| `Ctrl+N` | New skill template (prompt) |
| `?` / `F1` | Help |
| `Esc` | Home / cancel |
| `Ctrl+C` | Quit |

---

## MCP config discovery

| Tool | Typical paths |
|------|----------------|
| **Cursor** | `.cursor/mcp.json`, `~/.cursor/mcp.json` |
| **Claude Code** | `.mcp.json`, `.claude/mcp.json`, `~/.claude.json` (`projects.*.mcpServers`) |
| **Codex** | `.codex/config.toml`, `~/.codex/config.toml` |
| **Windsurf** | `~/.codeium/windsurf/mcp_config.json` |

**Bind** merges a server into the target agent config. **Toggle** and **remove** edit the underlying JSON/TOML in place.

---

## Skills compatibility

Aligned with the [vercel-labs/skills](https://github.com/vercel-labs/skills) model:

- Standard `SKILL.md` layout · project vs global scope
- Agent-specific install directories · install / update / remove flows

---

## Documentation

- [TUI 现代化改造路线图](docs/ui-modernization-roadmap.md) — Bubble Tea 示例对照、分阶段计划与完成状态
- [GitHub 发现性设置](docs/github-setup.md) — About、Topics、Social preview（维护者）
- [推广短文大纲](docs/promotion-post.md) — 中英帖模板与投放顺序

## Architecture

```text
cmd/skill-man          CLI entry
internal/app           Bubble Tea UI (panels, keys, layout)
  └── panel/           Skills & MCP tab strategies
internal/usecase       Use cases (no UI deps)
  ├── extension/       Remove, disable, add, init, update
  └── bind/            Agent bind/unbind for skills and MCP
internal/domain        Skill, MCP, Agent, Extension
internal/service
  ├── skill/           Scan, install, preview, update
  ├── mcp/             Scan, parse (JSON/TOML), mutate
  └── manager/         Generic extension scanner
```

---

## Development

**npm package (from repo):**

```bash
npm install -g .   # postinstall builds ./cmd/skill-man into dist/
```

**Go:**

```bash
make test          # unit tests (+ race in CI)
make test-cover    # coverage
make fmt vet       # format & vet
make lint          # golangci-lint (optional)
```

---

## Roadmap

- [x] Publish module path `github.com/JoeHe0x/skill-man`
- [x] Tag `v0.1.0` on GitHub
- [x] Demo GIF in README
- [x] CI (test + vet on PR)
- [ ] golangci-lint in CI (optional)
- [x] npm package (`npm install -g @joehe0x/skill-man`)
- [ ] Publish to npm registry on tag
- [ ] Homebrew formula (optional)
- [ ] Hooks / sub-agent tabs

**Want a feature?** [Open an issue](https://github.com/JoeHe0x/skill-man/issues) or send a PR — see [Contributing](#contributing).

---

## Contributing

1. Fork [JoeHe0x/skill-man](https://github.com/JoeHe0x/skill-man)
2. Branch: `git checkout -b feat/your-idea`
3. `make test` — keep PRs focused
4. Open a pull request

---

## License

[MIT](LICENSE) © JoeHe0x

---

## 中文简介

**skill-man** 是用 Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) 打造的终端工作台：一个界面管理 **Agent Skills** 和 **MCP**，不用在十几个路径里 `find`、手改 JSON。

**和别的工具差在哪？**

| 场景 | skill-man |
|------|-----------|
| 只用 [skills CLI](https://github.com/vercel-labs/skills) 装 skill | 同样认 `SKILL.md`；额外管 MCP、跨 Agent 绑定 |
| 用脚本/YAML 同步多份配置 | 交互式：列表 + 预览 + 确认后再写盘 |
| 普通编辑器改单个 `mcp.json` | 按 Agent 扫描、过滤，Skills 与 MCP 同一套快捷键 |

**为什么值得 Star？**

- 左侧列表、右侧实时预览 `SKILL.md` / MCP 配置  
- **Skills / MCP** 双 Tab，`Tab` 切换，`Ctrl+R` 一键重扫  
- 绑定、启用/禁用、删除（带确认），改的是真实配置文件  
- **Cursor、Claude Code、Codex、Windsurf** 等常见路径  

**快速开始：**

```bash
npm install -g @joehe0x/skill-man
# 或 Go：go install github.com/JoeHe0x/skill-man/cmd/skill-man@v0.1.0
# 或源码：git clone ... && make install
cd 你的项目目录 && skill-man
```

| 按键 | 作用 |
|------|------|
| `Tab` | Skills ↔ MCP |
| `X` / `B` / `Del` | 禁用·绑定·删除 |
| `Ctrl+F` | 搜索 |
| `Ctrl+R` | 重新扫描 |

觉得有用的话，欢迎 **[点 Star ⭐](https://github.com/JoeHe0x/skill-man)**，让更多人发现这个项目。

---

<p align="center">
  <a href="https://github.com/JoeHe0x/skill-man/stargazers">⭐ Star skill-man</a>
  ·
  <a href="https://github.com/JoeHe0x/skill-man/issues">Report issue</a>
  ·
  <sub>Built with Charm · Happy shipping</sub>
</p>
