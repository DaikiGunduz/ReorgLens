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
