# skill-man Detailed Design

- Status: Draft v0.1
- Author: Codex
- Date: 2026-05-15
- Target: `skill-man` TUI skill manager
- Upstream Compatibility Target: `vercel-labs/skills`

## 1. Document Purpose

本文定义 `skill-man` 的详细设计，目标不是写一份概念稿，而是给后续实现提供可以直接拆分模块、落代码、写测试的工程说明。

`skill-man` 是一个基于 Go 和 Charmbracelet 生态的现代化 TUI，用于管理兼容 `vercel-labs/skills` 规范的 Skills。它以 Slack 风格的 `/command` 为主交互，以 Gemini CLI 风格的头部状态条和实时预览为核心体验，替代传统 CLI 的一次性命令输出。

本文覆盖：

- 产品边界与非目标
- 与 `vercel-labs/skills` 的兼容模型
- TUI 布局与状态机
- 命令系统与自动补全
- 数据模型与本地存储
- Bubble Tea 架构与模块拆分
- 异步任务、错误处理、测试策略与里程碑

## 2. Background and Goals

### 2.1 Problem Statement

现有 `vercel-labs/skills` CLI 已经具备较完整的命令面，包括 `add`、`list`、`find`、`remove`、`update`、`init`。但其交互仍然以传统命令行为主，存在以下问题：

- 状态是瞬时输出，用户无法持续停留在管理界面中工作
- 浏览、查找、安装、删除、更新之间切换成本高
- Skills 内容预览弱，通常需要跳出到编辑器或文件系统
- 参数发现成本高，命令记忆负担重
- 项目级和全局级、agent 级关联关系缺乏可视化

### 2.2 Product Goals

`skill-man` v1 的目标：

- 提供常驻式 TUI 工作台，而不是一次性命令执行器
- 保持与 `vercel-labs/skills` 的概念兼容，至少覆盖核心命令语义
- 用 `/command` 统一操作入口，支持 Tab 自动补全和上下文提示
- 提供左侧列表、右侧预览的 master-detail 管理体验
- 让项目级、全局级、agent 绑定、安装方式、来源地址可视化
- 支持现代终端体验：键盘优先、鼠标可选、即时反馈、无闪烁布局

### 2.3 Non-Goals

以下内容不纳入 v1 必做范围：

- 完整替代上游所有边角选项与内部实现细节
- 内建完整聊天式 Skill 调试器
- 远程 marketplace 排行榜或账号体系
- Windows GUI、Web UI、移动端
- 多用户协作和服务端同步

### 2.4 Success Criteria

如果满足以下条件，可视为 v1 成功：

- 用户无需离开 TUI 即可完成 `list/find/add/remove/update/init` 主流程
- 对已安装 Skill 的预览响应时间稳定在可接受范围内
- Tab 补全、命令提示、确认对话框能明显降低误操作
- 在项目级与全局级切换时，界面和数据模型一致
- 对 Bubble Tea 常见布局问题有明确规约，避免高度错位与文本换行破坏 UI

## 3. Upstream Compatibility Model

### 3.1 Reference Surface

基于 `vercel-labs/skills` 当前公开 README，`skill-man` 需要优先兼容以下命令族：

- `add`
- `list` / `ls`
- `find`
- `remove` / `rm`
- `update`
- `init`

其中上游已具备的重要概念包括：

- 本地路径、GitHub shorthand、完整 Git URL、目录内单 Skill 路径
- 项目级安装与全局安装
- agent 过滤
- 交互式选择与非交互式确认
- 技能源列表与局部选择

### 3.2 Compatibility Philosophy

`skill-man` 不要求完全复制上游命令行参数，但必须保持以下兼容层：

- 概念兼容：source、scope、agent、skill selector、install method
- 数据兼容：能够识别标准 `SKILL.md`
- 操作兼容：主要动作的结果应与上游一致
- 迁移兼容：用户从 `npx skills ...` 切换到 `skill-man` 时无需重新学习领域概念

### 3.3 Command Mapping

`skill-man` 的斜杠命令与上游映射如下：

