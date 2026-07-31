package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const (
	cleanSample   = "../../samples/clean.jsonl"
	reorgSample   = "../../samples/reorged.jsonl"
	invalidSample = "../../samples/invalid-lines.jsonl"
)

func runArgs(args ...string) (int, string, string) {
	var out, errOut bytes.Buffer
	code := run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestNoArgumentsExitsTwo(t *testing.T) {
	code, _, errOut := runArgs()
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "usage:") {
		t.Errorf("stderr = %q, want the usage text", errOut)
	}
}

func TestUnknownCommandExitsTwo(t *testing.T) {
	code, _, errOut := runArgs("frobnicate")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, `unknown command "frobnicate"`) {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestVersion(t *testing.T) {
	code, out, _ := runArgs("version")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if out != "reorglens 0.1.0\n" {
		t.Errorf("out = %q", out)
	}
}

func TestVersionRejectsArguments(t *testing.T) {
	if code, _, _ := runArgs("version", "extra"); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
}

func TestHelpExitsZero(t *testing.T) {
	code, out, _ := runArgs("help")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(out, "commands:") {
		t.Errorf("out = %q", out)
	}
}

func TestExplain(t *testing.T) {
	code, out, _ := runArgs("explain")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(out, "parent-mismatch") {
		t.Errorf("out = %q", out)
	}
}

func TestExplainRejectsArguments(t *testing.T) {
	if code, _, _ := runArgs("explain", "extra"); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
}

func TestAnalyzeCleanExitsZero(t *testing.T) {
	code, out, errOut := runArgs("analyze", cleanSample)
	if code != 0 {
		t.Errorf("exit = %d, want 0, stderr %q", code, errOut)
	}
	if !strings.Contains(out, "result      clean") {
		t.Errorf("out = %q", out)
	}
}

func TestAnalyzeReorgedExitsOne(t *testing.T) {
	code, out, errOut := runArgs("analyze", reorgSample)
	if code != 1 {
		t.Errorf("exit = %d, want 1, stderr %q", code, errOut)
	}
	if !strings.Contains(out, "findings    5  (2 error, 3 warn)") {
		t.Errorf("out = %q", out)
	}
}

func TestFailOnErrorStillExitsOneForAReorgedCapture(t *testing.T) {
	if code, _, _ := runArgs("analyze", "--fail-on", "error", reorgSample); code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
}

func TestFailOnErrorIgnoresWarnings(t *testing.T) {
	// invalid-lines.jsonl has five warn findings and no error findings.
	code, _, errOut := runArgs("analyze", "--fail-on", "error", invalidSample)
	if code != 0 {
		t.Errorf("exit = %d, want 0, stderr %q", code, errOut)
	}
}

func TestFailOnNeverAlwaysExitsZero(t *testing.T) {
	if code, _, _ := runArgs("analyze", "--fail-on", "never", reorgSample); code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
}

func TestAnalyzeRejectsUnknownFormat(t *testing.T) {
	code, _, errOut := runArgs("analyze", "--format", "yaml", cleanSample)
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "--format must be text or json") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestAnalyzeRejectsUnknownFailOn(t *testing.T) {
	code, _, errOut := runArgs("analyze", "--fail-on", "maybe", cleanSample)
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "--fail-on must be any, error or never") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestAnalyzeRejectsNonNumericRange(t *testing.T) {
	code, _, errOut := runArgs("analyze", "--from", "twelve", cleanSample)
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "--from must be an unsigned integer") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestAnalyzeMissingFileExitsTwo(t *testing.T) {
	code, _, errOut := runArgs("analyze", "samples/does-not-exist.jsonl")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "reorglens:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestAnalyzeWithoutAFileExitsTwo(t *testing.T) {
	if code, _, _ := runArgs("analyze"); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if code, _, _ := runArgs("analyze", "a.jsonl", "b.jsonl"); code != 2 {
		t.Errorf("exit = %d, want 2 for two files", code)
	}
}

func TestAnalyzeEmptyWindowExitsTwo(t *testing.T) {
	code, _, errOut := runArgs("analyze", "--from", "999999", "--to", "1", cleanSample)
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "empty window") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestAnalyzeJsonOutputIsEncodable(t *testing.T) {
	code, out, _ := runArgs("analyze", "--format", "json", reorgSample)
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !json.Valid([]byte(out)) {
		t.Fatalf("output is not valid JSON:\n%s", out)
	}
	var rep map[string]any
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if rep["tool"] != "reorglens" {
		t.Errorf("tool = %v", rep["tool"])
	}
	if rep["exitCode"] != float64(1) {
		t.Errorf("exitCode = %v, want 1", rep["exitCode"])
	}
	if rep["result"] != "findings present" {
		t.Errorf("result = %v", rep["result"])
