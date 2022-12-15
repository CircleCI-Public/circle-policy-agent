/*
package cpa/tester simply re-exports the necessary public interface from the internal/tester package
*/
package tester

import "github.com/CircleCI-Public/circle-policy-agent/internal/tester"

type (
	RunnerOptions = tester.RunnerOptions
	Runner        = tester.Runner
)

var NewRunner = tester.NewRunner
