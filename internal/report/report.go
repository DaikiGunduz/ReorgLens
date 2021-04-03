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