| `skill-man` | 上游概念 | 说明 |
| --- | --- | --- |
| `/add <source>` | `skills add <source>` | 安装一个 source，可选进入交互选择 |
| `/list` | `skills list` | 展示已安装 Skills |
| `/find [query]` | `skills find [query]` | 在已安装或远程候选中检索 |
| `/remove [skill]` | `skills remove [skill]` | 删除一个或多个技能 |
| `/update [skill]` | `skills update [skill]` | 更新全部或部分技能 |
| `/init [name]` | `skills init [name]` | 创建新的 Skill 模板 |
| `/inspect [skill]` | TUI 增强命令 | 聚焦一个 Skill 并打开详细预览 |
| `/help` | TUI 增强命令 | 展示命令与键位帮助 |
| `/reload` | TUI 增强命令 | 重扫磁盘与缓存 |

### 3.4 Command Aliases

建议保留以下别名：

- `/ls` -> `/list`
- `/rm` -> `/remove`
- `/q` -> `/quit`
- `/h` -> `/help`

## 4. User Experience Overview

### 4.1 Primary Interaction Pattern

用户始终停留在一个常驻 TUI 中。界面结构如下：

```text
╭──────────────────────────────────────────────────────────────────────╮
│                                                                      │
│   ███████╗██╗  ██╗██╗██╗     ██╗         ███╗   ███╗ █████╗ ███╗   ██╗
│   ██╔════╝██║ ██╔╝██║██║     ██║         ████╗ ████║██╔══██╗████╗  ██║
│   ███████╗█████╔╝ ██║██║     ██║         ██╔████╔██║███████║██╔██╗ ██║
│   ╚════██║██╔═██╗ ██║██║     ██║         ██║╚██╔╝██║██╔══██║██║╚██╗██║
│   ███████║██║  ██╗██║███████╗███████╗    ██║ ╚═╝ ██║██║  ██║██║ ╚████║
│   ╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝    ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝
│                                                                      │
│   scope: project │ agents: codex, claude │ skills: 12 │ ready        │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Skills                            Preview                           │
│  ──────                            ───────                           │
│  ● web-design-guidelines           # Web Design Guidelines           │
│    project | codex, claude                                           │
│  ● api-docs-generator              A comprehensive guide for...      │
│    global | codex                                                    │
│  ● test-fixture-builder            Tools: read_file, write_file      │
│    project | codex, claude         Source: github.com/vercel-labs/   │
│  ● error-classifier                                                  │
│    global | claude                                                   │
│                                                                      │
│                                                                      │
├──────────────────────────────────────────────────────────────────────┤
│ ❯ /list                                                              │
│ 12 skills | scope: project | ↑↓ navigate  Tab complete  /help        │
╰──────────────────────────────────────────────────────────────────────╯
```

默认模式下（home / listing）：

- 顶部为 ASCII art 品牌标识 + 紧凑状态条
- 主内容区为左右分栏：左侧 Skill 列表，右侧预览
- 底部输入栏始终可用
- 无静态命令引用面板 —— 命令通过 `/help` 和 Tab 补全发现

### 4.2 UX Principles

- Command-first: 所有功能都能通过 `/command` 进入，命令通过 Tab 补全和 `/help` 发现，不依赖静态命令面板
- Preview-first: 焦点变化立即带来右侧预览变化
- Keyboard-first: 所有核心操作不依赖鼠标
- Safety-first: 删除、覆盖、跨 scope 修改必须二次确认
- Recoverable: 操作失败应可重试，并能回到稳定状态
- Brand presence: ASCII art header 提供强品牌识别，与 Gemini CLI 体验一致

## 5. Functional Scope

### 5.1 v1 In Scope

- 启动时扫描本地已安装 Skills
- 项目级与全局级切换
- 按 agent 过滤
- 已安装 Skill 浏览、搜索、预览
- 从 source 安装 Skill
- 删除 Skill
- 更新 Skill
- 初始化 Skill 模板
- 自动补全、命令提示、确认弹窗

### 5.2 v1.1 Optional

- 最近命令历史
- 多选批量 remove/update
- 搜索远程仓库中的可安装 Skill 清单
- 预览切换 tab：`README` / `SKILL.md` / `bindings` / `raw`

