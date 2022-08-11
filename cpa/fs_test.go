package cpa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadPolicyFromFS(t *testing.T) {
	testcases := []struct {
		Name        string
		Path        string
		ExpectedErr string
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
			Name: "successfully parses given directoryPath",
			Path: "./testdata/multiple_policies",
		},
		{
			Name: "successfully parses given filePath",
			Path: "./testdata/multiple_policies/policy1.rego",
		},
		{
			Name:        "fails when loading non-rego file",
			Path:        "./testdata/mixed_ext/policy.text",
			ExpectedErr: "no rego policies found at path",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			policy, err := LoadPolicyFromFS(tc.Path)

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

func TestLoadPolicyFromFSFiltersRego(t *testing.T) {
	policy, err := LoadPolicyFromFS("testdata/mixed_ext")
	require.NoError(t, err)

	bundle := policy.Source()
	require.Len(t, bundle, 1)
	require.Contains(t, bundle, "rego")
}
