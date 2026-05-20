package manager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

// CommonBinding provides generic bind/unbind logic for any extension type
type CommonBinding[T extension.Extension] struct {
	Strategy ScanStrategy[T]
}

// BindAgent links the extension to an agent by creating a symlink
func (b *CommonBinding[T]) BindAgent(ext T, a agent.Agent, projectRoot, home string) error {
	var baseDir string
	if ext.GetScope() == extension.ScopeGlobal {
		if home == "" {
			return errors.New("home directory not available for global binding")
		}
		baseDir = home
	} else {
		if projectRoot == "" {
			return errors.New("project root not available for project binding")
		}
		baseDir = projectRoot
	}

	targetDir := filepath.Join(baseDir, b.Strategy.AgentDir(a))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create agent dir %s: %w", targetDir, err)
	}

	targetPath := filepath.Join(targetDir, filepath.Base(ext.GetPath()))

	// Check if already bound intrinsically
	if samePath(ext.GetPath(), targetPath) {
		return nil
	}

	// Check if target exists
	info, err := os.Lstat(targetPath)
	if err == nil {
		// Target exists, check if it points to us
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(targetPath)
			if err == nil {
				if !filepath.IsAbs(link) {
					link = filepath.Join(targetDir, link)
				}
				if samePath(link, ext.GetPath()) {
					return nil // Already bound
				}
			}
		}
		return fmt.Errorf("target %s already exists and is not a symlink to this extension", targetPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	relPath, err := filepath.Rel(targetDir, ext.GetPath())
	if err != nil {
		relPath = ext.GetPath()
	}

	return os.Symlink(relPath, targetPath)
}

// UnbindAgent removes the link between the extension and an agent
func (b *CommonBinding[T]) UnbindAgent(ext T, a agent.Agent, projectRoot, home string) error {
	var baseDir string
	if ext.GetScope() == extension.ScopeGlobal {
		if home == "" {
			return errors.New("home directory not available for global binding")
		}
		baseDir = home
	} else {
		if projectRoot == "" {
			return errors.New("project root not available for project binding")
		}
		baseDir = projectRoot
	}

	targetDir := filepath.Join(baseDir, b.Strategy.AgentDir(a))
	targetPath := filepath.Join(targetDir, filepath.Base(ext.GetPath()))

	if samePath(ext.GetPath(), targetPath) {
		return errors.New("cannot unbind extension from its primary location")
	}

	info, err := os.Lstat(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // Already unbound
		}
		return err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("target %s is not a symlink, refusing to delete", targetPath)
	}

	return os.Remove(targetPath)
}

// ToggleDisable renames the configuration file to add or remove the .disabled suffix.
// Uses on-disk state when the cached extension disagrees (e.g. fast double-toggle in the TUI).
func ToggleDisable(ext extension.Extension) error {
	configPath := ext.GetConfigPath()
	if configPath == "" {
		return errors.New("extension configuration path unknown")
	}

	enabledPath, disabledPath := togglePairPaths(configPath)

	_, enabledErr := os.Stat(enabledPath)
	_, disabledErr := os.Stat(disabledPath)

	switch {
	case enabledErr == nil && disabledErr == nil:
		return fmt.Errorf("both %s and %s exist; remove one manually",
			filepath.Base(enabledPath), filepath.Base(disabledPath))
	case enabledErr == nil:
		return os.Rename(enabledPath, disabledPath)
	case disabledErr == nil:
		return os.Rename(disabledPath, enabledPath)
	default:
		return fmt.Errorf("toggle disable: %w (tried %s and %s)",
			os.ErrNotExist, filepath.Base(enabledPath), filepath.Base(disabledPath))
	}
}

func togglePairPaths(configPath string) (enabledPath, disabledPath string) {
	if strings.HasSuffix(configPath, ".disabled") {
		return strings.TrimSuffix(configPath, ".disabled"), configPath
	}
	return configPath, configPath + ".disabled"
}

// RemoveExtension removes the extension folder and all symlinks
func (b *CommonBinding[T]) RemoveExtension(ext T, projectRoot, home string) error {
	cleanPath := filepath.Clean(ext.GetPath())

	// First remove symlinks
	for _, a := range agent.DefaultAgents() {
		for _, baseDir := range []string{projectRoot, home} {
			if baseDir == "" {
				continue
			}
			targetDir := filepath.Join(baseDir, b.Strategy.AgentDir(a))
			targetPath := filepath.Join(targetDir, filepath.Base(cleanPath))
			if samePath(targetPath, cleanPath) {
				continue // Skip the actual source dir for now
			}
			info, err := os.Lstat(targetPath)
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				link, err := os.Readlink(targetPath)
				if err == nil {
					if !filepath.IsAbs(link) {
						link = filepath.Join(targetDir, link)
					}
					if samePath(link, cleanPath) {
						if err := os.Remove(targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
							return fmt.Errorf("remove symlink: %w", err)
						}
					}
				}
			}
		}
	}

	return os.RemoveAll(cleanPath)
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}
