package chain

import (
	"fmt"
	"strings"
	"testing"

	"reorglens/internal/block"
)

// h returns a distinct, well formed 32 byte hash for a small integer.
func h(v byte) string {
	return "0x" + strings.Repeat(fmt.Sprintf("%02x", v), 32)
}

// addr returns a distinct, well formed 20 byte address for a small integer.
func addr(v byte) string {
	return "0x" + strings.Repeat(fmt.Sprintf("%02x", v), 20)
}

func blk(number uint64, hash, parent string, ts uint64, canonical bool) string {
	c := "true"
	if !canonical {
		c = "false"
	}
	return fmt.Sprintf(`{"number":%d,"hash":"%s","parentHash":"%s","timestamp":%d,"canonical":%s}`,
		number, hash, parent, ts, c)
}

func analyse(t *testing.T, lines []string, opts Options) (*Analysis, error) {
	t.Helper()
	input := ""
	if len(lines) > 0 {
		input = strings.Join(lines, "\n") + "\n"
	}
	recs, bad, err := block.ReadAll(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	return Analyse(recs, bad, opts)
}

// cleanChain is four contiguous, correctly linked canonical blocks.
func cleanChain() []string {
	return []string{
		blk(100, h(100), h(99), 1000, true),
		blk(101, h(101), h(100), 1012, true),
		blk(102, h(102), h(101), 1024, true),
		blk(103, h(103), h(102), 1036, true),
	}
}

func TestCleanChainHasNoFindings(t *testing.T) {
	a, err := analyse(t, cleanChain(), Options{})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if a.TotalFindings != 0 {
		t.Errorf("TotalFindings = %d, want 0", a.TotalFindings)
	}
	if a.From != 100 || a.To != 103 || a.Numbers != 4 {
		t.Errorf("window = %d..%d (%d), want 100..103 (4)", a.From, a.To, a.Numbers)
	}
	if a.Valid != 4 || a.Invalid != 0 || a.Outside != 0 {
		t.Errorf("records = %d/%d/%d, want 4/0/0", a.Valid, a.Invalid, a.Outside)
	}
	if a.Parent.Checked != 3 || a.Parent.Verified != 3 || a.Parent.Mismatch != 0 || a.Parent.Unverified != 0 {
		t.Errorf("parent = %+v, want 3 checked all verified", a.Parent)
	}
	if a.Timestamp.Checked != 3 || a.Timestamp.OutOfOrder != 0 {
		t.Errorf("timestamp = %+v, want 3 checked none out of order", a.Timestamp)
	}
	if a.CanonicalRecords != 4 || a.OrphanRecords != 0 {
		t.Errorf("canonical=%d orphan=%d, want 4 and 0", a.CanonicalRecords, a.OrphanRecords)
	}
	if a.HasProducer {
		t.Error("HasProducer = true, want false when the capture carries no producers")
	}
}

func TestMissingNumber(t *testing.T) {
	lines := []string{cleanChain()[0], cleanChain()[2], cleanChain()[3]}
	a, err := analyse(t, lines, Options{})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(a.Missing) != 1 || a.Missing[0] != 101 {
		t.Fatalf("Missing = %v, want [101]", a.Missing)
	}
	if a.Parent.Checked != 1 || a.Parent.Unverified != 1 {
		t.Errorf("parent = %+v, want 1 checked and 1 unverified across the gap", a.Parent)
	}
	if a.TotalFindings != 1 || a.WarnFindings != 1 || a.ErrorFindings != 0 {
		t.Errorf("findings = total %d warn %d error %d, want 1/1/0",
			a.TotalFindings, a.WarnFindings, a.ErrorFindings)
	}
}

func TestDuplicateNumber(t *testing.T) {
	lines := cleanChain()
	lines = append(lines, blk(101, h(201), h(100), 1012, false))
	a, err := analyse(t, lines, Options{})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(a.Duplicates) != 1 {
		t.Fatalf("Duplicates = %v, want one", a.Duplicates)
	}
	d := a.Duplicates[0]
	if d.Number != 101 || d.Records != 2 || len(d.Hashes) != 2 {
		t.Errorf("duplicate = %+v, want number 101 with 2 records and 2 hashes", d)
	}
	if d.Hashes[0] >= d.Hashes[1] {
		t.Errorf("hashes = %v, want ascending", d.Hashes)
	}
	if a.OrphanRecords != 1 {
		t.Errorf("OrphanRecords = %d, want 1", a.OrphanRecords)
	}
	if a.ErrorFindings != 0 || a.WarnFindings != 1 {
		t.Errorf("error=%d warn=%d, want 0 and 1", a.ErrorFindings, a.WarnFindings)
	}
}

func TestCanonicalConflict(t *testing.T) {
	lines := cleanChain()
	lines = append(lines, blk(101, h(201), h(100), 1012, true))
	a, err := analyse(t, lines, Options{})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(a.Conflicts) != 1 {
		t.Fatalf("Conflicts = %v, want one", a.Conflicts)
	}
	c := a.Conflicts[0]
	if c.Number != 101 || c.Records != 2 || len(c.Hashes) != 2 {
		t.Errorf("conflict = %+v, want number 101 with 2 canonical records", c)
	}
	if a.ErrorFindings != 1 || a.WarnFindings != 0 {
		t.Errorf("error=%d warn=%d, want 1 and 0", a.ErrorFindings, a.WarnFindings)
	}
}

func TestOrphanSegmentGroupsConsecutiveNumbers(t *testing.T) {
	lines := []string{
		blk(100, h(100), h(99), 1000, true),
		blk(101, h(101), h(100), 1012, false),
		blk(102, h(102), h(101), 1024, false),
	}
	a, err := analyse(t, lines, Options{})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(a.Orphans) != 1 {
		t.Fatalf("Orphans = %+v, want one segment", a.Orphans)
	}
	o := a.Orphans[0]
	if o.From != 101 || o.To != 102 || o.Records != 2 {
		t.Errorf("segment = %+v, want 101..102 with 2 records", o)
	}
	if a.CanonicalRecords != 1 || a.OrphanRecords != 2 {
		t.Errorf("canonical=%d orphan=%d, want 1 and 2", a.CanonicalRecords, a.OrphanRecords)
	}
	if a.Parent.Checked != 0 || a.Parent.Unverified != 0 {
		t.Errorf("parent = %+v, want no pairs at all", a.Parent)
	}
}

func TestParentMismatch(t *testing.T) {
	lines := []string{
		blk(100, h(100), h(99), 1000, true),
		blk(101, h(101), h(55), 1012, true),
		blk(102, h(102), h(101), 1024, true),
	}
	a, err := analyse(t, lines, Options{})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(a.Mismatches) != 1 {
		t.Fatalf("Mismatches = %+v, want one", a.Mismatches)
	}
	m := a.Mismatches[0]
	if m.Number != 101 || m.ParentNumber != 100 {
		t.Errorf("mismatch at %d from %d, want 101 from 100", m.Number, m.ParentNumber)
	}
	if m.Expected != h(100) || m.Observed != h(55) {
		t.Errorf("expected %s observed %s, want %s and %s", m.Expected, m.Observed, h(100), h(55))
	}
	if a.Parent.Checked != 2 || a.Parent.Verified != 1 || a.Parent.Mismatch != 1 {
		t.Errorf("parent = %+v, want 2 checked, 1 verified, 1 mismatch", a.Parent)
	}
	if a.ErrorFindings != 1 {
		t.Errorf("ErrorFindings = %d, want 1", a.ErrorFindings)
	}
}

func TestTimestampMustIncrease(t *testing.T) {
	lines := []string{
		blk(100, h(100), h(99), 1000, true),
		blk(101, h(101), h(100), 1000, true),
	}
	a, err := analyse(t, lines, Options{})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(a.Orders) != 1 {
		t.Fatalf("Orders = %+v, want one", a.Orders)
	}
	o := a.Orders[0]
	if o.Number != 101 || o.Timestamp != 1000 || o.ParentTimestamp != 1000 {
		t.Errorf("order = %+v, want the equal pair at 101", o)
	}
	if a.Timestamp.Checked != 1 || a.Timestamp.OutOfOrder != 1 {
		t.Errorf("timestamp = %+v, want 1 checked and 1 out of order", a.Timestamp)
	}
}

func TestWindowNarrowingCountsRecordsOutside(t *testing.T) {
	lines := []string{
		blk(100, h(100), h(99), 1000, true),
		blk(101, h(101), h(100), 1012, true),
		blk(102, h(102), h(101), 1024, true),
		blk(103, h(103), h(102), 1036, true),
		blk(104, h(104), h(103), 1048, true),
	}
	from, to := uint64(101), uint64(103)
	a, err := analyse(t, lines, Options{From: &from, To: &to})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if a.From != 101 || a.To != 103 || a.Numbers != 3 {
		t.Errorf("window = %d..%d (%d), want 101..103 (3)", a.From, a.To, a.Numbers)
	}
	if a.Valid != 3 || a.Outside != 2 {
		t.Errorf("valid=%d outside=%d, want 3 and 2", a.Valid, a.Outside)
	}
	if a.TotalFindings != 0 {
		t.Errorf("TotalFindings = %d, want 0", a.TotalFindings)
	}
}

func TestWindowWideningReportsTheEdgesAsMissing(t *testing.T) {
	from, to := uint64(98), uint64(105)
	a, err := analyse(t, cleanChain(), Options{From: &from, To: &to})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if a.Numbers != 8 {
		t.Errorf("Numbers = %d, want 8", a.Numbers)
	}
	want := []uint64{98, 99, 104, 105}
	if len(a.Missing) != len(want) {
		t.Fatalf("Missing = %v, want %v", a.Missing, want)
	}
	for i := range want {
		if a.Missing[i] != want[i] {
			t.Errorf("Missing[%d] = %d, want %d", i, a.Missing[i], want[i])
		}
	}
}

func TestEmptyWindowIsAnError(t *testing.T) {
	from, to := uint64(500), uint64(100)
	if _, err := analyse(t, cleanChain(), Options{From: &from, To: &to}); err == nil {
		t.Fatal("want an error when from is greater than to")
	}
}

func TestNoRecordsIsAnError(t *testing.T) {
	if _, err := analyse(t, nil, Options{}); err == nil {
		t.Fatal("want an error when there are no records")
	}
}

func TestWindowPastTheLimitIsAnError(t *testing.T) {
	from, to := uint64(0), uint64(MaxWindow)
	_, err := analyse(t, cleanChain(), Options{From: &from, To: &to})
	if err == nil {
		t.Fatal("want an error past MaxWindow")
	}
	if !strings.Contains(err.Error(), "the limit is") {
		t.Errorf("error = %q, want it to name the limit", err)
	}
}

func TestProducerStatsAreSorted(t *testing.T) {
	a, err := Analyse(recordsWithProducers(t), nil, Options{})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if !a.HasProducer {
		t.Fatal("HasProducer = false, want true")
	}
	if len(a.Producers) != 2 {
		t.Fatalf("Producers = %d, want 2", len(a.Producers))
	}
	for i := 1; i < len(a.Producers); i++ {
		if a.Producers[i-1].Producer >= a.Producers[i].Producer {
			t.Errorf("producers not sorted: %v", a.Producers)
		}
	}
	if a.Producers[0].Producer != addr(1) || a.Producers[0].Total != 1 {
		t.Errorf("first producer = %+v, want %s with 1 record", a.Producers[0], addr(1))
	}
	if a.Producers[1].Producer != addr(2) || a.Producers[1].Total != 2 {
		t.Errorf("second producer = %+v, want %s with 2 records", a.Producers[1], addr(2))
	}
	if a.Producers[1].Canonical != 2 || a.Producers[1].Orphaned != 0 {
		t.Errorf("second producer split = %+v, want 2 canonical and 0 orphaned", a.Producers[1])
	}
}

// recordsWithProducers builds three linked blocks whose producer addresses are
// deliberately not in sorted order in the input.
func recordsWithProducers(t *testing.T) []*block.Record {
	t.Helper()
	producers := []string{addr(2), addr(1), addr(2)}
	base := []string{
		blk(100, h(100), h(99), 1000, true),
		blk(101, h(101), h(100), 1012, true),
		blk(102, h(102), h(101), 1024, true),
	}
	out := make([]string, 0, len(base))
	for i, l := range base {
		out = append(out, strings.TrimSuffix(l, "}")+`,"producer":"`+producers[i]+`"}`)
	}
