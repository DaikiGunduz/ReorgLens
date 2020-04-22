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
func fixture() *chain.Analysis {
	return &chain.Analysis{
		From: 100, To: 104, Numbers: 5,
		Valid: 6, Invalid: 1, Outside: 0,
		CanonicalRecords: 5, OrphanRecords: 1,
		Parent:    chain.ParentStats{Checked: 4, Verified: 3, Mismatch: 1, Unverified: 0},
