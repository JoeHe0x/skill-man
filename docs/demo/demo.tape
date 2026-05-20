# Short README demo: isolated HOME, launch hidden → Skills → toggle → MCP tab
# Record: make demo

Output docs/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1920
Set Height 1080
Set Theme "TokyoNight"
Set FontFamily "DejaVu Sans Mono"
Set Padding 12
Set TypingSpeed 40ms
Set Framerate 20

Hide
Type "cd docs/demo/fixture"
Enter
# Only fixture skills — avoid scanning ~/.agents/skills on the host
Type "HOME=$PWD/.record-home ../../../bin/skill-man"
Enter
Sleep 2.2s
Show

# Skills: file tree peek
Enter
Sleep 700ms
Escape
Sleep 300ms

# Toggle commit-helper (↓ skips api-docs); wait for rescan before restore
Down@200ms 1
Sleep 300ms
Type "x"
Sleep 1.2s
Type "x"
Sleep 300ms

# MCP tab
Tab
Sleep 1s
Down@200ms 2
Sleep 500ms

Ctrl+C