### 5.3 v2 Candidate

- Skill 对话测试面板
- 命令录制与回放
- marketplace 集成
- telemetry opt-in

## 6. Information Architecture

### 6.1 Screen Regions

#### Header

Header 采用 Gemini CLI 风格，由两部分组成：ASCII art 品牌标识 + 紧凑状态条。

**ASCII art 品牌标识：**

- 启动时渲染 "SKILL MAN" ASCII art 大字（通过 `figlet` 或 lipgloss 内置风格生成）
- 终端宽度 >= 80 时渲染完整 6 行 ASCII art
- 终端宽度 < 80 或高度不足时，自动退化为单行 text logo：`skill-man v0.1`
- ASCII art 仅在 home 态和 listing 态完整展示；modal 打开或高度紧张时可折叠

**状态条（ASCII art 下方）：**

紧凑单行，字段以 `│` 分隔：

- `scope: project|global`
- `agents: all|codex,claude-code`
- `skills: N`
- `status: ready|loading|error|confirm`
- 长任务时在 status 位显示 spinner

状态条始终可见，不随内容区滚动。

#### Left Panel（Content List）

用途：

- 显示 Skill 列表（`/list` 结果）
- 显示 find 搜索结果
- 显示 add source 后的远程候选列表
- 显示日志或帮助菜单

左侧是动态内容区 —— 仅在需要展示列表数据时出现，不永久占用。默认 home 态展示最近使用的 skills 或空状态提示。不再显示静态命令引用（命令通过 `/help` 和 Tab 补全发现）。

#### Right Panel

用途：

- 预览 Skill README
- 预览 `SKILL.md` frontmatter
- 展示工具列表、来源、安装方式、agent 绑定
- 错误详情与解决建议

#### Footer

用途：

- 输入 slash command
- 展示自动补全 ghost text
- 展示参数提示
- 展示当前模式帮助

### 6.2 Responsive Behavior

**Header 自适应：**

- `width >= 80 && height >= 24`: 完整 ASCII art（6 行）+ 状态条
- `width < 80 || height < 24`: 折叠为单行 text logo + 状态条

**内容区自适应：**

宽终端优先用左右分栏；窄终端退化成上下分栏。

- `width >= 120`: 左 35%，右 65%
- `80 <= width < 120`: 左 40%，右 60%
- `width < 80`: 上下堆叠，列表在上，预览在下

Bubble Tea 布局必须遵守以下硬规则：

- 所有边框高度先减 2 再计算内容区
- 有边框面板内禁止依赖自动换行
- 所有列表项、帮助文字、标题必须截断
- 水平布局用 X 坐标命中，垂直布局用 Y 坐标命中

## 7. Interaction Model

### 7.1 Session States

定义一个明确的会话状态机：

```go
type SessionState int

const (
    StateHome SessionState = iota
    StateListing
    StateSearching
    StateInstalling
    StateConfirming
    StateViewingHelp
    StateViewingLog
    StateError
)
```

状态语义：

- `StateHome`: 默认态，输入框活跃，展示欢迎信息或最近 skills，ASCII art header 完整渲染
- `StateListing`: 主内容区显示 Skill 列表 + 右侧预览
- `StateSearching`: 输入框处于搜索驱动模式，结果实时过滤
- `StateInstalling`: source 解析中或安装流程中
- `StateConfirming`: 展示危险操作确认框
- `StateViewingHelp`: 帮助面板（全宽覆盖主内容区）
- `StateViewingLog`: 操作日志面板
- `StateError`: 错误展示态，可返回上一稳定态

### 7.2 Focus Model

焦点要与状态分离，不要混在一个字段里。

```go
type FocusTarget int

const (
    FocusInput FocusTarget = iota
    FocusList
    FocusPreview
    FocusModal
)
```

核心原则：

- 默认焦点在 `input`
- `Up/Down` 可在不丢失输入焦点的前提下控制列表
- `PgUp/PgDn` 控制预览滚动
- 模态框打开时焦点强制切到 `modal`

### 7.3 Input Routing

按键分发优先级：

