package overlay

import (
	"fmt"
	"os"
)

var DebugEnabled = false

func debug(args ...string) {
	if DebugEnabled {
		for _, arg := range args {
			fmt.Fprintln(os.Stderr, arg)
		}
	}
}
