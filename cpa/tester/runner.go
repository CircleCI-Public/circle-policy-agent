package tester

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/CircleCI-Public/circle-policy-agent/cpa"
	"github.com/CircleCI-Public/circle-policy-agent/internal"
	"github.com/open-policy-agent/opa/tester"
	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

type Runner struct {
	include *regexp.Regexp
	folders []string
	compile func([]byte, map[string]any) ([]byte, error)
}

type RunnerOptions struct {
	Path    string
	Include *regexp.Regexp
	Compile func([]byte, map[string]any) ([]byte, error)
}

var ErrNoTests = errors.New("no tests")

func NewRunner(opts RunnerOptions) (*Runner, error) {
	if opts.Path == "" {
		opts.Path = "./..."
	}

	folders, err := getTestFolders(opts.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup test folders: %w", err)
	}

	return &Runner{
		include: opts.Include,
		folders: folders,
		compile: opts.Compile,
	}, nil
}

func (runner *Runner) Run() <-chan Result {
	results := make(chan Result)

	go func() {
		runner.runOpaTests(results)
		for _, folder := range runner.folders {
			runner.runFolder(folder, results)
		}
		close(results)
	}()

	return results
}

func (runner *Runner) RunAndHandleResults(handler ResultHandler) bool {
	return handler.HandleResults(runner.Run())
}

func (runner *Runner) runOpaTests(results chan<- Result) {
	root := runner.folders[0]

	policy, err := cpa.LoadPolicyFromFS(root)
	if err != nil {
		if errors.Is(err, cpa.ErrNoPolicies) {
			return
		}
		results <- Result{
			Group: "<opa.tests>",
			Err:   err,
		}
		return
	}

	for r := range internal.Must2(tester.NewRunner().Run(context.Background(), policy.Modules())) {
		name := r.Package + "." + r.Name
		if runner.include != nil && !runner.include.MatchString(name) {
			continue
		}
		results <- Result{
			Group:   "<opa.tests>",
			Name:    name,
			Passed:  r.Pass(),
			Elapsed: r.Duration,
			Err:     r.Error,
		}
	}
}

func (runner *Runner) runFolder(folder string, results chan<- Result) {
	policy, err := cpa.LoadPolicyFromFS(folder)
	if err != nil {
		results <- Result{
			Group:  folder,
			Err:    err,
			Passed: errors.Is(err, cpa.ErrNoPolicies),
		}
		return
	}

	nameSet := map[string]struct{}{}
	var namedTests []NamedTest

	entries, err := os.ReadDir(folder)
	if err != nil {
		results <- Result{
			Group: folder,
			Err:   err,
		}
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if name := entry.Name(); !strings.HasSuffix(name, "_test.yaml") && !strings.HasSuffix(name, "_test.yml") {
			continue
		}

		testPath := filepath.Join(folder, entry.Name())

		tests, err := loadTests(testPath)
		if err != nil {
			results <- Result{
				Group: folder,
				Err:   err,
			}
			return
		}

		for name, test := range tests {
			if _, ok := nameSet[name]; ok {
				results <- Result{
					Group: folder,
					Err:   fmt.Errorf("test name conflict: %q", name),
				}
				return
			}
			nameSet[name] = struct{}{}
			namedTests = append(namedTests, NamedTest{name, test})
		}
	}

	if len(namedTests) == 0 {
		results <- Result{
			Group:  folder,
			Passed: true,
			Err:    ErrNoTests,
		}
		return
	}

	slices.SortFunc(namedTests, func(a, b NamedTest) bool { return a.Name < b.Name })

	for _, t := range namedTests {
		runner.runTest(policy, results, t, folder, ParentTestContext{})
	}
}

func (runner *Runner) runTest(policy *cpa.Policy, results chan<- Result, t NamedTest, group string, parent ParentTestContext) {
	input := func() any {
		if t.Input == nil {
			return parent.Input
		}
		return internal.Merge(parent.Input, t.Input)
	}()

	meta := func() any {
		if t.Meta == nil {
			return parent.Meta
		}
		return internal.Merge(parent.Meta, t.Meta)
	}()

	decision := func() any {
		if t.Decision == nil {
			return parent.Decision
		}
		return internal.Merge(parent.Decision, t.Decision)
	}()

	compile := func() bool {
		if t.Compile != nil {
			return *t.Compile
		}
		return parent.Compile
	}()

	pipelineParams := func() map[string]any {
		if t.PipelineParameters == nil {
			return parent.PipelineParameters
		}
		return internal.MergeMaps(parent.PipelineParameters, t.PipelineParameters)
	}()

	name := t.Name
	if parent.Name != "" {
		name = parent.Name + "/" + name
	}

	if decision != nil && (runner.include == nil || runner.include.MatchString(name)) {
		func() {
			if compile || pipelineParams != nil {
				if runner.compile == nil {
					results <- Result{
						Group: group,
						Name:  name,
						Err:   errors.New("test set compile to true but no compiler was provided"),
					}
					return
				}

				inputData := internal.Must2(yaml.Marshal(input))

				compiledData, err := runner.compile(inputData, pipelineParams)
				if err != nil {
					results <- Result{
						Group: group,
						Name:  name,
						Err:   fmt.Errorf("failed to compile test input: %w", err),
						Ctx: map[string]any{
							"input":               input,
							"pipeline-parameters": pipelineParams,
						},
					}
					return
				}

				input = func() any {
					i := input.(map[string]any)
					var compiled any
					internal.Must(yaml.Unmarshal(compiledData, &compiled))
					i["_compiled_"] = compiled
					return i
				}()
			}

			eval, _ := policy.Eval(context.Background(), "data", input, cpa.Meta(meta))

			start := time.Now()
			var actualDecision any = internal.Must2(policy.Decide(context.Background(), input, cpa.Meta(meta)))
			elapsed := time.Since(start)

			actualDecision = internal.Must2(internal.ToRawInterface(actualDecision))

			diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(internal.Must2(yamlfy(decision))),
				FromFile: "Expected",
				B:        difflib.SplitLines(internal.Must2(yamlfy(actualDecision))),
				ToFile:   "Actual",
				Context:  1,
			})

			results <- Result{
				Group:  group,
				Name:   name,
				Passed: diff == "",
				Err: func() error {
					if diff == "" {
						return nil
					}
					return errors.New(diff)
				}(),
				Elapsed: elapsed,
				Ctx: map[string]any{
					"input":      input,
					"decision":   actualDecision,
					"evaluation": eval,
				},
			}
		}()
	}

	for _, subtest := range t.NamedCases() {
		runner.runTest(policy, results, subtest, group, ParentTestContext{
			Name:               name,
			Input:              input,
			Meta:               meta,
			Decision:           decision,
			Compile:            compile,
			PipelineParameters: pipelineParams,
		})
	}
}

func yamlfy(value any) (string, error) {
	raw, err := yaml.Marshal(value)
	return string(raw), err
}
