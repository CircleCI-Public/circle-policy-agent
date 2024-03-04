package tester

import (
	"bytes"
	"encoding/xml"
	"os"
	"regexp"
	"testing"
	"time"

	_ "embed"

	"github.com/CircleCI-Public/circle-policy-agent/internal/junit"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
		"policies/common/enable_hard",
		"policies/common/error",
		"policies/common/no_enabled_rules",
		"policies/common/reason_types",
		"policies/common/soft_and_hard_fail_together",
		"policies/common/structure",
		"policies/helpers",
		"policies/helpers/contexts",
		"policies/helpers/orbs",
		"policies/helpers/orbs/allowlist",
		"policies/helpers/orbs/ban_version",
		"policies/helpers/runner",
		"policies/helpers/utils",
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

		MakeJUnitResultHandlerWithGetTime(opts, func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2024-03-04T10:50:05Z")
			return t
		}).HandleResults(runner.Run())

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

func TestCompilePolicies(t *testing.T) {
	t.Run("with compiler", func(t *testing.T) {
		options := RunnerOptions{
			Path: "./compiler_policies/compile/...",
			Include: func() *regexp.Regexp {
				run := os.Getenv("RUN")
				if run == "" {
					return nil
				}
				return regexp.MustCompile(run)
			}(),
			Compile: func(b []byte, m map[string]any) ([]byte, error) {
				var data map[string]any
				if err := yaml.Unmarshal(b, &data); err != nil {
					return nil, err
				}
				return yaml.Marshal(data["compiled_definition"])
			},
		}

		runner, err := NewRunner(options)
		require.NoError(t, err)

		require.True(t, runner.RunAndHandleResults(MakeDefaultResultHandler(ResultHandlerOptions{
			Verbose: os.Getenv("VERBOSE") == "true",
			Debug:   os.Getenv("DEBUG") == "true",
			Dst:     os.Stdout,
		})))
	})

	t.Run("without compiler", func(t *testing.T) {
		options := RunnerOptions{
			Path: "./compiler_policies/compile/...",
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
			    "Group": "compiler_policies/compile",
			    "Name": "test_compiler",
			    "Elapsed": "0s",
			    "ElapsedMS": 0,
			    "Err": "test set compile to true but no compiler was provided"
			  },
			  {
			    "Passed": false,
			    "Group": "compiler_policies/compile",
			    "Name": "test_compiler/inherits_compile_option",
			    "Elapsed": "0s",
			    "ElapsedMS": 0,
			    "Err": "test set compile to true but no compiler was provided"
			  },
			  {
			    "Passed": true,
			    "Group": "compiler_policies/compile",
			    "Name": "test_compiler/overrides_compile_option",
			    "Elapsed": "0s",
			    "ElapsedMS": 0
			  }
			]`,
			buf.String(),
		)
	})
}

func TestPipelineParameters(t *testing.T) {
	options := RunnerOptions{
		Path: "./compiler_policies/parameters/...",
		Include: func() *regexp.Regexp {
			run := os.Getenv("RUN")
			if run == "" {
				return nil
			}
			return regexp.MustCompile(run)
		}(),
		Compile: func(b []byte, m map[string]any) ([]byte, error) {
			return yaml.Marshal(m)
		},
	}

	runner, err := NewRunner(options)
	require.NoError(t, err)

	require.True(t, runner.RunAndHandleResults(MakeDefaultResultHandler(ResultHandlerOptions{
		Verbose: os.Getenv("VERBOSE") == "true",
		Debug:   os.Getenv("DEBUG") == "true",
		Dst:     os.Stdout,
	})))
}
