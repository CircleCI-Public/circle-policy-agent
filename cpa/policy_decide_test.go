package cpa

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type DecideTestCase struct {
	Name     string
	Document string
	Config   string
	Error    error
	Metadata map[string]interface{}
	Decision *Decision
}

var lintingCases = []DecideTestCase{
	{
		Name: "error if package name is not org",
		Document: `
			package foo
			policy_name["test"]
		`,
		Error: errors.New("no org policy evaluations found"),
	},
}

func TestDecide(t *testing.T) {
	testGroups := []struct {
		Group string
		Cases []DecideTestCase
	}{
		{
			Group: "linting",
			Cases: lintingCases,
		},
	}

	for _, group := range testGroups {
		t.Run(group.Group, func(t *testing.T) {
			for _, tc := range group.Cases {
				t.Run(tc.Name, func(t *testing.T) {
					var config interface{}
					if err := yaml.Unmarshal([]byte(tc.Config), &config); err != nil {
						t.Fatalf("invalid config: %v", err)
					}

					bundle := map[string]string{}
					if tc.Document != "" {
						bundle["test.rego"] = tc.Document
					}

					doc, err := parseBundle(bundle)
					if err != nil {
						t.Fatalf("failed to parse rego document for testing: %v", err)
					}

					decision, err := doc.Decide(context.Background(), config, Meta(tc.Metadata))
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
						return
					}

					require.EqualValues(t, tc.Decision, decision)
				})
			}
		})
	}
}
