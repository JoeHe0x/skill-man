# skill-man domain glossary

Terms used in architecture reviews and ADRs. Prefer these names for new modules.

## Extension

Any manageable artifact on disk: **Skill**, **MCP server**, and (future) Hook / Sub-Agent. All implement `extension.Extension`.

## Skill

Agent skill backed by `SKILL.md`, installed under agent-specific directories. Scanned and mutated via `ExtensionManager[*Skill]`.

## MCP

Model Context Protocol server entries aggregated from JSON/TOML config files (Cursor, Claude, Codex, Windsurf, …). Mutated via `servicemcp.Manager`.

## Agent

A coding agent toolchain (Cursor, Claude Code, Codex, Windsurf). Drives scan paths, bind targets, and list filtering.

## Panel

One extension tab in the TUI (Skills, MCP). Owns list rows (`panel.Item`), scan, and preview for that extension type.

`panel.Core` (alias `panel.Panel`) is Bubble Tea–free; `panel.ScanCmd`, `panel.SyncPreviewCmd`, and `panel.ScanAllCmd` in `panel/tea.go` adapt scan/preview to `tea.Cmd` for the app layer.

## Bind

Linking an extension into an agent’s config layout (symlink for skills; config merge for MCP).

## Scan

Discovering extensions from disk and refreshing in-memory panel state (`panel.ScannedMsg`, `scanCoordinator`).

## Scope

`project` vs `global` installation scope (`extension.Scope`).

## Extension mutation

A use-case operation that changes disk or config: remove, toggle disable, add, init, update. Implemented in `internal/usecase/extension`; the TUI adapts via `internal/app/command`.

## Outcome

Result of an extension mutation (`usecase/extension.Outcome`): `Kind` (Skill vs MCP), `AffectedName`, `Message`, `Err`. Mapped to UI reselection in `mutation_lifecycle.go` (no `panel.Tab` in use case layer).

## Bind choice

One row in the agent bind dialog (`usecase/bind.Choice`): tracks `Initial` vs `Desired` bind state per agent or shared skills directory. Apply flows run through `usecase/bind.Binder` (not `app/command`).

## Feature host

Narrow interface (`confirmHost`, `installHost`, `bindHost`, …) between Bubble Tea features and `Model`. Features depend on hosts instead of reading arbitrary `Model` fields.

## TUI app layout (`internal/app`)

| Package | Role |
|---------|------|
| `feature/*` | Overlay features (install, bind, palette, help, prompt, confirm); registered on `Model.features`, consumed first in `Update` |
| `state/listing` | Home/listing/search key routing |
| `state/inspect` | Skill file-tree inspect keys |
| `state/filtering` | Agent filter overlay keys |
| `state/listfilter` | Inline `/` filter on main list |
| `state/installing` | Install wizard keys (when wizard open) |
| `state/fallback` | Resize, preview loaded, spinner, mutation fallthrough |
| `list` | Embedded `Pane` (main/agent lists, preview, tree); `bridge.go` syncs selection preview and reselect-by-name |
| `state_hosts.go` | Model adapters implementing `state/*` and `list.BridgeHost` host interfaces |
| `session` | Session state enum + `CanTransition` |

## Preview markdown

Service layers build markdown via `PreviewMarkdown` / `KeyPreviewMarkdown`; terminal rendering happens in `internal/app/panel` or `internal/render` (not in `service/*`).
