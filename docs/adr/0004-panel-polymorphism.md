# ADR-0004: Eliminate Kind Switching via Panel Polymorphism

## Status

Implemented (pragmatic scope)

**2026-05-20**: `app.listItem` type deleted — unified with `panel.Item`. Action methods
(CanInspect, InspectTarget, CanDisable, DisableTarget, etc.) moved to `panel.Item`.
`panel_bridge.go` conversion functions removed; only `panelToListItems` slice conversion
remains. Switch-on-kind now confined to `panel.Item` methods in `panel/item.go`. The
full ListItem interface from the original proposal was not needed — the consolidated
tagged union achieved the elimination of conversion code and scattered switches.

## Context

当前 `panel.Panel` 接口已经存在，但 app 层仍大量使用 **switch-on-kind** 反模式：

```go
// update.go — 散落各处的 kind 判断
if selected.kind == itemKindSkill {
    m.lastState = m.state
    m.state = stateInspecting
    m.tree.setRoot(selected.skill.Path)
}
if selected.kind == itemKindMCP && len(selected.mcpMembers) > 0 {
    return m, m.activePanel().SyncPreview(listItemToPanel(selected), width, &m.previewGen)
}

// model.go
if selected.kind == itemKindCommand { ... }
if selected.kind == itemKindSkill { ... }

// bind.go
if selected.kind == itemKindMCP { ... }
if selected.kind != itemKindSkill { ... }
```

`listItem` 是一个 tagged union：

```go
type listItem struct {
    kind       itemKind
    skill      *skilldomain.Skill   // nil for non-skill
    mcp        *mcpdomain.Server    // nil for non-mcp
    mcpMembers []*mcpdomain.Server  // nil for non-mcp-key
    command    commandSpec          // zero for non-command
    // ...
}
```

问题：
- **违反开闭原则**：加一种新的 Item 类型（如 Hook、SubAgent）需要改所有 switch
- **空指针风险**：访问 `selected.skill.Path` 前必须确认 `kind == itemKindSkill`
- **转换损耗**：`panelToListItem` 和 `listItemToPanel` 双向转换，信息在两种表示间复制
- **Panel 抽象泄漏**：Panel 接口存在，但 app 层做了 Panel 该做的分发

## Decision

**扩展 Panel 接口，让 Panel 处理自己的 Item 语义**，消除 app 层的 switch-on-kind：

```go
// ListItem 替代现有的 listItem tagged union
// 封装面板相关的数据和操作
type ListItem interface {
    // 显示
    Title() string
    Description() string
    Meta() string

    // 操作（多态分发，不再需要 switch-on-kind）
    Inspect(appContext) tea.Cmd
    HandleRemove(appContext) tea.Cmd
    HandleDisable(appContext) tea.Cmd
    HandleBind(appContext) tea.Cmd

    // 过滤
    MatchesAgent(agentIDs []string) bool
    MatchesSearch(query string) bool
}
```

**具体实现：**

```go
// skillItem 实现 ListItem
type skillItem struct {
    skill   *skilldomain.Skill
    manager manager.ExtensionManager[*skilldomain.Skill]
}

func (si *skillItem) Inspect(ctx appContext) tea.Cmd {
    ctx.OpenTree(si.skill.Path)
    return nil
}
func (si *skillItem) HandleRemove(ctx appContext) tea.Cmd {
    ctx.RequestConfirm(NewRemoveSkillCommand(si.skill, si.manager))
    return nil
}
func (si *skillItem) HandleDisable(ctx appContext) tea.Cmd {
    return NewToggleDisableSkillCommand(si.skill, si.manager).AsTeaCmd()
}

// mcpItem 实现 ListItem
type mcpItem struct {
    key     string
    members []*mcpdomain.Server
    manager *servicemcp.Manager
}

func (mi *mcpItem) Inspect(ctx appContext) tea.Cmd {
    return ctx.SyncPreview(mi.key, mi.members)
}
func (mi *mcpItem) HandleRemove(ctx appContext) tea.Cmd {
    ctx.RequestConfirm(NewRemoveMCPKeyCommand(mi.members, mi.manager))
    return nil
}
```

**Panel 接口扩展：**

```go
type Panel interface {
    // 已有
    Name() string
    Tab() Tab
    Capabilities() Capabilities

    // 新增：返回 ListItem 而不是 domain 实体
    ListItems(agentIDs []string) []ListItem
    SearchItems(query string, agentIDs []string) []ListItem

    // 预览
    SyncPreview(item ListItem, width int, gen *int) tea.Cmd
    StaticPreview() string

    // 操作
    HandleAction(item ListItem, action Action, ctx appContext) tea.Cmd

    // 扫描
    ScanCmd(projectRoot, home string, agents []agent.Agent) tea.Cmd
    ApplyScan(msg interface{})
}
```

**app 层调用变为：**

```go
func (m *Model) handleInspectSelected() (tea.Model, tea.Cmd) {
    item, ok := m.selectedListItem()
    if !ok {
        return m, nil
    }
    if !m.activePanel().Capabilities().Inspect {
        m.setFooterContext("Inspect is not available for this tab")
        return m, nil
    }
    return m, item.Inspect(m.appContext)  // 多态分发，无 switch
}
```

## Rationale

1. **多态替代 switch**：ListItem 接口的每个方法替代一个 switch 分支
2. **开放-封闭**：加 Hook/SubAgent 类型时，只需实现 ListItem 接口，不碰任何现有代码
3. **消除空指针**：每种 Item 类型有自己的 struct，字段总是有效的
4. **与现有 Panel 抽象对齐**：Panel 已经在做 Scan/Preview/Capabilities 的分发，ListItem 把分发延伸到操作层
5. **`panel.Item` 已经是这个方向的雏形**——它有 Kind 字段来区分类型，只是没有把行为也封装进去

## 备选方案

### Option 2: Visitor 模式
在 ListItem 上定义 `Accept(Visitor)`，操作都做成 Visitor。

```go
type ListItemVisitor interface {
    VisitSkill(*skilldomain.Skill) tea.Cmd
    VisitMCP(*mcpdomain.Server, []*mcpdomain.Server) tea.Cmd
}
```

- **优点**: 操作集中在一个 Visitor 里，方便批量处理
- **缺点**: 加新操作 = 改 Visitor 接口 + 改所有实现；Go 不支持方法重载，VisitXxx 名字不统一

### Option 3: 纯函数式分发
保持 `listItem` tagged union，但把 switch 集中到一个 `dispatch` 函数。

- **优点**: 不用改 listItem
- **缺点**: 仍然违反开闭原则，只是藏起来了

## Rationale for choosing ListItem 多态 over Visitor

Visitor 适合"操作频繁变化，Item 类型稳定"的场景。这里 Item 类型会随着扩展而增长（Hook、SubAgent、未来更多），操作相对固定（Inspect、Remove、Disable、Bind），所以把操作绑在 Item 上更合适。

## Consequences

### Positive
- 消除 app 层 ~30 处 `if kind == itemKindXxx` 分支
- 加新扩展类型时，只需新文件实现 ListItem 接口
- 消除 `listItem` 和 `panel.Item` 的双向转换

### Negative
- ListItem 接口含 bubbletea 依赖（返回 `tea.Cmd`），不是纯领域层
- 每个 Item 类型需要持有 Manager 引用（或从 appContext 获取）

### Risks
- ListItem 接口膨胀：随着操作增多，接口越来越大
- Mitigation: 如果未来超过 8 个方法，分离成 ListItem + Actionable 两个接口
