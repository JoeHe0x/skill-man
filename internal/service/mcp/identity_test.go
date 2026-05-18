package mcp

import "testing"

func TestInferImplementationNameFromNPX(t *testing.T) {
	t.Parallel()

	name := InferImplementationName("npx", []string{
		"-y", "@modelcontextprotocol/server-filesystem", "/tmp/workspace",
	}, "")
	if name != "server-filesystem" {
		t.Fatalf("got %q, want server-filesystem", name)
	}
}

func TestInferImplementationNameFallsBackToConfigKey(t *testing.T) {
	t.Parallel()

	if got := InferImplementationName("echo", []string{"hello"}, ""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestWorkspaceRootFromArgs(t *testing.T) {
	t.Parallel()

	root := WorkspaceRootFromArgs([]string{
		"-y", "@modelcontextprotocol/server-filesystem", "/mnt/c/Code/skill-man",
	})
	if root != "/mnt/c/Code/skill-man" {
		t.Fatalf("got %q", root)
	}
}

func TestInferImplementationNameFromURL(t *testing.T) {
	t.Parallel()

	name := InferImplementationName("", nil, "https://mcp.github.com/api")
	if name != "github" {
		t.Fatalf("got %q, want github", name)
	}
}
