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
