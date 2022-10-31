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
	Name          string
	Document      string
	Config        string
	ParseError    error
	DecisionError error
	Metadata      map[string]interface{}
	Decision      *Decision
}

var lintingCases = []DecideTestCase{
	{
		Name: "error if package name is not org",
		Document: `
			package foo
			policy_name["test"]
		`,
		ParseError: errors.New(`invalid package name "data.foo" must be one of org, branch["{expression}"] or project["{expression}"]`),
	},
}

var outputStructureCases = []DecideTestCase{
	{
		Name:          "trivial pass if no policy",
		Document:      "",
		Config:        "input: any",
		DecisionError: nil,
		Decision:      &Decision{Status: StatusPass},
	},
	{
		Name: "hard failure when an enabled hard_fail rule is violated",
		Document: `
			package org
			policy_name["test"]
			enable_rule["name_is_bob"]
			hard_fail["name_is_bob"]
			name_is_bob = "name must be bob!" {	input.name != "bob" }
		`,
		Config: `name: joe`,
		Decision: &Decision{
			Status:       "HARD_FAIL",
			EnabledRules: []string{"org.name_is_bob"},
			HardFailures: []Violation{
				{Rule: "org.name_is_bob", Reason: "name must be bob!"},
			},
		},
	},
	{
		Name: "soft failure when an enabled rule is violated",
		Document: `
			package org
			policy_name["test"]
			enable_rule["name_is_bob"]
			name_is_bob = "name must be bob!" {	input.name != "bob" }
		`,
		Config: "name: joe",
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.name_is_bob"},
			SoftFailures: []Violation{
				{Rule: "org.name_is_bob", Reason: "name must be bob!"},
			},
		},
	},
	{
		Name: "pass when no rule is violated",
		Document: `
			package org
			policy_name["test"]
			enable_rule["name_is_bob"]
			name_is_bob = "name must be bob!" {	input.name != "bob" }
		`,
		Config: `name: bob`,
		Decision: &Decision{
			Status:       StatusPass,
			EnabledRules: []string{"org.name_is_bob"},
		},
	},
	{
		Name: "decision status is hard fail when there are violations for hard and soft fail rules",
		Document: `
			package org
			policy_name["test"]
			enable_rule := ["name_is_bob", "type_is_person"]
			hard_fail := ["type_is_person"]
			name_is_bob = "name must be bob!" {	input.name != "bob" }
			type_is_person = "type must be person" { input.type != "person" }
		`,
		Config: `{
			"name": "sasha",
			"type": "scooter"
		}`,
		Decision: &Decision{
			Status:       "HARD_FAIL",
			EnabledRules: []string{"org.name_is_bob", "org.type_is_person"},
			HardFailures: []Violation{
				{Rule: "org.type_is_person", Reason: "type must be person"},
			},
			SoftFailures: []Violation{
				{Rule: "org.name_is_bob", Reason: "name must be bob!"},
			},
		},
	},
	{
		Name: "decision status is pass when no rule is enabled",
		Document: `
			package org
			policy_name["test"]
			hard_fail := ["type_is_person"]
			name_is_bob = "name must be bob!" {	input.name != "bob" }
			type_is_person = "type must be person" { input.type != "person" }
		`,
		Config: `{
			"name": "sasha",
			"type": "scooter"
		}`,
		Decision: &Decision{Status: "PASS"},
	},
	{
		Name: "violation reason can be parsed when reason is a map[string]interface{}",
		Document: `
			package org
			policy_name["test"]
			enable_rule["name_must_be_bob"]
			name_must_be_bob[name] = reason {
				name := input.names[_]
				name != "bob"
				reason := sprintf("%s is not bob", [name])
			}
		`,
		Config: `names: ["alice", "bob", "charlie"]`,
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.name_must_be_bob"},
			SoftFailures: []Violation{
				{Rule: "org.name_must_be_bob", Reason: "alice is not bob"},
				{Rule: "org.name_must_be_bob", Reason: "charlie is not bob"},
			},
		},
	},
	{
		Name: "violation reason can be parsed when reason is a static string",
		Document: `
			package org
			policy_name["test"]
			enable_rule["name_is_bob"]
			name_is_bob = "name must be bob" { input.name != "bob" }
		`,
		Config: "name: joe",
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.name_is_bob"},
			SoftFailures: []Violation{
				{Rule: "org.name_is_bob", Reason: "name must be bob"},
			},
		},
	},
	{
		Name: "violation reason can be parsed when reason is []interface{}",
		Document: `
			package org
			policy_name["test"]
			enable_rule["name_starts_with_a_or_b"]
			name_starts_with_a_or_b = reason {
				not startswith(input.name, "a")
				reason := ["input does not start with a", "input does not start with b"]
			}
		`,
		Config: `name: charlie`,
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.name_starts_with_a_or_b"},
			SoftFailures: []Violation{
				{Rule: "org.name_starts_with_a_or_b", Reason: "input does not start with a"},
				{Rule: "org.name_starts_with_a_or_b", Reason: "input does not start with b"},
			},
		},
	},
}

