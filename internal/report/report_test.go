package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"reorglens/internal/block"
	"reorglens/internal/chain"
)

// fixture carries one finding of every class so that a single assertion pass
// covers every rendering path.