1. 如果 `modal` 打开，优先由 modal 接收
2. 如果是全局快捷键，直接由顶层处理
3. 如果是导航键，派发到 list 或 preview
4. 其他可打印字符发给 input

建议全局键位：

- `Ctrl+C`: 退出
- `Esc`: 关闭 modal / 清空当前命令上下文 / 返回默认态
- `Tab`: 自动补全
- `Shift+Tab`: 反向补全或循环候选
- `Ctrl+J` / `Ctrl+K`: 列表上下移动
- `Ctrl+L`: 清空日志或刷新当前 view
- `F1` 或 `?`: 帮助

## 8. Slash Command Engine

### 8.1 Command Registry

命令系统不应写死在 `switch` 中，建议注册式设计：

```go
type CommandSpec struct {
    Name        string
    Aliases     []string
    Usage       string
    Summary     string
    Args        []ArgSpec
    Dangerous   bool
    Handler     CommandHandler
}
```

```go
type ArgSpec struct {
    Name        string
    Required    bool
    Repeated    bool
    Suggestions SuggestionProvider
}
```

### 8.2 Parsing Strategy

解析规则：

- 只有以 `/` 开头才进入命令模式
- 支持双引号包裹参数
- 支持布尔 flag 和枚举 flag
- 空输入不执行
- 未知命令进入错误态，但保留当前输入值

示例：

- `/add vercel-labs/agent-skills`
- `/add https://github.com/vercel-labs/agent-skills --list`
- `/remove "web design guidelines"`
- `/list --scope global --agent codex`

### 8.3 Autocomplete

自动补全分两层：

- 命令补全
- 参数补全

命令补全规则：

- 输入 `/` 时显示全部命令
- 输入 `/f` 时匹配 `/find`
- 唯一匹配时，`Tab` 直接补全
- 多匹配时，显示候选列表，不自动覆盖

参数补全规则：

- `/remove ` 后补全已安装 skill 名称
- `/list --scope ` 后补全 `project/global/all`
- `/list --agent ` 后补全当前识别到的 agent

视觉策略：

- 输入值保留在高亮色
- ghost text 用低对比灰色
- 当前参数 help 在 Footer 第二行显示

### 8.4 Execution Model

命令执行分三段：

1. Parse
2. Validate
3. Execute

执行必须异步，避免阻塞主循环：

```go
type CommandStartedMsg struct{ ID string }
type CommandFinishedMsg struct{ ID string; Result CommandResult; Err error }
```

## 9. Core User Flows

### 9.1 App Startup

流程：

1. 读取配置
2. 识别当前工作目录是否为 project scope
3. 扫描 global 和 project skill 根目录
4. 构建索引
5. 初始化命令注册器
6. 渲染默认首页（ASCII art header + 状态条 + 欢迎信息）

失败策略：

- 配置读取失败时使用默认配置继续启动
- 某个 Skill 解析失败时记录 error item，不中断整体启动

### 9.2 `/list`

目标：

- 展示已安装 Skills
- 按 scope、agent、source、install method 过滤

行为：

- 左侧切换为 skill list
- 右侧预览第一个条目
- Footer 显示过滤提示

扩展过滤：

- `/list`
- `/list --scope project`
- `/list --scope global --agent codex`

### 9.3 `/find`

支持两种模式：

- 无参进入交互式搜索
- 有参直接执行并展示结果

搜索范围建议分层：

- 已安装 skill name
- description
- tags
- tools
- source repo

v1 默认先搜本地索引；如果未来支持远程检索，可加显式 `--remote`。

### 9.4 `/add`

流程：

1. 用户输入 source
2. source resolver 识别为 local path / github shorthand / git url / repo skill path
3. 拉取元数据或枚举可安装 skills
4. 若 source 含多个 skills，则进入候选列表
5. 用户确认 scope、agent、install method
6. 执行安装
7. 更新索引和界面

需要明确的安装决策：

- `scope`: `project` 或 `global`
- `agent`: 单个、多个、全部
- `method`: `symlink` 或 `copy`

建议将它做成一个多步 modal，而不是要求用户一次写完所有 flag。

### 9.5 `/remove`

流程：

