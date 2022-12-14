package cpa

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadPolicyFromFS(t *testing.T) {
	testcases := []struct {
		Name             string
		Path             string
		ExpectedErr      string
		ExpectedPolicies []string
	}{
		{
			Name:        "fails on non-existing directoryPath",
			Path:        "./testdata/does_not_exist",
			ExpectedErr: "failed to walk root",
		},
		{
			Name:        "fails on non-existing filePath",
			Path:        "./testdata/does_not_exist",
			ExpectedErr: "failed to walk root",
		},
		{
			Name:             "successfully parses given directoryPath",
			Path:             "./testdata/multiple_policies",
			ExpectedPolicies: []string{"policy_1", "policy_2", "policy_3"},
		},
		{
			Name:             "successfully parses given filePath",
			Path:             "./testdata/multiple_policies/policy1.rego",
			ExpectedPolicies: []string{"policy_1"},
		},
		{
			Name:        "fails when loading non-rego file",
			Path:        "./testdata/mixed_ext/policy.text",
			ExpectedErr: "no rego policies found",
		},
		{
			Name:             "only load rego files",
			Path:             "./testdata/mixed_ext",
			ExpectedPolicies: []string{"rego"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			policy, err := LoadPolicyFromFS(tc.Path)

			if tc.ExpectedErr != "" {
				require.NotNil(t, err, "expected error to not be nil")
				require.Contains(t, err.Error(), tc.ExpectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)

			var policies []string
			for name := range policy.Source() {
				policies = append(policies, name)
			}
			sort.StringSlice(policies).Sort()
			sort.StringSlice(tc.ExpectedPolicies).Sort()

			require.Equal(t, tc.ExpectedPolicies, policies)
		})
	}
}
