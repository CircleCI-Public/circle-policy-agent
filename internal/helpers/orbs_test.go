package helpers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/open-policy-agent/opa/rego"
	"github.com/stretchr/testify/require"
)

func TestOrbHelper(t *testing.T) {
	mod := configHelpersMap["circleci_orbs_helper.rego"]

	input := map[string]interface{}{
		"orbs": map[string]string{
			"security": "circleci/security@1.2.3",
			"test":     "circleci/test@0.0.0",
		},
	}

	result, err := rego.
		New(rego.ParsedModule(mod), rego.Query("data"), rego.Input(input)).
		Eval(context.Background())

	require.NoError(t, err)

	output := result[0].Expressions[0].Value

	b, err := json.Marshal(output)
	require.NoError(t, err)

	require.EqualValues(t, `{"circleci":{"config":{"orbs":{"circleci/security":"1.2.3","circleci/test":"0.0.0"}}}}`, string(b))
}