1. 解析目标 skill
2. 如果无参数，则进入多选删除列表
3. 如果有参数，定位目标 skill
4. 弹出确认 modal
5. 执行删除
6. 刷新索引与预览

确认文案必须足够具体：

- skill 名称
- 影响 scope
- 影响 agent 数
- 删除的是 link 还是 copy

### 9.6 `/update`

流程与 `/remove` 类似，但确认信息需要额外展示：

- 当前版本或 revision
- 来源仓库
- 可更新项数量
- 是否有本地变更风险

如果没有可更新项，应显示非错误态提示。

### 9.7 `/init`

目标：

- 创建标准 `SKILL.md` 模板
- 可选初始化目录结构

流程：

1. 接受名称或目标目录
2. 检查目录冲突
3. 生成模板内容
4. 成功后在右侧预览新建文件

## 10. Data Model

### 10.1 Core Domain Types

```go
type Skill struct {
    ID            string
    Name          string
    Description   string
    Tags          []string
    Tools         []string
    Path          string
    Scope         Scope
    Agents        []AgentID
    InstallMethod InstallMethod
    Source        SkillSource
    Manifest      SkillManifest
    ReadmePath    string
    SkillFilePath string
    UpdatedAt     time.Time
    Broken        bool
    BrokenReason  string
}
```

```go
type SkillSource struct {
    Kind      SourceKind
    Locator   string
    RepoURL   string
    RepoRef   string
    SkillPath string
}
```

```go
type SkillManifest struct {
    Frontmatter map[string]any
    RawMarkdown string
}
```

### 10.2 Supporting Types

```go
type Scope string

const (
    ScopeProject Scope = "project"
    ScopeGlobal  Scope = "global"
)
```

```go
type InstallMethod string

const (
    InstallSymlink InstallMethod = "symlink"
    InstallCopy    InstallMethod = "copy"
)
```

### 10.3 UI Projection Types

列表层不直接依赖 domain model，建议引入 view model：

```go
type SkillItem struct {
    Skill      Skill
    TitleText  string
    DescText   string
    MetaText   string
    MatchScore int
}
```

预览层引入统一文档模型：

```go
type PreviewDoc struct {
    Title       string
    Body        string
    Format      PreviewFormat
    SourceLabel string
}
```

### 10.4 Local Index

为避免每次操作都全盘扫描，建议维护内存索引：

```go
type SkillIndex struct {
    ByID    map[string]Skill
    ByName  map[string][]string
    ByTool  map[string][]string
    ByScope map[Scope][]string
}
```

如需持久化缓存，可增加 `cache/index.json`，但 v1 可以只做内存态。

## 11. Filesystem and Storage Model

### 11.1 Scope Roots

需要抽象 scope 根目录，而不是把路径散落在各处。

```go
type ScopeRoots struct {
    ProjectRoot string
    GlobalRoot  string
    AgentRoots  map[AgentID]string
}
```

### 11.2 Canonical Store vs Agent Bindings

设计上建议区分：

- canonical skill store
- agent-visible bindings

原因：

- 同一个 Skill 可能绑定多个 agent
- symlink/copy 策略应该是绑定层语义，而不是 Skill 本体语义
- remove/update 时需要知道删的是 canonical source 还是 agent binding

### 11.3 Suggested Runtime Files

建议引入本地配置目录：

- `~/.config/skill-man/config.yaml`
- `~/.local/share/skill-man/history.json`
- `~/.cache/skill-man/preview/`

项目内不强制生成额外文件，除非未来需要 lockfile 或 workspace metadata。

## 12. Bubble Tea Application Architecture

### 12.1 High-Level Module Layout

建议目录：

```text
cmd/skill-man/
  main.go

internal/app/
  bootstrap.go
  model.go
  update.go
  view.go
  layout.go
  styles.go
  messages.go

internal/commands/
  registry.go
  parser.go
  add.go
  list.go
  find.go
  remove.go
  update.go
  init.go
  help.go

internal/domain/
  skill.go
  source.go
  scope.go

internal/service/
  scanner.go
  installer.go
  remover.go
  updater.go
  initializer.go
  preview.go
  search.go

internal/store/
  config_store.go
  history_store.go
  index_store.go

internal/ui/
  header.go          // ASCII art logo + status bar
  footer.go
  list_item.go
  preview.go
  modal_confirm.go
  modal_wizard.go
```

