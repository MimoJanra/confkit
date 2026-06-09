package interpolation

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var interpolationPattern = regexp.MustCompile(`\$\{([^}|]+)(?:\|([^}]*))?\}`)

type Resolver struct {
	envMap     map[string]string
	config     map[string]string
	visited    map[string]bool
	maxDepth   int
	currentKey string
}

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

func (r *Resolver) Resolve(value string, fieldPath string) (string, error) {
	if !strings.Contains(value, "${") {
		return value, nil
	}

	r.currentKey = fieldPath
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

		if defaultVal != "" {
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

func (r *Resolver) SetConfigValue(key string, value string) {
	r.config[key] = value
}
