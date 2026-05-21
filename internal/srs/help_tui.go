package srs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	helpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true).
			MarginBottom(1)

	helpActionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Width(16)

	helpDescBodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	helpKeybindStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)

	helpDividerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("237"))

	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 3).
			MarginTop(1)
)

// ---------------------------------------------------------------------------
// Help content
// ---------------------------------------------------------------------------

type helpEntry struct {
	action   string
	desc     string
	keybinds string
}

var helpEntries = []helpEntry{
	{
		action:   "Add Entry",
		desc:     "Open the editor to capture a new learning entry. Fill in the body and a tag, then save.",
		keybinds: "tab: switch field • ctrl+s / enter: save • esc: back",
	},
	{
		action:   "List / Browse",
		desc:     "Paginated view of all your entries (20 per page). Select an entry to view details, edit, or delete it.",
		keybinds: "↑/↓: navigate • enter: view • e: edit • d: delete • ←/→: page • esc: back",
	},
	{
		action:   "Review",
		desc:     "Spaced repetition session for entries that are due. Rate each card — Again, Hard, Good, or Easy — to schedule the next review.",
		keybinds: "enter/space: reveal • 1-4 / arrows: rate • q: quit session",
	},
	{
		action:   "Delete",
		desc:     "Choose to delete a single entry (by selecting from the list) or wipe all entries at once. Both require confirmation.",
		keybinds: "↑/↓: choose • enter: confirm • esc: back",
	},
	{
		action:   "Stats",
		desc:     "Overview of your learning database — totals, reviewed vs unreviewed, due today, and a bar chart of entries per tag.",
		keybinds: "r: refresh • esc: back",
	},
	{
		action:   "Help",
		desc:     "You're looking at it.",
		keybinds: "esc: back",
	},
	{
		action:   "Quit",
		desc:     "Exit the TIL TUI and return to the terminal.",
		keybinds: "q / ctrl+c",
	},
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type helpScreenModel struct{}

func newHelpScreenModel() helpScreenModel {
	return helpScreenModel{}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (a AppModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			a.current = screenMenu
			return a, nil
		}
	}
	return a, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (h helpScreenModel) View() string {
	title := appTitleStyle.Render("  Help  ")

	divider := helpDividerStyle.Render("  " + repeat("─", 58))

	var rows string
	for i, entry := range helpEntries {
		action := helpActionStyle.Render(entry.action)
		desc := helpDescBodyStyle.Render(entry.desc)
		keys := helpKeybindStyle.Render("  " + entry.keybinds)

		block := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top, action, desc),
			keys,
		)
		rows += block
		if i < len(helpEntries)-1 {
			rows += "\n" + divider + "\n"
		}
	}

	box := helpBoxStyle.Render(rows)
	footer := appHelpStyle.Render("esc: back to menu")

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		box,
		footer,
	)

	return appDocStyle.Render(body)
}

// ---------------------------------------------------------------------------
// repeat — local string repeat helper (avoids importing strings for one call)
// ---------------------------------------------------------------------------

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}