var orbCases = []DecideTestCase{
	{
		Name: "circleci require orb helper passes when orb is present",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_orbs"]
			require_security_orbs = config.require_orbs(["circleci/security"])
		`,
		Config: `{
			"orbs": {
				"security": "circleci/security@1.2.3"
			}
		}`,
		Decision: &Decision{
			Status:       "PASS",
			EnabledRules: []string{"org.require_security_orbs"},
		},
	},
	{
		Name: "circleci require orb helper fails when orb is absent",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_orbs"]
			require_security_orbs = config.require_orbs(["circleci/security"])
		`,
		Config: `{
			"orbs": {
				"security": "circleci/casual-checks@1.2.3"
			}
		}`,
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.require_security_orbs"},
			SoftFailures: []Violation{
				{Rule: "org.require_security_orbs", Reason: "circleci/security orb is required"},
			},
		},
	},
	{
		Name: "circleci require orb version helper passes when orb is present",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_orbs"]
			require_security_orbs = config.require_orbs_version(["circleci/security@1.2.3"])
		`,
		Config: `{
			"orbs": {
				"security": "circleci/security@1.2.3"
			}
		}`,
		Decision: &Decision{
			Status:       "PASS",
			EnabledRules: []string{"org.require_security_orbs"},
		},
	},
	{
		Name: "circleci require orb version helper fails when orb has wrong version",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_orbs"]
			require_security_orbs = config.require_orbs_version(["circleci/security@1.2.3"])
		`,
		Config: `{
			"orbs": {
				"security": "circleci/security@0.0.0"
			}
		}`,
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.require_security_orbs"},
			SoftFailures: []Violation{
				{Rule: "org.require_security_orbs", Reason: "circleci/security@1.2.3 orb is required"},
			},
		},
	},
	{
		Name: "circleci ban orb helper passes when orb is absent",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["ban_orbs"]
			ban_orbs = config.ban_orbs(["evilcorp/evil"])
		`,
		Config: `{
			"orbs": {
				"security": "circleci/security@1.2.3"
			}
		}`,
		Decision: &Decision{
			Status:       "PASS",
			EnabledRules: []string{"org.ban_orbs"},
		},
	},
	{
		Name: "circleci ban orb helper fails when orb is present",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["ban_orbs"]
			ban_orbs = config.ban_orbs(["foo/bar", "evilcorp/evil"])
		`,
		Config: `{
			"orbs": {
				"foobar": "foo/bar@1.2.3",
				"evil": "evilcorp/evil@1.2.3"
			}
		}`,
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.ban_orbs"},
			SoftFailures: []Violation{
				{Rule: "org.ban_orbs", Reason: "evilcorp/evil orb is not allowed in CircleCI configuration"},
				{Rule: "org.ban_orbs", Reason: "foo/bar orb is not allowed in CircleCI configuration"},
			},
		},
	},
	{
		Name: "circleci ban orb version helper passes when orb is absent",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["ban_orbs"]
			ban_orbs = config.ban_orbs_version(["evilcorp/evil@1.2.3", "security@0.0.0"])
		`,
		Config: `{
			"orbs": {
				"security": "circleci/security@1.2.3"
			}
		}`,
		Decision: &Decision{
			Status:       "PASS",
			EnabledRules: []string{"org.ban_orbs"},
		},
	},
	{
		Name: "circleci ban orb version helper fails when orb version is present",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["ban_orbs_version"]
			ban_orbs_version = config.ban_orbs_version(["foo/bar@1.2.3", "evilcorp/evil@4.5.6"])
		`,
		Config: `{
			"orbs": {
				"foobar": "foo/bar@1.2.3",
				"evil": "evilcorp/evil@4.5.6"
			}
		}`,
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.ban_orbs_version"},
			SoftFailures: []Violation{
				{Rule: "org.ban_orbs_version", Reason: "evilcorp/evil@4.5.6 orb is not allowed in CircleCI configuration"},
				{Rule: "org.ban_orbs_version", Reason: "foo/bar@1.2.3 orb is not allowed in CircleCI configuration"},
			},
		},
	},
}

