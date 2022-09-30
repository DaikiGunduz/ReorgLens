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
