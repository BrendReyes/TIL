package srs

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Menu items
// ---------------------------------------------------------------------------

type menuItem struct {
	label    string
	shortcut string
	desc     string
}

var menuItems = []menuItem{
	{"Add Entry", "1", "Capture a new learning entry"},
	{"List / Browse", "2", "Browse and manage your entries"},
	{"Review", "3", "Start a spaced repetition review session"},
	{"Delete", "4", "Delete entries"},
	{"Stats", "5", "View statistics about your learning"},
	{"Help", "6", "Show descriptions for each action"},
	{"Quit", "7", "Exit TIL"},
}

// ---------------------------------------------------------------------------
// Menu styles
// ---------------------------------------------------------------------------

var (
	menuSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Padding(0, 1).
				Bold(true)

	menuNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	menuShortcutStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true)

	menuDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	menuBannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true)
)

const menuBanner = `
 ████████╗    ██╗    ██╗     
    ██╔══╝    ██║    ██║     
    ██║       ██║    ██║     
    ██║       ██║    ██║     
    ██║       ██║    ███████╗
    ╚═╝       ╚═╝    ╚══════╝
  Today I Learned
`

// ---------------------------------------------------------------------------
// menuModel
// ---------------------------------------------------------------------------

type menuModel struct {
	cursor int
}

func newMenuModel() menuModel {
	return menuModel{cursor: 0}
}

func (m menuModel) View() string {
	banner := menuBannerStyle.Render(menuBanner)

	var rows string
	for i, item := range menuItems {
		shortcut := menuShortcutStyle.Render(fmt.Sprintf("[%s]", item.shortcut))
		desc := menuDescStyle.Render("  " + item.desc)

		var label string
		if i == m.cursor {
			label = menuSelectedStyle.Render(fmt.Sprintf("▶  %s", item.label))
		} else {
			label = menuNormalStyle.Render(fmt.Sprintf("   %s", item.label))
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			shortcut,
			" ",
			label,
			desc,
		)
		rows += row + "\n"
	}

	help := appHelpStyle.Render("↑/↓: navigate • enter/number: select • ctrl+c: quit")

	body := lipgloss.JoinVertical(lipgloss.Left,
		banner,
		rows,
		help,
	)

	return appDocStyle.Render(body)
}

// ---------------------------------------------------------------------------
// updateMenu
// ---------------------------------------------------------------------------

func (a AppModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit

		case "up", "k":
			if a.menu.cursor > 0 {
				a.menu.cursor--
			}

		case "down", "j":
			if a.menu.cursor < len(menuItems)-1 {
				a.menu.cursor++
			}

		case "enter", " ":
			return a.selectMenuItem(a.menu.cursor)

		// number shortcuts
		case "1":
			return a.selectMenuItem(0)
		case "2":
			return a.selectMenuItem(1)
		case "3":
			return a.selectMenuItem(2)
		case "4":
			return a.selectMenuItem(3)
		case "5":
			return a.selectMenuItem(4)
		case "6":
			return a.selectMenuItem(5)
		case "7":
			return a.selectMenuItem(6)
		}
	}
	return a, nil
}

func (a AppModel) selectMenuItem(idx int) (tea.Model, tea.Cmd) {
	switch idx {
	case 0: // Add
		a.add = newAddScreenModel()
		a.current = screenAdd
	case 1: // List
		a.list = newListScreenModel(a.db)
		a.current = screenList
		return a, fetchEntriesCmd(a.db)
	case 2: // Review
		return a, fetchDueEntriesCmd(a.db)
	case 3: // Delete — open directly to paginated list
		a.delete = newDeleteScreenModel()
		a.current = screenDelete
		return a, fetchEntriesCmd(a.db)
	case 4: // Stats
		a.stats = newStatsScreenModel()
		a.current = screenStats
		return a, fetchStatsCmd(a.db)
	case 5: // Help
		a.help = newHelpScreenModel()
		a.current = screenHelp
	case 6: // Quit
		return a, tea.Quit
	}
	return a, nil
}