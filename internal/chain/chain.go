// Package chain turns validated block records into a window analysis.
//
// The rules are structural only. Nothing here knows about a node, a network,
// or a clock: it reads parent links, canonical flags, and timestamps out of
// the capture and reports where the picture does not hold together.
package chain

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"reorglens/internal/block"
)

// MaxWindow bounds how many block numbers a single run will enumerate. The
// range flags can widen the window beyond the capture, and this is the guard
// that stops a typo turning into a ten minute loop.
const MaxWindow = 10_000_000

// Finding classes. These strings are part of the output contract.
const (
	ClassMissing           = "missing"
	ClassDuplicate         = "duplicate"
	ClassCanonicalConflict = "canonical-conflict"
	ClassOrphanSegment     = "orphan-segment"
	ClassParentMismatch    = "parent-mismatch"
	ClassTimestampOrder    = "timestamp-order"
)

// Severities. error breaks the shape of the canonical chain, warn breaks the
// completeness or ambiguity of the capture.
const (
	SeverityError = "error"
	SeverityWarn  = "warn"
)

// Options narrows or widens the window beyond the range of the capture.
type Options struct {
	From *uint64
	To   *uint64
}

// Duplicate is a number that carries more than one distinct hash.
type Duplicate struct {
	Number  uint64   `json:"number"`
	Records int      `json:"records"`
	Hashes  []string `json:"hashes"`
}

// Conflict is a number that carries more than one record flagged canonical.
type Conflict struct {
	Number  uint64   `json:"number"`
	Records int      `json:"records"`
	Hashes  []string `json:"hashes"`
}

// OrphanSegment is a maximal run of numbers that has records but no canonical
// record, meaning the capture observed a branch the chain did not keep.
type OrphanSegment struct {
	From    uint64 `json:"from"`
	To      uint64 `json:"to"`
	Records int    `json:"records"`
}

// ParentMismatch is a point where two adjacent canonical records do not link.
type ParentMismatch struct {
	Number       uint64 `json:"number"`
	ParentNumber uint64 `json:"parentNumber"`
	Expected     string `json:"expected"`
	Observed     string `json:"observed"`
}

// TimestampOrder is a point where a canonical record's timestamp is not
// strictly greater than its parent's.
type TimestampOrder struct {
	Number          uint64 `json:"number"`
	ParentNumber    uint64 `json:"parentNumber"`
	Timestamp       uint64 `json:"timestamp"`
	ParentTimestamp uint64 `json:"parentTimestamp"`
}

// ProducerStats tallies the records attributed to one producer address.
type ProducerStats struct {
	Producer  string `json:"producer"`
	Canonical int    `json:"canonical"`
	Orphaned  int    `json:"orphaned"`
	Total     int    `json:"total"`
}

// ParentStats counts the adjacent canonical pairs that could be checked.
type ParentStats struct {
	Checked    int `json:"checked"`
	Verified   int `json:"verified"`
	Mismatch   int `json:"mismatch"`
	Unverified int `json:"unverified"`
}

// TimestampStats counts the adjacent canonical pairs and the ordering faults.
type TimestampStats struct {
	Checked    int `json:"checked"`
	OutOfOrder int `json:"outOfOrder"`
}

// Analysis is the full result of one run. Every slice is sorted ascending so
// that identical input produces byte identical output.
type Analysis struct {
	From    uint64 `json:"from"`
	To      uint64 `json:"to"`
	Numbers uint64 `json:"numbers"`

	Valid   int `json:"valid"`
	Invalid int `json:"invalid"`
	Outside int `json:"outside"`

	CanonicalRecords int `json:"canonicalRecords"`
	OrphanRecords    int `json:"orphanRecords"`

	Parent    ParentStats    `json:"parent"`
	Timestamp TimestampStats `json:"timestamp"`

	Missing    []uint64           `json:"missing"`
	Duplicates []Duplicate        `json:"duplicates"`
	Conflicts  []Conflict         `json:"conflicts"`
	Orphans    []OrphanSegment    `json:"orphanSegments"`
	Mismatches []ParentMismatch   `json:"parentMismatches"`
	Orders     []TimestampOrder   `json:"timestampOrders"`
	Producers  []ProducerStats    `json:"producers"`
	HasProducer bool              `json:"hasProducer"`

