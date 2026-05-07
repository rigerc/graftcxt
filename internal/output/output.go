package output

import (
	"fmt"
	"io"
)

var silent bool

func SetSilent(s bool) { silent = s }

func IsSilent() bool { return silent }

func Fprintf(w io.Writer, format string, args ...any) {
	if !silent {
		fmt.Fprintf(w, format, args...)
	}
}

func Fprintln(w io.Writer, args ...any) {
	if !silent {
		fmt.Fprintln(w, args...)
	}
}
