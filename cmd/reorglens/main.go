// Command reorglens analyses a captured block window and reports where the
// canonical chain, the parent links, or the timestamps do not hold together.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"reorglens/internal/block"
	"reorglens/internal/chain"
	"reorglens/internal/report"
)

const version = "0.1.0"

const usageText = `usage: reorglens <command> [flags]

commands:
  analyze   read a JSONL block export and report findings
  explain   print the finding classes, severities and exit codes
  version   print the version

run 'reorglens analyze -h' for the flags of the main command

exit codes: 0 clean, 1 findings, 2 usage or input error
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		io.WriteString(stderr, usageText)
		return 2
	}
	switch args[0] {
	case "analyze":
		return cmdAnalyze(args[1:], stdout, stderr)
	case "explain":
		return cmdExplain(args[1:], stdout, stderr)
	case "version":
		if len(args) > 1 {
			fmt.Fprintln(stderr, "reorglens: version takes no arguments")
			return 2
		}
		fmt.Fprintf(stdout, "reorglens %s\n", version)
		return 0
	case "help", "--help", "-h":
		io.WriteString(stdout, usageText)
		return 0
	default:
		fmt.Fprintf(stderr, "reorglens: unknown command %q\n\n", args[0])
		io.WriteString(stderr, usageText)
		return 2
	}
}

func cmdAnalyze(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fromFlag := fs.String("from", "", "start of the window as a block number")
	toFlag := fs.String("to", "", "end of the window as a block number")
	format := fs.String("format", "text", "output format: text or json")
	failOn := fs.String("fail-on", "any", "exit 1 threshold: any, error or never")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: reorglens analyze [flags] <file>")
		fs.PrintDefaults()
		fmt.Fprintln(stderr, "\nexit codes: 0 clean, 1 findings, 2 usage or input error")
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(stderr, "reorglens: --format must be text or json, got %q\n", *format)
		return 2
	}
	if *failOn != "any" && *failOn != "error" && *failOn != "never" {
		fmt.Fprintf(stderr, "reorglens: --fail-on must be any, error or never, got %q\n", *failOn)
		return 2
	}
	from, err := parseUint("from", *fromFlag)
	if err != nil {
		fmt.Fprintf(stderr, "reorglens: %v\n", err)
		return 2
	}
	to, err := parseUint("to", *toFlag)
	if err != nil {
		fmt.Fprintf(stderr, "reorglens: %v\n", err)
		return 2
	}

	path := fs.Arg(0)
	records, bad, err := read(path)
	if err != nil {
		fmt.Fprintf(stderr, "reorglens: %v\n", err)
		return 2
	}

	analysis, err := chain.Analyse(records, bad, chain.Options{From: from, To: to})
	if err != nil {
		fmt.Fprintf(stderr, "reorglens: %v\n", err)
		return 2
	}

	code := exitCode(analysis, *failOn)
	meta := report.Meta{
		Version:  version,
		Command:  "analyze",
		Input:    path,
		Detail:   true,
		ExitCode: code,
	}
	if err := write(stdout, *format, meta, analysis); err != nil {
		fmt.Fprintf(stderr, "reorglens: %v\n", err)
		return 2
	}
	return code
}

func cmdExplain(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		fmt.Fprintln(stderr, "reorglens: explain takes no arguments")
		return 2
	}
	if err := report.Explain(stdout); err != nil {
		fmt.Fprintf(stderr, "reorglens: %v\n", err)
		return 2
	}
	return 0
}

func read(path string) ([]*block.Record, []block.BadLine, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	records, bad, err := block.ReadAll(f)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", path, err)
	}
	return records, bad, nil
}

func write(w io.Writer, format string, meta report.Meta, a *chain.Analysis) error {
	if format == "json" {
		return report.JSON(w, meta, a)
	}
	return report.Text(w, meta, a)
}

func exitCode(a *chain.Analysis, failOn string) int {
	switch failOn {
	case "never":
		return 0
	case "error":
		if a.ErrorFindings > 0 {
			return 1
		}
		return 0
	default:
		if a.TotalFindings > 0 {
			return 1
		}
		return 0
	}
}

func parseUint(name, v string) (*uint64, error) {
	if v == "" {
		return nil, nil
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("--%s must be an unsigned integer, got %q", name, v)
	}
	return &n, nil
}
