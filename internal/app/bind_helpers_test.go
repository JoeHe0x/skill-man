package app

import (
	"os"
	"testing"
)

func writeMCPJSON(t *testing.T, path, key, command string, args []string) {
	t.Helper()
	content := `{"mcpServers":{"` + key + `":{"command":"` + command + `","args":[`
	for i, a := range args {
		if i > 0 {
			content += ","
		}
		content += `"` + a + `"`
	}
	content += `]}}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
