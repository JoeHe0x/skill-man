# Record: make demo
# Shows: Skills list → Enter inspect → X toggle → Ctrl+A agent filter → Ctrl+D install → B bind → MCP tab

Output docs/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1400
Set Height 800
Set Theme "Catppuccin Mocha"
Set FontFamily "DejaVu Sans Mono"
Set Padding 32
Set TypingSpeed 40ms

Hide
Type "cd docs/demo/fixture"
Enter
Type "../../../bin/skill-man"
Enter
Show

# Wait for scan + full header
Sleep 3s

# --- Skills: inspect file tree (Enter) ---
Enter
Sleep 1.5s
Down@350ms 3
Sleep 600ms
Enter
Sleep 1.2s
Escape
Sleep 1s

# --- Skills: toggle disable (X), then restore ---
Set TypingSpeed 0
Type "x"
Set TypingSpeed 40ms
Sleep 1.8s
Set TypingSpeed 0
Type "x"
Set TypingSpeed 40ms
Sleep 1.2s

# --- Skills: agent filter (Ctrl+A) → pick agent → reset to All ---
Ctrl+A
Sleep 1.2s
Down@250ms 4
Sleep 600ms
Enter
Sleep 1s
Ctrl+A
Sleep 1s
Up@250ms 4
Sleep 500ms
Enter
Sleep 1s

# --- Skills: Search & Install (Ctrl+D) → search → browse → Esc cancel ---
Ctrl+D
Sleep 1.2s
Set TypingSpeed 50ms
Type "commit"
Set TypingSpeed 40ms
Sleep 600ms
Enter
Sleep 4s
Down@250ms 2
Sleep 800ms
Escape
Sleep 1s

# --- Skills: bind agents (B) → Space toggle → Escape cancel ---
Set TypingSpeed 0
Type "b"
Set TypingSpeed 40ms
Sleep 1.2s
Down@250ms 8
Sleep 500ms
Space
Sleep 900ms
Space
Sleep 700ms
Escape
Sleep 1s

# --- MCP tab ---
Tab
Sleep 1.8s
Down@350ms
Sleep 800ms

# MCP bind UI
Set TypingSpeed 0
Type "b"
Set TypingSpeed 40ms
Sleep 1.2s
Down@250ms 4
Sleep 500ms
Space
Sleep 900ms
Escape
Sleep 800ms

# MCP toggle disable
Set TypingSpeed 0
Type "x"
Set TypingSpeed 40ms
Sleep 1.5s
Set TypingSpeed 0
Type "x"
Set TypingSpeed 40ms
Sleep 1s

# Back to Skills
Shift+Tab
Sleep 1s

Ctrl+C
