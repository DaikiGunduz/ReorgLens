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
