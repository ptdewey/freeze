package pretty

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ptdewey/freeze/internal/diff"
	"github.com/ptdewey/freeze/internal/files"
)

func NewSnapshotBox(snap *files.Snapshot) string {
	return newSnapshotBoxInternal(snap)
}

func DiffSnapshotBox(old, newSnapshot *files.Snapshot, diffLines []diff.DiffLine) string {
	width := TerminalWidth()
	snapshotFileName := files.SnapshotFileName(newSnapshot.Test) + ".snap"

	var sb strings.Builder
	sb.WriteString("─── " + "Review Snapshot " + strings.Repeat("─", width-20) + "\n\n")

	// TODO: maybe make helper functions for this, swap coloring between the key and the value
	// TODO: maybe show the snapshot file name in gray next to the "a/r/s" options
	// (i.e. "a accept -> snap_file_name.snap", "reject" w/strikethrough?, skip, keeps "*snap.new")
	if newSnapshot.Title != "" {
		sb.WriteString(Blue("  title: ") + newSnapshot.Title + "\n")
	}
	sb.WriteString(Blue("  test: ") + newSnapshot.Test + "\n")
	sb.WriteString(Blue("  file: ") + snapshotFileName + "\n")
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")

	// Calculate max line numbers for proper spacing
	maxOldNum, maxNewNum := 0, 0
	for _, dl := range diffLines {
		if dl.OldNumber > maxOldNum {
			maxOldNum = dl.OldNumber
		}
		if dl.NewNumber > maxNewNum {
			maxNewNum = dl.NewNumber
		}
	}
	oldWidth := len(fmt.Sprintf("%d", maxOldNum))
	newWidth := len(fmt.Sprintf("%d", maxNewNum))

	for _, dl := range diffLines {
		var oldNumStr, newNumStr string
		var prefix string
		var formatted string

		switch dl.Kind {
		case diff.DiffOld:
			oldNumStr = fmt.Sprintf("%*d", oldWidth, dl.OldNumber)
			newNumStr = strings.Repeat(" ", newWidth)
			prefix = Red("−")
			formatted = Red(dl.Line)
		case diff.DiffNew:
			oldNumStr = strings.Repeat(" ", oldWidth)
			newNumStr = fmt.Sprintf("%*d", newWidth, dl.NewNumber)
			prefix = Green("+")
			formatted = Green(dl.Line)
		case diff.DiffShared:
			oldNumStr = fmt.Sprintf("%*d", oldWidth, dl.OldNumber)
			newNumStr = fmt.Sprintf("%*d", newWidth, dl.NewNumber)
			prefix = " "
			formatted = dl.Line
		}

		linePrefix := fmt.Sprintf("%s %s %s", Gray(oldNumStr), Gray(newNumStr), prefix)
		display := fmt.Sprintf("%s %s", linePrefix, formatted)

		// Adjust for actual display length considering ANSI codes
		if len(dl.Line) > width-oldWidth-newWidth-8 {
			formatted = formatted[:width-oldWidth-newWidth-11] + "..."
			display = fmt.Sprintf("%s %s", linePrefix, formatted)
		}

		sb.WriteString(fmt.Sprintf("  %s\n", display))
	}

	sb.WriteString(strings.Repeat("─", width) + "\n")
	return sb.String()
}

func newSnapshotBoxInternal(snap *files.Snapshot) string {
	width := TerminalWidth()

	var sb strings.Builder
	sb.WriteString("─── " + "New Snapshot " + strings.Repeat("─", width-15) + "\n\n")

	if snap.Title != "" {
		sb.WriteString(Blue("  title: ") + snap.Title + "\n")
		// sb.WriteString(fmt.Sprintf("  title: %s\n", Blue(snap.Title)))
	}
	if snap.Test != "" {
		// sb.WriteString(fmt.Sprintf("  test: %s\n", Blue(snap.Test)))
		sb.WriteString(Blue("  test: ") + snap.Test + "\n")
	}
	if snap.FileName != "" {
		// sb.WriteString(fmt.Sprintf("  file: %s\n", Gray(snap.FileName)))
		sb.WriteString(Blue("  file: ") + snap.FileName + "\n")
	}
	sb.WriteString("\n")

	lines := strings.Split(snap.Content, "\n")
	numLines := len(lines)
	lineNumWidth := len(strconv.Itoa(numLines))

	topBar := strings.Repeat("─", lineNumWidth+3) + "┬" +
		strings.Repeat("─", width-lineNumWidth-2) + "\n"
	sb.WriteString(topBar)

	for i, line := range lines {
		lineNum := fmt.Sprintf("%*d", lineNumWidth, i+1)
		prefix := fmt.Sprintf("%s %s", Green(lineNum), Green("+"))

		if len(line) > width-len(prefix)-4 {
			line = line[:width-len(prefix)-7] + "..."
		}

		display := fmt.Sprintf("%s %s", prefix, Green(line))
		sb.WriteString(fmt.Sprintf("  %s\n", display))
	}

	bottomBar := strings.Repeat("─", lineNumWidth+3) + "┴" +
		strings.Repeat("─", width-lineNumWidth-2) + "\n"
	sb.WriteString(bottomBar)

	return sb.String()
}
