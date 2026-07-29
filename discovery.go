package confkit

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotFound is reported by the Source returned from FindSource when no config
// file could be located. Test for it with errors.Is.
var ErrNotFound = errors.New("confkit: config file not found")

// DefaultSearchDirs returns the standard config search directories for appName.
// Order: ./ → ./config/ → /etc/<appName>/ → ~/.<appName>/
func DefaultSearchDirs(appName string) []string {
	dirs := []string{
		"./",
		"./config/",
		"/etc/" + appName + "/",
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, home+"/."+appName+"/")
	}
	return dirs
}

// FindFile returns the first existing path for name in the given dirs.
// If name has no extension, it tries .yaml, .yml, .json, .toml in order, so a
// directory holding both config.yaml and config.yml resolves to config.yaml.
// Returns ("", false) if nothing is found.
func FindFile(name string, dirs ...string) (string, bool) {
	hasExt := filepath.Ext(name) != ""

	for _, dir := range dirs {
		if hasExt {
			p := filepath.Clean(filepath.Join(dir, name))
			if _, err := os.Stat(p); err == nil {
				return p, true
			}
		} else {
			for _, ext := range []string{".yaml", ".yml", ".json", ".toml"} {
				p := filepath.Clean(filepath.Join(dir, name+ext))
				if _, err := os.Stat(p); err == nil {
					return p, true
				}
			}
		}
	}
	return "", false
}

// FindSource locates name via FindFile and returns a ready-to-use Source.
// Returns an error source wrapping ErrNotFound if the file is not found.
func FindSource(name string, dirs ...string) Source {
	path, ok := FindFile(name, dirs...)
	if !ok {
		return &errorSource{err: ErrNotFound}
	}
	switch filepath.Ext(path) {
	case ".json":
		return FromJSON(path)
	case ".toml":
		return FromTOML(path)
	default:
		return FromYAML(path)
	}
}