var jobCases = []DecideTestCase{
	{
		Name: "circleci require job helper passes when job is present",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_jobs"]
			require_security_jobs = config.require_jobs(["security-job"])
		`,
		Config: `{
			"workflows": {
				"security-workflow": {
				  	"jobs": [
						"security-job"
				  	]
				}
			}
		}`,
		Decision: &Decision{
			Status:       StatusPass,
			EnabledRules: []string{"org.require_security_jobs"},
		},
	},
	{
		Name: "circleci require job helper passes when jobs are present and jobs are in string and map format",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_jobs"]
			require_security_jobs = config.require_jobs(["security-job","vulnerability-job"])
		`,
		Config: `{
			"workflows": {
				"security-workflow": {
				  	"jobs": [
						"security-job",
						"vulnerability-job": {
							"context": "dockerhub-readonly",
							"requires": [
								"security-job"
							]
						}
				  	]
				}
			}
		}`,
		Decision: &Decision{
			Status:       StatusPass,
			EnabledRules: []string{"org.require_security_jobs"},
		},
	},
	{
		Name: "circleci require job helper soft fails when all required jobs are absent",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_jobs"]
			require_security_jobs = config.require_jobs(["security-job","vulnerability-scan-job"])
		`,
		Config: `{
			"workflows": {
				"security-workflow": {
				  	"jobs": [
						"evil-job"
				  	]
				}
			}
		}`,
		Decision: &Decision{
			Status:       StatusSoftFail,
			EnabledRules: []string{"org.require_security_jobs"},
			SoftFailures: []Violation{
				{Rule: "org.require_security_jobs", Reason: "security-job job is required"},
				{Rule: "org.require_security_jobs", Reason: "vulnerability-scan-job job is required"},
			},
		},
	},
	{
		Name: "circleci require job helper fails when one required job is present and one required job is absent",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_jobs"]
			require_security_jobs = config.require_jobs(["security-job","vulnerability-scan-job"])
		`,
		Config: `{
			"workflows": {
				"security-workflow": {
				  	"jobs": [
						"security-job"
				  	]
				}
			}
		}`,
		Decision: &Decision{
			Status:       StatusSoftFail,
			EnabledRules: []string{"org.require_security_jobs"},
			SoftFailures: []Violation{
				{Rule: "org.require_security_jobs", Reason: "vulnerability-scan-job job is required"},
			},
		},
	},
	{
		Name: "circleci require job helper properly handles multiple jobs in multiple workflows",
		Document: `
			package org
			policy_name["test"]
			import data.circleci.config
			enable_rule["require_security_jobs"]
			require_security_jobs = config.require_jobs(["security-job","vulnerability-scan-job"])
		`,
		Config: `{
			"workflows": {
				"workflow-one": {
				  	"jobs": [
						"security-job",
						"foo-job"
				  	]
				},
				"workflow-two": {
				  	"jobs": [
						"bar-job"
				  	]
				}
			}
		}`,
		Decision: &Decision{
			Status:       StatusSoftFail,
			EnabledRules: []string{"org.require_security_jobs"},
			SoftFailures: []Violation{
				{Rule: "org.require_security_jobs", Reason: "vulnerability-scan-job job is required"},
			},
		},
	},
}

