package console

import (
	"fmt"
	"io"

	"github.com/mattn/go-colorable"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Italic  = "\033[3m"
	Underln = "\033[4m"

	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
)

// Stdout is a colorable stdout writer for Windows compatibility.
var Stdout io.Writer = colorable.NewColorableStdout()

// Stderr is a colorable stderr writer for Windows compatibility.
var Stderr io.Writer = colorable.NewColorableStderr()

// Colorf prints formatted text with the given color code to colorable stdout.
func Colorf(color, format string, a ...any) {
	fmt.Fprintf(Stdout, "%s%s%s", color, fmt.Sprintf(format, a...), Reset)
}

// Infof prints an info-styled message (cyan label + message).
func Infof(format string, a ...any) {
	fmt.Fprintf(Stdout, "%s%sINFO%s %s\n", Bold, Cyan, Reset, fmt.Sprintf(format, a...))
}

// Successf prints a success-styled message (green).
func Successf(format string, a ...any) {
	fmt.Fprintf(Stdout, "%s%s%s\n", Green, fmt.Sprintf(format, a...), Reset)
}

// Warnf prints a warning-styled message (yellow).
func Warnf(format string, a ...any) {
	fmt.Fprintf(Stdout, "%s%s%s\n", Yellow, fmt.Sprintf(format, a...), Reset)
}

// Errorf prints an error-styled message (red) to stderr.
func Errorf(format string, a ...any) {
	fmt.Fprintf(Stderr, "%s%s%s\n", Red, fmt.Sprintf(format, a...), Reset)
}

// Banner prints a prominent banner box.
func Banner(color, title string, lines []string) {
	divider := "========================================"
	fmt.Fprintf(Stdout, "\n%s%s%s%s\n", Bold, color, divider, Reset)
	fmt.Fprintf(Stdout, "%s%s  %s%s\n", Bold, color, title, Reset)

	for _, line := range lines {
		fmt.Fprintf(Stdout, "%s  %s%s\n", color, line, Reset)
	}

	fmt.Fprintf(Stdout, "%s%s%s%s\n\n", Bold, color, divider, Reset)
}

// ColorableStdout returns the colorable stdout writer (for use with zerolog ConsoleWriter).
func ColorableStdout() io.Writer {
	return Stdout
}
