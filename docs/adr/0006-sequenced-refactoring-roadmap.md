# ADR-0006: Sequenced Refactoring Roadmap

## Status

In Progress (Phase 1-4 complete, Phase 5 deferred)

## Context

`internal/app/` 当前处于半重构状态。5 个 ADR (0001–0005) 在同一天提出，但全部**并行部分实施**，没有一个完成：

```
ADR-0001 (Command Pattern)       → command/ dir exists, 4 cmds, rest on Model
ADR-0002 (State Machine)         → session.go has table, handlers in update.go
ADR-0003 (Decompose God Model)   → feature/ dir exists, zero features registered
ADR-0004 (Panel Polymorphism)    → Panel interface exists, switch-on-kind everywhere
ADR-0005 (Plugin Architecture)   → nothing done
```

**结果:** 代码库比开始前更差——架构框架引入了额外的抽象层（feature dispatch、command interface、state transition table），但没有足够的实现来简化任何代码。实际逻辑仍在 `Model` 方法中，`update.go` 仍然是 795 行，`model.go` 仍然是 30+ 字段。

### 根本原因

1. **ADR 之间有依赖关系**——不按顺序做会产生死代码
2. **按基础设施优先的方式推进**（先建接口、注册表、分发器）而不是由具体用例驱动
3. **没有一个 ADR 被推动到完成**——每个都在第一步或第二步之后停止

### ADR 依赖图

```
0005 (Plugin Architecture)       ← 最高层，需要 ListItem 多态
  └── depends on: 0004, 0001
0004 (Panel Polymorphism)        ← 需要 ListItem 接口，Command 用于操作
  └── depends on: 0001, 0003
0003 (Decompose God Model)       ← 需要从 Model 中提取 Feature
  └── depends on: 0001, 0002
0002 (State Machine)             ← 需要将状态处理程序移出 update.go
  └── depends on: 0001
0001 (Command Pattern)           ← 基础层，无内部依赖
```

## Decision

**自底向上完成 ADR，一次一个。** 顺序和具体范围：

### Phase 1: 完成 ADR-0001 — Command 模式 (预计: 1-2 次提交)

**目标:** 将所有操作移出 Model，停止添加新的 Model 方法。

具体范围 — 为每个现有操作创建命令：

| Model 方法 | 新命令 |
|-------------|-------------|
| `removeSkillCmd()` | `RemoveSkill` (已经存在) |
| `toggleDisableSkillCmd()` | `ToggleDisableSkill` (已经存在) |
| `removeMCPKeyCmd()` | `RemoveMCPKey` (已经存在) |
| `mcpKeyMutationCmd()` | `MCPKeyMutation` (已经存在) |
| `updateSkillCmd()` | `UpdateSkill` |
| `updateAllSkillsCmd()` | `UpdateAllSkills` |
| `addSkillCmd()` | `AddSkill` |
| `initSkillCmd()` | `InitSkill` |

然后从 Model 中删除这些方法。`command/` 包不应再向 `app` 导入任何内容。

**完成标准:** 零个操作保留为 Model 方法。Model 对 `command` 包的唯一引用是通过 `Command` 接口。

### Phase 2: 完成 ADR-0002 — 状态机 (预计: 2-3 次提交)

**目标:** 将按键处理程序移出 `update.go`。

将 `update.go` 中的 ~20 个 `handleXxxKeys` 方法移动为每个状态的文件：

```
internal/app/state/
├── state.go          # State 接口 + 注册表
├── listing.go        # handleListKeys, handleListEnter, ...
├── installing.go     # handleInstallKeys, handleInstallUpdate, ...
├── binding.go        # handleBindKeys, handleBindUpdate, ...
├── inspecting.go     # handleInspectKeys, handleTreeUpdate, ...
├── confirming.go     # handleConfirmKeys
├── filtering.go      # handleAgentFilterKeys
├── palette.go        # handleCommandPaletteKeys
└── help.go           # handleHelpKeys
```

**完成标准:** `update.go` 降到 100 行以下。每个按键组合的状态特定处理存在于其状态文件中。

