package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// confirmIn reads the answers. One reader is kept for the whole run: a fresh one per question
// would discard whatever the user typed ahead into its buffer.
var confirmIn = bufio.NewReader(os.Stdin)

// confirmOut receives the question. It bypasses the logger so the answer is typed on the same
// line as the question, without a log prefix in front of the cursor.
var confirmOut io.Writer = os.Stdout

// confirmIsInteractive reports whether someone is there to answer.
var confirmIsInteractive = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Confirm asks a yes/no question, taking an empty answer as yes. Piped or redirected input has
// nobody to answer, so the default stands and the question is never printed.
func Confirm(question string) bool {
	if !confirmIsInteractive() {
		return true
	}
	for {
		_, _ = fmt.Fprintf(confirmOut, "%s [Y/n] ", question)
		line, err := confirmIn.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "", "y", "yes":
			return true
		case "n", "no":
			return false
		}
		// The answer was gibberish and there is no more input to ask for; take the default.
		if err != nil {
			return true
		}
	}
}
