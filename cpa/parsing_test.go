package cpa

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
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
				policy_name["test"]
				p[x] { foo[x]; not ba[x]; x >= min_x }
				min_x = 100 { true }
			`,
			Error: errors.New("failed to compile policy: 1 error occurred: test.rego:5: rego_unsafe_var_error: var ba is unsafe"),
		},
		{
			Name: "succeeds on valid document",
			Document: `
				package opa.example
				import data.foo
				import input.bar
				policy_name["test"]
				p[x] { foo[x]; not bar[x]; x >= min_x }
				min_x = 100 { true }
			`,
			Error: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := parseBundle(map[string]string{"test.rego": tc.Document}, tc.LintRules...)
			if tc.Error != nil {
				require.ErrorContains(t, err, tc.Error.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegoLinting(t *testing.T) {
	testCases := []struct {
		Name      string
		Document  string
		LintRules []LintRule
		Error     error
	}{
		{
			Name: "fails package name linting",
			Document: `
				package evil
				policy_name["test"]
			`,
			LintRules: []LintRule{AllowedPackages("good", "righteous")},
			//nolint
			Error: errors.New(`failed policy linting: "test": invalid package name: expected one of packages [good, righteous] but got "package evil"`),
		},
		{
			Name: "passes package name linting",
			Document: `
				package good
				policy_name["test"]
			`,
			LintRules: []LintRule{AllowedPackages("good", "righteous")},
			Error:     nil,
		},
		{
			Name: "fails if data.meta.branch is used",
			Document: `
				package org
				policy_name["test"]
				rule {
					data.meta.branch == "main"
				}	
				`,
			LintRules: []LintRule{DisallowMetaBranch()},
			Error:     errors.New("failed policy linting: \"test\": test.rego:5: invalid use of data.meta.branch use data.meta.vcs.branch instead"),
		},
		{
			Name: "fails if data.meta.branch multi bodies",
			Document: `
				package org
				policy_name["test"]
				rule {
					data.meta.vcs.branch == "main"
				} {
					data.meta.branch == "main"
				}	
				`,
			LintRules: []LintRule{DisallowMetaBranch()},
			Error:     errors.New("failed policy linting: \"test\": test.rego:7: invalid use of data.meta.branch use data.meta.vcs.branch instead"),
		},
		{
			Name: "passes if data.meta.vcs.branch is used",
			Document: `
				package org
				policy_name["test"]
				rule {
					data.meta.vcs.branch == "main"
				}	
				`,
			LintRules: []LintRule{DisallowMetaBranch()},
			Error:     nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := parseBundle(map[string]string{"test.rego": tc.Document}, tc.LintRules...)
			if tc.Error != nil {
				require.EqualError(t, err, tc.Error.Error())
				require.True(t, errors.Is(err, ErrLint))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
