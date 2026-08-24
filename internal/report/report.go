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
			OrphanSegments:   len(a.Orphans),
			Missing:          len(a.Missing),
			Duplicates:       len(a.Duplicates),
			Conflicts:        len(a.Conflicts),
		},
		Parent:    a.Parent,
		Timestamp: a.Timestamp,
		Producers: producers,
		Findings: jsonFindings{
			Total:   a.TotalFindings,
			Error:   a.ErrorFindings,
			Warn:    a.WarnFindings,
			ByClass: byClass(a),
		},
		Detail:   detail,
		BadLines: bad,
		Result:   Result(a),
		ExitCode: m.ExitCode,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

func summaryLines(a *chain.Analysis) []string {
	segment := "segments"
	if len(a.Orphans) == 1 {
		segment = "segment"
	}
	lines := []string{
		fmt.Sprintf("%-12s%d..%d  (%d numbers)", "window", a.From, a.To, a.Numbers),
		fmt.Sprintf("%-12s%d valid  %d invalid  %d outside", "records", a.Valid, a.Invalid, a.Outside),
		fmt.Sprintf("%-12s%d records", "canonical", a.CanonicalRecords),
		fmt.Sprintf("%-12s%d records in %d %s", "orphaned", a.OrphanRecords, len(a.Orphans), segment),
		fmt.Sprintf("%-12s%d checked  %d mismatch  %d unverified", "parent", a.Parent.Checked, a.Parent.Mismatch, a.Parent.Unverified),
		fmt.Sprintf("%-12s%d checked  %d out of order", "timestamp", a.Timestamp.Checked, a.Timestamp.OutOfOrder),
	}
	if a.HasProducer {
		lines = append(lines, fmt.Sprintf("%-12s%d", "producers", len(a.Producers)))
	}
	lines = append(lines,
		fmt.Sprintf("%-12s%d  (%d error, %d warn)", "findings", a.TotalFindings, a.ErrorFindings, a.WarnFindings))
	return lines
}

// findingLines keeps a fixed order: the three error classes first, then the
// three warn classes, each sorted by number because Analyse already sorts.
func findingLines(a *chain.Analysis) []string {
	lines := make([]string, 0, a.TotalFindings)
	for _, c := range a.Conflicts {
		lines = append(lines, fmt.Sprintf("%-5s  %-18s number=%d canonical=%d hashes=%s",
			chain.SeverityError, chain.ClassCanonicalConflict,
			c.Number, c.Records, strings.Join(c.Hashes, ",")))
	}
	for _, m := range a.Mismatches {
		lines = append(lines, fmt.Sprintf("%-5s  %-18s number=%d parent=%d expected=%s observed=%s",
			chain.SeverityError, chain.ClassParentMismatch,
			m.Number, m.ParentNumber, m.Expected, m.Observed))
	}
	for _, t := range a.Orders {
		lines = append(lines, fmt.Sprintf("%-5s  %-18s number=%d parent=%d timestamp=%d parentTimestamp=%d",
			chain.SeverityError, chain.ClassTimestampOrder,
			t.Number, t.ParentNumber, t.Timestamp, t.ParentTimestamp))
	}
	for _, n := range a.Missing {
		lines = append(lines, fmt.Sprintf("%-5s  %-18s number=%d",
			chain.SeverityWarn, chain.ClassMissing, n))
	}
	for _, d := range a.Duplicates {
		lines = append(lines, fmt.Sprintf("%-5s  %-18s number=%d records=%d hashes=%s",
			chain.SeverityWarn, chain.ClassDuplicate,
			d.Number, d.Records, strings.Join(d.Hashes, ",")))
	}
	for _, o := range a.Orphans {
		lines = append(lines, fmt.Sprintf("%-5s  %-18s from=%d to=%d records=%d",
			chain.SeverityWarn, chain.ClassOrphanSegment, o.From, o.To, o.Records))
	}
	return lines
}

// findings builds the JSON detail list in the same order as findingLines.
func findings(a *chain.Analysis) []Finding {
	out := make([]Finding, 0, a.TotalFindings)
	add := func(class, severity string, fields map[string]any) {
		out = append(out, Finding{Class: class, Severity: severity, Fields: fields})
	}
	for _, c := range a.Conflicts {
		add(chain.ClassCanonicalConflict, chain.SeverityError, map[string]any{
			"number": c.Number, "canonicalRecords": c.Records, "hashes": c.Hashes,
		})
	}
	for _, m := range a.Mismatches {
		add(chain.ClassParentMismatch, chain.SeverityError, map[string]any{
			"number": m.Number, "parentNumber": m.ParentNumber,
			"expected": m.Expected, "observed": m.Observed,
		})
	}
	for _, t := range a.Orders {
		add(chain.ClassTimestampOrder, chain.SeverityError, map[string]any{
			"number": t.Number, "parentNumber": t.ParentNumber,
			"timestamp": t.Timestamp, "parentTimestamp": t.ParentTimestamp,
		})
	}
	for _, n := range a.Missing {
		add(chain.ClassMissing, chain.SeverityWarn, map[string]any{"number": n})
	}
	for _, d := range a.Duplicates {
		add(chain.ClassDuplicate, chain.SeverityWarn, map[string]any{
			"number": d.Number, "records": d.Records, "hashes": d.Hashes,
		})
	}
	for _, o := range a.Orphans {
		add(chain.ClassOrphanSegment, chain.SeverityWarn, map[string]any{
			"from": o.From, "to": o.To, "records": o.Records,
		})
	}
	return out
}

func byClass(a *chain.Analysis) map[string]int {
	m := map[string]int{
		chain.ClassMissing:           len(a.Missing),
		chain.ClassDuplicate:         len(a.Duplicates),
		chain.ClassCanonicalConflict: len(a.Conflicts),
		chain.ClassOrphanSegment:     len(a.Orphans),
		chain.ClassParentMismatch:    len(a.Mismatches),
		chain.ClassTimestampOrder:    len(a.Orders),
	}
	return m
}

