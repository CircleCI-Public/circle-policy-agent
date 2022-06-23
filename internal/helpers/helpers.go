package helpers

import (
	_ "embed"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
)

var (
	//go:embed rego/jobs.rego
	JobsRego string

	//go:embed rego/orbs.rego
	OrbsRego string
)

var configHelpers = map[string]string{
	"circleci_jobs_helper.rego": JobsRego,
	"circleci_orbs_helper.rego": OrbsRego,
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
