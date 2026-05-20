# Demo recording

Fixture project under `fixture/` includes sample Skills and Cursor MCP config for recordings.

**Recording uses `HOME=fixture/.record-home`** so the GIF only lists fixture skills (not your real `~/.agents/skills`). Twelve sample skills under `fixture/.agents/skills/` — all **enabled** (`SKILL.md`) so toggle demos do not error on pre-disabled entries.

## Prerequisites

Install [vhs](https://github.com/charmbracelet/vhs) and its runtime dependencies:

```bash
go install github.com/charmbracelet/vhs@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`.

**Linux / WSL:**

```bash
brew install ffmpeg ttyd
# Chromium libs for vhs (first run downloads a browser; on Debian/Ubuntu also run):
sudo apt-get install -y libnss3 libnspr4 libatk1.0-0 libatk-bridge2.0-0 \
  libcups2 libdrm2 libxkbcommon0 libgbm1 libasound2t64
```

**macOS:**

```bash
brew install ffmpeg ttyd
```

Then build the app:

```bash
make build
```

## Recording density (why the UI looked “huge”)

VHS `Set Width` / `Set Height` are the **video frame** size in pixels. How much of skill-man fits on screen is driven by **terminal rows × columns**, which shrink when `FontSize` or `Padding` is large.

| Setting | Role |
|---------|------|
| `FontSize` | Main knob — lower = more list rows visible (demo uses 14) |
| `Width` / `Height` | Larger frame — more cells at the same font size |
| `Padding` | Border around the terminal — lower = less wasted space |

The demo tape uses `FontSize 14`, `1920×1080`, `Framerate 20`. Launch + scan run **under `Hide`** (`Sleep 2.2s` before `Show`) so the GIF opens on a loaded Skills list. If the first frame is still empty on a slow machine, bump to `2.5s` in [demo.tape](./demo.tape).

## Color (why the GIF can look washed out)

GIF uses ~256 colors per frame, so lipgloss ANSI colors get quantized and can look paler than your terminal. The tape uses VHS theme **TokyoNight** (not Catppuccin Mocha) for stronger contrast in recordings.

## Record

From the repository root (with `vhs`, `ffmpeg`, and `ttyd` available):

```bash
make demo
```

This writes `docs/demo.gif` (~15s, no network). The short script shows:

1. **Enter** — peek skill file tree, **Esc** back
2. **↓** then **X** twice on `commit-helper` (skips `api-docs`; 1.2s between — waits for disk + rescan)
3. **Tab** — MCP tab, browse one server

To record a longer walkthrough (agent filter, install, bind), extend [demo.tape](./demo.tape) locally — keep the committed GIF short for README load time.

Commit the GIF when it looks good.

## Troubleshooting

| Error | Fix |
|-------|-----|
| `vhs: No such file or directory` | `go install github.com/charmbracelet/vhs@latest` and add `GOPATH/bin` to `PATH` |
| `ffmpeg is not installed` | `brew install ffmpeg` |
| `ttyd is not installed` | `brew install ttyd` |
| `libnss3.so: cannot open shared object` | Install apt packages listed above (WSL/Debian) |
| `Invalid command` in tape | Update vhs (`go install ...@latest`); see [demo.tape](./demo.tape) syntax |
| Red error in status bar during **X** | Second **X** before rescan finishes, or toggling a pre-disabled skill; tape waits 1.2s and uses isolated `HOME` — re-run `make demo` |
| List shows dozens of skills | Recording without `HOME=.record-home` — scans your real home directory |
