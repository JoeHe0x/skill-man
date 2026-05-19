# Demo recording

Fixture project under `fixture/` includes sample Skills and Cursor MCP config for recordings.

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

The demo tape uses `FontSize 14`, `1920×1080`, `Padding 12` so the GIF shows a fuller skills list, not a zoomed-in few items.

## Color (why the GIF can look washed out)

GIF uses ~256 colors per frame, so lipgloss ANSI colors get quantized and can look paler than your terminal. The tape uses **Tokyo Night** (not Catppuccin Mocha) for stronger contrast in recordings.

## Record

From the repository root (with `vhs`, `ffmpeg`, and `ttyd` available):

```bash
make demo
```

This writes `docs/demo.gif`. The script demonstrates:

1. **Enter** — inspect skill file tree (with preview)
2. **X** — toggle disable / enable (shown twice so the fixture stays unchanged)
3. **Ctrl+A** — agent filter dialog (only agents with a local skills directory; Enter to apply, then reset to All)
4. **Ctrl+D** — Search & Install dialog (type a keyword, search skills.sh, browse results, Esc to cancel without installing)
5. **B** — agent bind UI (Space to toggle, Esc to cancel without saving)
6. **Tab** — switch to MCP and repeat bind / toggle

**Ctrl+D** needs outbound network access to skills.sh during recording. If search fails, you still get the install dialog and error state in the GIF; re-run when online for results.

Commit the GIF when it looks good.

## Troubleshooting

| Error | Fix |
|-------|-----|
| `vhs: No such file or directory` | `go install github.com/charmbracelet/vhs@latest` and add `GOPATH/bin` to `PATH` |
| `ffmpeg is not installed` | `brew install ffmpeg` |
| `ttyd is not installed` | `brew install ttyd` |
| `libnss3.so: cannot open shared object` | Install apt packages listed above (WSL/Debian) |
| `Invalid command` in tape | Update vhs (`go install ...@latest`); see [demo.tape](./demo.tape) syntax |
