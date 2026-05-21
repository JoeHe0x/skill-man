package install

import (
	"fmt"
	"regexp"
	"strings"
)

var registrySourcePattern = regexp.MustCompile(`^([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)(?:@([A-Za-z0-9_.:@+-]+))?$`)

// parseRegistrySource splits owner/repo@skill (skills.sh registry id).
func parseRegistrySource(source string) (owner, repo, skill string, err error) {
	source = strings.TrimSpace(source)
	m := registrySourcePattern.FindStringSubmatch(source)
	if len(m) != 4 {
		return "", "", "", fmt.Errorf("invalid registry source %q (expected owner/repo or owner/repo@skill)", source)
	}
	owner, repo = m[1], m[2]
	skill = strings.TrimSpace(m[3])
	if skill == "" {
		return "", "", "", fmt.Errorf("registry install requires owner/repo@skill, got %q", source)
	}
	return owner, repo, skill, nil
}

// skillSlug matches vercel-labs/skills toSkillSlug for download API paths.
func skillSlug(name string) string {
	name = strings.ToLower(name)
	name = strings.NewReplacer(" ", "-", "_", "-").Replace(name)
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-':
			if !lastDash {
				b.WriteRune(r)
			}
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "skill"
	}
	return out
}

func formatInstallCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM installs", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK installs", float64(n)/1_000)
	case n == 1:
		return "1 install"
	default:
		return fmt.Sprintf("%d installs", n)
	}
}
