package runner_test

import (
	"os"
	"testing"

	"github.com/CircleCI-Public/circle-policy-agent/cpa/tester"
	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	options := tester.RunnerOptions{
		Path:    "./...",
		Dst:     os.Stdout,
		Verbose: true,
		Debug:   false,
		Include: nil,
	}

	runner, err := tester.NewRunner(options)
	require.NoError(t, err)

	require.True(t, runner.Run())
}
