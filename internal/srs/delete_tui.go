package srs

import (
	"context"
	"fmt"
	"strings"

	"github.com/brendreyes/til/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

// ---------------------------------------------------------------------------
// Delete sub-screens
// ---------------------------------------------------------------------------

type deleteSubScreen int

const (
	deleteSubList    deleteSubScreen = iota // paginated list
	deleteSubConfirm                        // single-entry confirm
	deleteSubResult                         // delete-all result
)

// Button indices for the list bar
const (
	deleteBtnDeleteAll = 0
	deleteBtnBack      = 1
)

// Button indices for the confirm bar
const (
	deleteConfirmBtnYes = 0
	deleteConfirmBtnNo  = 1
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
	deleteResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true)

	deleteConfirmBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("9")).
				Padding(1, 3).
				MarginTop(1)

	deleteWarningTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true)
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type deleteScreenModel struct {
	// list state
	entries   []database.Entry
	loading   bool
	err       error
	cursor    int
	page      int
	totalPage int
	zone      focusZone  // zoneRows or zoneButtons
	listBar   buttonBar  // [Delete All] [Back]

	// confirm state
	sub         deleteSubScreen
	selected    database.Entry
	confirmBar  buttonBar // [Yes] [No]

	// result state
	deleted int64
	delErr  error
}

func newDeleteScreenModel() deleteScreenModel {
	listBar := newButtonBar([]string{"Delete All", "Back"})
	listBar.SetDanger(deleteBtnDeleteAll)
	listBar.cursor = deleteBtnBack // default to safe option

	confirmBar := newButtonBar([]string{"Yes, Delete", "Cancel"})
	confirmBar.SetDanger(deleteConfirmBtnYes)
	confirmBar.cursor = deleteConfirmBtnNo // default safe

	return deleteScreenModel{
		loading:    true,
		sub:        deleteSubList,
		listBar:    listBar,
		confirmBar: confirmBar,
	}
}

// ---------------------------------------------------------------------------
// Pagination helpers (mirrors listScreenModel)
// ---------------------------------------------------------------------------

func (d *deleteScreenModel) currentPageEntries() []database.Entry {
	start := d.page * pageSize
	if start >= len(d.entries) {
		return nil
	}
	end := start + pageSize
	if end > len(d.entries) {
		end = len(d.entries)
	}
	return d.entries[start:end]
}

func (d *deleteScreenModel) clampCursor() {
	page := d.currentPageEntries()
	if len(page) == 0 {
		d.cursor = 0
		return
	}
	if d.cursor >= len(page) {
		d.cursor = len(page) - 1
	}
	if d.cursor < 0 {
		d.cursor = 0
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
	// Async results — handle regardless of sub-screen
	switch ev := msg.(type) {
	case entriesFetchedMsg:
		if ev.err != nil {
			a.delete.err = ev.err
			a.delete.loading = false
			return a, nil
		}
		a.delete.entries = ev.entries
		a.delete.loading = false
		a.delete.totalPage = (len(ev.entries) + pageSize - 1) / pageSize
		if a.delete.totalPage == 0 {
			a.delete.totalPage = 1
		}
		a.delete.clampCursor()
		return a, nil

	case entryDeletedMsg:
		if ev.err != nil {
			a.delete.err = ev.err
			a.delete.sub = deleteSubList
			return a, nil
		}
		// Remove from local slice, stay on list
		updated := make([]database.Entry, 0, len(a.delete.entries))
		for _, e := range a.delete.entries {
			if e.ID != ev.id {
				updated = append(updated, e)
			}
		}
		a.delete.entries = updated
		a.delete.totalPage = (len(updated) + pageSize - 1) / pageSize
		if a.delete.totalPage == 0 {
			a.delete.totalPage = 1
		}
		if a.delete.page >= a.delete.totalPage {
			a.delete.page = a.delete.totalPage - 1
		}
		a.delete.clampCursor()
		a.delete.sub = deleteSubList
		// Reset confirm bar for next use
		a.delete.confirmBar.cursor = deleteConfirmBtnNo
		return a, nil

	case allDeletedMsg:
		a.delete.delErr = ev.err
		a.delete.deleted = ev.count
		a.delete.sub = deleteSubResult
		return a, nil
	}

	switch a.delete.sub {
	case deleteSubConfirm:
		return a.updateDeleteConfirm(msg)
	case deleteSubResult:
		return a.updateDeleteResult(msg)
	default:
		return a.updateDeleteList(msg)
	}
}

// ── List sub-screen ──────────────────────────────────────────────────────────

func (a AppModel) updateDeleteList(msg tea.Msg) (tea.Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch kMsg.String() {
	case "esc", "q":
		a.current = screenMenu
		return a, nil

	case "tab":
		if a.delete.zone == zoneRows {
			a.delete.zone = zoneButtons
			a.delete.listBar.focused = true
		} else {
			a.delete.zone = zoneRows
			a.delete.listBar.focused = false
		}
		return a, nil
	}

	// ── Button bar focused ───────────────────────────────────────────────────
	if a.delete.zone == zoneButtons {
		a.delete.listBar.HandleKey(kMsg.String())
		if a.delete.listBar.Activated() {
			switch a.delete.listBar.ActiveIndex() {
			case deleteBtnDeleteAll:
				a.confirm = newConfirmModel(
					"Delete ALL entries?",
					[]string{
						"This will permanently erase your entire learning database.",
						"This action cannot be undone.",
					},
				)
				a.confirmReturn = screenDelete
				a.current = screenConfirm
				return a, nil
			case deleteBtnBack:
				a.current = screenMenu
				return a, nil
			}
		}
		return a, nil
	}

	// ── Row navigation ───────────────────────────────────────────────────────
	switch kMsg.String() {
	case "up", "k":
		if a.delete.cursor > 0 {
			a.delete.cursor--
		}

	case "down", "j":
		page := a.delete.currentPageEntries()
		if a.delete.cursor < len(page)-1 {
			a.delete.cursor++
		}

	case "left", "h":
		if a.delete.page > 0 {
			a.delete.page--
			a.delete.cursor = 0
		}

	case "right", "l":
		if a.delete.page < a.delete.totalPage-1 {
			a.delete.page++
			a.delete.cursor = 0
		}

	case "enter", " ":
		page := a.delete.currentPageEntries()
		if len(page) == 0 || a.delete.cursor >= len(page) {
			return a, nil
		}
		a.delete.selected = page[a.delete.cursor]
		a.delete.confirmBar.cursor = deleteConfirmBtnNo
		a.delete.confirmBar.focused = true
		a.delete.sub = deleteSubConfirm
	}

	return a, nil
}

// ── Confirm sub-screen ───────────────────────────────────────────────────────

func (a AppModel) updateDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch kMsg.String() {
	case "esc", "q":
		a.delete.sub = deleteSubList
		a.delete.confirmBar.focused = false
		return a, nil
	}

	a.delete.confirmBar.HandleKey(kMsg.String())
	if a.delete.confirmBar.Activated() {
		switch a.delete.confirmBar.ActiveIndex() {
		case deleteConfirmBtnYes:
			id := a.delete.selected.ID
			return a, deleteEntryCmd(a.db, id)
		case deleteConfirmBtnNo:
			a.delete.sub = deleteSubList
			a.delete.confirmBar.focused = false
		}
	}

	return a, nil
}