var runnerCases = []DecideTestCase{
	{
		Name: "detects resource_class violations",
		Document: `
			package org
			import data.circleci.config
			policy_name["runner_test"]

			enable_rule["check_resource_class"]

			check_resource_class = config.resource_class_by_project({
				"large": {"A"},
				"medium": {"B"},
				"small": {"C"},
			})
		`,
		Config: `{
			"jobs": {
				"lint": {"resource_class": "medium"},
				"test": {"resource_class": "large"},
				"build": {"resource_class": "small"}
			}
		}`,
		Metadata: map[string]interface{}{
			"project_id": "B",
		},
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.check_resource_class"},
			SoftFailures: []Violation{
				{Rule: "org.check_resource_class", Reason: "project is not allowed to use resource_class \"large\" declared in job \"test\""},
				{Rule: "org.check_resource_class", Reason: "project is not allowed to use resource_class \"small\" declared in job \"build\""},
			},
		},
	},
	{
		Name: "does not affect unspecified resource classes",
		Document: `
			package org
			import data.circleci.config
			policy_name["runner_test"]

			enable_rule["check_resource_class"]

			check_resource_class = config.resource_class_by_project({"large": {"A"}})
		`,
		Config: `{
			"jobs": { "lint": { "resource_class": "medium" } }
		}`,
		Metadata: map[string]interface{}{
			"project_id": "B",
		},
		DecisionError: nil,
		Decision:      &Decision{Status: "PASS", EnabledRules: []string{"org.check_resource_class"}},
	},
}

var contextCases = []DecideTestCase{
	{
		Name: "blocklist/should ban from project",
		Document: `
			package org
			import data.circleci.config

			policy_name["test"]

			enable_rule["ban_private_context"]

			ban_private_context = config.context_blocklist_by_project("public", "private")
		`,
		Config: `{
			"workflows": {
				"main": {
				  	"jobs": [
						"good-job",
						"bad-job": { "context": "private" }
				  	]
				}
			}
		}`,
		Metadata: map[string]interface{}{
			"project_id": "public",
		},
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.ban_private_context"},
			SoftFailures: []Violation{
				{
					Rule:   "org.ban_private_context",
					Reason: "context \"private\" used in job \"bad-job\" has been banned from current project",
				},
			},
		},
	},
	{
		Name: "blocklist/should not ban from valid project",
		Document: `
			package org
			import data.circleci.config

			policy_name["test"]

			enable_rule["ban_private_context"]

			ban_private_context = config.context_blocklist_by_project("public", "private")
		`,
		Config: `{
			"workflows": {
				"main": {
				  	"jobs": [
						"good-job",
						"bad-job": { "context": "private" }
				  	]
				}
			}
		}`,
		Metadata: map[string]interface{}{
			"project_id": "cool project",
		},
		Decision: &Decision{
			Status:       "PASS",
			EnabledRules: []string{"org.ban_private_context"},
		},
	},
	{
		Name: "blocklist/should ban multiple contexts",
		Document: `
			package org
			import data.circleci.config

			policy_name["test"]

			enable_rule["ban_private_context"]

			ban_private_context = config.context_blocklist_by_project("public", {"private", "secret"})
		`,
		Config: `{
			"workflows": {
				"main": {
				  	"jobs": [
						"good-job",
						"bad-job": { "context": ["private", "secret"] }
				  	]
				}
			}
		}`,
		Metadata: map[string]interface{}{
			"project_id": "public",
		},
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.ban_private_context"},
			SoftFailures: []Violation{
				{Rule: "org.ban_private_context", Reason: "context \"private\" used in job \"bad-job\" has been banned from current project"},
				{Rule: "org.ban_private_context", Reason: "context \"secret\" used in job \"bad-job\" has been banned from current project"},
			},
		},
	},
	{
		Name: "allowlist/should block contexts not in list",
		Document: `
			package org
			import data.circleci.config

			policy_name["test"]

			enable_rule["context_check"]

			context_check = config.context_allowlist_by_project("prjkt", {"one", "two"})
		`,
		Config: `{
			"workflows": {
				"main": {
				  	"jobs": [
						"good-job",
						"bad-job": { "context": ["one", "three", "four"] }
				  	]
				}
			}
		}`,
		Metadata: map[string]interface{}{
			"project_id": "prjkt",
		},
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"org.context_check"},
			SoftFailures: []Violation{
				{Rule: "org.context_check", Reason: "context \"four\" used in job \"bad-job\" is not part of allowed list of contexts for project"},
				{Rule: "org.context_check", Reason: "context \"three\" used in job \"bad-job\" is not part of allowed list of contexts for project"},
			},
		},
	},

	{
		Name: "allowlist/should pass if all contexts are in allow list",
		Document: `
			package org
			import data.circleci.config

			policy_name["test"]

			enable_rule["context_check"]

			context_check = config.context_allowlist_by_project("prjkt", {"one", "two"})
		`,
		Config: `{
			"workflows": {
				"main": {
				  	"jobs": [
						"good-job",
						"bad-job": { "context": "one" }
				  	]
				}
			}
		}`,
		Metadata: map[string]interface{}{
			"project_id": "prjkt",
		},
		Decision: &Decision{
			Status:       "PASS",
			EnabledRules: []string{"org.context_check"},
		},
	},
}

