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