// ── Result screen ────────────────────────────────────────────────────────────

func (a AppModel) updateDeleteResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "esc", "q", "enter", " ":
			a.delete = newDeleteScreenModel()
			a.current = screenMenu
			return a, nil
		}
	}
	return a, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (d deleteScreenModel) View() string {
	title := appTitleStyle.Render("  Delete  ")
	switch d.sub {
	case deleteSubConfirm:
		return d.viewConfirm(title)
	case deleteSubResult:
		return d.viewResult(title)
	default:
		return d.viewList(title)
	}
}

// ── List view ────────────────────────────────────────────────────────────────

func (d deleteScreenModel) viewList(title string) string {
	if d.loading {
		return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			title, "", appMutedStyle.Render("Loading...")))
	}
	if d.err != nil {
		return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			title, "", appErrorStyle.Render("Error: "+d.err.Error())))
	}

	if len(d.entries) == 0 {
		return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			title, "",
			appMutedStyle.Render("No entries to delete."),
			"",
			appHelpStyle.Render("esc: back to menu"),
		))
	}

	page := d.currentPageEntries()

	header := listHeaderStyle.Render(fmt.Sprintf(
		"  %-5s %-42s %-14s %s",
		"ID", "Body", "Tag", "Created",
	))
	divider := listRowMutedStyle.Render(strings.Repeat("─", 72))

	var rows string
	for i, entry := range page {
		body := entry.Body
		if len(body) > 40 {
			body = body[:40] + "…"
		}
		tag := entry.Tag
		if len(tag) > 12 {
			tag = tag[:12] + "…"
		}
		created := humanize.Time(entry.CreatedAt)
		if len(created) > 14 {
			created = created[:14]
		}
		line := fmt.Sprintf("  %-5d %-42s %-14s %s",
			entry.ID, body, tag, created)

		if d.zone == zoneRows && i == d.cursor {
			rows += listRowSelectedStyle.Render("▶ "+line[2:]) + "\n"
		} else {
			rows += listRowNormalStyle.Render(line) + "\n"
		}
	}

	pagination := appMutedStyle.Render(fmt.Sprintf(
		"  Page %d/%d  (%d entries)",
		d.page+1, d.totalPage, len(d.entries),
	))

	warning := deleteWarningTextStyle.Render("  ⚠  Select an entry to delete it, or use the buttons below.")

	bar := d.listBar.View()

	var help string
	if d.zone == zoneButtons {
		help = appHelpStyle.Render("←/→: choose action • enter: activate • tab: back to list")
	} else {
		help = appHelpStyle.Render("↑/↓: navigate • enter: select to delete • ←/→: page • tab: actions • esc: menu")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		warning,
		"",
		header,
		divider,
		rows,
		divider,
		pagination,
		bar,
		help,
	)

	return appDocStyle.Render(body)
}

// ── Confirm view ─────────────────────────────────────────────────────────────

func (d deleteScreenModel) viewConfirm(title string) string {
	entry := d.selected

	preview := entry.Body
	if len(preview) > 58 {
		preview = preview[:58] + "…"
	}

	warning := deleteWarningTextStyle.Render("⚠  Delete this entry permanently?")

	idLine := listRowMutedStyle.Render(fmt.Sprintf("Entry #%d", entry.ID))
	tagLine := lipgloss.JoinHorizontal(lipgloss.Top,
		listRowMutedStyle.Render("Tag: "),
		listTagStyle.Render(entry.Tag),
	)
	bodyBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2).
		Width(60).
		Render(preview)

	bar := d.confirmBar.View()

	box := deleteConfirmBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		warning,
		"",
		idLine,
		tagLine,
		"",
		bodyBox,
		"",
		bar,
	))

	help := appHelpStyle.Render("←/→: choose • enter: confirm • esc: back to list")

	return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		title,
		box,
		help,
	))
}

// ── Result view ──────────────────────────────────────────────────────────────

func (d deleteScreenModel) viewResult(title string) string {
	var content string
	if d.delErr != nil {
		content = appErrorStyle.Render("✗ Error: " + d.delErr.Error())
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

	return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		title,
		box,
		help,
	))
}