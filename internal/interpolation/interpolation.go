// Package interpolation expands ${VAR} references inside configuration values.
//
// It is internal: the substitution rules are an implementation detail of confkit and
// are not part of its public API.
package interpolation

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var interpolationPattern = regexp.MustCompile(`\$\{([^}|]+)(?:\|([^}]*))?\}`)

// Resolver expands ${VAR} references, drawing values first from the environment
// captured at construction and then from other configuration fields.
//
// A Resolver is not safe for concurrent use: it keeps per-resolution state to detect
// reference cycles.
type Resolver struct {
	envMap   map[string]string
	config   map[string]string
	visited  map[string]bool
	maxDepth int
}

// NewResolver returns a Resolver that snapshots the current environment and refuses
// to nest references more than maxDepth deep. A maxDepth of zero or less becomes 10.
func NewResolver(maxDepth int) *Resolver {
	if maxDepth <= 0 {
		maxDepth = 10
	}

	envMap := make(map[string]string)
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	return &Resolver{
		envMap:   envMap,
		config:   make(map[string]string),
		visited:  make(map[string]bool),
		maxDepth: maxDepth,
	}
}

// Resolve expands every reference in value, using fieldPath only to describe the
// location in error messages.
//
// A reference may carry a default after a pipe, as in ${PORT|8080}; the default may be
// empty (${PORT|}), which is distinct from having none. A reference that resolves to
// nothing and has no default is an error, as is a cycle or exceeding the maximum
// depth. Write $$ for a literal dollar sign.
func (r *Resolver) Resolve(value string, fieldPath string) (string, error) {
	// "$$" must be unescaped even when the value contains no ${...} placeholder.
	if !strings.Contains(value, "${") && !strings.Contains(value, "$$") {
		return value, nil
	}

	r.visited = make(map[string]bool)
	result, err := r.resolve(value, 0)
	if err != nil {
		return "", fmt.Errorf("interpolation in %s: %w", fieldPath, err)
	}
	return result, nil
}

func (r *Resolver) resolve(value string, depth int) (string, error) {
	if depth > r.maxDepth {
		return "", fmt.Errorf("max interpolation depth exceeded")
	}

	placeholder := "\x00ESCAPED_DOLLAR\x00"
	protected := strings.ReplaceAll(value, "$$", placeholder)

	result := interpolationPattern.ReplaceAllStringFunc(protected, func(match string) string {
		parts := interpolationPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		varName := parts[1]
		defaultVal := ""
		if len(parts) > 2 {
			defaultVal = parts[2]
		}
		// A "|" in the match means a default was supplied, possibly an empty one
		// ("${VAR|}"), which is distinct from no default at all ("${VAR}").
		hasDefault := strings.Contains(match, "|")

		if r.visited[varName] {
			return match
		}

		if envVal, ok := r.envMap[varName]; ok {
			return envVal
		}

		if val, ok := r.config[varName]; ok {
			r.visited[varName] = true
			resolved, err := r.resolve(val, depth+1)
			delete(r.visited, varName)
			if err == nil {
				return resolved
			}
		}

		if hasDefault {
			return defaultVal
		}
		return match
	})

	if interpolationPattern.MatchString(result) {
		matches := interpolationPattern.FindAllStringSubmatch(result, -1)
		for _, m := range matches {
			if len(m) > 1 {
				return "", fmt.Errorf("circular reference or undefined variable: ${%s}", m[1])
			}
		}
	}

	result = strings.ReplaceAll(result, placeholder, "$")

	return result, nil
}

// SetConfigValue registers a value that other fields may reference by key, which is a
// field path such as "DB.Host". Environment variables take precedence over these.
func (r *Resolver) SetConfigValue(key string, value string) {
	r.config[key] = value
}
