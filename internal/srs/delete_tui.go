package srs

import (
	"context"
	"fmt"
	"github.com/brendreyes/til/internal/database"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Delete sub-screens
// ---------------------------------------------------------------------------

type deleteSubScreen int

const (
	deleteSubMenu   deleteSubScreen = iota // main submenu
	deleteSubResult                        // result screen after delete-all
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

type allDeletedMsg struct {
	count int64
	err   error
}

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	deleteMenuSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("9")).
				Padding(0, 1).
				Bold(true)

	deleteMenuNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 1)

	deleteResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true)
)

// ---------------------------------------------------------------------------
// Delete screen items
// ---------------------------------------------------------------------------

type deleteMenuItem struct {
	label string
	desc  string
}

var deleteMenuItems = []deleteMenuItem{
	{"Delete by Selecting", "Browse entries and delete one at a time"},
	{"Delete All", "Permanently remove every entry from your database"},
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type deleteScreenModel struct {
	cursor  int
	sub     deleteSubScreen
	result  string
	err     error
	deleted int64
}

func newDeleteScreenModel() deleteScreenModel {
	return deleteScreenModel{
		cursor: 0,
		sub:    deleteSubMenu,
	}
}

// ---------------------------------------------------------------------------
// Command
// ---------------------------------------------------------------------------

func deleteAllEntriesCmd(db *database.Queries) tea.Cmd {
	return func() tea.Msg {
		count, err := db.DeleteAllEntries(context.Background())
		return allDeletedMsg{count: count, err: err}
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (a AppModel) updateDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle async delete-all result regardless of sub-screen
	if ev, ok := msg.(allDeletedMsg); ok {
		if ev.err != nil {
			a.delete.err = ev.err
			a.delete.sub = deleteSubResult
			return a, nil
		}
		a.delete.deleted = ev.count
		a.delete.sub = deleteSubResult
		return a, nil
	}

	switch a.delete.sub {
	case deleteSubResult:
		return a.updateDeleteResult(msg)
	default:
		return a.updateDeleteMenu(msg)
	}
}

// ── Submenu ──────────────────────────────────────────────────────────────────

func (a AppModel) updateDeleteMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			a.current = screenMenu
			return a, nil

		case "up", "k":
			if a.delete.cursor > 0 {
				a.delete.cursor--
			}

		case "down", "j":
			if a.delete.cursor < len(deleteMenuItems)-1 {
				a.delete.cursor++
			}

		case "1":
			return a.selectDeleteItem(0)
		case "2":
			return a.selectDeleteItem(1)

		case "enter", " ":
			return a.selectDeleteItem(a.delete.cursor)
		}
	}
	return a, nil
}

func (a AppModel) selectDeleteItem(idx int) (tea.Model, tea.Cmd) {
	switch idx {
	case 0: // Delete by selecting — go to list screen
		a.list = newListScreenModel(a.db)
		a.current = screenList
		return a, fetchEntriesCmd(a.db)

	case 1: // Delete all — open confirm modal
		a.confirm = newConfirmModel(
			"Delete ALL entries?",
			[]string{
				"This will permanently erase your entire learning database.",
				"This action cannot be undone.",
			},
		)
		a.confirmReturn = screenDelete
		a.current = screenConfirm
	}
	return a, nil
}

// ── Result screen ─────────────────────────────────────────────────────────────

func (a AppModel) updateDeleteResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter", " ":
			a.delete = newDeleteScreenModel()
			a.current = screenMenu
			return a, nil
		}
	}
	_ = msg
	return a, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (d deleteScreenModel) View() string {
	title := appTitleStyle.Render("  Delete  ")

	switch d.sub {
	case deleteSubResult:
		return d.viewResult(title)
	default:
		return d.viewMenu(title)
	}
}

// ── Submenu view ─────────────────────────────────────────────────────────────

func (d deleteScreenModel) viewMenu(title string) string {
	var rows string
	for i, item := range deleteMenuItems {
		shortcut := menuShortcutStyle.Render(fmt.Sprintf("[%d]", i+1))
		desc := menuDescStyle.Render("  " + item.desc)

		var label string
		if i == d.cursor {
			label = deleteMenuSelectedStyle.Render(fmt.Sprintf("▶  %s", item.label))
		} else {
			label = deleteMenuNormalStyle.Render(fmt.Sprintf("   %s", item.label))
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, shortcut, " ", label, desc)
		rows += row + "\n"
	}

	warning := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		MarginTop(1).
		Render("⚠  Deletions are permanent and cannot be undone.")

	help := appHelpStyle.Render("↑/↓: navigate • enter/number: select • esc: back to menu")

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		rows,
		warning,
		"",
		help,
	)

	return appDocStyle.Render(body)
}

// ── Result view ───────────────────────────────────────────────────────────────

func (d deleteScreenModel) viewResult(title string) string {
	var content string
	if d.err != nil {
		content = appErrorStyle.Render("✗ Error: " + d.err.Error())
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			deleteResultStyle.Render(fmt.Sprintf("✓ Deleted %d entries.", d.deleted)),
			"",
			appMutedStyle.Render("Your learning database has been cleared."),
		)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 3).
		MarginTop(1).
		Render(content)

	help := appHelpStyle.Render("press any key to return to menu")

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		box,
		help,
	)

	return appDocStyle.Render(body)
}

// ---------------------------------------------------------------------------
// View helper — delete submenu and result only; single-entry confirm is
// handled inline in list_tui.go (viewDeleteConfirm).
// ---------------------------------------------------------------------------