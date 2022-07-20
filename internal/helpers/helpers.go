package helpers

import (
	_ "embed"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
)

var (
	//go:embed rego/jobs.rego
	jobsRego string

	//go:embed rego/orbs.rego
	orbsRego string

	//go:embed rego/runner.rego
	runnerRego string
)

var configHelpers = map[string]string{
	"circleci_jobs_helper.rego":   jobsRego,
	"circleci_orbs_helper.rego":   orbsRego,
	"circleci_runner_helper.rego": runnerRego,
}

var configHelpersMap = make(map[string]*ast.Module, len(configHelpers))

func init() {
	for filename, rego := range configHelpers {
		mod, err := ast.ParseModule(filename, rego)
		if err != nil {
			panic(err)
		}
		configHelpersMap[filename] = mod
	}
}

func AppendCircleCIConfigHelpers(mods map[string]*ast.Module) error {
	for filename, helper := range configHelpersMap {
		if _, ok := mods[filename]; ok {
			return fmt.Errorf("policy filename %q uses reserved circleci_ prefix", filename)
		}
		mods[filename] = helper
	}

	return nil
}
