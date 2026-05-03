// Package ui provides helper functions for formatted terminal output.
package ui

import (
	"fmt"
	"os"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
	Bold   = "\033[1m"
)

// Success prints a success message with a green checkmark.
func Success(format string, a ...any) {
	fmt.Printf("%s%s✔%s %s\n", Green, Bold, Reset, fmt.Sprintf(format, a...))
}

// Info prints an informational message in blue.
func Info(format string, a ...any) {
	fmt.Printf("%s%sℹ%s %s\n", Blue, Bold, Reset, fmt.Sprintf(format, a...))
}

// Warn prints a warning message in yellow.
func Warn(format string, a ...any) {
	fmt.Printf("%s%s⚠%s %s\n", Yellow, Bold, Reset, fmt.Sprintf(format, a...))
}

// Error prints an error and a suggestion to stderr.
func Error(err error, help string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%s✖ Error:%s %v\n", Red, Bold, Reset, err)
	}
	if help != "" {
		fmt.Fprintf(os.Stderr, "%s%s💡 Suggestion:%s %s\n", Cyan, Bold, Reset, help)
	}
}

// Fatal prints error and exits the program.
func Fatal(err error, help string) {
	Error(err, help)
	os.Exit(1)
}

// Section prints a section header in purple.
func Section(name string) {
	fmt.Printf("\n%s%s===> %s%s\n", Purple, Bold, name, Reset)
}
