package cpa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
)

type Policy struct {
	compiler *ast.Compiler
	source   map[string]string
}

// Source returns a map of policy_name to normalized rego source code used to build the policy
func (policy Policy) Source() map[string]string {
	return policy.source
}

// Eval will run native OPA query against your document, input, and apply any evaluation options.
// It returns raw OPA expression values.
func (policy Policy) Eval(ctx context.Context, query string, input interface{}, opts ...EvalOption) (interface{}, error) {
	var options evalOptions
	for _, apply := range opts {
		apply(&options)
	}
	return policy.eval(ctx, query, input, options)
}

func (policy Policy) eval(ctx context.Context, query string, input interface{}, options evalOptions) (interface{}, error) {
	input, err := convertYAMLMapKeyTypes(input, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid value: %w", err)
	}

	regoOptions := []func(*rego.Rego){
		rego.Compiler(policy.compiler),
		rego.Query(query),
		rego.Input(input),
	}

	if options.storage != nil {
		regoOptions = append(regoOptions, rego.Store(inmem.NewFromObject(options.storage)))
	}

	q, err := rego.New(regoOptions...).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare context for evaluation: %w", err)
	}

	result, err := q.Eval(ctx)
	if err != nil {
		return nil, err
	}

	var values []interface{}
	for _, r := range result {
		for _, exp := range r.Expressions {
			values = append(values, exp.Value)
		}
	}

	if len(values) == 1 {
		return values[0], nil
	}

	return values, nil
}

// Decide takes an input and evaluates it against a policy. Evaluation options will be passed down to policy.Eval
func (policy Policy) Decide(ctx context.Context, input interface{}, opts ...EvalOption) (*Decision, error) {
	if len(policy.compiler.Modules) == 0 {
		return &Decision{Status: StatusPass}, nil
	}

	var options evalOptions
	for _, apply := range opts {
		apply(&options)
	}

	data, err := policy.eval(ctx, "data", input, options)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate the query: %w", err)
	}

	output, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected opa output")
	}

	var decision Decision

	meta := fromJsonRepresentation(options.storage["meta"])

	projectID := asString(meta["project_id"])
	project := asMap(output["project"])

	if err := decision.evaluate(dotJoin("project", projectID), asMap(project[projectID])); err != nil {
		return nil, err
	}

	slug := asString(meta["project_slug"])
	for key, value := range project {
		if matched, _ := path.Match(key, slug); !matched {
			continue
		}
		if err := decision.evaluate(dotJoin("project", key), asMap(value)); err != nil {
			return nil, err
		}
	}

	branch := asString(meta["branch"])
	for key, value := range asMap(output["branch"]) {
		if matched, _ := path.Match(key, branch); !matched {
			continue
		}
		if err := decision.evaluate(dotJoin("branch", key), asMap(value)); err != nil {
			return nil, err
		}
	}

	if err := decision.evaluate("org", asMap(output["org"])); err != nil {
		return nil, err
	}

	decision.finalize()

	return &decision, nil
}

func asStringSlice(value interface{}) ([]string, error) {
	if value == nil {
		return nil, nil
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, errors.New("value is not a slice")
	}

	result := make([]string, 0, len(values))
	for _, v := range values {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}

	return result, nil
}

func as[T any](value any) T {
	result, _ := value.(T)
	return result
}

var (
	asMap    = as[map[string]any]
	asString = as[string]
)

func dotJoin(values ...string) string {
	return strings.Join(values, ".")
}

func fromJsonRepresentation(value any) (result map[string]any) {
	result = map[string]any{}
	if value == nil {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	_ = json.Unmarshal(data, &result)
	return result
}
