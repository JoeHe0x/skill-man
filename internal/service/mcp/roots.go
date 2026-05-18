package mcp

import mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"

// WorkspaceRoots collects unique workspace roots from all bindings on a server.
func WorkspaceRoots(srv *mcpdomain.Server) []string {
	seen := map[string]bool{}
	var roots []string
	for _, b := range srv.AllBindings() {
		root := WorkspaceRootFromArgs(b.Args)
		if root == "" || seen[root] {
			continue
		}
		seen[root] = true
		roots = append(roots, root)
	}
	return roots
}
