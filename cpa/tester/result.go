package tester

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/CircleCI-Public/circle-policy-agent/cpa"
	"github.com/CircleCI-Public/circle-policy-agent/internal"
	"gopkg.in/yaml.v3"
)

type Result struct {
	Group string
	Name  string
	Ok    bool
	Err   error

	Elapsed time.Duration

	Ctx any
}

func (r Result) MarshalJSON() ([]byte, error) {
	value := struct {
		Group     string
		Name      string
		Ok        bool
		Err       string
		ElapsedMS int64
		Ctx       any
	}{
		Group: r.Group,
		Name:  r.Name,
		Ok:    r.Ok,
		Err: func() string {
			if r.Err == nil {
				return ""
			}
			return r.Err.Error()
		}(),
		ElapsedMS: r.Elapsed.Milliseconds(),
		Ctx:       r.Ctx,
	}

	return json.Marshal(value)
}

type ResultHandler interface {
	HandleResults(c <-chan Result) (success bool)
}
type resultHandler struct {
	table   internal.TableWriter
	verbose bool
	debug   bool
}

func (rh resultHandler) HandleResults(c <-chan Result) bool {
	type Group struct {
		Name    string
		Status  string
		Elapsed time.Duration
	}

	var (
		group       Group
		failed      int
		passed      int
		errorGroups int
		totalTime   time.Duration
		result      Result
	)

	for result = range c {
		totalTime += result.Elapsed

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
			continue
		}

		if result.Group != group.Name {
			if group.Name != "" {
				rh.table.Row(group.Status, group.Name, fmt.Sprintf("%.3fs", group.Elapsed.Seconds()))
			}
			group = Group{Status: "ok", Name: result.Group}
		}

		group.Elapsed += result.Elapsed
		if result.Ok {
			passed++
			if rh.verbose {
				rh.table.Row("ok", result.Name, fmt.Sprintf("%.3fs", result.Elapsed.Seconds()))
			}
		} else {
			group.Status = "fail"
			failed++
			rh.table.Row("FAIL", result.Name, fmt.Sprintf("%.3fs", result.Elapsed.Seconds()))
			rh.table.Textln(result.Err.Error())
		}
		if rh.debug {
			rh.table.Textln("---- Debug Test Context ----")
			_ = yaml.NewEncoder(rh.table).Encode(result.Ctx)
			rh.table.Textln("---- End of Test Context ---")
		}
	}

	if result.Name != "" {
		rh.table.Row(group.Status, group.Name, fmt.Sprintf("%.3fs", group.Elapsed.Seconds()))
	}

	rh.table.Textf("\n%d/%d tests passed (%.3fs)\n", passed, passed+failed, totalTime.Seconds())

	return failed == 0 && errorGroups == 0
}

type ResultHandlerOptions struct {
	Verbose bool
	Debug   bool
	Dst     io.WriteCloser
}

func MakeDefaultResultHandler(opts ResultHandlerOptions) ResultHandler {
	if opts.Dst == nil {
		opts.Dst = os.Stderr
	}
	if opts.Debug {
		opts.Verbose = true
	}
	return resultHandler{
		table:   internal.MakeTableWriter(opts.Dst),
		verbose: opts.Verbose,
		debug:   opts.Debug,
	}
}
