# ADR-0005: Plugin Architecture for Extension Types

## Status

Not Implemented

**2026-05-20**: `newPanelRegistry()` still hardcodes two types (Skill and MCP).
No `ExtensionPlugin` interface or `PluginRegistry` exists. Adding a new extension
type would still require changes across 5 packages.

## Context

当前系统支持两种扩展类型：**Skill** 和 **MCP**。但架构已经暗示会有更多：

- Sub-Agent（`ExtensionManager` 文档中提到）
- Hook（`ExtensionManager` 文档中提到）
- 未来可能的类型：Rule、Prompt、Command 等

当前每种新类型需要：

1. `domain/<type>/` — 领域实体
2. `service/<type>/` — 扫描、安装、预览逻辑
3. `service/manager/` — 手动注册 `ScanStrategy`
4. `internal/app/panel/` — 手动添加 Panel 实现
5. `internal/app/` — 手动处理 listItem kind 分发

**每个新类型要改 5 个包**。这不是插件架构。

当前 `panel.Registry` 的初始化：

```go
func newPanelRegistry() *panel.Registry {
    return panel.NewRegistry(
        panel.SkillDeps{Manager: manager.NewManager[*skilldomain.Skill](skill.SkillScanStrategy{})},
        panel.MCPDeps{},
    )
}
```

每加一个类型就要改这个函数。

## Decision Drivers

- 新扩展类型应通过注册机制添加，不改已有代码
- 每种类型的扫描、预览、CRUD 应自包含
- Panel、Command、Manager 等应统一从同一个注册入口获取

## Decision

采用 **插件注册模式**，每种扩展类型注册为一组能力：

```go
// ExtensionPlugin 描述一种扩展类型的完整能力
type ExtensionPlugin struct {
    // 元数据
    ID          string // "skill", "mcp", "hook", "subagent"
    DisplayName string
    TabLabel    string // Tab 标题

    // 工厂
    NewManager  func() interface{} // 返回 ExtensionManager[T]
    NewScanner  func() Scanner
    NewInstaller func() Installer

    // 列表和预览
    ListItems    func(agentIDs []string) []ListItem
    SearchItems  func(query string, agentIDs []string) []ListItem
    SyncPreview  func(item ListItem, width int, gen *int) tea.Cmd
    DefaultPreview func() string

    // 能力声明
    Capabilities Capabilities
}

// Registry 管理所有注册的插件
type PluginRegistry struct {
    plugins []ExtensionPlugin
    byID    map[string]*ExtensionPlugin
    byTab   map[TabID]*ExtensionPlugin
}

func (r *PluginRegistry) Register(plugin ExtensionPlugin) {
    r.plugins = append(r.plugins, plugin)
    r.byID[plugin.ID] = &r.plugins[len(r.plugins)-1]
    // 自动分配 Tab
    r.byTab[TabID(len(r.byTab))] = &r.plugins[len(r.plugins)-1]
}

func (r *PluginRegistry) Plugins() []ExtensionPlugin { return r.plugins }
func (r *PluginRegistry) TabCount() int               { return len(r.byTab) }
```

**注册时机（init 或 main 组装）：**

```go
// cmd/skill-man/plugins.go
func RegisterAllPlugins(registry *PluginRegistry) {
    registry.Register(skill.NewPlugin())
    registry.Register(mcp.NewPlugin())
    // 未来：
    // registry.Register(hook.NewPlugin())
    // registry.Register(subagent.NewPlugin())
}

// internal/domain/skill/plugin.go
func NewPlugin() ExtensionPlugin {
    return ExtensionPlugin{
        ID:          "skill",
        DisplayName: "Skills",
        TabLabel:    "skills",
        NewManager: func() interface{} {
            return manager.NewManager[*Skill](SkillScanStrategy{})
        },
        Capabilities: Capabilities{
            Inspect:       true,
            Bind:          true,
            Update:        true,
            Find:          true,
            Init:          true,
            Disable:       true,
            SearchInstall: true,
            Remove:        true,
        },
        // ... ListItems, SearchItems, SyncPreview 等方法
    }
}
```

**app.New 变为：**

```go
func New(cwd, home string) *Model {
    registry := NewPluginRegistry()
    RegisterAllPlugins(registry)

    m := Model{
        plugins: registry,
        tabs:    buildTabsFromPlugins(registry),
        // ...
    }
    return &m
}
```

## Rationale

1. **开放的扩展系统**：加新类型 = 新建一个包 + 调用一次 `registry.Register()`
2. **统一入口**：Panel、Manager、Scanner 都从同一个 Plugin 获取，不再需要分别构造
3. **自然对齐现有设计**：当前 `panel.SkillDeps` 和 `panel.MCPDeps` 已经是 plugin 的胚胎，只是没有抽象成接口
4. **Tab 自动管理**：Plugin 注册顺序决定 Tab 顺序，不需要手动管理

## 当前设计与目标的差距

| 当前 | 目标 |
|------|------|
| `panel.SkillDeps{Manager: ...}` 手动传入 | `skill.NewPlugin()` 自包含所有依赖 |
| `panel.MCPDeps{}` 空壳 | MCP 也有自己的 Manager 和完整 Plugin |
| `newPanelRegistry()` 硬编码两种类型 | `PluginRegistry.Register()` 注册任意类型 |
| switch-on-kind 手动分发 | Panel 从 Plugin 获取 Strategy，多态调用 |

## Consequences

### Positive
- 新扩展类型只需新建一个包，不改任何框架代码
- Tab 数量、顺序由注册决定
- 每个 Plugin 可独立测试（mock registry 即可）
- 为未来的社区插件机制打基础

### Negative
- 引入新的抽象层，初期理解成本
- 如果最终只有 2 种类型，过度设计
- `NewManager` 返回 `interface{}` 丢失类型安全

### Risks
- 泛型擦除：`NewManager` 返回 `interface{}` 需要 type assertion
- Mitigation: Plugin 提供类型化的 `Manager[T]()` 方法，用 Go 泛型约束

## 与设计模式的对应

| 模式 | 应用 |
|------|------|
| **Strategy** | `ScanStrategy` — 已有，每种类型的扫描逻辑 |
| **Factory Method** | `NewManager()` / `NewScanner()` — 创建类型对应的服务 |
| **Registry** | `PluginRegistry` — 运行时发现所有类型 |
| **Template Method** | `genericManager[T]` — 已有，共用 CRUD 骨架 |
