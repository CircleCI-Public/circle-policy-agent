package cpa

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

type LintRule func(*ast.Module) error

func AllowedPackages(names ...string) LintRule {
	return func(m *ast.Module) error {
		packageName := strings.TrimPrefix(m.Package.String(), "package ")
		for _, name := range names {
			if name == packageName {
				return nil
			}
		}
		return lintErrf(
			"invalid package name: expected one of packages [%s] but got %q",
			strings.Join(names, ", "),
			m.Package,
		)
	}
}

func DisallowMetaBranch() LintRule {
	return func(m *ast.Module) error {
		for _, rule := range m.Rules {
			for _, expr := range []*ast.Expr(rule.Body) {
				terms, ok := expr.Terms.([]*ast.Term)
				if !ok {
					continue
				}
				for _, term := range terms {
					if term != nil && term.String() == "data.meta.branch" {
						return lintErrf("%s: invalid use of data.meta.branch use data.meta.vcs.branch instead", term.Location.String())
					}
				}
			}
		}
		return nil
	}
}

type LintError string

var ErrLint = LintError("linting error")

func (e LintError) Error() string {
	return string(e)
}

func (LintError) Is(target error) bool {
	_, ok := target.(LintError)
	return ok
}

func lintErrf(format string, args ...any) LintError {
	return LintError(fmt.Sprintf(format, args...))
}
