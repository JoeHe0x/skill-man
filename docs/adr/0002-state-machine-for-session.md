# ADR-0002: State Machine for Session State Management

## Status

Implemented

**2026-05-20**: `session.go` provides the transition table, `transitionTo()`, and
`exitState()`/`enterState()` hooks. State-specific key handling extracted from
`update.go` (796→98 lines) into per-state files. `dispatchKey()` routes to the
correct state handler.

**2026-05-21**: Key/msg handlers live in `internal/app/state/{listing,inspect,fallback,filtering,listfilter,installing}/`; Model bridges in `state_hosts.go`. Overlay states (bind, confirm, palette, help, prompt) stay in `feature/*` via `dispatchToFeatures`.

## Context

当前 `SessionState` 是一个 flat enum：

```go
const (
    stateHome SessionState = iota
    stateListing
    stateSearching
    stateInstalling
    stateConfirming
    stateHelpOverlay
    stateBindingAgent
    stateFilteringAgent
    stateInspecting
    stateCommandPalette
)
```

状态转换散落在 `update.go` 各处，以 `m.state = stateX` 直接赋值，没有任何校验：

```go
// model.go:457 — 切换 tab 时暴力重置
if m.state == stateInspecting || m.state == stateBindingAgent || ... {
    m.state = stateListing
    m.clearInstallFlow()
}

// update.go:478 — 搜索完成直接覆盖
m.state = stateListing
m.lastState = stateListing
```

问题：
- **无校验**：从 installing 直接跳到 helpOverlay 没人拦
- **副作用散落**：进入/退出每个状态的清理逻辑分散在各处
- **lastState 不可靠**：被随意覆盖，回退行为不定
- **多状态叠加**：`stateInstalling` 时 prompt 也能打开，实际形成了隐含的复合状态

## Decision Drivers

- 状态转换必须显式、可追踪
- 每个状态的进入/退出副作用必须集中管理
- 必须防止非法状态转换

## Decision

采用 **State 模式**，将状态建模为独立的 state handler，用显式的转换表管理：

```go
// State 是一个 session 状态
type State interface {
    ID() SessionStateID

    // OnEnter 状态进入时的副作用（设置 footer、调整 UI 等）
    OnEnter(m *appContext)
    // OnExit 状态退出时的清理
    OnExit(m *appContext)

    // HandleKey 处理按键，返回下个状态和命令
    HandleKey(m *appContext, msg tea.KeyMsg) (State, tea.Cmd)
    // HandleMsg 处理非按键消息
    HandleMsg(m *appContext, msg tea.Msg) (State, tea.Cmd)

    // AllowedTransitions 返回可以转换到的状态集合
    AllowedTransitions() []SessionStateID
}
```

**具体状态：**

```go
type ListingState struct{}

func (s *ListingState) ID() SessionStateID { return StateIDListing }
func (s *ListingState) OnEnter(m *appContext) {
    m.setFooterContext(fmt.Sprintf("%d %s · agents: %s", ...))
}
func (s *ListingState) OnExit(m *appContext) {
    // nothing to clean
}
func (s *ListingState) HandleKey(m *appContext, msg tea.KeyMsg) (State, tea.Cmd) {
    switch {
    case key.Matches(msg, keys.Quit):
        return nil, tea.Quit
    case key.Matches(msg, keys.Install):
        return GetState(StateIDInstalling), nil
    case key.Matches(msg, keys.Bind):
        return GetState(StateIDBinding), nil
    // ...
    }
    return s, nil
}
func (s *ListingState) AllowedTransitions() []SessionStateID {
    return []SessionStateID{StateIDHome, StateIDSearching, StateIDInstalling, StateIDBinding, StateIDInspecting, StateIDConfirming, StateIDHelpOverlay, StateIDCommandPalette, StateIDFilteringAgent}
}

// 状态注册表
var stateTransitions = map[SessionStateID][]SessionStateID{
    StateIDHome:       {StateIDListing, StateIDHelpOverlay, StateIDCommandPalette},
    StateIDListing:    {StateIDHome, StateIDSearching, StateIDInstalling, StateIDBinding, StateIDInspecting, StateIDConfirming, StateIDHelpOverlay, StateIDCommandPalette, StateIDFilteringAgent},
    StateIDInstalling: {StateIDListing, StateIDHome}, // ← 安装中只能回列表或退出
    // ...
}
```

**Model 中的 Update 简化为：**

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        next, cmd := m.currentState.HandleKey(m.appContext, msg)
        return m.transition(next, cmd)
    default:
        next, cmd := m.currentState.HandleMsg(m.appContext, msg)
        return m.transition(next, cmd)
    }
}

func (m *Model) transition(next State, cmd tea.Cmd) (tea.Model, tea.Cmd) {
    if next == nil {
        return m, tea.Quit
    }
    if next == m.currentState {
        return m, cmd
    }
    if !m.currentState.canTransitionTo(next.ID()) {
        return m, cmd // silently reject invalid transition
    }
    m.currentState.OnExit(m.appContext)
    m.currentState = next
    m.currentState.OnEnter(m.appContext)
    return m, cmd
}
```

## Rationale

1. **校验内置**：`canTransitionTo` 阻止非法转换，不需要散落的 if 检查
2. **副作用集中**：OnEnter/OnExit 将清理逻辑从 update.go 的 10 个不同分支集中到各状态文件
3. **可测试**：每个 State 独立测试按键响应
4. **与 Command 模式互补**：State 管 UI 交互流，Command 管业务操作执行
5. **当前 `handleXxxKeys` 方法已经是半成品 State**——每个方法就是一个状态的按键处理，只是没有显式建模

## Consequences

### Positive
- 所有状态转换显式、可追踪
- 新加状态只需新建一个文件 + 注册转换表
- 消除 `lastState` hack（回退目标由转换表 + 栈管理）
- 每个状态 Handler 可独立测试

### Negative
- 每个状态一个文件（~10 个新文件）
- 简单状态的 boilerplate 略多
- 需要定义 `appContext` 暴露共享的 UI 能力

### Risks
- 过度设计：如果状态交互简单，State 模式的优势不明显
- Mitigation: 先用 3 个状态（Listing、Installing、Binding）验证，其他状态保持 `m.state` 赋值兼容
