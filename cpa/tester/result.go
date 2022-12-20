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
		Ok        bool
		Group     string
		Name      string `json:",omitempty"`
		Elapsed   string
		ElapsedMS int64
		Err       string `json:",omitempty"`
		Ctx       any    `json:",omitempty"`
	}{
		Ok:        r.Ok,
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
	)

	for result := range c {
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
			if result.Err != nil {
				rh.table.Textln(result.Err.Error())
			}
		}
		if rh.debug {
			rh.table.Textln("---- Debug Test Context ----")
			_ = yaml.NewEncoder(rh.table).Encode(result.Ctx)
			rh.table.Textln("---- End of Test Context ---")
		}
	}

	if group.Name != "" {
		rh.table.Row(group.Status, group.Name, fmt.Sprintf("%.3fs", group.Elapsed.Seconds()))
	}

	rh.table.Textf("\n%d/%d tests passed (%.3fs)\n", passed, passed+failed, totalTime.Seconds())

	return failed == 0 && errorGroups == 0
}

type jsonResultHandler struct {
	w     io.Writer
	debug bool
}

func (jrh jsonResultHandler) HandleResults(c <-chan Result) bool {
	var (
		ok      = true
		results = []Result{}
	)

	for result := range c {
		if !result.Ok {
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

type ResultHandlerOptions struct {
	Verbose bool
	Debug   bool
	Dst     io.Writer
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

func MakeJSONResultHandler(opts ResultHandlerOptions) ResultHandler {
	if opts.Dst == nil {
		opts.Dst = os.Stderr
	}
	return jsonResultHandler{opts.Dst, opts.Debug}
}
