package confkit

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/MimoJanra/confkit/structtags"
)

// FromFlags reads values from the command line (os.Args[1:]).
//
// A field is matched by its `flag` tag, or failing that by its kebab-cased name,
// its snake_cased name, or its `short` tag. Both --name=value and --name value
// are accepted, a flag with no value becomes "true", and a repeated flag is
// joined with commas so it can populate a slice.
func FromFlags() Source {
	return &flagsSource{args: os.Args[1:]}
}

// FromFlagsWithArgs behaves like FromFlags but parses the supplied arguments,
// which is useful in tests and for wrapping a subcommand's own argument list.
func FromFlagsWithArgs(args []string) Source {
	return &flagsSource{args: args}
}

type flagsSource struct {
	args   []string
	parsed map[string]string
	once   sync.Once
}

func (f *flagsSource) Name() string {
	return "flag"
}

func (f *flagsSource) parseArgs() {
	f.once.Do(func() {
		f.parsed = make(map[string]string)
		args := f.args
		for i := 0; i < len(args); i++ {
			arg := args[i]
			if strings.HasPrefix(arg, "--") {
				name := arg[2:]
				if eq := strings.Index(name, "="); eq != -1 {
					appendFlag(f.parsed, name[:eq], name[eq+1:])
				} else if i+1 < len(args) && !isFlagToken(args[i+1]) {
					appendFlag(f.parsed, name, args[i+1])
					i++
				} else {
					appendFlag(f.parsed, name, "true")
				}
			} else if strings.HasPrefix(arg, "-") && len(arg) == 2 {
				name := arg[1:]
				if i+1 < len(args) && !isFlagToken(args[i+1]) {
					appendFlag(f.parsed, name, args[i+1])
					i++
				} else {
					appendFlag(f.parsed, name, "true")
				}
			} else if strings.HasPrefix(arg, "-") && len(arg) > 2 && arg[2] == '=' {
				name := arg[1:2]
				appendFlag(f.parsed, name, arg[3:])
			}
		}
	})
}

func isFlagToken(s string) bool {
	if !strings.HasPrefix(s, "-") {
		return false
	}
	if len(s) > 1 && (s[1] >= '0' && s[1] <= '9' || s[1] == '.') {
		return false
	}
	return true
}

func appendFlag(m map[string]string, key, val string) {
	if existing, ok := m[key]; ok {
		m[key] = existing + "," + val
	} else {
		m[key] = val
	}
}

func (f *flagsSource) Lookup(_ context.Context, field *FieldInfo) (any, bool, error) {
	f.parseArgs()

	flagName := field.Tags["flag"]
	if flagName == "" {
		flagName = toKebabCase(field.Name)
	}

	if val, ok := f.parsed[flagName]; ok {
		return val, true, nil
	}

	snakeName := structtags.SnakeCase(field.Name)
	if snakeName != flagName {
		if val, ok := f.parsed[snakeName]; ok {
			return val, true, nil
		}
	}

	if short := field.Tags["short"]; short != "" {
		if val, ok := f.parsed[short]; ok {
			return val, true, nil
		}
	}

	return "", false, nil
}

func toKebabCase(s string) string {
	snake := structtags.SnakeCase(s)
	return strings.ReplaceAll(snake, "_", "-")
}
