package utils

import (
	"fmt"
)

// Logger will print a string to console when verbose flag is set.
// Verbose flag can be overwritten (true) to log to console.
func Logger(message string, verbose bool) {
	// Block console logging when not verbose mode
	if !verbose {
		return
	}

	// Write to console
	_, _ = fmt.Printf("%s", message)
}
