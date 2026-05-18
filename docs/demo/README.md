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

## Record GIF

From the repository root (with `vhs`, `ffmpeg`, and `ttyd` available):

```bash
make demo
```

This writes `docs/demo.gif`. The script demonstrates:

1. **Enter** — inspect skill file tree (with preview)
2. **X** — toggle disable / enable (shown twice so the fixture stays unchanged)
3. **B** — agent bind UI (Space to toggle, Esc to cancel without saving)
4. **Tab** — switch to MCP and repeat bind / toggle

Commit the GIF when it looks good.

## Troubleshooting

| Error | Fix |
|-------|-----|
| `vhs: No such file or directory` | `go install github.com/charmbracelet/vhs@latest` and add `GOPATH/bin` to `PATH` |
| `ffmpeg is not installed` | `brew install ffmpeg` |
| `ttyd is not installed` | `brew install ttyd` |
| `libnss3.so: cannot open shared object` | Install apt packages listed above (WSL/Debian) |
| `Invalid command` in tape | Update vhs (`go install ...@latest`); see [demo.tape](./demo.tape) syntax |
