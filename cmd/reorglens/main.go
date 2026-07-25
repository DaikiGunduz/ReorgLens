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
