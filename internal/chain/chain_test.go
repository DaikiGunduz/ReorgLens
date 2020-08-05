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
