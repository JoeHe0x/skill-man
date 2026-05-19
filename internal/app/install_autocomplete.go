package app

import (
	"strings"
)

// Common skills.sh search terms for inline completion.
var defaultInstallSuggestions = []string{
	"react",
	"testing",
	"go",
	"golang",
	"typescript",
	"python",
	"git",
	"docker",
	"api",
	"lint",
	"docs",
	"security",
	"review",
	"refactor",
}

func configureInstallSearchInput(flow *installFlow) {
	flow.searchInput.ShowSuggestions = true
	flow.searchInput.SetSuggestions(defaultInstallSuggestions)
}

func (flow *installFlow) rememberSearchQuery(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}
	for _, q := range flow.recentQueries {
		if strings.EqualFold(q, query) {
			return
		}
	}
	flow.recentQueries = append([]string{query}, flow.recentQueries...)
	if len(flow.recentQueries) > 8 {
		flow.recentQueries = flow.recentQueries[:8]
	}
}

func (flow *installFlow) syncSearchSuggestions() {
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

	for _, q := range flow.recentQueries {
		add(q)
	}
	for _, s := range defaultInstallSuggestions {
		add(s)
	}
	for _, c := range flow.results {
		add(c.Name)
		add(c.Source)
		if c.URL != "" {
			add(c.URL)
		}
	}

	flow.searchInput.SetSuggestions(out)
}
