package tester

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	options := RunnerOptions{
		Path: "./...",
		Include: func() *regexp.Regexp {
			run := os.Getenv("RUN")
			if run == "" {
				return nil
			}
			return regexp.MustCompile(run)
		}(),
	}

	runner, err := NewRunner(options)
	require.NoError(t, err)

	require.Equal(t, []string{
		".",
		"policies",
		"policies/common",
		"policies/common/base",
		"policies/common/error",
		"policies/common/no_rules",
		"policies/common/reason_types",
		"policies/common/soft_and_hard_fail_together",
		"policies/helpers",
		"policies/helpers/jobs",
		"policies/helpers/orbs",
		"policies/helpers/orbs/ban_version",
		"policies/helpers/orbs/require_version",
		"policies/helpers/runner",
		"policies/multifile",
		"policies/multifile/sub0",
		"policies/multifile/sub1",
		"policies/multifile/sub1/sub1_0",
	}, runner.folders)

	require.True(t, runner.RunAndHandleResults(MakeDefaultResultHandler(ResultHandlerOptions{
		Verbose: os.Getenv("VERBOSE") == "true",
		Debug:   os.Getenv("DEBUG") == "true",
		Dst:     os.Stdout,
	})))
}
