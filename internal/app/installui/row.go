package installui

import (
	"strings"
)

// Row is a list entry in the install wizard (no dependency on app/panel).
type Row struct {
	Title       string
	Desc        string
	Meta        string
	DetailLines []string
}

func (r Row) FilterValue() string {
	parts := []string{r.Title, r.Desc, r.Meta}
	parts = append(parts, r.DetailLines...)
	return strings.ToLower(strings.Join(parts, " "))
}
