package tester

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/CircleCI-Public/circle-policy-agent/cpa"
	"github.com/CircleCI-Public/circle-policy-agent/internal"
	"github.com/yazgazan/jaydiff/diff"
	"golang.org/x/exp/slices"
)

type Test struct {
	Input    any
	Meta     any
	Decision any
	Cases    map[string]*Test
}

func (t Test) NamedCases() []NamedTest {
	result := make([]NamedTest, 0, len(t.Cases))
	for name, test := range t.Cases {
		result = append(result, NamedTest{name, test})
	}
	slices.SortFunc(result, func(a, b NamedTest) bool { return a.Name < b.Name })
	return result
}

type NamedTest struct {
	Name string
	*Test
}

type ParentTestContext struct {
	Name  string
	Input any
	Meta  any
}

type TestRunOptions struct {
	Parent  ParentTestContext
	Include *regexp.Regexp
}

func (t NamedTest) Run(policy *cpa.Policy, opts TestRunOptions) []Result {
	input := func() any {
		if t.Input == nil {
			return opts.Parent.Input
		}
		return internal.Merge(opts.Parent.Input, t.Input)
	}()

	meta := func() any {
		if t.Meta == nil {
			return opts.Parent.Meta
		}
		return internal.Merge(opts.Parent.Meta, t.Meta)
	}()

	name := t.Name
	if opts.Parent.Name != "" {
		name = opts.Parent.Name + "/" + name
	}

	var results []Result

	if opts.Include == nil || opts.Include.MatchString(name) {
		eval, _ := policy.Eval(context.Background(), "data", input, cpa.Meta(meta))

		start := time.Now()
		var decision any = internal.Must(policy.Decide(context.Background(), input, cpa.Meta(meta)))
		elapsed := time.Since(start)

		decision = internal.Must(internal.ToRawInterface(decision))

		d := internal.Must(diff.Diff(decision, t.Decision))

		if d.Diff() != diff.Identical {
			fmt.Printf("")
		}

		results = append(results, Result{
			Name: name,
			Ok:   d.Diff() == diff.Identical,
			Err: func() error {
				if d.Diff() == diff.Identical {
					return nil
				}
				return errors.New(d.StringIndent("", "  ", diff.Output{
					Indent:     "  ",
					Colorized:  true,
					JSON:       true,
					JSONValues: true,
				}))
			}(),
			Elapsed: elapsed,
			Ctx: map[string]any{
				"input":      input,
				"meta":       meta,
				"decision":   decision,
				"evaluation": eval,
			},
		})
	}

	for _, subtest := range t.NamedCases() {
		results = append(results, subtest.Run(policy, TestRunOptions{
			Parent: ParentTestContext{
				Name:  name,
				Input: input,
				Meta:  meta,
			},
			Include: opts.Include,
		})...)
	}

	return results
}
