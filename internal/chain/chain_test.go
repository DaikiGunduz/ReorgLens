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
