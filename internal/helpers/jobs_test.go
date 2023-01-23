package helpers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/open-policy-agent/opa/rego"
	"github.com/stretchr/testify/require"
)

func TestJobsHelper(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          string
		ExpectedOutput string
	}{
		{
			Name: "jobs can be parsed when the jobs are strings",
			Input: `{
				"workflows": {
					"workflow-one": {
						"jobs": ["job-a", "job-b"]
					}
				}
			}`,
			ExpectedOutput: `{"circleci":{"config":{"jobs":["job-a","job-b"]}}}`,
		},
		{
			Name: "jobs can be parsed when the jobs are maps",
			Input: `{
			  "workflows": {
			    "workflow-one": {
			      "jobs": [
			        {
			          "job-a": { "context": "dockerhub-readonly"}
			        },
			        {
			          "job-b": {
			            "context": "dockerhub-readonly",
			            "filters": {
			              "branches": {
			                "only": [
			                  "main"
			                ]
			              }
			            },
			            "requires": ["job-a"]
			          }
			        }
			      ]
			    }
			  }
			}`,
			ExpectedOutput: `{"circleci":{"config":{"jobs":["job-a","job-b"]}}}`,
		},
		{
			Name: "jobs can be parsed when jobs are both strings and maps",
			Input: `{
				"workflows": {
				  "workflow-one": {
				    "jobs": [
				      "job-a",
				      {
				        "job-b": {
				          "context": "dockerhub-readonly"
				        }
				      }
				    ]
				  }
				}
			}`,
			ExpectedOutput: `{"circleci":{"config":{"jobs":["job-a","job-b"]}}}`,
		},
		{
			Name: "jobs can be parsed when there are multiple workflows",
			Input: `{
					  "workflows": {
					    "workflow-one": {
					      "jobs": [
					        "job-a",
					        {
					          "job-b": {
					            "context": "dockerhub-readonly"
					          }
					        }
					      ]
					    },
					    "workflow-two": {
					      "jobs": [
					        "job-a",
					        "job-c"
					      ]
					    }
					  }
					}`,
			ExpectedOutput: `{"circleci":{"config":{"jobs":["job-a","job-b","job-c"]}}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mod := helpers["config"]["circleci/rego/config/jobs.rego"]

			var input map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(tc.Input), &input))

			result, err := rego.
				New(rego.ParsedModule(mod), rego.Query("data"), rego.Input(input)).
				Eval(context.Background())

			require.NoError(t, err)

			output := result[0].Expressions[0].Value

			b, err := json.Marshal(output)
			require.NoError(t, err)

			require.EqualValues(t, tc.ExpectedOutput, string(b))
		})
	}
}
