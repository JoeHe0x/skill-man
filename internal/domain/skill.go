package domain

import "time"

type Scope string

const (
	ScopeProject Scope = "project"
	ScopeGlobal  Scope = "global"
)

type Skill struct {
	ID            string
	Name          string
	Description   string
	Tools         []string
	Path          string
	ReadmePath    string
	SkillFilePath string
	UpdatedAt     time.Time
	Managed       bool
	SourceKind    string
	SourcePath    string
	MetadataPath  string
	Scope         Scope
	Agents        []string
	Disabled      bool
}
