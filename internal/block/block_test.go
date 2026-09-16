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
	}
	if r.Line != 1 {
		t.Errorf("Line = %d, want 1", r.Line)
	}
}

func TestBlankLinesAreSkipped(t *testing.T) {
	recs, bad, err := read(t, "\n  \n"+goodLine+"\n\n")
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(recs) != 1 || len(bad) != 0 {
		t.Fatalf("records = %d, bad = %d, want 1 and 0", len(recs), len(bad))
	}
	if recs[0].Line != 3 {
		t.Errorf("Line = %d, want 3 (blank lines still count)", recs[0].Line)
	}
}

func TestCanonicalDefaultsTrue(t *testing.T) {
	line := `{"number":1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 + `","timestamp":1}`
	recs, bad, err := read(t, line+"\n")
	if err != nil || len(bad) != 0 || len(recs) != 1 {
		t.Fatalf("err=%v bad=%v recs=%d", err, bad, len(recs))
	}
	if !recs[0].Canonical {
		t.Error("Canonical = false, want the default true")
	}
}

func TestCanonicalFalseHonoured(t *testing.T) {
	line := `{"number":1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 +
		`","timestamp":1,"canonical":false}`
	recs, _, err := read(t, line+"\n")
	if err != nil || len(recs) != 1 {
		t.Fatalf("err=%v recs=%v", err, len(recs))
	}
	if recs[0].Canonical {
		t.Error("Canonical = true, want false")
	}
}

func TestBadLines(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{"missing number", `{"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 + `","timestamp":1}`, `missing field "number"`},
		{"missing hash", `{"number":1,"parentHash":"0x` + parent64 + `","timestamp":1}`, `missing field "hash"`},
		{"missing parent", `{"number":1,"hash":"0x` + fixed64 + `","timestamp":1}`, `missing field "parentHash"`},
		{"missing timestamp", `{"number":1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 + `"}`, `missing field "timestamp"`},
		{"short hash", `{"number":1,"hash":"0xabc","parentHash":"0x` + parent64 + `","timestamp":1}`, "want 64 hex digits after 0x, got 3"},
		{"hash without prefix", `{"number":1,"hash":"` + fixed64 + `","parentHash":"0x` + parent64 + `","timestamp":1}`, "want a 32 byte hash as 0x prefixed hex"},
		{"hash with bad digit", `{"number":1,"hash":"0x` + strings.Repeat("z", 64) + `","parentHash":"0x` + parent64 + `","timestamp":1}`, "want hex digits only, found \"z\" at offset 0"},
		{"timestamp as string", `{"number":1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 + `","timestamp":"soon"}`, `field "timestamp": want an unsigned integer`},
		{"timestamp null", `{"number":1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 + `","timestamp":null}`, `field "timestamp": value is null`},
		{"number negative", `{"number":-1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 + `","timestamp":1}`, `field "number": want an unsigned integer`},
		{"producer short", `{"number":1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 + `","timestamp":1,"producer":"0x1234"}`, "want 40 hex digits after 0x, got 4"},
		{"canonical not a bool", `{"number":1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 + `","timestamp":1,"canonical":"yes"}`, `field "canonical": want true or false`},
		{"not json", `{this is not json`, "invalid JSON: invalid character"},
		{"not an object", `[1,2,3]`, "invalid JSON: want an object"},
		{"trailing garbage", goodLine + ` x`, "invalid JSON"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, bad, err := read(t, c.line+"\n")
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}
			if len(bad) != 1 {
				t.Fatalf("bad lines = %d, want 1", len(bad))
			}
			if !strings.Contains(bad[0].Reason, c.want) {
				t.Errorf("Reason = %q, want it to contain %q", bad[0].Reason, c.want)
			}
			if bad[0].Line != 1 {
				t.Errorf("Line = %d, want 1", bad[0].Line)
			}
		})
	}
}

func TestBadLineKeepsLineNumbers(t *testing.T) {
	input := "nope\n" + goodLine + "\nalso nope\n"
	_, bad, err := read(t, input)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(bad) != 2 {
		t.Fatalf("bad lines = %d, want 2", len(bad))
	}
	if bad[0].Line != 1 || bad[1].Line != 3 {
		t.Errorf("lines = %d, %d, want 1 and 3", bad[0].Line, bad[1].Line)
	}
}

func TestUnknownFieldsIgnored(t *testing.T) {
	line := `{"number":1,"hash":"0x` + fixed64 + `","parentHash":"0x` + parent64 +
		`","timestamp":1,"extra":"ignored"}`
	recs, bad, err := read(t, line+"\n")
	if err != nil || len(bad) != 0 || len(recs) != 1 {
		t.Fatalf("err=%v bad=%v recs=%d", err, bad, len(recs))
	}
}

func TestOverlongLineIsAReadError(t *testing.T) {
	_, _, err := read(t, strings.Repeat("a", MaxLineBytes+1)+"\n")
	if err == nil {
		t.Fatal("want an error for a line past MaxLineBytes")
	}
}

func TestShorten(t *testing.T) {
	if got := shorten("short"); got != "short" {
		t.Errorf("shorten(short) = %q", got)
	}
	long := strings.Repeat("x", maxTextBytes+50)
	got := shorten(long)
	if len(got) != maxTextBytes+3 {
		t.Errorf("len = %d, want %d", len(got), maxTextBytes+3)
