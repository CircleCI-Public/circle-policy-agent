package cpa

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"

	"github.com/CircleCI-Public/circle-policy-agent/internal/helpers"
)

// This function has been stolen from the OPA codebase, ast/parser.go:2210 for
// converting yaml decoded types to types that OPA can handle
func convertYAMLMapKeyTypes(x interface{}, path []string) (interface{}, error) {
	var err error
	switch x := x.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(x))
		for k, v := range x {
			str, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("invalid map key type(s): %v", strings.Join(path, "/"))
			}
			result[str], err = convertYAMLMapKeyTypes(v, append(path, str))
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case []interface{}:
		for i := range x {
			x[i], err = convertYAMLMapKeyTypes(x[i], append(path, fmt.Sprintf("%d", i)))
			if err != nil {
				return nil, err
			}
		}
		return x, nil
	default:
		return x, nil
	}
}

type LintRule func(*ast.Module) error

func AllowedPackages(names ...string) LintRule {
	return func(m *ast.Module) error {
		packageName := strings.TrimPrefix(m.Package.String(), "package ")
		for _, name := range names {
			if name == packageName {
				return nil
			}
		}
		return fmt.Errorf(
			"invalid package name: expected one of packages [%s] but got %q",
			strings.Join(names, ", "),
			m.Package,
		)
	}
}

func shouldImportConfigHelpers(mods map[string]*ast.Module) bool {
	for _, m := range mods {
		for _, i := range m.Imports {
			if i.Path.String() == "data.circleci.config" {
				return true
			}
		}
	}
	return false
}

// parseBundle will parse multiple rego files together into a bundle
func parseBundle(bundle map[string]string, rules ...LintRule) (*Policy, error) {
	moduleMap := make(map[string]*ast.Module, len(bundle))

	var multiErr MultiError
	for file, rego := range bundle {
		mod, err := ast.ParseModule(file, rego)
		if err != nil {
			multiErr = append(multiErr, fmt.Errorf("failed to parse file %q: %w", file, err))
		} else {
			moduleMap[file] = mod
		}
	}

	if len(multiErr) > 0 {
		return nil, fmt.Errorf("failed to parse policy file(s): %w", multiErr)
	}

	for filename, mod := range moduleMap {
		for _, rule := range rules {
			if err := rule(mod); err != nil {
				multiErr = append(multiErr, fmt.Errorf("lint error: %q: %w", filename, err))
			}
		}
	}

	if len(multiErr) > 0 {
		return nil, fmt.Errorf("failed policy linting: %w", multiErr)
	}

	if shouldImportConfigHelpers(moduleMap) {
		if err := helpers.AppendCircleCIConfigHelpers(moduleMap); err != nil {
			return nil, fmt.Errorf("failed to import helper functions")
		}
	}

	compiler := ast.NewCompiler()
	compiler.WithStrict(true)

	if compiler.Compile(moduleMap); compiler.Failed() {
		return nil, fmt.Errorf("failed to compile policy: %w", compiler.Errors)
	}

	return &Policy{compiler}, nil
}

//nolint:lll
// ParseBundle will restrict package name to 'org'. This allows us to more easily extract information from the OPA output after evaluating a
// policy, because we know what the keys will be in the map that contains the results (e.g., map["org"]["enable_rule"] to find enabled rules).
func ParseBundle(files map[string]string) (*Policy, error) {
	return parseBundle(files, AllowedPackages("org"))
}

type MultiError []error

func (err MultiError) Error() string {
	messages := make([]string, len(err))
	for i, e := range err {
		messages[i] = e.Error()
	}

	switch len(messages) {
	case 0:
		return "no errors"
	case 1:
		return messages[0]
	default:
		return fmt.Sprintf("%d error(s) occurred: %s", len(err), strings.Join(messages, "; "))
	}
}
