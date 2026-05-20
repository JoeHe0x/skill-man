# ADR-0001: Command Pattern for Extension Operations

## Status

Implemented

**2026-05-20**: All 8 extension operations are implemented as `command.Cmd` structs:
`RemoveSkill`, `ToggleDisableSkill`, `AddSkill`, `InitSkill`, `UpdateSkill`,
`UpdateAllSkills`, `RemoveMCPKey`, `ToggleDisableMCPKey`. Zero operation logic
remains on Model — every mutation flows through `runCommand(cmd command.Cmd)`.
Model no longer imports `sync` or `errgroup`.

## Context

当前 `internal/app/model.go` 中，所有扩展操作（remove、disable、bind、update、install）都以
`Model` 的方法存在：

```go
func (m *Model) removeSkillCmd(skill *skilldomain.Skill) tea.Cmd
func (m *Model) toggleDisableSkillCmd(skill *skilldomain.Skill) tea.Cmd
func (m *Model) removeMCPKeyCmd(members []*mcpdomain.Server) tea.Cmd
func (m *Model) updateSkillCmd(skill *skilldomain.Skill) tea.Cmd
func (m *Model) updateAllSkillsCmd() tea.Cmd
func (m *Model) addSkillCmd(source string) tea.Cmd
func (m *Model) initSkillCmd(name string) tea.Cmd
func (m *Model) mcpKeyMutationCmd(members []*mcpdomain.Server, apply func(*mcpdomain.Server) error, label string) tea.Cmd
```

问题：
- **Model 持续膨胀**：每加一个操作就要在 Model 上加一个方法
- **无法复用**：操作逻辑和 UI 生命周期（flash、rescan、reselect）耦合
- **测试困难**：测试一个操作需要构造完整的 Model
- **违反开闭原则**：加新操作必须改 Model

## Decision Drivers

- 操作必须可独立测试
- Model 不应知道每个操作的具体实现
- 操作的生命周期（验证 → 执行 → 后处理）应该标准化

## Decision

采用 **Command 模式**，将每个操作封装为独立命令对象：

```go
// Command 一次幂等的扩展操作
type Command interface {
    // Label 返回操作名称（用于 footer flash 和日志）
    Label() string
    // Validate 在执行前校验操作是否合法
    Validate(m *appContext) error
    // Execute 执行操作，返回结果
    Execute(ctx context.Context) Result
    // After 在扫描完成后执行 UI 收尾（reselect、flash）
    After(m *appContext) tea.Cmd
}

type Result struct {
    AffectedName string
    Message      string
    Err          error
    TargetTab    panel.Tab
}
```

**具体命令示例：**

```go
type RemoveSkillCommand struct {
    skill   *skilldomain.Skill
    manager manager.ExtensionManager[*skilldomain.Skill]
}

func (c *RemoveSkillCommand) Validate(m *appContext) error {
    if c.skill == nil {
        return errors.New("no skill selected")
    }
    return nil
}

func (c *RemoveSkillCommand) Execute(ctx context.Context) Result {
    err := c.manager.Remove(ctx, c.skill, /*...*/)
    if err != nil {
        return Result{Err: err}
    }
    return Result{
        AffectedName: c.skill.GetName(),
        Message:      fmt.Sprintf("removed %s", c.skill.GetName()),
    }
}

func (c *RemoveSkillCommand) After(m *appContext) tea.Cmd {
    // flash + rescan + reselect
    return tea.Batch(
        m.flashFooter(c.Label()),
        tea.Sequence(
            m.scanAllCmd(),
            func() tea.Msg { return reselectMsg{kind: "skill", name: c.skill.GetName()} },
        ),
    )
}
```

**Model 中的调用简化为：**

```go
func (m *Model) handleRemoveSelected() (tea.Model, tea.Cmd) {
    cmd := NewRemoveSkillCommand(selected.skill, m.skillManager)
    m.pendingCommand = cmd
    m.state = stateConfirming
    return m, nil
}

func (m *Model) handleConfirm() (tea.Model, tea.Cmd) {
    cmd := m.pendingCommand
    m.pendingCommand = nil
    return m, m.executeCommand(cmd)
}

func (m *Model) executeCommand(cmd Command) tea.Cmd {
    return func() tea.Msg {
        return cmd.Execute(context.Background())
    }
}
```

## 备选方案

### Option 2: 保持现状，用辅助函数提取
只把重复的逻辑（flash+rescan+reselect）提取成 helper 函数，不改 Model 结构。
- **优点**: 改动最小
- **缺点**: 不解决 Model 膨胀的根本问题

### Option 3: 每个操作独立 Bubble Tea Model
把 remove flow、bind flow 等做成独立的 `tea.Model`。
- **优点**: 完全解耦
- **缺点**: Bubble Tea 的嵌套 Model 通信复杂，对当前问题过度设计

## Rationale

选 Command 模式因为：
1. **轻量**：只是一个接口 + 结构体，不引入框架依赖
2. **可组合**：Command 可以包装（带确认的 Command、带日志的 Command）
3. **可测试**：每个 Command 可以独立单元测试
4. **渐进迁移**：可以从最复杂的操作开始改，不影响其他代码
5. **MCP 的 `mcpKeyMutationCmd` 已经是半个 Command**——它接受一个 `apply func` 作为执行逻辑，这就是 Command 模式的雏形

## Consequences

### Positive
- Model 减少 ~200 行方法定义
- 每个操作可独立测试
- 加新操作只需新建一个文件实现 Command 接口
- 操作的 validate → execute → after 生命周期显式化

### Negative
- 文件数量增加（每个 Command 一个文件）
- 需要额外的 appContext 来暴露 scan/flash 等 UI 能力
- Command 接口的 After 签名依赖 bubbletea，不是纯领域逻辑

### Risks
- 如果 Command 接口设计得太重，会成为新瓶颈
- Mitigation: 先抽 3 个 Command 验证接口设计，确认够用后再全量迁移
