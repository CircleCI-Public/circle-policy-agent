package cpa

import (
	"context"
	"errors"
	"fmt"

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
	input, err := convertYAMLMapKeyTypes(input, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid value: %w", err)
	}

	var options evalOptions
	for _, apply := range opts {
		apply(&options)
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

	data, err := policy.Eval(ctx, "data", input, opts...)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return &Decision{Status: StatusError, Cause: err.Error()}, nil
	}

	output, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected opa output")
	}

	org, ok := output["org"].(map[string]interface{})
	if !ok {
		return nil, errors.New("no org policy evaluations found")
	}

	enabledRules, err := asStringSlice(org["enable_rule"])
	if err != nil {
		return nil, fmt.Errorf("invalid enable_rule: %w", err)
	}

	hardFailRules, err := asStringSlice(org["hard_fail"])
	if err != nil {
		return nil, fmt.Errorf("invalid hard_fail: %w", err)
	}

	hardFailMap := make(map[string]struct{})
	for _, rule := range hardFailRules {
		hardFailMap[rule] = struct{}{}
	}

	decision := Decision{EnabledRules: enabledRules}

	for _, rule := range enabledRules {
		if _, ok := hardFailMap[rule]; ok {
			decision.HardFailures = append(decision.HardFailures, extractViolations(org, rule)...)
		} else {
			decision.SoftFailures = append(decision.SoftFailures, extractViolations(org, rule)...)
		}
	}

	switch {
	case len(decision.HardFailures) > 0:
		decision.Status = StatusHardFail
	case len(decision.SoftFailures) > 0:
		decision.Status = StatusSoftFail
	default:
		decision.Status = StatusPass
	}

	decision.sort()

	return &decision, nil
}

func extractViolations(data map[string]interface{}, rule string) []Violation {
	var violations []Violation

	switch reasonsType := data[rule].(type) {
	case []interface{}:
		reasons, err := asStringSlice(reasonsType)
		if err != nil {
			break // TODO: should we fail if rules return non string reasons? Should we only report string reasons?
		}
		for _, reason := range reasons {
			violations = append(violations, Violation{Rule: rule, Reason: reason})
		}
	case map[string]interface{}:
		for _, value := range reasonsType {
			reason, ok := value.(string)
			if !ok {
				continue // TODO what to do about non-string reasons?
			}
			violations = append(violations, Violation{Rule: rule, Reason: reason})
		}
	case string:
		violations = append(violations, Violation{Rule: rule, Reason: reasonsType})
	}

	return violations
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
