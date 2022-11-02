package output

import (
	"io"

	"github.com/muncus/gh-slocheck/search"
)

type Formatter interface {
	Print(io.Writer, search.PRInfo) error
	ToString(search.PRInfo) string
}
