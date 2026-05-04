package confkit

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type InterpolationResolver struct {
	env        []string
	config     map[string]string
	visited    map[string]bool
	maxDepth   int
	currentKey string
}

func NewInterpolationResolver(maxDepth int) *InterpolationResolver {
	if maxDepth <= 0 {
		maxDepth = 10
	}
	return &InterpolationResolver{
		env:      os.Environ(),
		config:   make(map[string]string),
		visited:  make(map[string]bool),
		maxDepth: maxDepth,
	}
}

func (r *InterpolationResolver) Resolve(value string, fieldPath string) (string, error) {
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

func (r *InterpolationResolver) resolve(value string, depth int) (string, error) {
	if depth > r.maxDepth {
		return "", fmt.Errorf("max interpolation depth exceeded")
	}

	placeholder := "\x00ESCAPED_DOLLAR\x00"
	protected := strings.ReplaceAll(value, "$$", placeholder)

	pattern := regexp.MustCompile(`\$\{([^}|]+)(?:\|([^}]*))?\}`)
	result := pattern.ReplaceAllStringFunc(protected, func(match string) string {
		parts := pattern.FindStringSubmatch(match)
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

		for _, envPair := range r.env {
			if idx := strings.Index(envPair, "="); idx != -1 {
				key := envPair[:idx]
				val := envPair[idx+1:]
				if key == varName {
					return val
				}
			}
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

	if pattern.MatchString(result) {
		pattern2 := regexp.MustCompile(`\$\{([^}|]+)`)
		matches := pattern2.FindAllStringSubmatch(result, -1)
		for _, m := range matches {
			if len(m) > 1 {
				return "", fmt.Errorf("circular reference or undefined variable: ${%s}", m[1])
			}
		}
	}

	result = strings.ReplaceAll(result, placeholder, "$")

	return result, nil
}

func (r *InterpolationResolver) SetConfigValue(key string, value string) {
	r.config[key] = value
}
