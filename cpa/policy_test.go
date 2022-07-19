package cpa

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePolicy(t *testing.T) {
	testCases := []struct {
		Name           string
		DocumentBundle map[string]string
		Error          error
	}{
		{
			Name:           "succeeds with no policies",
			DocumentBundle: map[string]string{},
			Error:          nil,
		},
		{
			Name: "succeeds with proper policy",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name = "test"
				`,
			},
			Error: nil,
		},
		{
			Name: "Successfully parses policy bundle when package name is org for all documents in the bundle",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name = "test_1"
				`,
				"foo.rego": `
					package org
					policy_name = "test_2"
				`,
				"bar.rego": `
					package org
					policy_name = "test_3"
				`,
			},
			Error: nil,
		},
		{
			Name: "Error when package name is not org for a document in the bundle",
			DocumentBundle: map[string]string{
				"bad.rego": `
					package bad
					policy_name = "test_1"
				`,
				"foo.rego": `
					package org
					policy_name = "test_2"
				`,
				"bar.rego": `
					package org
					policy_name = "test_3"
				`,
			},
			//nolint
			Error: errors.New(`failed policy linting: lint error: "bad.rego": invalid package name: expected one of packages [org] but got "package bad"`),
		},
		{
			Name: "Successfully parses policy bundle when helper functions are added to the rego",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					import future.keywords
					import data.circleci.config
					policy_name = "test"
					my_orbs := config.orbs
				`,
			},
			Error: nil,
		},
		{
			Name: "fails if no policy name",
			DocumentBundle: map[string]string{
				"test.rego": "package org",
			},
			Error: errors.New(`failed to parse policy file(s): failed to parse file: "test.rego": must declare rule "policy_name" but module contains no rules`),
		},
		{
			Name: "fails if policy_name empty",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name = ""
				`,
			},
			Error: errors.New(`failed to parse policy file(s): failed to parse file: "test.rego": policy_name must not be empty`),
		},
		{
			Name: "fails if policy_name not the first rule",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					first_rule = "first"
					policy_name = "policy"
				`,
			},
			Error: errors.New(`first rule declaration must be "policy_name" but found "first_rule"`),
		},
		{
			Name: "fails if policy_name is declared more than once",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name = "test"
				`,
				"test2.rego": `
					package org
					policy_name = "test"
				`,
			},
			Error: errors.New(`failed to parse bundle: policy "test" declared 2 times`),
		},
		{
			Name: "fails if policy_name is not a string",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name = 3.14159
				`,
			},
			Error: errors.New(`failed to parse file: "test.rego": invalid policy_name: json: cannot unmarshal number into Go value of type string`),
		},
		{
			Name: "fails if policy_name is invalid string",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name = "!@3"
				`,
			},
			Error: errors.New(`failed to parse policy file(s): failed to parse file: "test.rego": "policy_name" must use alphanumeric and underscore characters only`),
		},
		{
			Name: "fails if policy_name is too long",
			DocumentBundle: func() map[string]string {
				return map[string]string{
					"test.rego": fmt.Sprintf(`
						package org
						policy_name = %q
					`, strings.Repeat("a", 81)),
				}
			}(),
			Error: errors.New(`failed to parse policy file(s): failed to parse file: "test.rego": policy_name must be maximum 80 characters but got 81`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := ParseBundle(tc.DocumentBundle)

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

func TestDocumentQuery(t *testing.T) {
	type Member struct {
		Name      string `json:"name"`
		Developer bool   `json:"developer"`
	}

	type Input struct {
		Team []Member `json:"team"`
	}

	input := Input{
		Team: []Member{
			{
				Name:      "Sagar",
				Developer: true,
			},
			{
				Name:      "Idoh",
				Developer: false,
			},
		},
	}

	doc, err := parseBundle(map[string]string{"test.rego": `
		package test
		import future.keywords

		policy_name = "test"

		names[name] {
			name := input.team[_].name
		}

		product[name] {
			some i; input.team[i].developer == false
			name := input.team[i].name
		}

		devs[name] {
			some name in names;
			not product[name]
		}
	`})
	if err != nil {
		t.Fatalf("failed to parse rego document for testing: %v", err)
	}

	result, err := doc.Eval(context.Background(), "data", input)
	if err != nil {
		t.Fatalf("failed to query document: %v", err)
	}

	require.EqualValues(
		t,
		map[string]interface{}{
			"test": map[string]interface{}{
				"policy_name": "test",
				"devs":        []interface{}{"Sagar"},
				"names":       []interface{}{"Idoh", "Sagar"},
				"product":     []interface{}{"Idoh"},
			},
		},
		result,
	)
}

func TestBundleQuery(t *testing.T) {
	type Member struct {
		Name      string `json:"name"`
		Developer bool   `json:"developer"`
	}

	type Input struct {
		Team []Member `json:"team"`
	}

	input := Input{
		Team: []Member{
			{
				Name:      "Sagar",
				Developer: true,
			},
			{
				Name:      "Idoh",
				Developer: false,
			},
		},
	}

	doc, err := parseBundle(map[string]string{
		"helper.rego": `
			package helper
			policy_name := "helper"
			names[name] {
				name := input.team[_].name
			}
		`,
		"product.rego": `
			package team
			policy_name := "prod"
			product[name] {
				some i; input.team[i].developer == false
				name := input.team[i].name
			}
		`,
		"devs.rego": `
			package devs
			import future.keywords
			import data.helper
			import data.team
			policy_name := "dev"
			devs[name] {
				some name in helper.names
				not team.product[name]
			}
		`,
	})
	if err != nil {
		t.Fatalf("failed to parse bundle for testing: %v", err)
	}

	result, err := doc.Eval(context.Background(), "data", input)
	if err != nil {
		t.Fatalf("failed to query document: %v", err)
	}

	require.EqualValues(
		t,
		map[string]interface{}{
			"devs": map[string]interface{}{
				"devs":        []interface{}{"Sagar"},
				"policy_name": "dev",
			},
			"helper": map[string]interface{}{
				"names":       []interface{}{"Idoh", "Sagar"},
				"policy_name": "helper",
			},
			"team": map[string]interface{}{
				"policy_name": "prod",
				"product":     []interface{}{"Idoh"},
			},
		},
		result,
	)
}

func TestMeta(t *testing.T) {
	policy, err := ParseBundle(map[string]string{
		"test.rego": `
			package org

			policy_name = "test"
			
			meta = data.meta
		`,
	})

	require.NoError(t, err)

	metadata := map[string]interface{}{
		"key":  "value",
		"test": true,
	}

	value, err := policy.Eval(context.Background(), "data.org.meta", nil, Meta(metadata))
	require.NoError(t, err)

	require.EqualValues(t, metadata, value)
}

func TestGetSource(t *testing.T) {
	testcases := []struct {
		Name   string
		Bundle map[string]string
		Source map[string]string
	}{
		{
			Name:   "empty source",
			Bundle: map[string]string{},
			Source: map[string]string{},
		},
		{
			Name: "gets source",
			Bundle: map[string]string{
				"test.rego": `
					package org
					policy_name = "name_test"
					# some comment
				`,
			},
			Source: map[string]string{
				"name_test": "package org\n\npolicy_name = \"name_test\" { true }",
			},
		},
		{
			Name: "multiple source files",
			Bundle: map[string]string{
				"test1.rego": `
					package org
					policy_name = "test1"
				`,
				"test2.rego": `
					package org
					policy_name = "test2"
				`,
			},
			Source: map[string]string{
				"test1": "package org\n\npolicy_name = \"test1\" { true }",
				"test2": "package org\n\npolicy_name = \"test2\" { true }",
			},
		},
		{
			Name: "bundle links helpers",
			Bundle: map[string]string{
				"test.rego": `
					package org
					import data.circleci.config
					policy_name = "orbs"
					versions = config.require_orbs_version([])
				`,
			},
			Source: map[string]string{
				"orbs": "package org\n\nimport data.circleci.config\n\npolicy_name = \"orbs\" { true }\nversions = config.require_orbs_version([]) { true }",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			policy, err := ParseBundle(tc.Bundle)
			require.NoError(t, err)
			require.EqualValues(t, tc.Source, policy.Source())
		})
	}
}