### 12.2 Bubble Tea Model

建议顶层 `model`：

```go
type model struct {
    state       SessionState
    focus       FocusTarget
    width       int
    height      int

    input       textinput.Model
    list        list.Model
    preview     viewport.Model
    spinner     spinner.Model

    skills      []Skill
    filtered    []SkillItem
    selected    *Skill

    registry    *CommandRegistry
    config      Config
    runtime     RuntimeState

    confirm     *ConfirmState
    wizard      *WizardState
    status      StatusBarState
    logs        []LogEntry
    lastStable  SessionState
}
```

### 12.3 Messages

消息类型建议显式声明，不要滥用匿名 struct：

```go
type SkillsScannedMsg struct {
    Skills []Skill
    Err    error
}

type PreviewLoadedMsg struct {
    SkillID string
    Doc     PreviewDoc
    Err     error
}

type SearchResultsMsg struct {
    Query   string
    Results []SkillItem
}

type CommandCompletedMsg struct {
    Name   string
    Result CommandResult
    Err    error
}
```

### 12.4 Update Loop Responsibilities

顶层 `Update()` 只做 4 件事：

- 处理全局消息
- 按状态做键盘路由
- 触发异步命令
- 更新视图模型

不要在顶层 `Update()` 中塞入大量业务逻辑。安装、删除、搜索、扫描都应下沉到 service。

## 13. Rendering Design

### 13.1 Header Rendering

Header 分为两层：

**Layer 1 — ASCII art 品牌标识：**

- 使用 `figlet` 或 charmbracelet `lipgloss` 内置风格渲染 "SKILL MAN" 大字
- 默认使用 ANSI Shadow 字体（6 行高），颜色使用渐变或品牌色
- 终端宽度 >= 80 且高度 >= 24 时渲染完整 ASCII art
- 终端较小时自动折叠为单行：`skill-man v0.1`
- ASCII art 行不参与内容滚动

**Layer 2 — 紧凑状态条：**

- 固定 1 行，字段以 `│` 分隔，位于 ASCII art 下方
- 字段：`scope: project|global │ agents: codex,claude │ skills: N │ ready`
- 长任务时 status 位显示 spinner（如 `installing...`）
- 错误态时 status 位变红
- 状态条始终可见，不随内容区滚动

Header 总高度：完整模式 7-8 行（ASCII art + 间距 + 状态条），紧凑模式 1 行。

### 13.2 Content List Rendering（原 Left Panel）

内容列表不再包含静态命令引用。列表项类型由当前状态决定：

- `StateHome`：不渲染列表或展示最近使用的 skills 快捷入口
- `StateListing`：Skill 列表
- `StateSearching`：实时过滤的搜索结果
- `StateInstalling`：source 候选列表

左侧推荐复用 `bubbles/list`，但要注意默认样式需要重做：

- 去掉过强的默认装饰
- 高亮当前选中项
- 显示副标题和元数据
- 截断超长文本

建议每个 skill item 三段信息：

- 名称
- 一行描述
- 一行元信息：`scope | agents | method | source`

### 13.3 Right Panel Rendering

预览渲染管线：

1. 选中 Skill
2. preview service 读取 README 或 `SKILL.md`
3. 如果存在 Markdown，使用 `glamour`
4. 如果不存在 README，则回退到 metadata summary
5. 将渲染后的 ANSI 内容写入 `viewport`

预览优先级：

1. `README.md`
2. `SKILL.md`
3. 结构化 metadata summary
4. 错误说明

### 13.4 Footer Rendering

Footer 至少两行：

- 第一行：命令输入
- 第二行：hint / autocomplete / validation error
- 备注：Hint 信息必须基于当前 `State` 和列表选中项动态计算生成。例如，只有在列表中选中具体的 Skill 时才展示 `Enter:Inspect`, `X:Toggle`, `B:Bind`, `Del:Remove` 等针对单个 Skill 的快捷键提示；首页或无选中项时仅保留全局快捷键提示。

