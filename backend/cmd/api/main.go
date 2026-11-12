// Command api is the calculator HTTP service entry point.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

func main() {
	os.Exit(run(os.Stderr))
}

// run returns an exit code instead of calling os.Exit so the failure path
// stays testable.
func run(stderr io.Writer) int {
	if _, err := config.Load(); err != nil {
		// A failed stderr write leaves nothing else to report to, so the error is dropped.
		_, _ = fmt.Fprintf(stderr, constants.MsgConfigLoadFailed, err)
		return constants.ExitCodeConfigInvalid
	}
	return constants.ExitCodeOK
}
