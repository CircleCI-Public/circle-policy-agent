package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/stretchr/testify/require"
)

func TestOrbHelper(t *testing.T) {
	mod := helpers["config"]["circleci/rego/config/orbs.rego"]

	input := map[string]interface{}{
		"orbs": map[string]string{
			"security": "circleci/security@1.2.3",
			"test":     "circleci/test@0.0.0",
		},
	}

	validateOrbTypes := rego.Function2(
		&rego.Function{
			Name: "validate_orb_types",
			Decl: types.NewFunction(types.Args(types.A, types.A), types.A),
		},
		func(_ rego.BuiltinContext, allowedOrbType, orbsUsed *ast.Term) (*ast.Term, error) {
			dummyOrbRegistry := map[string]string{
				"certified/abc": "certified",
				"partner/abc":   "partner",
				"public/abc":    "public",
			}

			return ast.MustParseTerm(fmt.Sprintf("%v", dummyOrbRegistry)), nil
		})

	result, err := rego.
		New(rego.ParsedModule(mod), rego.Query("data"), rego.Input(input), validateOrbTypes).
		Eval(context.Background())

	require.NoError(t, err)

	output := result[0].Expressions[0].Value

	b, err := json.Marshal(output)
	require.NoError(t, err)

	require.EqualValues(t, `{"circleci":{"config":{"orbs":{"circleci/security":"1.2.3","circleci/test":"0.0.0"}}}}`, string(b))
}
