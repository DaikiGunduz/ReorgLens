package block

import (
	"strings"
	"testing"
)

const goodLine = `{"number":19000100,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 +
	`","timestamp":1712000000,"producer":"0xABCDEF0123456789ABCDEF0123456789ABCDEF01","canonical":true}`

const fixed64 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const parent64 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func read(t *testing.T, input string) ([]*Record, []BadLine, error) {
	t.Helper()
	return ReadAll(strings.NewReader(input))
}

func TestReadAllValid(t *testing.T) {
	recs, bad, err := read(t, goodLine+"\n")
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(bad) != 0 {
		t.Fatalf("bad lines = %d, want 0", len(bad))
	}
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1", len(recs))
	}
	r := recs[0]
	if r.Number != 19000100 {
		t.Errorf("Number = %d, want 19000100", r.Number)
	}
	if r.Hash != "0x"+fixed64 {
		t.Errorf("Hash = %q", r.Hash)
	}
	if r.ParentHash != "0x"+parent64 {
		t.Errorf("ParentHash = %q", r.ParentHash)
	}
	if r.Timestamp != 1712000000 {
		t.Errorf("Timestamp = %d, want 1712000000", r.Timestamp)
	}
	if r.Producer != "0xabcdef0123456789abcdef0123456789abcdef01" {
		t.Errorf("Producer = %q, want lowercase", r.Producer)
	}
	if !r.Canonical {
		t.Error("Canonical = false, want true")
