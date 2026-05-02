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

func Success(format string, a ...any) {
	fmt.Printf("%s%s✔%s %s\n", Green, Bold, Reset, fmt.Sprintf(format, a...))
}

func Info(format string, a ...any) {
	fmt.Printf("%s%sℹ%s %s\n", Blue, Bold, Reset, fmt.Sprintf(format, a...))
}

func Warn(format string, a ...any) {
	fmt.Printf("%s%s⚠%s %s\n", Yellow, Bold, Reset, fmt.Sprintf(format, a...))
}

func Error(err error, help string) {
	fmt.Fprintf(os.Stderr, "%s%s✖ Error:%s %v\n", Red, Bold, Reset, err)
	if help != "" {
		fmt.Fprintf(os.Stderr, "%s%s💡 Suggestion:%s %s\n", Cyan, Bold, Reset, help)
	}
}

func Fatal(err error, help string) {
	Error(err, help)
	os.Exit(1)
}

func Section(name string) {
	fmt.Printf("\n%s%s===> %s%s\n", Purple, Bold, name, Reset)
}
