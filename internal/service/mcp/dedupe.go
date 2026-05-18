package mcp

import mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"

func dedupeByConfigLocation(servers []*mcpdomain.Server) []*mcpdomain.Server {
	type key struct {
		configPath string
		configKey  string
		scope      string
	}
	seen := map[key]int{}
	var out []*mcpdomain.Server

	for _, srv := range servers {
		if srv == nil {
			continue
		}
		k := key{configPath: srv.ConfigPath, configKey: srv.ConfigKey, scope: string(srv.Scope)}
		if idx, ok := seen[k]; ok {
			out[idx].Agents = mergeAgentIDs(out[idx].Agents, srv.Agents)
			continue
		}
		seen[k] = len(out)
		out = append(out, srv)
	}
	return out
}

func dedupeByName(servers []*mcpdomain.Server) []*mcpdomain.Server {
	seen := map[string]int{}
	var out []*mcpdomain.Server

	for _, srv := range servers {
		if srv == nil {
			continue
		}
		name := srv.GetName()
		if name == "" {
			out = append(out, srv)
			continue
		}
		if idx, ok := seen[name]; ok {
			mergeServersByName(out[idx], srv)
			continue
		}
		seen[name] = len(out)
		srv.Bindings = bindingsFromServer(srv)
		srv.SyncAggregatedFields()
		out = append(out, srv)
	}
	return out
}

func mergeServersByName(dst, src *mcpdomain.Server) {
	for _, b := range bindingsFromServer(src) {
		appendBinding(dst, b)
	}
	dst.SyncAggregatedFields()
}

func bindingsFromServer(srv *mcpdomain.Server) []mcpdomain.Binding {
	raw := srv.AllBindings()
	out := make([]mcpdomain.Binding, len(raw))
	copy(out, raw)
	return out
}

func appendBinding(dst *mcpdomain.Server, b mcpdomain.Binding) {
	for i, existing := range dst.Bindings {
		if existing.ConfigPath == b.ConfigPath && existing.ConfigKey == b.ConfigKey {
			dst.Bindings[i].Agents = mergeAgentIDs(existing.Agents, b.Agents)
			return
		}
	}
	dst.Bindings = append(dst.Bindings, b)
}
