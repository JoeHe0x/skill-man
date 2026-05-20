package installui

import "strings"

var defaultSuggestions = []string{
	"react", "testing", "go", "golang", "typescript", "python",
	"git", "docker", "api", "lint", "docs", "security", "review", "refactor",
}

func configureSearchInput(m *Model) {
	m.searchInput.ShowSuggestions = true
	m.searchInput.SetSuggestions(defaultSuggestions)
}

func (m *Model) rememberSearchQuery(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}
	for _, q := range m.recentQueries {
		if strings.EqualFold(q, query) {
			return
		}
	}
	m.recentQueries = append([]string{query}, m.recentQueries...)
	if len(m.recentQueries) > 8 {
		m.recentQueries = m.recentQueries[:8]
	}
}

func (m *Model) syncSearchSuggestions() {
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, q := range m.recentQueries {
		add(q)
	}
	for _, s := range defaultSuggestions {
		add(s)
	}
	for _, c := range m.results {
		add(c.Name)
		add(c.Source)
		if c.URL != "" {
			add(c.URL)
		}
	}
	m.searchInput.SetSuggestions(out)
}
