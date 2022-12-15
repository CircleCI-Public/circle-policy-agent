package tester

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/CircleCI-Public/circle-policy-agent/cpa"
	"github.com/CircleCI-Public/circle-policy-agent/internal"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/tester"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

type Runner struct {
	writer internal.TableWriter
	opts   RunnerOptions

	total      int
	failed     int
	folders    []string
	hasErrored bool
}

type RunnerOptions struct {
	Path    string
	Dst     io.Writer
	Verbose bool
	Debug   bool
	Include *regexp.Regexp
}

func NewRunner(opts RunnerOptions) (*Runner, error) {
	if opts.Path == "" {
		opts.Path = "./..."
	}
	if opts.Dst == nil {
		opts.Dst = os.Stderr
	}

	folder, err := getTestFolders(opts.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup test folders: %w", err)
	}

	return &Runner{
		writer:  internal.MakeTableWriter(opts.Dst),
		folders: folder,
		opts:    opts,
	}, nil
}

func (runner *Runner) Run() bool {
	start := time.Now()

	runner.runOpaTests()
	for _, folder := range runner.folders {
		runner.runFolder(folder)
	}

	runner.writer.Textf("\n%d/%d tests passed (%.3fs)\n", runner.total-runner.failed, runner.total, time.Since(start).Seconds())

	return !runner.hasErrored && runner.failed == 0
}

func (runner *Runner) runOpaTests() {
	root := runner.folders[0]

	policy, err := cpa.LoadPolicyFromFS(root)
	if err != nil {
		if errors.Is(err, cpa.ErrNoPolicies) {
			return
		}
		runner.hasErrored = true
		runner.writer.Row("FAIL", "<opa.tests>", err)
		return
	}

	modules := make(map[string]*ast.Module, len(policy.Source()))
	for key, value := range policy.Source() {
		modules[key] = ast.MustParseModule(value)
	}

	start := time.Now()
	status := "ok"

	var count int
	for r := range internal.Must(tester.NewRunner().Run(context.Background(), modules)) {
		name := "<opa.tests>/" + r.Package + "." + r.Name
		if runner.opts.Include != nil && !runner.opts.Include.MatchString(name) {
			continue
		}
		count++
		runner.total++
		if !r.Pass() {
			status = "fail"
			runner.failed++
		}
		runner.printResult(Result{
			Name:    name,
			Ok:      r.Pass(),
			Elapsed: r.Duration,
			Err:     r.Error,
		})
	}

	if count == 0 {
		runner.writer.Row(status, "<opa.tests>", "no tests")
		return
	}

	runner.writer.Row(status, "<opa.tests>", fmt.Sprintf("%.3fs", time.Since(start).Seconds()))
}

func (runner *Runner) runFolder(folder string) {
	start := time.Now()
	results, err := runner.execFolderTests(folder)
	elapsed := time.Since(start)

	runner.total += len(results)

	// Flush before test block
	if len(results) > 0 && (runner.opts.Verbose || runner.opts.Debug) {
		runner.writer.Flush()
	}

	status := "ok"
	for _, result := range results {
		if !result.Ok {
			runner.failed++
			status = "FAIL"
		}
		runner.printResult(result)
	}

	// flush after test block
	if len(results) > 0 && (runner.opts.Verbose || runner.opts.Debug) {
		runner.writer.Flush()
	}

	if err != nil {
		if errors.Is(err, cpa.ErrNoPolicies) {
			runner.writer.Row("ok", folder, "no policies")
			return
		}
		runner.hasErrored = true
		runner.writer.Row("fail", folder, err)
		return
	}

	if len(results) == 0 {
		runner.writer.Row("ok", folder, "no tests")
		return
	}

	runner.writer.Row(status, folder, fmt.Sprintf("%.3fs", elapsed.Seconds()))
}

func (runner Runner) execFolderTests(folder string) ([]Result, error) {
	policy, err := cpa.LoadPolicyFromFS(folder)
	if err != nil {
		return nil, err
	}

	nameSet := map[string]struct{}{}
	var namedTests []NamedTest

	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		for name, test := range tests {
			if _, ok := nameSet[name]; ok {
				return nil, fmt.Errorf("test name conflict: %q", name)
			}
			nameSet[name] = struct{}{}
			namedTests = append(namedTests, NamedTest{name, test})
		}
	}

	slices.SortFunc(namedTests, func(a, b NamedTest) bool { return a.Name < b.Name })

	var results []Result
	for _, t := range namedTests {
		results = append(results, t.Run(policy, testRunOptions{
			Parent:  parentCtx{},
			Include: runner.opts.Include,
		})...)
	}

	return results, nil
}

func loadTests(path string) (tests map[string]*Test, err error) {
	//nolint gosec
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &tests)
	if err != nil {
		return
	}
	for _, t := range tests {
		sanitizeTest(t)
	}
	return
}

func getTestFolders(path string) (folders []string, err error) {
	if path != "./..." {
		path = strings.TrimPrefix(path, "./")
	}
	if !strings.HasSuffix(path, "/...") {
		return []string{path}, nil
	}
	err = filepath.WalkDir(path[:len(path)-4], func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if name := d.Name(); len(name) > 1 && name[0] == '.' {
			return filepath.SkipDir
		}
		folders = append(folders, path)
		return nil
	})
	return
}

func sanitizeTest(t *Test) {
	if t == nil {
		return
	}
	t.Decision = internal.Must(internal.ConvertYAMLMapKeyTypes(t.Decision))
	t.Input = internal.Must(internal.ConvertYAMLMapKeyTypes(t.Input))
	t.Meta = internal.Must(internal.ConvertYAMLMapKeyTypes(t.Meta))
	for _, t := range t.Cases {
		sanitizeTest(t)
	}
}