### Phase 3: 完成 ADR-0003 — Feature 组件 (预计: 3-4 次提交)

**目标:** 将大功能（安装流程、绑定、命令面板）提取为 Feature。

此时 ADR-0001 和 ADR-0002 已经完成，因此 Feature 可以干净地使用 Command 模式进行变异，并依赖正确建模的状态机进行 UI 状态转换。

```
internal/app/feature/install/     ← 从 install_flow.go 迁移
internal/app/feature/bind/        ← 从 bind.go 迁移
internal/app/feature/inspect/     ← 从 tree.go 迁移
internal/app/feature/confirm/     ← 新：通用确认对话框
internal/app/feature/palette/     ← 从 command_palette.go 迁移
internal/app/feature/help/        ← 从 help_overlay.go 迁移
```

**完成标准:** 零个 Feature 逻辑作为 Model 方法。Model 降到 15 个字段以下。

### Phase 4: 完成 ADR-0004 — Panel 多态 (预计: 2-3 次提交)

**目标:** 消除所有 `if kind == itemKindXxx` 分支。

在 app 层用 `ListItem` 接口替换 `listItem` tagged union。操作（Inspect、Remove、Bind 等）成为 ListItem 多态方法，由每个 Panel 类型实现。

**完成标准:** `update.go`、`model.go`、`bind.go`、`items.go` 中零个 switch-on-kind 分支。

### Phase 5: ADR-0005 — 插件架构 (未来)

推迟。只有在至少有一种新的扩展类型需要添加时才实施。有了 0001–0004 完成后建立的模式，0005 的实施将是直接的。

## 备选方案

### Option 2: 并行继续，缓慢修复

保持当前方式：随时可以时增量移动逻辑。这就是导致当前状态的策略。**拒绝。**

### Option 3: 高优先实施最薄弱的 ADR-0003

直接跳到 Feature 分解，将其他 ADR 作为嵌套步骤。具有吸引力，因为 Feature 是用户最明显的地方，但在没有完成 Command 模式（用于 Feature 触发变异）和状态机（用于 UI 转换）的情况下无法干净地工作。

### Option 4: 暂停重构，合并到 Model 中

回退 `command/`、`feature/`、`session.go` 的更改，接受上帝 Model 作为设计选择。**拒绝**——当前架构在纸面上是合理的；执行才是问题。

## 为什么选择自底向上

1. **每个 Phase 都减少混乱**——不会增加未使用的抽象
2. **ADR-0001 首先完成** 因为它是严格的基础：Command 模式不需要任何其他 ADR，但所有其他 ADR 都需要它
3. **Phase 2 依赖于 Phase 1**：状态处理程序需要发出命令
4. **Phase 3 依赖于 Phase 1 和 2**：Feature 使用 Command 进行变异，并使用状态机进行 UI 转换
5. **Phase 4 依赖于 Phase 1 和 3**：多态 ListItem 操作由 Feature 组件调用，触发 Command 执行

## 指导原则

- **一次一个 Phase。** 不部分实施后续 Phase。
- **每个 Phase 减少总代码量。** 目标不仅仅是移动代码；而是删除 boilerplate。
- **删除已失效的代码。** 如果某个抽象（feature/、command/ 骨架）在 Phase 完成后被废弃——删除它。
- **之后统一测试。** 不要为不会存活的过渡性代码编写测试。在 Phase 完成、API 稳定后编写测试。

## Consequences

### Positive
- 可预测的进度：每个 Phase 有明确的完成标准
- 不会再有半成品的抽象
- 每个 Phase 后代码库包含更少的 bug 容纳点

### Negative
- 在 Phase 3 之前不添加新功能
- 每个 Phase 在审查期间阻塞后续工作
- 如果未来扩展类型被宣布，可能会有时间压力以更快地达到 Phase 5

### Risks
- 范围蔓延：每个 Phase 可能变成"摸一下一切"的重构
- 缓解：完成标准被严格定义。当标准满足时 Phase 结束——不完美，不全面。
