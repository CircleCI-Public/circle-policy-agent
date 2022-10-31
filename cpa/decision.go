package cpa

import (
	"fmt"
	"sort"
)

type Status string

const (
	StatusPass     Status = "PASS"
	StatusSoftFail Status = "SOFT_FAIL"
	StatusHardFail Status = "HARD_FAIL"
)

type Violation struct {
	Rule   string `json:"rule"`
	Reason string `json:"reason"`
}

// Decision is a circleci flavoured output representing a policy decision.
type Decision struct {
	Status       Status      `json:"status"`
	EnabledRules []string    `json:"enabled_rules,omitempty"`
	HardFailures []Violation `json:"hard_failures,omitempty"`
	SoftFailures []Violation `json:"soft_failures,omitempty"`
}

// sort will sort the decision's enabled rules and hard/soft violations
// to make decision output more predictable for users and more easily testable.
// Values are sorted lexicographically.
// Violations sorted by the combination of their rule and reason.
func (d Decision) sort() {
	sort.StringSlice(d.EnabledRules).Sort()
	for _, violations := range [][]Violation{d.SoftFailures, d.HardFailures} {
		sort.SliceStable(violations, func(i, j int) bool {
			left, right := violations[i], violations[j]
			return left.Rule+left.Reason < right.Rule+right.Reason
		})
	}
}

func (d *Decision) evaluate(contextName string, result map[string]any) error {
	if result == nil {
		return nil
	}

	enabled, err := asStringSlice(result["enable_rule"])
	if err != nil {
		return fmt.Errorf("invalid enable_rule: %w", err)
	}

	hardFailRules, err := asStringSlice(result["hard_fail"])
	if err != nil {
		return fmt.Errorf("invalid hard_fail: %w", err)
	}

	hardFailMap := make(map[string]struct{}, len(hardFailRules))
	for _, rule := range hardFailRules {
		hardFailMap[rule] = struct{}{}
	}

	for _, rule := range enabled {
		d.EnabledRules = append(d.EnabledRules, contextName+"."+rule)
		if _, ok := hardFailMap[rule]; ok {
			d.HardFailures = append(d.HardFailures, extractViolations(result, rule, contextName)...)
		} else {
			d.SoftFailures = append(d.SoftFailures, extractViolations(result, rule, contextName)...)
		}
	}

	return nil
}

func (d *Decision) finalize() {
	if d == nil {
		return
	}

	switch {
	case len(d.HardFailures) > 0:
		d.Status = StatusHardFail
	case len(d.SoftFailures) > 0:
		d.Status = StatusSoftFail
	default:
		d.Status = StatusPass
	}

	d.sort()
}

func extractViolations(data map[string]interface{}, rule, contextName string) []Violation {
	var violations []Violation

	qualifiedRule := contextName + "." + rule

	switch reasonsType := data[rule].(type) {
	case []interface{}:
		reasons, err := asStringSlice(reasonsType)
		if err != nil {
			break // TODO: should we fail if rules return non string reasons? Should we only report string reasons?
		}
		for _, reason := range reasons {
			violations = append(violations, Violation{Rule: qualifiedRule, Reason: reason})
		}
	case map[string]interface{}:
		for _, value := range reasonsType {
			reason, ok := value.(string)
			if !ok {
				continue // TODO what to do about non-string reasons?
			}
			violations = append(violations, Violation{Rule: qualifiedRule, Reason: reason})
		}
	case string:
		violations = append(violations, Violation{Rule: qualifiedRule, Reason: reasonsType})
	}

	return violations
}
