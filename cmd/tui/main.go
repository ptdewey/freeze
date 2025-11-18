package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ptdewey/freeze/internal/diff"
	"github.com/ptdewey/freeze/internal/files"
	"github.com/ptdewey/freeze/internal/pretty"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	counterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("230"))

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

type model struct {
	snapshots    []string
	current      int
	newSnap      *files.Snapshot
	accepted     *files.Snapshot
	diffLines    []pretty.DiffLine
	choice       string
	done         bool
	err          error
	acceptedAll  int
	rejectedAll  int
	skippedAll   int
	actionResult string
	viewport     viewport.Model
	ready        bool
	width        int
	height       int
}

func initialModel() (model, error) {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return model{}, err
	}

	if len(snapshots) == 0 {
		return model{done: true}, nil
	}

	m := model{
		snapshots: snapshots,
		current:   0,
	}

	if err := m.loadCurrentSnapshot(); err != nil {
		return model{}, err
	}

	return m, nil
}

func (m *model) loadCurrentSnapshot() error {
	if m.current >= len(m.snapshots) {
		m.done = true
		return nil
	}

	testName := m.snapshots[m.current]

	newSnap, err := files.ReadSnapshot(testName, "new")
	if err != nil {
		return err
	}
	m.newSnap = newSnap

	accepted, err := files.ReadSnapshot(testName, "accepted")
	if err == nil {
		m.accepted = accepted
		diffLines := computeDiffLines(accepted, newSnap)
		m.diffLines = diffLines
	} else {
		m.accepted = nil
		m.diffLines = nil
	}

	return nil
}

func computeDiffLines(old, new *files.Snapshot) []pretty.DiffLine {
	diffLines := diff.Histogram(old.Content, new.Content)
	result := make([]pretty.DiffLine, len(diffLines))
	for i, dl := range diffLines {
		result[i] = pretty.DiffLine{
			OldNumber: dl.OldNumber,
			NewNumber: dl.NewNumber,
			Line:      dl.Line,
			Kind:      pretty.DiffKind(dl.Kind),
		}
	}
	return result
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 2
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
			m.updateViewportContent()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit

		case "a":
			// Accept current snapshot
			testName := m.snapshots[m.current]
			if err := files.AcceptSnapshot(testName); err != nil {
				m.err = err
			} else {
				m.actionResult = "Snapshot accepted"
				m.current++
				if err := m.loadCurrentSnapshot(); err != nil {
					m.err = err
				}
				m.updateViewportContent()
			}

		case "r":
			// Reject current snapshot
			testName := m.snapshots[m.current]
			if err := files.RejectSnapshot(testName); err != nil {
				m.err = err
			} else {
				m.actionResult = "Snapshot rejected"
				m.current++
				if err := m.loadCurrentSnapshot(); err != nil {
					m.err = err
				}
				m.updateViewportContent()
			}

		case "s":
			// Skip current snapshot
			m.actionResult = "Snapshot skipped"
			m.current++
			if err := m.loadCurrentSnapshot(); err != nil {
				m.err = err
			}
			m.updateViewportContent()

		case "A":
			// Accept all remaining
			for i := m.current; i < len(m.snapshots); i++ {
				if err := files.AcceptSnapshot(m.snapshots[i]); err != nil {
					m.err = err
					break
				}
				m.acceptedAll++
			}
			m.done = true
			return m, tea.Quit

		case "R":
			// Reject all remaining
			for i := m.current; i < len(m.snapshots); i++ {
				if err := files.RejectSnapshot(m.snapshots[i]); err != nil {
					m.err = err
					break
				}
				m.rejectedAll++
			}
			m.done = true
			return m, tea.Quit

		case "S":
			// Skip all remaining
			m.skippedAll = len(m.snapshots) - m.current
			m.done = true
			return m, tea.Quit
		}
	}

	// Handle viewport scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateViewportContent() {
	if !m.ready {
		return
	}

	var b strings.Builder

	// Show diff or new snapshot
	if m.accepted != nil && m.diffLines != nil {
		b.WriteString(pretty.DiffSnapshotBox(m.accepted, m.newSnap, m.diffLines))
	} else {
		if m.newSnap != nil {
			if m.newSnap.FuncName != "" {
				b.WriteString(pretty.NewSnapshotBoxFunc(m.newSnap))
			} else {
				b.WriteString(pretty.NewSnapshotBox(m.newSnap))
			}
		}
	}

	if m.actionResult != "" {
		b.WriteString("\n\n")
		b.WriteString(pretty.Success("✓ " + m.actionResult))
	}

	m.viewport.SetContent(contentStyle.Render(b.String()))
	m.viewport.GotoTop()
}