Footer 必须常驻，不能被 modal 覆盖后完全消失；modal 出现时，Footer 可降亮但要保留上下文。

### 13.5 Modal Rendering

两类 modal：

- confirm modal
- wizard modal

confirm modal 要求：

- 固定宽度
- 红色或高对比边框
- 明确展示影响对象
- `y` 确认，`n` 取消，`Enter` 默认安全路径

wizard modal 用于 `/add` 和可能的 `/init`。

## 14. Search and Filtering

### 14.1 Search Modes

`/find` 建议支持三种模式：

- exact prefix
- fuzzy name match
- metadata match

### 14.2 Search Ranking

简单可行的排序规则：

1. 名称前缀完全匹配
2. 名称包含匹配
3. 工具名匹配
4. 描述匹配
5. tag 匹配

### 14.3 Incremental Search

如果处于 `StateSearching`：

- 输入每次变化都触发本地过滤
- 使用 debounce，避免大数据量下每键重算
- 查询为空时恢复完整列表

### 14.4 Filter Chips

虽然是 TUI，但可以文本化展示过滤条件：

- `scope:project`
- `agent:codex`
- `method:symlink`

这些条件展示在 Header 或 Footer hint 中，而不是塞进列表正文。

## 15. Install / Remove / Update Service Design

### 15.1 Scanner Service

职责：

- 扫描本地 scope roots
- 识别有效 Skill
- 解析 `SKILL.md`
- 建立 agent 绑定关系

接口建议：

```go
type Scanner interface {
    Scan(ctx context.Context, opts ScanOptions) ([]Skill, error)
}
```

### 15.2 Installer Service

职责：

- 解析 source
- 拉取远程仓库或读取本地目录
- 枚举 skill candidates
- 安装 canonical copy
- 创建 agent bindings

关键难点：

- source 解析归一化
- 单仓库多个 skills 的选择
- symlink 与 copy 策略
- 覆盖安装确认

### 15.3 Remover Service

职责：

- 删除指定 skill 或 binding
- 清理失效链接
- 回收索引

删除策略需要明确：

- remove binding only
- remove canonical and all bindings

v1 可以先默认删除当前 scope 下的目标 skill 及其相关绑定。

### 15.4 Updater Service

职责：

- 检查 source 是否可更新
- 拉取新 revision
- 重新同步 binding
- 保持原先 scope 和 agent 配置

## 16. Error Handling and Recovery

### 16.1 Error Categories

建议错误分类：

- user input error
- validation error
- filesystem error
- network/source resolution error
- parse/render error
- unexpected internal error

### 16.2 Display Strategy

不同错误用不同表现：

- 输入错误：Footer 内联提示
- 命令失败：Header 状态变红 + 日志面板记录
- 大错误：右侧预览显示详细说明
- 危险错误：modal 拦截

### 16.3 Recovery Rules

- 任意错误后，用户都能用 `Esc` 回到稳定态
- 异步任务失败不能让 UI 失焦或卡死
- 预览渲染失败时回退为纯文本

## 17. Logging and Observability

### 17.1 In-App Log

建议维护最近 N 条操作日志，用于：

- 命令成功/失败记录
- source 解析记录
- 安装/删除摘要

日志展示入口：

- `/logs`
- `StateViewingLog`

### 17.2 Debug Mode

提供 `--debug` 启动参数，输出：

- 当前 roots
- 扫描结果
- 命令解析细节
- service 执行耗时

## 18. Configuration Design

### 18.1 Config File

建议配置项：

```yaml
theme: auto
default_scope: project
default_install_method: symlink
preview:
  preferred_source: readme
  glamour_style: dark
search:
  fuzzy: true
  debounce_ms: 80
ui:
  show_help_footer: true
  compact_header: false
agents:
  preferred:
    - codex
    - claude-code
```

### 18.2 Config Precedence

优先级：

1. runtime flags
2. environment variables
3. config file
4. defaults

## 19. Keybinding Specification

建议键位矩阵：

