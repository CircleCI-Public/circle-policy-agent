package cpa

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
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
					policy_name["test"]
				`,
			},
			Error: nil,
		},
		{
			Name: "Successfully parses policy bundle when package name is org for all documents in the bundle",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name["test_1"]
				`,
				"foo.rego": `
					package org
					policy_name["test_2"]
				`,
				"bar.rego": `
					package org
					policy_name["test_3"]
				`,
			},
			Error: nil,
		},
		{
			Name: "Error when package name is not org for a document in the bundle",
			DocumentBundle: map[string]string{
				"bad.rego": `
					package bad
					policy_name["test_1"]
				`,
				"foo.rego": `
					package org
					policy_name["test_2"]
				`,
				"bar.rego": `
					package org
					policy_name["test_3"]
				`,
			},
			//nolint
			Error: errors.New(`failed policy linting: lint error: "test_1": invalid package name: expected one of packages [org] but got "package bad"`),
		},
		{
			Name: "Successfully parses policy bundle when helper functions are added to the rego",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					import future.keywords
					import data.circleci.config
					policy_name["test"]
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
			Name: "fails if policy_name is not declared as  key",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name = "not_a_key"
				`,
			},
			Error: errors.New("failed to parse policy file(s): failed to parse file: \"test.rego\": invalid policy_name declaration: must declare as key"),
		},
		{
			Name: "fails if policy_name empty",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name[""]
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
					policy_name["policy"]
				`,
			},
			Error: errors.New(`first rule declaration must be "policy_name" but found "first_rule"`),
		},
		{
			Name: "fails if policy_name is declared more than once",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name["test"]
				`,
				"test2.rego": `
					package org
					policy_name["test"]
				`,
			},
			Error: errors.New(`failed to parse bundle: policy "test" declared 2 times`),
		},
		{
			Name: "fails if policy_name is not a string",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name[3.14159]
				`,
			},
			Error: errors.New(`failed to parse file: "test.rego": invalid policy_name: json: cannot unmarshal number into Go value of type string`),
		},
		{
			Name: "fails if policy_name is invalid string",
			DocumentBundle: map[string]string{
				"test.rego": `
					package org
					policy_name["!@3"]
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
						policy_name[%q]
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

		policy_name["test"]

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
				"policy_name": []interface{}{"test"},
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
			policy_name["helper"]
			names[name] {
				name := input.team[_].name
			}
		`,
		"product.rego": `
			package team
			policy_name["prod"]
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
			policy_name["dev"]
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
				"policy_name": []interface{}{"dev"},
			},
			"helper": map[string]interface{}{
				"names":       []interface{}{"Idoh", "Sagar"},
				"policy_name": []interface{}{"helper"},
			},
			"team": map[string]interface{}{
				"policy_name": []interface{}{"prod"},
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

			policy_name["test"]
			
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
		Name       string
		Bundle     map[string]string
		KeyMapping map[string]string
	}{
		{
			Name:       "empty source",
			Bundle:     map[string]string{},
			KeyMapping: map[string]string{},
		},
		{
			Name: "gets source",
			Bundle: map[string]string{
				"test.rego": `
					package org
					policy_name["name_test"]
					# some comment
				`,
			},
			KeyMapping: map[string]string{
				"name_test": "test.rego",
			},
		},
		{
			Name: "multiple source files",
			Bundle: map[string]string{
				"test1.rego": `
					package org
					policy_name["test1"]
				`,
				"test2.rego": `
					package org
					policy_name["test2"]
				`,
			},
			KeyMapping: map[string]string{
				"test1": "test1.rego",
				"test2": "test2.rego",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			policy, err := ParseBundle(tc.Bundle)
			sourceKeys := maps.Keys(policy.Source())
			modulesKeys := maps.Keys(policy.Modules())
			sort.Strings(sourceKeys)
			sort.Strings(modulesKeys)

			require.NoError(t, err)
			require.Equal(t, sourceKeys, modulesKeys)
			require.Equal(t, len(tc.Bundle), len(policy.Source()))
			for policyName, content := range policy.Source() {
				require.Equal(t, content, tc.Bundle[tc.KeyMapping[policyName]])
			}
		})
	}
}

func TestBundleLinksHelpers(t *testing.T) {
	policyName := "orbs"
	content := fmt.Sprintf(`
		package org
		import data.circleci.config
		policy_name["%s"]
		orbs = config.orbs
		`, policyName,
	)

	policy, err := parseBundle(map[string]string{"test.rego": content})

	require.NoError(t, err)
	require.EqualValues(t, content, policy.Source()[policyName])

	expectedModules := []string{
		"circleci/rego/config/contexts.rego",
		"circleci/rego/config/orbs.rego",
		"circleci/rego/config/runner.rego",
		"circleci/rego/utils/utils.rego",
		"orbs",
	}
	modules := maps.Keys(policy.Modules())
	sort.Strings(modules)

	require.EqualValues(t, expectedModules, modules)
}

func TestHttpBlocked(t *testing.T) {
	policy, err := ParseBundle(map[string]string{
		"policy.rego": `
			package org
			policy_name["http_test"]
			test = http.send({
				"url": "https://localhost:3000",
				"method": "GET"
			})
		`,
	})
	require.ErrorContains(t, err, "undefined function http.send")
	require.Nil(t, policy)
}

func TestNetLookupBlocked(t *testing.T) {
	policy, err := ParseBundle(map[string]string{
		"policy.rego": `
			package org
			policy_name["http_test"]
			test = net.lookup_ip_addr("localhost")
		`,
	})
	require.ErrorContains(t, err, "undefined function net.lookup_ip_addr")
	require.Nil(t, policy)
}

func TestPolicyRuntimeError(t *testing.T) {
	policy, err := ParseBundle(map[string]string{
		"policy.rego": `
			package org
			policy_name["http_test"]
			rule = "yes"
			rule = "no"
		`,
	})
	require.NoError(t, err)

	decision, err := policy.Decide(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, StatusError, decision.Status)
	require.Equal(
		t,
		"policy.rego:5: eval_conflict_error: complete rules must not produce multiple outputs",
		decision.Reason,
	)
}

func TestEnableHard(t *testing.T) {
	policy, err := ParseBundle(map[string]string{
		"policy.rego": `
			package org
			policy_name["policy"]
			some_rule = "my rule"
			enable_hard["some_rule"]
		`,
	})
	require.NoError(t, err)

	v := []Violation{
		{
			Rule:   "some_rule",
			Reason: "my rule",
		},
	}

	decision, err := policy.Decide(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, StatusHardFail, decision.Status)
	require.Contains(t, "some_rule", decision.Reason)
	require.Equal(t, v, decision.HardFailures)
}
