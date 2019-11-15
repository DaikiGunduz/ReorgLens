// Package chain turns validated block records into a window analysis.
//
// The rules are structural only. Nothing here knows about a node, a network,
// or a clock: it reads parent links, canonical flags, and timestamps out of
// the capture and reports where the picture does not hold together.
package chain

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"reorglens/internal/block"
)

// MaxWindow bounds how many block numbers a single run will enumerate. The
// range flags can widen the window beyond the capture, and this is the guard
// that stops a typo turning into a ten minute loop.
const MaxWindow = 10_000_000

