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
