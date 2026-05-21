# ADR-0003: Decompose Model into Feature Components

## Status

Partially Implemented

**2026-05-20**: 7 feature wrapper structs implement `feature.Feature` and are registered
in `Model.features`. `dispatchToFeatures()` routes messages to active features before
the main Update switch. `dispatchKey()` simplified from 9 checks to 3.

**2026-05-21 audit** — dispatch is real, depth is not yet:

| Criterion (original) | Current |
|----------------------|---------|
| Feature logic not on Model | **Mostly** — install UI/completion in `feature_install.go`; palette/help/bind/confirm/prompt in feature files; Model keeps thin delegates (`startInstallFlow`, list helpers) |
| Model ≤ 15 fields | **Partial** — `Core` + embedded `listPane` (main/overlay lists, preview, tree); `Model` still owns session, features, panels, managers |
| Feature-owned state | **Mostly** — install flow + background progress on `installFeature`; palette/help/bind/confirm/prompt owned by features |
| Deletion test on `feature/` | Removing `installFeature` would pull `installui` + background overlay back onto Model; removing `listPane` would scatter list/preview/tree fields |

**2026-05-21 (candidates 1+2):** bind, confirm, prompt, item effects — see ADR-0004/0006.

**2026-05-21 (follow-up):** `cmdPalette`, `helpScreen` features; `mutation_lifecycle.go` (`runCommand` + `applyMutationResult`).

**2026-05-21 (list/install split):** `list_pane.go` (`listPane`: main list, overlay `agentList`, preview, inspect tree); `feature_install.go` (wizard + `installBackground`); deleted `install_bridge.go`.

**2026-05-21 (state handlers, ADR-0002 deepening):** Per-state key/msg logic under `internal/app/state/{listing,inspect,fallback,filtering,listfilter,installing}/`; `Model` bridges via `state_*_host.go`. Bind/confirm/palette/help remain in `feature/*` (consumed by `dispatchToFeatures` before `dispatchKey`).

**2026-05-21 (state hosts + list bridge):** Merged `state_*_host.go` → `state_hosts.go`; selection/preview/reselect helpers in `list/bridge.go` (`BridgeHost`).

Next deepening: move agent-filter open/render into `state/filtering` or `feature/filter`; ADR-0005 when a third extension type lands.

## Context

`Model` struct 有 30+ 字段，承担 7 个独立功能：

```
Model
├── 核心 (width, height, cwd, home, status, errMsg, darkTheme)  ← 基础
├── 列表 (list, listDelegate, tree, focusedPane)                ← 主列表
├── 预览 (preview, previewBody, previewGen)                     ← 预览面板
├── 标签页 (activeTab, panels)                                   ← Tab 管理
├── 技能操作 (skillManager, binds)                               ← 技能 CRUD
├── MCP 操作 (mcpManager)                                        ← MCP CRUD
├── 安装流 (installFlow, installCancel, pending)                 ← 搜索安装
├── Agent (allAgents, agentIDs, agentList, agentListDelegate)    ← Agent 过滤
├── 命令面板 (palette, registry)                                  ← 命令搜索
├── 帮助 (help, helpOverlay)                                     ← 帮助覆盖层
├── UI 装饰 (spinner, styles, footerFlash, footerFlashTag, footerContext)
└── 提示 (prompt)                                                ← 临时文本输入
```

问题：
- **认知负载高**：任何修改都需要理解整个 Model
- **改一个功能可能影响其他**：install flow 改 footer 和 agent filter 改 footer 互相冲突
- **无封装**：所有字段都是 exported 或 package-private，任何方法可以碰任何字段
- **测试耦合**：`model_test.go` 测一个功能要初始化整个 Model

## Decision

采用 **组合模式**，将 Model 拆分为自包含的功能组件（Feature Component），每个组件有自己的字段和方法：

```go
// Model 退化为 orchestrator
type Model struct {
    core      *Core          // width, height, cwd, home, theme, status
    list      *ListComponent // 主列表 + 树
    preview   *PreviewComponent
    tabs      *TabComponent
    header    *HeaderComponent  // 包含 agent filter
    footer    *FooterComponent  // 包含 flash

    // 功能组件以 stack 形式管理（当前活跃的顶层）
    features  []Feature
}

// Feature 是一个自包含的 UI 功能
type Feature interface {
    Name() string
    // Active 返回当前是否拦截事件
    Active() bool
    // Update 处理消息；如果不关心，返回 nil
    Update(msg tea.Msg) (tea.Cmd, bool)  // bool = 事件已消费
    // View 返回此功能的渲染内容（如果 Visible）
    View(width, height int) string
    // Visible 是否需要渲染
    Visible() bool
}
```

