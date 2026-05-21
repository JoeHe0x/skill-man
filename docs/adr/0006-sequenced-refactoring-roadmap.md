# ADR-0006: Sequenced Refactoring Roadmap

## Status

Phase 1–4 complete (with documented caveats). Phase 5 deferred until a third extension type is needed.

## Context

`internal/app/` 曾在 2026-05-20 处于**半重构**状态（见下方历史快照）。按本 ADR 自底向上完成后，核心框架已落地；剩余问题从「死代码抽象」转为「接口仍浅、状态仍集中在 Model」。

### Historical snapshot (2026-05-20, pre-roadmap)

```
ADR-0001 (Command Pattern)       → command/ dir exists, 4 cmds, rest on Model
ADR-0002 (State Machine)         → session.go has table, handlers in update.go
ADR-0003 (Decompose God Model)   → feature/ dir exists, zero features registered
ADR-0004 (Panel Polymorphism)    → Panel interface exists, switch-on-kind everywhere
ADR-0005 (Plugin Architecture)   → nothing done
```

当时风险：框架抽象先行、实现滞后。该风险已通过 Phase 1–4 顺序执行缓解。

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

## Post-roadmap audit (2026-05-21)

Measured against Phase completion criteria and the **deletion test** (see improve-codebase-architecture skill).

### Phase outcomes

| Phase | ADR | Verdict | Evidence |
|-------|-----|---------|----------|
| 1 | 0001 Command | **Done** | 8 commands in `internal/app/command/`; mutations via `runCommand`; no `*Cmd()` methods left on Model |
| 2 | 0002 State machine | **Done (deepening)** | `update.go` ~70 lines; `state/{listing,inspect,fallback,filtering,listfilter,installing}/` + single `state_hosts.go`; `session.go` transition table |
| 3 | 0003 Features | **Mostly done** | 7 features in `feature/*`; list shell in `list/` (`Pane`, `bridge.go`); Model delegates remain for session/panels |
| 4 | 0004 Panel polymorphism | **Done (pragmatic)** | Unified `panel.Item`; app `itemKind` gone; Kind switches in `panel/item.go` + listing handlers |
| 5 | 0005 Plugin | **Deferred** | `newPanelRegistry()` still hardcodes Skill + MCP panels |

### Recommended deepening (ordered by leverage)

1. ~~**Feature-owned state**~~ — **Done (2026-05-21):** bind, confirm, prompt, `cmdPalette`, `helpScreen`.
2. ~~**Item action depth**~~ — **Done (2026-05-21):** `panel/item_effect.go` + `item_ops.go`.
3. ~~**Mutation lifecycle module**~~ — **Done (2026-05-21):** `mutation_lifecycle.go` (`runCommand`, `applyMutationResult`).
4. ~~**Scan coordinator**~~ — **Done (2026-05-21):** `panel.ScannedMsg`, `scan_coordinator.go`, unified `handleScanned`.
5. ~~**MCP parser registry**~~ — **Done (2026-05-21):** `parse_registry.go` with `configFileParsers` map.
6. **ADR-0005** — only when adding Hook / Sub-Agent / third tab; do not pre-build.
7. ~~**Core struct**~~ — **Done (2026-05-21):** embedded `Core` in `core.go` (size, paths, status/footer, agent filter, `scanCoordinator`).
8. ~~**List/preview shell**~~ — **Done (2026-05-21):** `list_pane.go` (`listPane`: lists, preview, tree; selection/preview helpers).
9. ~~**Install feature depth**~~ — **Done (2026-05-21):** `feature_install.go` (wizard, background install, `handleCompleted`); removed `install_bridge.go`.

### Domain glossary

No `CONTEXT.md` at repo root yet. Domain terms live in [skill-man-detailed-design.md](../skill-man-detailed-design.md) (Skill, MCP, Agent, Extension, Panel). New module names from future refactors should be added to `CONTEXT.md` when introduced.

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
