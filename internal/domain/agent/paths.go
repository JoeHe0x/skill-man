package agent

import (
	"os"
	"path/filepath"
)

// HasLocalEntityDir reports whether the agent's entity directory exists under projectRoot or home.
func HasLocalEntityDir(a Agent, entity EntityType, projectRoot, home string) bool {
	dir := ""
	if a.EntityDirs != nil {
		dir = a.EntityDirs[entity]
	}
	if dir == "" {
		return false
	}
	if projectRoot != "" && localDirExists(filepath.Join(projectRoot, dir)) {
		return true
	}
	if home != "" && localDirExists(filepath.Join(home, dir)) {
		return true
	}
	return false
}

// HasLocalSkillDir reports whether the agent's skills directory exists locally.
func HasLocalSkillDir(a Agent, projectRoot, home string) bool {
	return HasLocalEntityDir(a, EntitySkill, projectRoot, home)
}

// AgentsWithLocalSkillDir returns agents whose skill directory exists under projectRoot or home.
func AgentsWithLocalSkillDir(agents []Agent, projectRoot, home string) []Agent {
	out := make([]Agent, 0, len(agents))
	for _, a := range agents {
		if HasLocalSkillDir(a, projectRoot, home) {
			out = append(out, a)
		}
	}
	return out
}

func localDirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
