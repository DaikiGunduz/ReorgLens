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
