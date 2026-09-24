package utils

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// withAnswers drives Confirm from a canned script instead of the terminal, and returns the
// buffer the questions are written to.
func withAnswers(t *testing.T, input string) *bytes.Buffer {
	t.Helper()
	out := &bytes.Buffer{}
	originalIn, originalOut, originalInteractive := confirmIn, confirmOut, confirmIsInteractive
	confirmIn = bufio.NewReader(strings.NewReader(input))
	confirmOut = out
	confirmIsInteractive = func() bool { return true }
	t.Cleanup(func() {
		confirmIn, confirmOut, confirmIsInteractive = originalIn, originalOut, originalInteractive
	})
	return out
}

func TestConfirm(t *testing.T) {
	t.Run("yes answers", func(t *testing.T) {
		for _, answer := range []string{"y\n", "Y\n", "yes\n", "YES\n", " y \n", "\n"} {
			withAnswers(t, answer)
			assert.True(t, Confirm("Install?"), "answer %q", answer)
		}
	})

	t.Run("no answers", func(t *testing.T) {
		for _, answer := range []string{"n\n", "N\n", "no\n", "No\n", " n \n"} {
			withAnswers(t, answer)
			assert.False(t, Confirm("Install?"), "answer %q", answer)
		}
	})

	t.Run("question is printed", func(t *testing.T) {
		out := withAnswers(t, "n\n")
		Confirm("Install brew: neovim?")
		assert.Equal(t, "Install brew: neovim? [Y/n] ", out.String())
	})

	t.Run("an unusable answer is asked again", func(t *testing.T) {
		out := withAnswers(t, "maybe\nn\n")
		assert.False(t, Confirm("Install?"))
		assert.Equal(t, 2, strings.Count(out.String(), "[Y/n]"))
	})

	t.Run("exhausted input takes the default", func(t *testing.T) {
		withAnswers(t, "maybe")
		assert.True(t, Confirm("Install?"))
	})

	t.Run("no terminal means no question", func(t *testing.T) {
		out := withAnswers(t, "n\n")
		confirmIsInteractive = func() bool { return false }
		assert.True(t, Confirm("Install?"))
		assert.Empty(t, out.String())
	})
}
