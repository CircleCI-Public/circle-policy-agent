package cpa

import "sort"

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
