// Package report renders an Analysis as line oriented text, as JSON, or as
// the class reference printed by the explain command.
//
// Text output is ordered and padded so that two runs against the same input
// produce byte identical bytes and a diff of two exports stays readable.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"reorglens/internal/block"
	"reorglens/internal/chain"
)

// Meta describes the run that produced the analysis.
type Meta struct {
	Version  string
	Command  string
	Input    string
	Detail   bool
	ExitCode int
}

// Finding is one entry of the JSON detail list. The fields map holds only the
// keys that apply to the class, and encoding/json sorts those keys, so the
// encoding stays deterministic.
type Finding struct {
	Class    string         `json:"class"`
	Severity string         `json:"severity"`
	Fields   map[string]any `json:"fields"`
}

type jsonReport struct {
	Tool      string             `json:"tool"`
	Version   string             `json:"version"`
	Command   string             `json:"command"`
	Input     string             `json:"input"`
	Window    jsonWindow         `json:"window"`
	Records   jsonRecords        `json:"records"`
	Chain     jsonChain          `json:"chain"`
	Parent    chain.ParentStats  `json:"parent"`
	Timestamp chain.TimestampStats `json:"timestamp"`
	Producers []chain.ProducerStats `json:"producers"`
	Findings  jsonFindings       `json:"findings"`
	Detail    []Finding          `json:"detail"`
	BadLines  []block.BadLine    `json:"badLines"`
	Result    string             `json:"result"`
	ExitCode  int                `json:"exitCode"`
}

type jsonWindow struct {
	From    uint64 `json:"from"`
	To      uint64 `json:"to"`
	Numbers uint64 `json:"numbers"`
}

type jsonRecords struct {
	Valid   int `json:"valid"`
	Invalid int `json:"invalid"`
	Outside int `json:"outside"`
}

type jsonChain struct {
	CanonicalRecords int `json:"canonicalRecords"`
	OrphanRecords    int `json:"orphanRecords"`
	OrphanSegments   int `json:"orphanSegments"`
	Missing          int `json:"missing"`
	Duplicates       int `json:"duplicates"`
	Conflicts        int `json:"conflicts"`
}

type jsonFindings struct {
	Total   int            `json:"total"`
	Error   int            `json:"error"`
	Warn    int            `json:"warn"`
	ByClass map[string]int `json:"byClass"`
}

// Result names the state of the data, independent of the exit threshold.
func Result(a *chain.Analysis) string {
	if a.TotalFindings == 0 {
		return "clean"
	}
	return "findings present"
}

// Text writes the line oriented report.
func Text(w io.Writer, m Meta, a *chain.Analysis) error {
	var b strings.Builder

	fmt.Fprintf(&b, "reorglens %s %s %s\n", m.Version, m.Command, m.Input)
	b.WriteByte('\n')
	for _, line := range summaryLines(a) {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	if m.Detail {
		b.WriteByte('\n')
		if a.TotalFindings == 0 {
			b.WriteString("findings: none\n")
		} else {
			for _, line := range findingLines(a) {
				b.WriteString(line)
				b.WriteByte('\n')
			}
		}
		if len(a.BadLines) > 0 {
			b.WriteByte('\n')
			fmt.Fprintf(&b, "invalid lines (%d):\n", len(a.BadLines))
			for _, bad := range a.BadLines {
				fmt.Fprintf(&b, "  line %d: %s  %s\n", bad.Line, bad.Reason, bad.Text)
			}
		}
		b.WriteByte('\n')
		fmt.Fprintf(&b, "%-12s%s\n", "result", Result(a))
		fmt.Fprintf(&b, "%-12s%d\n", "exit code", m.ExitCode)
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// JSON writes the machine readable report with fixed field names.
func JSON(w io.Writer, m Meta, a *chain.Analysis) error {
	producers := a.Producers
	if producers == nil {
		producers = make([]chain.ProducerStats, 0)
	}
	detail := make([]Finding, 0)
	if m.Detail {
		detail = findings(a)
	}
	bad := a.BadLines
	if bad == nil {
		bad = make([]block.BadLine, 0)
	}

	rep := jsonReport{
		Tool:    "reorglens",
		Version: m.Version,
		Command: m.Command,
		Input:   m.Input,
		Window: jsonWindow{
			From: a.From, To: a.To, Numbers: a.Numbers,
		},
		Records: jsonRecords{
			Valid: a.Valid, Invalid: a.Invalid, Outside: a.Outside,
		},
		Chain: jsonChain{
			CanonicalRecords: a.CanonicalRecords,
			OrphanRecords:    a.OrphanRecords,
