package cpa

import (
	"errors"
	"strings"
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
			_, err := parseBundle(map[string]string{"test.rego": tc.Document}, tc.LintRules...)

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

func TestLoadPolicyFile(t *testing.T) {
	testcases := []struct {
		Name        string
		FilePath    string
		ExpectedErr string
	}{
		{
			Name:        "fails on non-existing filePath",
			FilePath:    "./testdata/does_not_exist",
			ExpectedErr: "failed to read file: open testdata/does_not_exist: no such file or directory",
		},
		{
			Name:        "fails if filePath is a directory",
			FilePath:    "./testdata",
			ExpectedErr: "failed to read file: read testdata: is a directory",
		},
		{
			Name:     "successfully parses given filePath",
			FilePath: "./testdata/multiple_policies/policy1.rego",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			policy, err := LoadPolicyFile(tc.FilePath)

			if tc.ExpectedErr == "" {
				require.NoError(t, err)
				require.NotNil(t, policy)
			} else {
				require.EqualError(t, err, tc.ExpectedErr)
			}
		})
	}
}

func TestLoadPolicyDirectory(t *testing.T) {
	testcases := []struct {
		Name          string
		DirectoryPath string
		ExpectedErr   string
	}{
		{
			Name:          "fails on non-existing directoryPath",
			DirectoryPath: "./testdata/does_not_exist",
			ExpectedErr:   "failed to get list of policy files: open ./testdata/does_not_exist: no such file or directory",
		},
		{
			Name:          "fails if directoryPath is a file",
			DirectoryPath: "./testdata/multiple_policies/policy1.rego",
			ExpectedErr:   "./testdata/multiple_policies/policy1.rego: not a directory",
		},
		{
			Name:          "successfully parses given directoryPath",
			DirectoryPath: "./testdata/multiple_policies",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			policy, err := LoadPolicyDirectory(tc.DirectoryPath)

			if tc.ExpectedErr == "" {
				require.NoError(t, err)
				require.NotNil(t, policy)
			} else {
				require.NotNil(t, err, "expected error to not be nil")
				require.Contains(t, err.Error(), tc.ExpectedErr)
			}
		})
	}
}
