package tester

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/CircleCI-Public/circle-policy-agent/cpa"
	"github.com/CircleCI-Public/circle-policy-agent/internal"
	"github.com/CircleCI-Public/circle-policy-agent/internal/junit"

	"gopkg.in/yaml.v3"
)

var (
	clock Clock = StandardClock{}
)

type Clock interface {
	Now() time.Time
}

type StandardClock struct{}

func (StandardClock) Now() time.Time { return time.Now() }

type MockedClock struct{}

func (MockedClock) Now() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-03-04T10:50:05Z")
	return t
}

type Result struct {
	Group  string
	Name   string
	Passed bool
	Err    error

	Elapsed time.Duration

	Ctx any
}

func (r Result) MarshalJSON() ([]byte, error) {
	value := struct {
		Passed    bool
		Group     string
		Name      string `json:",omitempty"`
		Elapsed   string
		ElapsedMS int64
		Err       string `json:",omitempty"`
		Ctx       any    `json:",omitempty"`
	}{
		Passed:    r.Passed,
		Group:     r.Group,
		Name:      r.Name,
		Elapsed:   r.Elapsed.String(),
		ElapsedMS: r.Elapsed.Milliseconds(),
		Err: func() string {
			if r.Err == nil {
				return ""
			}
			return r.Err.Error()
		}(),
		Ctx: r.Ctx,
	}

	return json.Marshal(value)
}

type ResultHandler interface {
	HandleResults(c <-chan Result) (success bool)
}
type StandardResultHandler struct {
	table   internal.TableWriter
	verbose bool
	debug   bool
}

func (rh StandardResultHandler) HandleResults(c <-chan Result) bool {
	type Group struct {
		Name    string
		Status  string
		Elapsed time.Duration
	}

	var (
		currentGroup Group
		failed       int
		passed       int
		errorGroups  int
		totalTime    time.Duration
	)

	for result := range c {
		totalTime += result.Elapsed

		// On group changes we must print the current group status before updating
		if result.Group != currentGroup.Name {
			if currentGroup.Name != "" {
				rh.table.Row(currentGroup.Status, currentGroup.Name, fmt.Sprintf("%.3fs", currentGroup.Elapsed.Seconds()))
			}
			currentGroup = Group{Status: "ok", Name: result.Group}
		}

		// Handle an Error Group
		if result.Name == "" {
			switch {
			case errors.Is(result.Err, cpa.ErrNoPolicies):
				rh.table.Row("?", result.Group, "no policies")
			case errors.Is(result.Err, ErrNoTests):
				rh.table.Row("?", result.Group, "no tests")
			default:
				errorGroups++
				rh.table.Row("fail", result.Group, result.Err)
			}
			// We have printed our group result since we knew it immediately since it was an error group,
			// thus we can reset group state to nothing so it doesn't get printed twice on group switches
			currentGroup = Group{}
			continue
		}

		currentGroup.Elapsed += result.Elapsed
		if result.Passed {
			passed++
			if rh.verbose {
				rh.table.Row("ok", result.Name, fmt.Sprintf("%.3fs", result.Elapsed.Seconds()))
			}
		} else {
			currentGroup.Status = "fail"
			failed++
			rh.table.Row("FAIL", result.Name, fmt.Sprintf("%.3fs", result.Elapsed.Seconds()))
			if result.Err != nil {
				rh.table.Textln("\n" + indent(result.Err.Error(), "    "))
			}
		}
		if rh.debug {
			rh.table.Textln("---- Debug Test Context ----")
			_ = yaml.NewEncoder(rh.table).Encode(result.Ctx)
			rh.table.Textln("---- End of Test Context ---\n")
		}
	}

	// Print the last group status after the result loop ends
	if currentGroup.Name != "" {
		rh.table.Row(currentGroup.Status, currentGroup.Name, fmt.Sprintf("%.3fs", currentGroup.Elapsed.Seconds()))
	}

	rh.table.Textf("\n%d/%d tests passed (%.3fs)\n", passed, passed+failed, totalTime.Seconds())

	return failed == 0 && errorGroups == 0
}

