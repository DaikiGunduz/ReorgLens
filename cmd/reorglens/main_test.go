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
