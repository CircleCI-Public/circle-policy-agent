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

	require.True(t, runner.RunAndHandleResults(MakeDefaultResultHandler(ResultHandlerOptions{
		Verbose: os.Getenv("VERBOSE") == "true",
		Debug:   os.Getenv("DEBUG") == "true",
		Dst:     os.Stdout,
	})))
}