func (m model) View() string {
	if m.done {
		if len(m.snapshots) == 0 {
			return pretty.Success("✓ No new snapshots to review\n")
		}

		if m.acceptedAll > 0 {
			return pretty.Success(fmt.Sprintf("✓ Accepted %d snapshot(s)\n", m.acceptedAll))
		}
		if m.rejectedAll > 0 {
			return pretty.Warning(fmt.Sprintf("⊘ Rejected %d snapshot(s)\n", m.rejectedAll))
		}
		if m.skippedAll > 0 {
			return pretty.Warning(fmt.Sprintf("⊘ Skipped %d snapshot(s)\n", m.skippedAll))
		}
		return pretty.Success("\n✓ Review complete\n")
	}

	if m.err != nil {
		return pretty.Error("Error: " + m.err.Error() + "\n")
	}

	if !m.ready {
		return "\n  Initializing..."
	}

	// Header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		titleStyle.Render("Review Snapshots"),
		counterStyle.Render(fmt.Sprintf("[%d/%d] %s", m.current+1, len(m.snapshots), m.snapshots[m.current])),
	)

	// Footer with help
	scrollInfo := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
	helpText := "↑/↓/scroll: navigate • [a]ccept • [r]eject • [s]kip • [A]ll Accept • [R]ll Reject • [q]uit"

	footerLeft := helpStyle.Render(helpText)
	footerRight := helpStyle.Render(scrollInfo)

	gap := max(m.width-lipgloss.Width(footerLeft)-lipgloss.Width(footerRight), 0)

	footer := lipgloss.JoinHorizontal(
		lipgloss.Left,
		footerLeft,
		strings.Repeat(" ", gap),
		footerRight,
	)

	// Main content with viewport
	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusBarStyle.Width(m.width).Render(header),
		m.viewport.View(),
		statusBarStyle.Width(m.width).Render(footer),
	)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "accept-all":
			if err := acceptAll(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "reject-all":
			if err := rejectAll(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "help", "-h", "--help":
			fmt.Println(`Usage: freeze-tui [COMMAND]

Commands:
  review      Review and accept/reject new snapshots (default)
  accept-all  Accept all new snapshots
  reject-all  Reject all new snapshots
  help        Show this help message

Interactive Controls:
  a           Accept current snapshot
  r           Reject current snapshot
  s           Skip current snapshot
  A           Accept all remaining snapshots
  R           Reject all remaining snapshots
  S           Skip all remaining snapshots
  q           Quit`)
			return
		}
	}

	m, err := initialModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if m.done && len(m.snapshots) == 0 {
		fmt.Println(m.View())
		return
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func acceptAll() error {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return err
	}

	for _, testName := range snapshots {
		if err := files.AcceptSnapshot(testName); err != nil {
			return err
		}
	}

	fmt.Printf(pretty.Success("✓ Accepted %d snapshot(s)\n"), len(snapshots))
	return nil
}

func rejectAll() error {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return err
	}

	for _, testName := range snapshots {
		if err := files.RejectSnapshot(testName); err != nil {
			return err
		}
	}

	fmt.Printf(pretty.Warning("⊘ Rejected %d snapshot(s)\n"), len(snapshots))
	return nil
}
