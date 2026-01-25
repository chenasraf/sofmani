package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBoxLine(t *testing.T) {
	t.Run("formats short text with padding", func(t *testing.T) {
		// formatBoxLine adds: "│ " + text + padding + " │"
		// where padding = innerWidth - len(text)
		result := formatBoxLine("Hello", 20)
		// padding = 20 - 5 = 15 spaces
		// result = "│ " + "Hello" + 15 spaces + " │"
		assert.Equal(t, "│ Hello                │", result)
	})

	t.Run("formats empty string", func(t *testing.T) {
		result := formatBoxLine("", 10)
		// padding = 10 - 0 = 10 spaces
		// result = "│ " + "" + 10 spaces + " │"
		assert.Equal(t, "│            │", result)
	})

	t.Run("truncates text longer than width", func(t *testing.T) {
		result := formatBoxLine("This is a very long text", 10)
		// text truncated to "This is a " (10 chars), padding = 0
		// result = "│ " + "This is a " + "" + " │"
		assert.Equal(t, "│ This is a  │", result)
	})

	t.Run("formats text exactly at width", func(t *testing.T) {
		result := formatBoxLine("1234567890", 10)
		// padding = 10 - 10 = 0 spaces
		assert.Equal(t, "│ 1234567890 │", result)
	})
}

func TestWrapText(t *testing.T) {
	t.Run("returns single line for short text", func(t *testing.T) {
		result := wrapText("Hello world", 20)
		assert.Equal(t, []string{"Hello world"}, result)
	})

	t.Run("wraps long text", func(t *testing.T) {
		result := wrapText("This is a longer sentence that needs wrapping", 20)
		assert.Len(t, result, 3)
		assert.Equal(t, "This is a longer", result[0])
		assert.Equal(t, "sentence that needs", result[1])
		assert.Equal(t, "wrapping", result[2])
	})

	t.Run("respects existing newlines", func(t *testing.T) {
		result := wrapText("Line one\nLine two\nLine three", 50)
		assert.Equal(t, []string{"Line one", "Line two", "Line three"}, result)
	})

	t.Run("preserves blank lines", func(t *testing.T) {
		result := wrapText("First paragraph\n\nSecond paragraph", 50)
		assert.Equal(t, []string{"First paragraph", "", "Second paragraph"}, result)
	})

	t.Run("wraps each paragraph separately", func(t *testing.T) {
		result := wrapText("Short\nThis is a longer line that will need to wrap", 20)
		assert.Len(t, result, 4)
		assert.Equal(t, "Short", result[0])
		assert.Equal(t, "This is a longer", result[1])
		assert.Equal(t, "line that will need", result[2])
		assert.Equal(t, "to wrap", result[3])
	})

	t.Run("handles empty input", func(t *testing.T) {
		result := wrapText("", 20)
		assert.Equal(t, []string{""}, result)
	})

	t.Run("handles single word longer than width", func(t *testing.T) {
		result := wrapText("Supercalifragilisticexpialidocious", 10)
		// Single word stays on one line even if longer than width
		assert.Equal(t, []string{"Supercalifragilisticexpialidocious"}, result)
	})

	t.Run("handles multiple spaces between words", func(t *testing.T) {
		result := wrapText("Word1   Word2   Word3", 50)
		// strings.Fields collapses multiple spaces
		assert.Equal(t, []string{"Word1 Word2 Word3"}, result)
	})
}

func TestBoxDrawingConstants(t *testing.T) {
	t.Run("default box width is 60", func(t *testing.T) {
		assert.Equal(t, 60, boxDefaultWidth)
	})

	t.Run("box characters are correct", func(t *testing.T) {
		assert.Equal(t, "┌", boxTopLeft)
		assert.Equal(t, "┐", boxTopRight)
		assert.Equal(t, "└", boxBottomLeft)
		assert.Equal(t, "┘", boxBottomRight)
		assert.Equal(t, "─", boxHorizontal)
		assert.Equal(t, "│", boxVertical)
		assert.Equal(t, "├", boxLeftT)
		assert.Equal(t, "┤", boxRightT)
	})
}

func TestGetBoxWidth(t *testing.T) {
	t.Run("returns positive width", func(t *testing.T) {
		width := getBoxWidth()
		assert.Greater(t, width, 0)
	})

	t.Run("returns at most default width", func(t *testing.T) {
		width := getBoxWidth()
		assert.LessOrEqual(t, width, boxDefaultWidth)
	})
}

func TestHighlight(t *testing.T) {
	t.Run("wraps text with highlight markers", func(t *testing.T) {
		result := Highlight("test")
		assert.Equal(t, highlightStart+"test"+highlightEnd, result)
	})

	t.Run("H is alias for Highlight", func(t *testing.T) {
		assert.Equal(t, Highlight("test"), H("test"))
	})
}

func TestStripHighlightMarkers(t *testing.T) {
	t.Run("removes highlight markers", func(t *testing.T) {
		input := highlightStart + "highlighted" + highlightEnd + " normal"
		result := stripHighlightMarkers(input)
		assert.Equal(t, "highlighted normal", result)
	})

	t.Run("handles text without markers", func(t *testing.T) {
		result := stripHighlightMarkers("plain text")
		assert.Equal(t, "plain text", result)
	})

	t.Run("handles multiple highlighted sections", func(t *testing.T) {
		input := highlightStart + "one" + highlightEnd + " and " + highlightStart + "two" + highlightEnd
		result := stripHighlightMarkers(input)
		assert.Equal(t, "one and two", result)
	})
}

func TestProcessHighlights(t *testing.T) {
	t.Run("converts markers to ANSI codes", func(t *testing.T) {
		input := highlightStart + "text" + highlightEnd
		result := processHighlights(input, ansiBlueBold)
		assert.Equal(t, ansiHighlight+"text"+ansiReset+ansiBlueBold, result)
	})

	t.Run("handles empty base color", func(t *testing.T) {
		input := highlightStart + "text" + highlightEnd
		result := processHighlights(input, "")
		assert.Equal(t, ansiHighlight+"text"+ansiReset, result)
	})
}
