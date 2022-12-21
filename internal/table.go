package internal

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type TableWriter struct {
	tabWriter *tabwriter.Writer
	w         io.Writer
}

func MakeTableWriter(w io.Writer) TableWriter {
	return TableWriter{
		tabWriter: tabwriter.NewWriter(w, 0, 0, 4, ' ', 0),
		w:         w,
	}
}

func (tw TableWriter) Write(data []byte) (int, error) {
	tw.Flush()
	return tw.w.Write(data)
}

func (tw TableWriter) Row(args ...any) {
	format := strings.Repeat("%v\t", len(args)) + "\n"
	_, _ = tw.tabWriter.Write([]byte(fmt.Sprintf(format, args...)))
}

func (tw TableWriter) Flush() {
	_ = tw.tabWriter.Flush()
}

func (tw TableWriter) Text(value string) {
	tw.Flush()
	_, _ = io.WriteString(tw.w, value)
}

func (tw TableWriter) Textf(format string, args ...any) {
	tw.Text(fmt.Sprintf(format, args...))
}

func (tw TableWriter) Textln(value string) {
	tw.Textf(value + "\n")
}
