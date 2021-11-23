// Command reorglens analyses a captured block window and reports where the
// canonical chain, the parent links, or the timestamps do not hold together.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"reorglens/internal/block"
	"reorglens/internal/chain"
	"reorglens/internal/report"
)

const version = "0.1.0"

const usageText = `usage: reorglens <command> [flags]

commands:
  analyze   read a JSONL block export and report findings
  explain   print the finding classes, severities and exit codes
