package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"reorglens/internal/block"
	"reorglens/internal/chain"
)

// fixture carries one finding of every class so that a single assertion pass
// covers every rendering path.
func fixture() *chain.Analysis {
	return &chain.Analysis{
		From: 100, To: 104, Numbers: 5,
		Valid: 6, Invalid: 1, Outside: 0,
		CanonicalRecords: 5, OrphanRecords: 1,
		Parent:    chain.ParentStats{Checked: 4, Verified: 3, Mismatch: 1, Unverified: 0},
		Timestamp: chain.TimestampStats{Checked: 4, OutOfOrder: 1},
		Missing:   []uint64{102},
		Duplicates: []chain.Duplicate{
			{Number: 103, Records: 2, Hashes: []string{"0xaa", "0xbb"}},
		},
		Conflicts: []chain.Conflict{
			{Number: 101, Records: 2, Hashes: []string{"0xaa", "0xbb"}},
		},
		Orphans: []chain.OrphanSegment{{From: 104, To: 104, Records: 1}},
		Mismatches: []chain.ParentMismatch{
			{Number: 103, ParentNumber: 102, Expected: "0xaa", Observed: "0xbb"},
		},
		Orders: []chain.TimestampOrder{
			{Number: 104, ParentNumber: 103, Timestamp: 5, ParentTimestamp: 5},
		},
		Producers: []chain.ProducerStats{
			{Producer: "0x01", Canonical: 5, Orphaned: 1, Total: 6},
		},
		HasProducer:    true,
		ErrorFindings:  3,
		WarnFindings:   3,
		TotalFindings:  6,
		BadLines: []block.BadLine{
			{Line: 4, Reason: `missing field "number"`, Text: `{"hash":"0x.."}`},
		},
	}
}

func meta() Meta {
	return Meta{
		Version:  "1.2.3",
		Command:  "analyze",
		Input:    "fixtures.jsonl",
		Detail:   true,
		ExitCode: 1,
	}
}

func TestTextReportsEveryFindingClass(t *testing.T) {
	var buf bytes.Buffer
	if err := Text(&buf, meta(), fixture()); err != nil {
		t.Fatalf("Text: %v", err)
	}
	out := buf.String()
	want := []string{
		"reorglens 1.2.3 analyze fixtures.jsonl",
		"window      100..104  (5 numbers)",
		"records     6 valid  1 invalid  0 outside",
		"canonical   5 records",
		"orphaned    1 records in 1 segment",
		"parent      4 checked  1 mismatch  0 unverified",
		"timestamp   4 checked  1 out of order",
		"producers   1",
		"findings    6  (3 error, 3 warn)",
		"canonical-conflict",
		"parent-mismatch",
		"timestamp-order",
		"missing",
		"duplicate",
		"orphan-segment",
		"invalid lines (1):",
		`line 4: missing field "number"`,
		"result      findings present",
		"exit code   1",
	}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q\n---\n%s", w, out)
		}
	}
}

func TestTextErrorsAreListedBeforeWarnings(t *testing.T) {
	var buf bytes.Buffer
	if err := Text(&buf, meta(), fixture()); err != nil {
		t.Fatalf("Text: %v", err)
	}
	out := buf.String()
	lastError := strings.LastIndex(out, "error ")
	firstWarn := strings.Index(out, "warn   ")
	if lastError < 0 || firstWarn < 0 {
		t.Fatalf("could not find the severity markers\n%s", out)
	}
	if lastError > firstWarn {
		t.Errorf("a warning is listed before an error\n%s", out)
	}
}

