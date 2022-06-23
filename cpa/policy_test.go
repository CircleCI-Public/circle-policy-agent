package cpa

import (
	"context"
	"errors"
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
			Name: "Successfully parses policy bundle when package name is org for all documents in the bundle",
			DocumentBundle: map[string]string{
				"test.rego": "package org",
				"foo.rego":  "package org",
				"bar.rego":  "package org",
			},
			Error: nil,
		},
		{
			Name: "Error when package name is not org for a document in the bundle",
			DocumentBundle: map[string]string{
				"bad.rego": "package bad",
				"foo.rego": "package org",
				"bar.rego": "package org",
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
					my_orbs := config.orbs
				`,
			},
			Error: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := ParsePolicy(tc.DocumentBundle)

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

	doc, err := ParseRego("test.rego", `
		package test
		import future.keywords

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
	`)

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
				"devs":    []interface{}{"Sagar"},
				"names":   []interface{}{"Idoh", "Sagar"},
				"product": []interface{}{"Idoh"},
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

	doc, err := ParseBundle(map[string]string{
		"helper.rego": `
			package helper
			names[name] {
				name := input.team[_].name
			}
		`,
		"product.rego": `
			package team
			product[name] {
				some i; input.team[i].developer == false
				name := input.team[i].name
			}
		`,
		"devs.rego": `
			package team
			import future.keywords
			import data.helper
			devs[name] {
				some name in helper.names
				not product[name]
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
			"helper": map[string]interface{}{
				"names": []interface{}{"Idoh", "Sagar"},
			},
			"team": map[string]interface{}{
				"devs":    []interface{}{"Sagar"},
				"product": []interface{}{"Idoh"},
			},
		},
		result,
	)
}

type DecideTestCase struct {
	Name     string
	Document string
	Config   string
	Error    error
	Decision *Decision
}