| Key | Scope | Behavior |
| --- | --- | --- |
| `Enter` | input | 执行命令 |
| `Tab` | input | 补全命令或参数 |
| `Esc` | global | 关闭 modal / 返回默认态 |
| `Ctrl+J` | global | 列表下移 |
| `Ctrl+K` | global | 列表上移 |
| `PgDn` | preview | 下滚预览 |
| `PgUp` | preview | 上滚预览 |
| `?` | global | 打开帮助 |
| `y` | confirm | 确认危险操作 |
| `n` | confirm | 取消危险操作 |
| `Ctrl+C` | global | 退出 |

## 20. Testing Strategy

### 20.1 Unit Tests

优先覆盖：

- 命令解析器
- 自动补全逻辑
- search ranking
- scope root 识别
- `SKILL.md` 解析

### 20.2 Service Tests

使用临时目录构造：

- local skill install
- symlink install
- copy install
- remove after multi-agent binding
- project/global scope isolation

### 20.3 TUI Tests

建议对以下内容做 golden tests：

- Header 渲染
- 列表和预览布局
- confirm modal
- 窄终端 fallback 布局

### 20.4 End-to-End Tests

通过脚本驱动：

- 启动应用
- 输入 `/list`
- 选择一个 Skill
- 执行 `/remove`
- 确认删除
- 验证索引刷新

## 21. Performance Considerations

### 21.1 Startup

目标是让启动尽快看到可交互 UI，因此建议：

- 先渲染 skeleton
- 后台扫描 skills
- 扫描完成后增量刷新列表

### 21.2 Preview

Markdown 渲染成本可能偏高，建议：

- 缓存最近预览结果
- 仅在焦点变化时重新渲染
- 同一 Skill 重复选中时直接复用缓存

### 21.3 Search

本地数据量通常不大，但仍建议：

- 维护内存索引
- 增量过滤
- 结果数量过多时限制预览更新频率

## 22. Security and Safety

### 22.1 Trust Boundary

Skill source 可能来自任意 Git 仓库或本地目录，因此必须假设：

- metadata 不可信
- README 不可信
- 路径可能恶意构造

### 22.2 Required Protections

- 所有路径先做 `Clean` 和 root boundary 校验
- 禁止目录穿越写入
- 删除前必须确认作用域
- 渲染 Markdown 时不执行外部内容

## 23. Implementation Plan

### Phase 1: Skeleton

- 建立 Bubble Tea app 骨架
- 建立 ASCII art Header（figlet/lipgloss）+ 状态条 + Content Area（列表+预览）+ Footer 布局
- 落地命令注册器与 `/help`

### Phase 2: Local Management

- 扫描本地 Skills
- 完成 `/list`、`/find`、`/inspect`
- 完成实时预览

### Phase 3: Mutations

- 完成 `/add`
- 完成 `/remove`
- 完成 `/update`
- 完成确认 modal 和 wizard

### Phase 4: Hardening

- 完成配置系统
- 完成日志与调试模式
- 完成 golden tests 和 e2e

## 24. Open Questions

以下问题应在实现前尽快定稿：

1. `skill-man` 是否直接复用上游 `vercel-labs/skills` 的内部安装逻辑，还是完全自实现？
2. v1 是否要求远程 source 浏览，也就是 `/add <repo> --list` 的仓库内 skill 枚举？
3. canonical store 的目录规范是否要显式对齐上游，以便双向兼容？
4. `/remove` 删除默认语义是“删当前 scope 的绑定”还是“删整个技能实体”？
5. 预览是否需要多 tab，还是单 viewport 即可？

## 25. Final Recommendation

`skill-man` 不应该只是把现有 CLI 包一层皮，而应明确定位成一个长期驻留的 skill workspace。工程上最关键的不是炫效果，而是三件事：

- 命令系统要可扩展，不能靠大 `switch`
- 数据层要把 `skill`、`binding`、`scope`、`source` 分离清楚
- 布局层要严格遵守 Bubble Tea 的边框和截断规则，否则后面 UI 会一直返工

如果按本文设计推进，第一版就能形成一个足够现代、可维护、可继续扩展的 skill 管理 TUI 基线。
