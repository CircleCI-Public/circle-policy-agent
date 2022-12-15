package tester

import (
	"fmt"
	"reflect"
	"time"

	"gopkg.in/yaml.v2"
)

type Result struct {
	Name string
	Ok   bool
	Err  error

	Elapsed time.Duration

	Ctx any
}

func (runner Runner) PrintResult(result Result) {
	if result.Ok && !runner.opts.Debug && !runner.opts.Verbose {
		return
	}

	status := "PASS"
	if !result.Ok {
		status = "FAIL"
	}

	runner.writer.Row(status, result.Name, fmt.Sprintf("%.3fs", result.Elapsed.Seconds()))

	if !result.Ok && result.Err != nil {
		runner.writer.Textln(result.Err.Error())
	}
	if runner.opts.Debug && !reflect.ValueOf(result.Ctx).IsZero() {
		runner.writer.Textln("\n------- Begin Test Context ------")
		_ = yaml.NewEncoder(runner.writer).Encode(result.Ctx)
		runner.writer.Textln("------- End Test Context --------\n")
	}
}
