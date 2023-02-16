package tester

import (
	"bytes"
	"encoding/xml"
	"os"
	"regexp"
	"testing"

	_ "embed"

	"github.com/CircleCI-Public/circle-policy-agent/internal/junit"
	"github.com/stretchr/testify/require"
)

var (
	//go:embed policies/results.json
	expectedJSON string

	//go:embed policies/results.xml
	expectedXML string
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
		"policies/common/structure",
		"policies/helpers",
		"policies/helpers/contexts",
		"policies/helpers/orbs",
		"policies/helpers/orbs/ban_version",
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
	t.Run("json", func(t *testing.T) {
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

		require.JSONEq(t, expectedJSON, buf.String())
	})

	t.Run("xml", func(t *testing.T) {
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

		buf := new(bytes.Buffer)
		opts := ResultHandlerOptions{Dst: buf}

		MakeJUnitResultHandler(opts).HandleResults(runner.Run())

		suites := junit.JUnitTestSuites{}
		require.NoError(t, xml.Unmarshal(buf.Bytes(), &suites))

		suites.Time = "0"
		for i := range suites.Suites {
			suites.Suites[i].Time = "0"
			for j := range suites.Suites[i].TestCases {
				suites.Suites[i].TestCases[j].Time = "0"
			}
		}

		sanitizedXML, err := xml.MarshalIndent(suites, "", "\t")
		require.NoError(t, err)

		require.Equal(t, expectedXML, string(sanitizedXML))
	})
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