var metaPackageCases = []DecideTestCase{
	{
		Name: "project_id",
		Document: `
			package project.__id__
			policy_name["prjkt_policies"]
			enable_rule["custom"]
			custom = "some error"
		`,
		Metadata: map[string]interface{}{"project_id": "__id__"},
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"project.__id__.custom"},
			SoftFailures: []Violation{
				{Rule: "project.__id__.custom", Reason: "some error"},
			},
		},
	},
	{
		Name: "project_id not matched",
		Document: `
			package project.__id__
			policy_name["prjkt_policies"]
			enable_rule["custom"]
			custom = "some error"
		`,
		Metadata: map[string]interface{}{"project_id": "__other__"},
		Decision: &Decision{Status: "PASS"},
	},
	{
		Name: "project slug",
		Document: `
			package project["slug-*"]
			policy_name["prjkt_policies"]
			enable_rule["custom"]
			custom = "some error"
		`,
		Metadata: map[string]interface{}{"project_slug": "slug-example"},
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"project.slug-*.custom"},
			SoftFailures: []Violation{
				{Rule: "project.slug-*.custom", Reason: "some error"},
			},
		},
	},
	{
		Name: "project slug not matched",
		Document: `
			package project["slug-*"]
			policy_name["prjkt_policies"]
			enable_rule["custom"]
			custom = "some error"
		`,
		Metadata: map[string]interface{}{"project_slug": "not-slug-example"},
		Decision: &Decision{Status: "PASS"},
	},
	{
		Name: "branch",
		Document: `
			package branch["feature-*"]
			policy_name["feature_policies"]
			enable_rule["custom"]
			custom = "some error"
		`,
		Metadata: map[string]interface{}{"branch": "feature-demo"},
		Decision: &Decision{
			Status:       "SOFT_FAIL",
			EnabledRules: []string{"branch.feature-*.custom"},
			SoftFailures: []Violation{
				{Rule: "branch.feature-*.custom", Reason: "some error"},
			},
		},
	},
	{
		Name: "branch not matched",
		Document: `
			package branch["feature-*"]
			policy_name["prjkt_policies"]
			enable_rule["custom"]
			custom = "some error"
		`,
		Metadata: map[string]interface{}{"branch": "main"},
		Decision: &Decision{Status: "PASS"},
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
		{
			Group: "output structure",
			Cases: outputStructureCases,
		},
		{
			Group: "orb helper",
			Cases: orbCases,
		},
		{
			Group: "job helper",
			Cases: jobCases,
		},
		{
			Group: "runner helper",
			Cases: runnerCases,
		},
		{
			Group: "context cases",
			Cases: contextCases,
		},
		{
			Group: "meta cases",
			Cases: metaPackageCases,
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

					doc, err := ParseBundle(bundle)
					if tc.ParseError == nil {
						require.NoError(t, err)
					} else {
						require.ErrorContains(t, err, tc.ParseError.Error())
						return
					}

					decision, err := doc.Decide(context.Background(), config, Meta(tc.Metadata))
					if tc.DecisionError == nil && err != nil {
						t.Fatalf("expected no error but got: %v", err)
					}

					if tc.DecisionError != nil {
						if err == nil {
							t.Fatalf("expected error %q but got none", tc.DecisionError.Error())
						}
						expected := tc.DecisionError.Error()
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