func TestTextCleanRun(t *testing.T) {
	a := fixture()
	a.Missing = nil
	a.Duplicates = nil
	a.Conflicts = nil
	a.Orphans = nil
	a.Mismatches = nil
	a.Orders = nil
	a.BadLines = nil
	a.ErrorFindings, a.WarnFindings, a.TotalFindings = 0, 0, 0
	a.Invalid = 0

	m := meta()
	m.ExitCode = 0

	var buf bytes.Buffer
	if err := Text(&buf, m, a); err != nil {
		t.Fatalf("Text: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "findings: none") {
		t.Errorf("want the explicit 'findings: none' marker\n%s", out)
	}
	if !strings.Contains(out, "result      clean") {
		t.Errorf("want result clean\n%s", out)
	}
	if strings.Contains(out, "invalid lines") {
		t.Errorf("must not print an invalid lines section\n%s", out)
	}
}

func TestTextWithoutDetailStopsAfterTheSummary(t *testing.T) {
	m := meta()
	m.Detail = false
	var buf bytes.Buffer
	if err := Text(&buf, m, fixture()); err != nil {
		t.Fatalf("Text: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "result") {
		t.Errorf("detail off must omit the result line\n%s", out)
	}
	if strings.Contains(out, "canonical-conflict") {
		t.Errorf("detail off must omit the finding list\n%s", out)
	}
	if !strings.Contains(out, "findings    6  (3 error, 3 warn)") {
		t.Errorf("the summary block must still be there\n%s", out)
	}
}

func TestTextOmitsProducersWhenTheCaptureHasNone(t *testing.T) {
	a := fixture()
	a.HasProducer = false
	var buf bytes.Buffer
	if err := Text(&buf, meta(), a); err != nil {
		t.Fatalf("Text: %v", err)
	}
	if strings.Contains(buf.String(), "producers") {
		t.Errorf("must not print a producers line with no producer data\n%s", buf.String())
	}
}

func TestJSONFieldNamesAreStable(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, meta(), fixture()); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	want := []string{
		"tool", "version", "command", "input", "window", "records", "chain",
		"parent", "timestamp", "producers", "findings", "detail", "badLines",
		"result", "exitCode",
	}
	for _, k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("missing top level field %q", k)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d top level fields, want %d", len(got), len(want))
	}
}

func TestJSONDetailOrderIsFixed(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, meta(), fixture()); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var rep struct {
		Detail []Finding `json:"detail"`
	}
	if err := json.Unmarshal(buf.Bytes(), &rep); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := []string{
		chain.ClassCanonicalConflict,
		chain.ClassParentMismatch,
		chain.ClassTimestampOrder,
		chain.ClassMissing,
		chain.ClassDuplicate,
		chain.ClassOrphanSegment,
	}
	if len(rep.Detail) != len(want) {
		t.Fatalf("detail = %d entries, want %d", len(rep.Detail), len(want))
	}
	for i, w := range want {
		if rep.Detail[i].Class != w {
			t.Errorf("detail[%d] = %q, want %q", i, rep.Detail[i].Class, w)
		}
	}
}

func TestJSONEncodesDeterministically(t *testing.T) {
	var a, b bytes.Buffer
	if err := JSON(&a, meta(), fixture()); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if err := JSON(&b, meta(), fixture()); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if a.String() != b.String() {
		t.Error("two encodings of the same analysis differ")
	}
}

func TestJSONDetailIsEmptyWhenDetailIsOff(t *testing.T) {
	m := meta()
	m.Detail = false
	var buf bytes.Buffer
	if err := JSON(&buf, m, fixture()); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var rep struct {
		Detail []Finding `json:"detail"`
	}
	if err := json.Unmarshal(buf.Bytes(), &rep); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(rep.Detail) != 0 {
		t.Errorf("detail = %d entries, want 0", len(rep.Detail))
	}
}

func TestResultNamesTheDataState(t *testing.T) {
	if got := Result(fixture()); got != "findings present" {
		t.Errorf("Result = %q", got)
	}
	a := fixture()
	a.TotalFindings = 0
	if got := Result(a); got != "clean" {
		t.Errorf("Result = %q, want clean", got)
	}
}

func TestExplainCoversEveryClass(t *testing.T) {
	var buf bytes.Buffer
	if err := Explain(&buf); err != nil {
		t.Fatalf("Explain: %v", err)
	}
	out := buf.String()
	for _, c := range []string{
		chain.ClassCanonicalConflict,
		chain.ClassParentMismatch,
		chain.ClassTimestampOrder,
		chain.ClassMissing,
		chain.ClassDuplicate,
		chain.ClassOrphanSegment,
	} {
		if !strings.Contains(out, c) {
			t.Errorf("explain does not mention %q", c)
		}
	}
	for _, c := range []string{"exit codes", "severity"} {
		if !strings.Contains(out, c) {
			t.Errorf("explain does not contain %q", c)
		}
	}
}
