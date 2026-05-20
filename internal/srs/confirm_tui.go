package srs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("9")).
			Padding(1, 4).
			MarginTop(1)

	confirmYesStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("9")).
			Padding(0, 2).
			Bold(true).
			MarginRight(2)

	confirmNoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("240")).
			Padding(0, 2).
			Bold(true)

	confirmWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true)

	confirmBodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	confirmCursorYes = 0
	confirmCursorNo  = 1
)

// ---------------------------------------------------------------------------
// confirmModel — reusable modal, embedded in AppModel
// ---------------------------------------------------------------------------

type confirmModel struct {
	title    string   // e.g. "Delete All Entries?"
	lines    []string // extra description lines shown in the box
	cursor   int      // 0 = Yes, 1 = No
	accepted bool
	resolved bool
}

func newConfirmModel(title string, lines []string) confirmModel {
	return confirmModel{
		title:  title,
		lines:  lines,
		cursor: confirmCursorNo, // default to No for safety
	}
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (c confirmModel) View() string {
	warning := confirmWarningStyle.Render("⚠  " + c.title)

	var descLines string
	for _, l := range c.lines {
		descLines += confirmBodyStyle.Render(l) + "\n"
	}

	var yesBtn, noBtn string
	if c.cursor == confirmCursorYes {
		yesBtn = confirmYesStyle.Render("  Yes  ")
		noBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 2).
			Render("  No  ")
	} else {
		yesBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 2).
			Render("  Yes  ")
		noBtn = confirmNoStyle.Render("  No  ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, yesBtn, noBtn)

	box := confirmBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		warning,
		"",
		descLines,
		"",
		buttons,
	))

	help := appHelpStyle.Render("←/→ or y/n: choose • enter: confirm • esc: cancel")

	body := lipgloss.JoinVertical(lipgloss.Left,
		box,
		help,
	)

	return appDocStyle.Render(body)
}

// ---------------------------------------------------------------------------
// updateConfirm — handles the modal when it's the active screen
// ---------------------------------------------------------------------------

func (a AppModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "right", "l", "tab":
			if a.confirm.cursor == confirmCursorYes {
				a.confirm.cursor = confirmCursorNo
			} else {
				a.confirm.cursor = confirmCursorYes
			}

		case "y", "Y":
			a.confirm.accepted = true
			a.confirm.resolved = true
			return a.resolveConfirm()

		case "n", "N", "esc", "q":
			a.confirm.accepted = false
			a.confirm.resolved = true
			return a.resolveConfirm()

		case "enter", " ":
			a.confirm.accepted = a.confirm.cursor == confirmCursorYes
			a.confirm.resolved = true
			return a.resolveConfirm()
		}
	}
	return a, nil
}

// resolveConfirm routes back to the screen that opened the modal.
func (a AppModel) resolveConfirm() (tea.Model, tea.Cmd) {
	if !a.confirm.accepted {
		// Cancelled — go back to wherever we came from
		a.current = a.confirmReturn
		return a, nil
	}
	// Accepted — fire the delete-all command and return to delete screen
	a.current = a.confirmReturn
	if a.confirmReturn == screenDelete {
		return a, deleteAllEntriesCmd(a.db)
	}
	return a, nil
}