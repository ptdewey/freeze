package freeze

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
)

func TerminalWidth() int {
	width := os.Getenv("COLUMNS")
	if w, err := strconv.Atoi(width); err == nil && w > 0 {
		return w
	}
	return 80
}

func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}

func ClearLine() {
	fmt.Print("\033[K")
}

func Red(s string) string {
	if !hasColor() {
		return s
	}
	return colorRed + s + colorReset
}

func Green(s string) string {
	if !hasColor() {
		return s
	}
	return colorGreen + s + colorReset
}

func Yellow(s string) string {
	if !hasColor() {
		return s
	}
	return colorYellow + s + colorReset
}

func Blue(s string) string {
	if !hasColor() {
		return s
	}
	return colorBlue + s + colorReset
}

func Gray(s string) string {
	if !hasColor() {
		return s
	}
	return colorGray + s + colorReset
}

func Bold(s string) string {
	if !hasColor() {
		return s
	}
	return colorBold + s + colorReset
}

func hasColor() bool {
	return os.Getenv("NO_COLOR") == ""
}

func NewSnapshotBox(snap *Snapshot) string {
	width := TerminalWidth()
	separator := strings.Repeat("─", width)

	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", width) + "╮\n")
	// FIX: this line is missing the '│' symbol at the end
	sb.WriteString(fmt.Sprintf("│ %s\n", Blue("New Snapshot")))
	sb.WriteString("├" + separator + "┤\n")

	lines := strings.Split(snap.Content, "\n")
	for _, line := range lines {
		if len(line) > width-4 {
			line = line[:width-7] + "..."
		}
		// TODO: added code lines in snapshots should be in green with "<line number> +" next to them
		// - line numbers should be left aligned with space padding
		// FIX: each of these lines is missing the '│' symbol at the end
		sb.WriteString(fmt.Sprintf("│ %s\n", line))
	}

	sb.WriteString("╰" + strings.Repeat("─", width) + "╯\n")
	return sb.String()
}

func DiffSnapshotBox(old, new *Snapshot) string {
	width := TerminalWidth()

	diffLines := Histogram(old.Content, new.Content)

	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", width-2) + "╮\n")
	sb.WriteString(fmt.Sprintf("│ %s\n", Blue("Snapshot Diff")))
	sb.WriteString("├" + strings.Repeat("─", width-2) + "┤\n")

	for _, dl := range diffLines {
		var prefix string
		var formatted string

		switch dl.Kind {
		case DiffOld:
			prefix = Red("−")
			formatted = Red(dl.Line)
		case DiffNew:
			prefix = Green("+")
			formatted = Green(dl.Line)
		case DiffShared:
			prefix = " "
			formatted = dl.Line
		}

		display := fmt.Sprintf("%s %s", prefix, formatted)
		if len(display) > width-4 {
			display = display[:width-7] + "..."
		}
		sb.WriteString(fmt.Sprintf("│ %s\n", display))
	}

	sb.WriteString("╰" + strings.Repeat("─", width-2) + "╯\n")
	return sb.String()
}

func FormatHeader(text string) string {
	return Bold(Blue(text))
}

func FormatSuccess(text string) string {
	return Green(text)
}

func FormatError(text string) string {
	return Red(text)
}

func FormatWarning(text string) string {
	return Yellow(text)
}
