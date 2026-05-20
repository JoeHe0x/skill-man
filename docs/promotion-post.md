# 推广短文大纲

目标：让人 **试用 `go install`**，而不是先 Star。Star 往往在「用顺了」之后。

---

## 中文帖（V2EX / 即刻 / 飞书群 / 知乎想法）

**标题候选：**

- 我在 4 个 AI 编程工具之间同步 Skills 和 MCP，写了个 TUI
- 不再手改五份 mcp.json：skill-man 把 Skills + MCP 收进一个终端

**结构（约 400～600 字）：**

1. **痛点（2～3 句）**  
   Cursor、Claude Code、Codex、Windsurf 各有一套 Skills / MCP 路径；改完不知道哪个 Agent 真的读到了。

2. **我试了啥（1 句）**  
   用 [vercel-labs/skills](https://github.com/vercel-labs/skills) 装 skill 很方便，但 MCP 还是要自己翻 JSON/TOML。

3. **skill-man 做什么（3 条 bullet）**  
   - 一个 TUI：左侧列表，右侧预览 `SKILL.md` 或 MCP 配置  
   - `Tab` 切 Skills / MCP，`B` 绑定到别的 Agent，`X` 启用/禁用  
   - 直接改磁盘上的真实配置文件，不是假 UI

4. **一张图**  
   贴 `docs/demo.gif` 或终端截图。

5. **安装（复制即用）**

   ```bash
   go install github.com/JoeHe0x/skill-man/cmd/skill-man@v0.1.0
   cd 你的项目 && skill-man
   ```

6. **结尾**  
   开源 MIT，欢迎 issue/PR。链接：https://github.com/JoeHe0x/skill-man  
   （不要写「求 star」，可写「觉得有用可以 star 让更多人看到」。）

---

## 英文帖（X / Reddit r/cursor / Hacker News Show HN）

**Title:** Show HN: skill-man – one TUI to browse, preview, and bind Agent Skills + MCP across Cursor, Claude, Codex, Windsurf

**Body skeleton:**

- Problem: N agents × M config formats = grep-driven ops.
- What it is: Keyboard-first Bubble Tea workbench; split list + preview; mutates real JSON/TOML on disk.
- Not trying to replace the skills CLI for installs—complements it with MCP + cross-agent bind.
- Demo: link to GIF in repo README.
- Try: `go install github.com/JoeHe0x/skill-man/cmd/skill-man@v0.1.0` then `skill-man` in your project root.
- Repo: https://github.com/JoeHe0x/skill-man

**HN 注意：** 评论区准备好回答「和 AgentX / agents repo 的区别」——见 README「How is this different?」表。

---

## 社区投放顺序（性价比）

| 顺序 | 渠道 | 说明 |
|------|------|------|
| 1 | 个人 X / 微博 | 零成本，附 GIF |
| 2 | Cursor / Claude Discord | 找 `#tools` 或 `#showcase` |
| 3 | V2EX「分享创造」 | 中文开发者密度高 |
| 4 | awesome 列表 PR | `awesome-mcp-servers`、Cursor 相关合集 |
| 5 | vercel-labs/skills Discussions | 说明是「管理端」，非竞品 |

---

## 一周后复盘

- GitHub → Insights → Traffic：views / clones 是否上升  
- 有没有人提 issue（哪怕 bug 也是信号）  
- 若 clones > 0 仍 0 star：正常，继续迭代功能或第二条短文
