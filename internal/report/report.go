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