**具体组件：**

```go
// install/flow.go — 独立包
type InstallFlow struct {
    provider serviceinstall.Provider
    step     installStep
    // ... install 专属字段
    cancel   context.CancelFunc
}

func (f *InstallFlow) Active() bool  { return f.step != stepIdle }
func (f *InstallFlow) Update(msg tea.Msg) (tea.Cmd, bool) {
    // 只处理 install 相关的消息
}
func (f *InstallFlow) View(w, h int) string {
    // 只渲染 install 对话框
}
```

```go
// bind/flow.go — 独立包
type BindFlow struct {
    skill      *skilldomain.Skill
    mcpMembers []*mcpdomain.Server
    choices    []agentBindChoice
    list       list.Model
}
func (f *BindFlow) Active() bool { return f.skill != nil || len(f.mcpMembers) > 0 }
```

**Model.Update 简化为：**

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 1. 顶层功能优先消费事件
    for _, f := range m.features {
        if f.Active() {
            cmd, consumed := f.Update(msg)
            if consumed {
                return m, cmd
            }
        }
    }
    // 2. 基础组件
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.core.Update(msg)
    case tea.KeyMsg:
        return m.core.HandleKey(msg)
    }
    // 3. List/Preview 常驻更新
    var cmds []tea.Cmd
    cmds = append(cmds, m.list.Update(msg))
    cmds = append(cmds, m.preview.Update(msg))
    return m, tea.Batch(cmds...)
}
```

## 文件结构目标

```
internal/app/
├── app.go               # Model + New() + Update + View（瘦身）
├── core/
│   ├── core.go          # 基础字段 + 主题 + 状态
│   └── core_test.go
├── feature/
│   ├── feature.go       # Feature 接口
│   ├── install/
│   │   ├── flow.go      # InstallFlow
│   │   ├── browse.go    # 浏览步骤
│   │   ├── agents.go    # 选择 Agent 步骤
│   │   ├── confirm.go   # 确认步骤
│   │   ├── render.go    # 渲染
│   │   └── flow_test.go
│   ├── bind/
│   │   ├── flow.go      # BindFlow
│   │   ├── skill.go     # 技能绑定
│   │   ├── mcp.go       # MCP 绑定
│   │   ├── render.go
│   │   └── flow_test.go
│   ├── inspect/
│   │   ├── flow.go      # InspectFlow (文件树)
│   │   └── flow_test.go
│   ├── confirm/
│   │   ├── confirm.go   # 通用确认对话框
│   │   └── confirm_test.go
│   ├── palette/
│   │   ├── palette.go   # 命令面板
│   │   └── palette_test.go
│   └── help/
│       ├── overlay.go   # 帮助覆盖层
│       └── overlay_test.go
├── component/
│   ├── list.go          # 列表组件
│   ├── preview.go       # 预览组件
│   ├── header.go        # 顶栏（Tab + agent filter）
│   └── footer.go        # 底栏（status + key hints + flash）
└── panel/
    └── ...              # 已有，保持不变
```

## Rationale

1. **单一职责**：每个组件只管一件事
2. **独立测试**：InstallFlow 可以独立测试，不需要构造 Model
3. **可选渲染**：Feature.Visible() 让 View() 变成简单的组合调用，不需要 `if m.state == stateInstalling` 分支
4. **事件消费链**：类似 HTTP 中间件，顶层功能优先拦截事件
5. **渐进迁移**：可以先把 `installFlow` 和 `bindSession` 迁出去，其他保持现状

## Consequences

### Positive
- 每个功能组件 < 200 行，可独立理解和修改
- 并行开发：不同的 Feature 互不冲突
- 测试隔离：每个 Feature 有独立的测试套件

### Negative
- 组件间通信需要显式设计（event bus / callback）
- 同个消息可能需要多个组件消费（需要明确优先级）
- 文件数量增加

### Risks
- Feature 栈的优先级管理可能变复杂
- Mitigation: 用 priority 字段排序栈，活跃的 Feature 优先消费按键
