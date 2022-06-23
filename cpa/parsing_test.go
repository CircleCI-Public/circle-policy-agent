package cpa

import (
	"errors"
	"strings"
	"testing"
)

func TestRegoParsing(t *testing.T) {
	testCases := []struct {
		Name      string
		Document  string
		LintRules []LintRule
		Error     error
	}{
		{
			Name:     "invalid document fails in parsing",
			Document: "Invalid Content",
			Error:    errors.New("test.rego:1: rego_parse_error: package expected"),
		},
		{
			Name: "invalid document-parses successfully, fails in compiling",
			Document: `
				package opa.example
				import data.foo
				p[x] { foo[x]; not ba[x]; x >= min_x }
				min_x = 100 { true }
			`,
			Error: errors.New("1 error occurred: test.rego:4: rego_unsafe_var_error: var ba is unsafe"),
		},
		{
			Name: "succeeds on valid document",
			Document: `
				package opa.example
				import data.foo
				import input.bar
				p[x] { foo[x]; not bar[x]; x >= min_x }
				min_x = 100 { true }
			`,
			Error: nil,
		},
		{
			Name: "fails package name linting",
			Document: `
				package evil
			`,
			LintRules: []LintRule{AllowedPackages("good", "righteous")},
			//nolint
			Error: errors.New(`failed policy linting: lint error: "test.rego": invalid package name: expected one of packages [good, righteous] but got "package evil"`),
		},
		{
			Name: "passes package name linting",
			Document: `
				package good
			`,
			LintRules: []LintRule{AllowedPackages("good", "righteous")},
			Error:     nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := ParseRego("test.rego", tc.Document, tc.LintRules...)

			if tc.Error == nil && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}

			if tc.Error != nil {
				if err == nil {
					t.Fatalf("expected error %q but got none", tc.Error.Error())
				}
				expected := tc.Error.Error()
				actual := err.Error()
				if !strings.Contains(actual, expected) {
					t.Fatalf("expected error %q but got %q", expected, actual)
				}
			}
		})
	}
}
