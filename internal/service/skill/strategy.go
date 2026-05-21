package skill

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

type SkillScanStrategy struct{}

func (s SkillScanStrategy) DefaultDir() string {
	return ".skills"
}

func (s SkillScanStrategy) AgentDir(a agent.Agent) string {
	if a.EntityDirs != nil {
		return a.EntityDirs[agent.EntitySkill]
	}
	return ""
}

func (s SkillScanStrategy) SkipDir(dirName string) bool {
	return dirName == ".git" || dirName == "node_modules"
}

func (s SkillScanStrategy) TargetFiles() []string {
	return []string{"SKILL.md", "SKILL.md.disabled"}
}

func (s SkillScanStrategy) ParseFile(filePath, projectRoot, home string, scope extension.Scope) (*skilldomain.Skill, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(filePath)
	if resolvedDir, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolvedDir
	}
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		readmePath = ""
	}

	name := filepath.Base(dir)
	description := ""
	tools := []string{}

	if fmName, fmDesc, ok := parseSkillFrontmatter(content); ok {
		if fmName != "" {
			name = fmName
		}
		if fmDesc != "" {
			description = fmDesc
		}
	}

	body := skillBodyAfterFrontmatter(content)
	lines := bytes.Split(body, []byte("\n"))
	for _, line := range lines {
		text := strings.TrimSpace(string(line))
		switch {
		case strings.HasPrefix(text, "- "):
			tools = append(tools, strings.TrimSpace(strings.TrimPrefix(text, "- ")))
		case strings.HasPrefix(text, "# ") && description == "":
			description = strings.TrimSpace(strings.TrimPrefix(text, "# "))
		}
		if description != "" && len(tools) >= 4 {
			break
		}
	}

	if description == "" {
		description = "No description found in SKILL.md."
	}

	hash := sha1.Sum([]byte(filePath))
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	meta, ok, err := readInstallMetadata(dir)
	if err != nil {
		return nil, err
	}

	managed := false
	sourceKind := ""
	sourcePath := ""
	metadataPath := ""
	if ok {
		managed = true
		sourceKind = meta.SourceKind
		sourcePath = meta.SourcePath
		metadataPath = metadataPathForDir(dir)
	}

	agents := agent.ResolveAgentIDs(dir, projectRoot, home)

	return &skilldomain.Skill{
		BaseExtension: extension.BaseExtension{
			ID:           fmt.Sprintf("%x", hash[:8]),
			Name:         strings.Trim(name, `"'`),
			Description:  strings.Trim(description, `"'`),
			Path:         dir,
			ReadmePath:   readmePath,
			ConfigPath:   filepath.Join(dir, filepath.Base(filePath)),
			UpdatedAt:    info.ModTime(),
			Managed:      managed,
			SourceKind:   sourceKind,
			SourcePath:   sourcePath,
			MetadataPath: metadataPath,
			Scope:        scope,
			Agents:       agents,
			Disabled:     strings.HasSuffix(filePath, ".disabled"),
		},
		Tools: tools,
	}, nil
}

func (s SkillScanStrategy) Dedupe(skills []*skilldomain.Skill) []*skilldomain.Skill {
	type key struct {
		path  string
		scope extension.Scope
	}
	seen := map[key]int{}
	var out []*skilldomain.Skill

	for _, sk := range skills {
		k := key{path: sk.Path, scope: sk.Scope}
		if idx, ok := seen[k]; ok {
			out[idx].Agents = mergeAgentIDs(out[idx].Agents, sk.Agents)
			out[idx] = preferSkillEntry(out[idx], sk)
			continue
		}
		seen[k] = len(out)
		out = append(out, sk)
	}
	return out
}

// preferSkillEntry picks the scan row that matches disk (SKILL.md vs SKILL.md.disabled).
func preferSkillEntry(a, b *skilldomain.Skill) *skilldomain.Skill {
	aOK := skillConfigExists(a)
	bOK := skillConfigExists(b)
	switch {
	case aOK && !bOK:
		return a
	case bOK && !aOK:
		return b
	case !a.IsDisabled() && b.IsDisabled():
		return a
	case a.IsDisabled() && !b.IsDisabled():
		return b
	default:
		return a
	}
}

func skillConfigExists(sk *skilldomain.Skill) bool {
	if sk == nil || sk.GetConfigPath() == "" {
		return false
	}
	_, err := os.Stat(sk.GetConfigPath())
	return err == nil
}

// parseSkillFrontmatter reads name/description only from the leading --- YAML block.
func parseSkillFrontmatter(content []byte) (name, description string, ok bool) {
	s := string(content)
	if !strings.HasPrefix(s, "---") {
		return "", "", false
	}
	rest := s
	if strings.HasPrefix(rest, "---\n") {
		rest = rest[4:]
	} else {
		rest = strings.TrimPrefix(rest, "---")
		rest = strings.TrimLeft(rest, "\r\n")
	}
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", "", false
	}
	for _, line := range strings.Split(rest[:end], "\n") {
		text := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(text, "name:") && name == "":
			name = trimYAMLScalar(strings.TrimPrefix(text, "name:"))
		case strings.HasPrefix(text, "description:") && description == "":
			description = trimYAMLScalar(strings.TrimPrefix(text, "description:"))
		}
	}
	return name, description, name != "" || description != ""
}

func skillBodyAfterFrontmatter(content []byte) []byte {
	s := string(content)
	if !strings.HasPrefix(s, "---") {
		return content
	}
	rest := s
	if strings.HasPrefix(rest, "---\n") {
		rest = rest[4:]
	} else {
		rest = strings.TrimPrefix(rest, "---")
		rest = strings.TrimLeft(rest, "\r\n")
	}
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return content
	}
	body := rest[end+4:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}
	return []byte(body)
}

func trimYAMLScalar(v string) string {
	return strings.Trim(strings.TrimSpace(v), `"'`)
}

func ParseSkillFile(filePath string) (*skilldomain.Skill, error) {
	return SkillScanStrategy{}.ParseFile(filePath, "", "", extension.ScopeProject)
}

func mergeAgentIDs(a, b []string) []string {
	set := map[string]bool{}
	for _, id := range a {
		set[id] = true
	}
	for _, id := range b {
		set[id] = true
	}
	var out []string
	for id := range set {
		out = append(out, id)
	}
	return out
}
