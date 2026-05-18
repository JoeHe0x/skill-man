package skill

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skill-man/internal/domain/agent"
	"skill-man/internal/domain/extension"
	skilldomain "skill-man/internal/domain/skill"
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

	lines := bytes.Split(content, []byte("\n"))
	for _, line := range lines {
		text := strings.TrimSpace(string(line))
		switch {
		case strings.HasPrefix(text, "name:") && name == filepath.Base(dir):
			name = strings.TrimSpace(strings.TrimPrefix(text, "name:"))
		case strings.HasPrefix(text, "description:") && description == "":
			description = strings.TrimSpace(strings.TrimPrefix(text, "description:"))
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
			continue
		}
		seen[k] = len(out)
		out = append(out, sk)
	}
	return out
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
