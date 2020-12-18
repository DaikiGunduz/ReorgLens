// Package block defines the record format that reorglens reads and the
// validation rules that decide whether a line can be analysed.
//
// The format is newline delimited JSON, one block observation per line. Every
// line is either a valid Record or a BadLine carrying its line number. The
// reader never aborts on the first bad line, because a broken export is
// exactly the thing worth reporting.
package block

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// MaxLineBytes bounds one input line so a broken or hostile export cannot
// exhaust memory. A line longer than this is a read error, not a bad line.
