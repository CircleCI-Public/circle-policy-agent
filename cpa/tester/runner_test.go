package tester

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	options := RunnerOptions{
		Path: "./policies/...",
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
		"policies",
		"policies/common",
		"policies/common/base",
		"policies/common/error",
		"policies/common/no_enabled_rules",
		"policies/common/reason_types",
		"policies/common/soft_and_hard_fail_together",
		"policies/helpers",
		"policies/helpers/contexts",
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

func TestRunnerResults(t *testing.T) {
	options := RunnerOptions{
		Path: "./policies/...",
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

	sanitizedResults := make(chan Result)
	go func() {
		for r := range runner.Run() {
			r.Elapsed = 0 // We cannot statically assert the elapsed time so we zero it out
			sanitizedResults <- r
		}
		close(sanitizedResults)
	}()

	buf := new(bytes.Buffer)
	opts := ResultHandlerOptions{Dst: buf}

	MakeJSONResultHandler(opts).HandleResults(sanitizedResults)

	require.JSONEq(t,
		`[
			{
			  "Passed": true,
			  "Group": "\u003copa.tests\u003e",
			  "Name": "data.org.test_get_job_name_object",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "\u003copa.tests\u003e",
			  "Name": "data.org.test_get_job_name_string",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies",
			  "Elapsed": "0s",
			  "ElapsedMS": 0,
			  "Err": "no tests"
			},
			{
			  "Passed": true,
			  "Group": "policies/common",
			  "Elapsed": "0s",
			  "ElapsedMS": 0,
			  "Err": "no tests"
			},
			{
			  "Passed": true,
			  "Group": "policies/common/base",
			  "Name": "test_base_policy",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/base",
			  "Name": "test_base_policy/not_bob",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/base",
			  "Name": "test_base_policy/not_bob/hard_fail",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/error",
			  "Name": "test_error",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/no_enabled_rules",
			  "Name": "test_no_enabled_rules",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/reason_types",
			  "Name": "test_reason_types",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/reason_types",
			  "Name": "test_reason_types/array",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/reason_types",
			  "Name": "test_reason_types/map",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/reason_types",
			  "Name": "test_reason_types/string",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/common/soft_and_hard_fail_together",
			  "Name": "test_soft_and_hard_fail_together",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers",
			  "Elapsed": "0s",
			  "ElapsedMS": 0,
			  "Err": "no tests"
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_allowlist",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_allowlist/invalid_context",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_allowlist/invalid_contexts",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_allowlist/test_multiple_invalid_contexts_in_job",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_allowlist/test_unaffected_project",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_blocked_contexts_list",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_blocked_contexts_list/blocked_context",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_blocked_contexts_list/invalid_contexts",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_blocked_contexts_list/unaffected_project",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_branch_allowlist",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_branch_allowlist/invalid_branch",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/contexts",
			  "Name": "test_branch_allowlist/unrestricted_by_branch",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/jobs",
			  "Name": "test_require_jobs",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/jobs",
			  "Name": "test_require_jobs/jobs_as_objects",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/jobs",
			  "Name": "test_require_jobs/missing_jobs",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs",
			  "Elapsed": "0s",
			  "ElapsedMS": 0,
			  "Err": "no tests"
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/ban_version",
			  "Name": "test_ban_orbs",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/ban_version",
			  "Name": "test_ban_orbs/orb_present",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/ban_version",
			  "Name": "test_ban_orbs_version",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/ban_version",
			  "Name": "test_ban_orbs_version/exact_match",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/ban_version",
			  "Name": "test_ban_orbs_version/wrong_version",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/require_version",
			  "Name": "test_require_orbs",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/require_version",
			  "Name": "test_require_orbs/orb_absent",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/require_version",
			  "Name": "test_require_orbs_version",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/require_version",
			  "Name": "test_require_orbs_version/absent",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/orbs/require_version",
			  "Name": "test_require_orbs_version/wrong_version",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/runner",
			  "Name": "test_runner_helper",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/runner",
			  "Name": "test_runner_helper/project_medium",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/helpers/runner",
			  "Name": "test_runner_helper/project_small",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/multifile",
			  "Name": "test_multifile_policy",
			  "Elapsed": "0s",
			  "ElapsedMS": 0
			},
			{
			  "Passed": true,
			  "Group": "policies/multifile/sub0",
			  "Elapsed": "0s",
			  "ElapsedMS": 0,
			  "Err": "no tests"
			},
			{
			  "Passed": true,
			  "Group": "policies/multifile/sub1",
			  "Elapsed": "0s",
			  "ElapsedMS": 0,
			  "Err": "no tests"
			},
			{
			  "Passed": true,
			  "Group": "policies/multifile/sub1/sub1_0",
			  "Elapsed": "0s",
			  "ElapsedMS": 0,
			  "Err": "no tests"
			}
		  ]`,
		buf.String(),
	)
}

func TestFailedPolicies(t *testing.T) {
	options := RunnerOptions{
		Path: "./failed_policies/...",
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

	sanitizedResults := make(chan Result)
	go func() {
		for r := range runner.Run() {
			r.Elapsed = 0 // We cannot statically assert the elapsed time so we zero it out
			sanitizedResults <- r
		}
		close(sanitizedResults)
	}()

	buf := new(bytes.Buffer)
	opts := ResultHandlerOptions{Dst: buf}

	MakeJSONResultHandler(opts).HandleResults(sanitizedResults)

	require.JSONEq(t, `[
		{
		  "Passed": false,
		  "Group": "\u003copa.tests\u003e",
		  "Name": "data.org.test_int_is_string",
		  "Elapsed": "0s",
		  "ElapsedMS": 0
		},
		{
		  "Passed": true,
		  "Group": "failed_policies",
		  "Elapsed": "0s",
		  "ElapsedMS": 0,
		  "Err": "no tests"
		}
	  ]`,
		buf.String(),
	)
}
