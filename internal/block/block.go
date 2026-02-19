// Package block defines the record format that reorglens reads and the
// validation rules that decide whether a line can be analysed.
//
// The format is newline delimited JSON, one block observation per line. Every
// line is either a valid Record or a BadLine carrying its line number. The
// reader never aborts on the first bad line, because a broken export is
// exactly the thing worth reporting.
package block

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// MaxLineBytes bounds one input line so a broken or hostile export cannot
// exhaust memory. A line longer than this is a read error, not a bad line.
const MaxLineBytes = 1 << 20

// maxTextBytes bounds the line text echoed back in a BadLine report.
const maxTextBytes = 120

// Record is one block observation that passed validation.
type Record struct {
	Number     uint64
	Hash       string
	ParentHash string
	Timestamp  uint64
	Producer   string
	Canonical  bool
	Line       int
}

// BadLine records a line that could not be validated, with enough context to
// show the reader exactly where the export is broken.
type BadLine struct {
	Line   int    `json:"line"`
	Reason string `json:"reason"`
	Text   string `json:"text"`
}

// ReadAll reads newline delimited JSON records from r. Blank lines are
// skipped. Every other line becomes either a *Record or a BadLine.
func ReadAll(r io.Reader) ([]*Record, []BadLine, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), MaxLineBytes)

	var (
		records []*Record
		bad     []BadLine
		n       int
	)
	for sc.Scan() {
		n++
		text := sc.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		rec, err := parse(text, n)
		if err != nil {
			bad = append(bad, BadLine{Line: n, Reason: err.Error(), Text: shorten(text)})
			continue
		}
		records = append(records, rec)
	}
	if err := sc.Err(); err != nil {
		return nil, nil, fmt.Errorf("reading input: %w", err)
	}
	return records, bad, nil
}

// parse validates one line. Fields are read out of a raw object rather than
// into a struct so that a wrong type produces a message about the field the
// reader wrote, not about an internal Go type name.
func parse(text string, line int) (*Record, error) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "{") {
		return nil, fmt.Errorf("invalid JSON: want an object, got %s", shorten(trimmed))
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	number, err := requireUint(raw, "number")
	if err != nil {
		return nil, err
	}
	timestamp, err := requireUint(raw, "timestamp")
	if err != nil {
		return nil, err
	}
	hash, err := requireHex(raw, "hash", 64, "a 32 byte hash")
	if err != nil {
		return nil, err
	}
	parent, err := requireHex(raw, "parentHash", 64, "a 32 byte hash")
	if err != nil {
		return nil, err
	}

	rec := &Record{
		Number:     number,
		Hash:       hash,
		ParentHash: parent,
		Timestamp:  timestamp,
		Canonical:  true,
		Line:       line,
	}

	if v, ok := raw["canonical"]; ok {
		var c bool
		if err := json.Unmarshal(v, &c); err != nil {
			return nil, fmt.Errorf("field \"canonical\": want true or false, got %s", shorten(string(v)))
		}
		rec.Canonical = c
	}
	if v, ok := raw["producer"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return nil, fmt.Errorf("field \"producer\": want a string, got %s", shorten(string(v)))
		}
		addr, err := hexBody(s, 40, "a 20 byte address")
		if err != nil {
			return nil, fmt.Errorf("field \"producer\": %v", err)
		}
		rec.Producer = "0x" + strings.ToLower(addr)
	}
	return rec, nil
}

func requireUint(raw map[string]json.RawMessage, key string) (uint64, error) {
	v, ok := raw[key]
	if !ok {
		return 0, fmt.Errorf("missing field %q", key)
	}
	if isNull(v) {
		return 0, fmt.Errorf("field %q: value is null", key)
	}
	var n uint64
	if err := json.Unmarshal(v, &n); err != nil {
		return 0, fmt.Errorf("field %q: want an unsigned integer, got %s", key, shorten(string(v)))
	}
	return n, nil
}

func requireHex(raw map[string]json.RawMessage, key string, digits int, what string) (string, error) {
	v, ok := raw[key]
	if !ok {
		return "", fmt.Errorf("missing field %q", key)
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return "", fmt.Errorf("field %q: want a string, got %s", key, shorten(string(v)))
	}
	body, err := hexBody(s, digits, what)
	if err != nil {
		return "", fmt.Errorf("field %q: %v", key, err)
	}
	return "0x" + strings.ToLower(body), nil
}

// hexBody checks the 0x prefix, the digit count, and the alphabet, and
// returns the digits without the prefix.
func hexBody(v string, want int, what string) (string, error) {
	if !strings.HasPrefix(v, "0x") && !strings.HasPrefix(v, "0X") {
		return "", fmt.Errorf("want %s as 0x prefixed hex, got %q", what, v)
	}
	body := v[2:]
	if len(body) != want {
		return "", fmt.Errorf("want %d hex digits after 0x, got %d", want, len(body))
	}
	for i := 0; i < len(body); i++ {
		c := body[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