type JSONResultHandler struct {
	w     io.Writer
	debug bool
}

func (jrh JSONResultHandler) HandleResults(c <-chan Result) bool {
	var (
		ok      = true
		results = []Result{}
	)

	for result := range c {
		if !result.Passed {
			ok = false
		}
		if !jrh.debug {
			result.Ctx = nil
		}
		results = append(results, result)
	}

	enc := json.NewEncoder(jrh.w)
	enc.SetIndent("", "  ")

	_ = enc.Encode(results)
	return ok
}

type JUnitResultHandler struct {
	w io.Writer
}

func (rh JUnitResultHandler) HandleResults(c <-chan Result) bool {
	var (
		root             = junit.JUnitTestSuites{Name: "root"}
		currentSuite     junit.JUnitTestSuite
		currentSuiteTime time.Duration
		totalTime        time.Duration
	)

	finalizeCurrentSuite := func() {
		if currentSuite.Name == "" {
			return
		}
		currentSuite.Time = fmt.Sprintf("%.3f", currentSuiteTime.Seconds())
		currentSuite.Timestamp = clock.Now().Format(time.RFC3339)
		currentSuiteTime = 0
		root.Suites = append(root.Suites, currentSuite)
	}

	for result := range c {
		if result.Group != currentSuite.Name {
			finalizeCurrentSuite()
			currentSuite = junit.JUnitTestSuite{Name: result.Group}
		}

		// Handle an Error Group
		if result.Name == "" {
			switch {
			case errors.Is(result.Err, cpa.ErrNoPolicies):
				currentSuite.Properties = []junit.JUnitProperty{{Name: "skipped", Value: "no policies"}}
			case errors.Is(result.Err, ErrNoTests):
				currentSuite.Properties = []junit.JUnitProperty{{Name: "skipped", Value: "no tests"}}
			default:
				root.Errors++
			}
			finalizeCurrentSuite()
			currentSuite = junit.JUnitTestSuite{}
			continue
		}

		totalTime += result.Elapsed
		currentSuiteTime += result.Elapsed

		currentSuite.Tests++
		root.Tests++

		if !result.Passed {
			root.Failures++
			currentSuite.Failures++
		}

		currentSuite.TestCases = append(currentSuite.TestCases, junit.JUnitTestCase{
			Classname: result.Group,
			Name:      result.Name,
			Time:      fmt.Sprintf("%.3f", result.Elapsed.Seconds()),
			Failure: func() *junit.JUnitFailure {
				if result.Passed {
					return nil
				}
				return &junit.JUnitFailure{
					Message: "failed",
					Contents: func() string {
						if result.Err == nil {
							return ""
						}
						return result.Err.Error()
					}(),
				}
			}(),
		})
	}

	// Print the last group status after the result loop ends
	finalizeCurrentSuite()

	root.Time = fmt.Sprintf("%.3f", totalTime.Seconds())

	encoder := xml.NewEncoder(rh.w)
	encoder.Indent("", "\t")

	_, _ = io.WriteString(rh.w, xml.Header)
	_ = encoder.Encode(root)
	_, _ = io.WriteString(rh.w, "\n")

	return root.Failures+root.Errors == 0
}

func MakeJUnitResultHandler(opts ResultHandlerOptions) JUnitResultHandler {
	return JUnitResultHandler{
		w: opts.Dst,
	}
}

type ResultHandlerOptions struct {
	Verbose bool
	Debug   bool
	Dst     io.Writer
}

func MakeDefaultResultHandler(opts ResultHandlerOptions) StandardResultHandler {
	if opts.Dst == nil {
		opts.Dst = os.Stderr
	}
	if opts.Debug {
		opts.Verbose = true
	}
	return StandardResultHandler{
		table:   internal.MakeTableWriter(opts.Dst),
		verbose: opts.Verbose,
		debug:   opts.Debug,
	}
}

func MakeJSONResultHandler(opts ResultHandlerOptions) JSONResultHandler {
	if opts.Dst == nil {
		opts.Dst = os.Stderr
	}
	return JSONResultHandler{opts.Dst, opts.Debug}
}

func indent(value, indent string) string {
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
