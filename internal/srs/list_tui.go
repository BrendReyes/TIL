package srs

import (
	"context"
	"fmt"
	"strings"

	"github.com/brendreyes/til/internal/database"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const pageSize = 20

// ---------------------------------------------------------------------------
// List sub-screens
// ---------------------------------------------------------------------------

type listSubScreen int

const (
	listSubMain listSubScreen = iota
	listSubDetail
	listSubEdit
	listSubDeleteConfirm
)

// Button indices for the detail bar
const (
	detailBtnEdit   = 0
	detailBtnDelete = 1
	detailBtnBack   = 2
)

// Button indices for the main list bar
const (
	mainBtnRefresh = 0
	mainBtnBack    = 1
)

// Button indices for the list inline delete confirm bar
const (
	listDeleteConfirmBtnYes = 0
	listDeleteConfirmBtnNo  = 1
)

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	listRowSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Bold(true)

	listRowNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	listRowMutedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	listTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	listDetailBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(1, 2).
				Width(60)

	listHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true)

	listDeleteWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true)
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type listScreenModel struct {
	db      *database.Queries
	entries []database.Entry
	loading bool
	err     error
	height  int // terminal height, set from AppModel

	cursor    int // index within current page
	page      int // 0-based page index
	totalPage int

	sub       listSubScreen
	editForm  *model // reuses edit_tui model
	editEntry database.Entry

	// detail sub-screen
	detailBar  buttonBar
	detailZone focusZone // zoneRows (metadata) or zoneButtons

	// main list bar
	mainBar  buttonBar
	mainZone focusZone // zoneRows (table) or zoneButtons

	// inline delete confirm bar
	deleteConfirmBar buttonBar
}

func newListScreenModel(db *database.Queries) listScreenModel {
	detailBar := newButtonBar([]string{"Edit", "Delete", "Back"})
	detailBar.SetDanger(detailBtnDelete)

	mainBar := newButtonBar([]string{"Refresh", "Back"})

	deleteConfirmBar := newButtonBar([]string{"Yes, Delete", "Cancel"})
	deleteConfirmBar.SetDanger(listDeleteConfirmBtnYes)
	deleteConfirmBar.cursor = listDeleteConfirmBtnNo // default to safe option
	deleteConfirmBar.focused = true

	return listScreenModel{
		db:               db,
		loading:          true,
		detailBar:        detailBar,
		mainBar:          mainBar,
		deleteConfirmBar: deleteConfirmBar,
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (l *listScreenModel) currentPageEntries() []database.Entry {
	start := l.page * pageSize
	if start >= len(l.entries) {
		return nil
	}
	end := start + pageSize
	if end > len(l.entries) {
		end = len(l.entries)
	}
	return l.entries[start:end]
}

func (l *listScreenModel) selectedEntry() (database.Entry, bool) {
	page := l.currentPageEntries()
	if len(page) == 0 || l.cursor >= len(page) {
		return database.Entry{}, false
	}
	return page[l.cursor], true
}

func (l *listScreenModel) visiblePageSize() int {
	const listChrome = 12
	max := pageSize
	if l.height > 0 {
		avail := l.height - listChrome
		if avail < 4 {
			avail = 4
		}
		if avail < max {
			max = avail
		}
	}
	return max
}

func (l *listScreenModel) clampCursor() {
	page := l.currentPageEntries()
	limit := l.visiblePageSize()
	if len(page) > limit {
		page = page[:limit]
	}
	if l.cursor >= len(page) {
		l.cursor = len(page) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
}

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

func deleteEntryCmd(db *database.Queries, id int64) tea.Cmd {
	return func() tea.Msg {
		_, err := db.DeleteEntry(context.Background(), id)
		return entryDeletedMsg{id: id, err: err}
	}
}

type entryDeletedMsg struct {
	id  int64
	err error
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (a AppModel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Keep list height in sync with terminal size.
	a.list.height = a.height

	switch a.list.sub {
	case listSubDetail:
		return a.updateListDetail(msg)
	case listSubEdit:
		return a.updateListEdit(msg)
	case listSubDeleteConfirm:
		return a.updateListDeleteConfirm(msg)
	default:
		return a.updateListMain(msg)
	}
}

// ── Main list ───────────────────────────────────────────────────────────────

func (a AppModel) updateListMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ── Async messages (zone-independent) ───────────────────────────────────
	switch msg := msg.(type) {
	case entriesFetchedMsg:
		if msg.err != nil {
			a.list.err = msg.err
			a.list.loading = false
			return a, nil
		}
		a.list.entries = msg.entries
		a.list.loading = false
		a.list.totalPage = (len(msg.entries) + pageSize - 1) / pageSize
		if a.list.totalPage == 0 {
			a.list.totalPage = 1
		}
		a.list.clampCursor()
		return a, nil

	case entryDeletedMsg:
		if msg.err != nil {
			a.list.err = msg.err
			return a, nil
		}
		updated := make([]database.Entry, 0, len(a.list.entries))
		for _, e := range a.list.entries {
			if e.ID != msg.id {
				updated = append(updated, e)
			}
		}
		a.list.entries = updated
		a.list.totalPage = (len(updated) + pageSize - 1) / pageSize
		if a.list.totalPage == 0 {
			a.list.totalPage = 1
		}
		if a.list.page >= a.list.totalPage {
			a.list.page = a.list.totalPage - 1
		}
		a.list.clampCursor()
		return a, nil
	}

	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	// esc / q always exit regardless of zone
	switch kMsg.String() {
	case "esc", "q":
		a.current = screenMenu
		a.list.mainZone = zoneRows
		a.list.mainBar.focused = false
		return a, nil
	}

	// ── Button bar zone ──────────────────────────────────────────────────────
	if a.list.mainZone == zoneButtons {
		switch kMsg.String() {
		case "tab", "shift+tab":
			a.list.mainZone = zoneRows
			a.list.mainBar.focused = false
			return a, nil
		}

		a.list.mainBar.HandleKey(kMsg.String())
		if a.list.mainBar.Activated() {
			switch a.list.mainBar.ActiveIndex() {
			case mainBtnRefresh:
				a.list.mainZone = zoneRows
				a.list.mainBar.focused = false
				a.list.loading = true
				return a, fetchEntriesCmd(a.db)
			case mainBtnBack:
				a.current = screenMenu
				a.list.mainZone = zoneRows
				a.list.mainBar.focused = false
			}
		}
		return a, nil
	}

	// ── Row zone ─────────────────────────────────────────────────────────────
	switch kMsg.String() {
	case "tab":
		a.list.mainZone = zoneButtons
		a.list.mainBar.focused = true

	case "up", "k":
		if a.list.cursor > 0 {
			a.list.cursor--
		}

	case "down", "j":
		page := a.list.currentPageEntries()
		limit := a.list.visiblePageSize()
		if len(page) > limit {
			page = page[:limit]
		}
		if a.list.cursor < len(page)-1 {
			a.list.cursor++
		}

	case "left", "h":
		if a.list.page > 0 {
			a.list.page--
			a.list.cursor = 0
		}

	case "right", "l":
		if a.list.page < a.list.totalPage-1 {
			a.list.page++
			a.list.cursor = 0
		}

	case "enter":
		if _, ok := a.list.selectedEntry(); ok {
			a.list.detailZone = zoneRows
			a.list.detailBar.focused = false
			a.list.detailBar.cursor = detailBtnEdit
			a.list.sub = listSubDetail
		}

	case "e":
		if entry, ok := a.list.selectedEntry(); ok {
			a.list.editEntry = entry
			a.list.editForm = newEditFormModel(entry.Body, entry.Tag)
			a.list.sub = listSubEdit
			return a, textarea.Blink
		}

	case "d":
		if _, ok := a.list.selectedEntry(); ok {
			a.list.deleteConfirmBar.cursor = listDeleteConfirmBtnNo
			a.list.sub = listSubDeleteConfirm
		}

	case "r":
		a.list.loading = true
		return a, fetchEntriesCmd(a.db)
	}

	return a, nil
}

// ── Detail sub-screen ───────────────────────────────────────────────────────

func (a AppModel) updateListDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	// esc always goes back, regardless of zone
	switch kMsg.String() {
	case "esc", "q":
		a.list.sub = listSubMain
		a.list.detailZone = zoneRows
		a.list.detailBar.focused = false
		a.list.detailBar.cursor = detailBtnEdit
		return a, nil
	}

	// ── Button bar focused ───────────────────────────────────────────────────
	if a.list.detailZone == zoneButtons {
		switch kMsg.String() {
		case "tab", "shift+tab":
			// tab out of bar → back to content zone
			a.list.detailZone = zoneRows
			a.list.detailBar.focused = false
			return a, nil
		}

		a.list.detailBar.HandleKey(kMsg.String())
		if a.list.detailBar.Activated() {
			switch a.list.detailBar.ActiveIndex() {
			case detailBtnEdit:
				if entry, ok := a.list.selectedEntry(); ok {
					a.list.editEntry = entry
					a.list.editForm = newEditFormModel(entry.Body, entry.Tag)
					a.list.detailZone = zoneRows
					a.list.detailBar.focused = false
					a.list.sub = listSubEdit
					return a, textarea.Blink
				}
			case detailBtnDelete:
				a.list.detailZone = zoneRows
				a.list.detailBar.focused = false
				a.list.deleteConfirmBar.cursor = listDeleteConfirmBtnNo
				a.list.sub = listSubDeleteConfirm
			case detailBtnBack:
				a.list.sub = listSubMain
				a.list.detailZone = zoneRows
				a.list.detailBar.focused = false
			}
		}
		return a, nil
	}

	// ── Content zone ─────────────────────────────────────────────────────────
	switch kMsg.String() {
	case "tab":
		a.list.detailZone = zoneButtons
		a.list.detailBar.focused = true
	case "enter", " ":
		a.list.sub = listSubMain
	case "e":
		if entry, ok := a.list.selectedEntry(); ok {
			a.list.editEntry = entry
			a.list.editForm = newEditFormModel(entry.Body, entry.Tag)
			a.list.sub = listSubEdit
			return a, textarea.Blink
		}
	case "d":
		a.list.deleteConfirmBar.cursor = listDeleteConfirmBtnNo
		a.list.sub = listSubDeleteConfirm
	}

	return a, nil
}

// ── Edit sub-screen ─────────────────────────────────────────────────────────

func (a AppModel) updateListEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.list.editForm == nil {
		a.list.sub = listSubMain
		return a, nil
	}

	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "esc":
			a.list.editForm = nil
			a.list.sub = listSubMain
			return a, nil

		case "ctrl+s", "ctrl+enter":
			return a.submitListEdit()

		case "tab":
			f := a.list.editForm
			switch f.focus {
			case editFocusBody:
				f.focus = editFocusTag
				f.bodyInput.Blur()
				f.bar.focused = false
				a.list.editForm = f
				return a, f.tagInput.Focus()
			case editFocusTag:
				f.focus = editFocusButtons
				f.tagInput.Blur()
				f.bar.focused = true
				a.list.editForm = f
				return a, nil
			case editFocusButtons:
				f.focus = editFocusBody
				f.bar.focused = false
				a.list.editForm = f
				return a, f.bodyInput.Focus()
			}

		case "shift+tab":
			f := a.list.editForm
			switch f.focus {
			case editFocusBody:
				f.focus = editFocusButtons
				f.bodyInput.Blur()
				f.bar.focused = true
				a.list.editForm = f
				return a, nil
			case editFocusTag:
				f.focus = editFocusBody
				f.tagInput.Blur()
				f.bar.focused = false
				a.list.editForm = f
				return a, f.bodyInput.Focus()
			case editFocusButtons:
				f.focus = editFocusTag
				f.bar.focused = false
				a.list.editForm = f
				return a, f.tagInput.Focus()
			}
		}

		// ── Button bar active ────────────────────────────────────────────────
		if a.list.editForm.focus == editFocusButtons {
			f := a.list.editForm
			f.bar.HandleKey(kMsg.String())
			if f.bar.Activated() {
				switch f.bar.ActiveIndex() {
				case editBtnSave:
					a.list.editForm = f
					return a.submitListEdit()
				case editBtnCancel:
					a.list.editForm = nil
					a.list.sub = listSubMain
					return a, nil
				}
			}
			a.list.editForm = f
			return a, nil
		}

		// ── enter on tag field saves ─────────────────────────────────────────
		if kMsg.String() == "enter" && a.list.editForm.focus == editFocusTag {
			return a.submitListEdit()
		}
	}

	// Handle save result
	if ev, ok := msg.(entrySavedMsg); ok {
		if ev.err != nil {
			a.list.editForm.err = ev.err
			return a, nil
		}
		for i, e := range a.list.entries {
			if e.ID == a.list.editEntry.ID {
				a.list.entries[i].Body = ev.body
				a.list.entries[i].Tag = ev.tag
				break
			}
		}
		a.list.editForm = nil
		a.list.sub = listSubMain
		return a, nil
	}

	f := a.list.editForm
	var cmd tea.Cmd
	if f.focus == editFocusBody {
		f.bodyInput, cmd = f.bodyInput.Update(msg)
	} else if f.focus == editFocusTag {
		f.tagInput, cmd = f.tagInput.Update(msg)
	}
	a.list.editForm = f
	return a, cmd
}

func (a AppModel) submitListEdit() (tea.Model, tea.Cmd) {
	f := a.list.editForm
	if strings.TrimSpace(f.bodyInput.Value()) == "" || strings.TrimSpace(f.tagInput.Value()) == "" {
		f.err = fmt.Errorf("Body and Tag cannot be empty")
		a.list.editForm = f
		return a, nil
	}
	body := strings.TrimSpace(f.bodyInput.Value())
	tag := strings.TrimSpace(f.tagInput.Value())
	id := a.list.editEntry.ID
	return a, saveEditCmd(a.db, id, body, tag)
}

// ── Delete confirm sub-screen ───────────────────────────────────────────────

func (a AppModel) updateListDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch kMsg.String() {
	case "esc", "q":
		a.list.deleteConfirmBar.cursor = listDeleteConfirmBtnNo
		a.list.sub = listSubMain
		return a, nil

	case "y", "Y":
		if entry, ok := a.list.selectedEntry(); ok {
			a.list.deleteConfirmBar.cursor = listDeleteConfirmBtnNo
			a.list.sub = listSubMain
			return a, deleteEntryCmd(a.db, entry.ID)
		}
	}

	a.list.deleteConfirmBar.HandleKey(kMsg.String())
	if a.list.deleteConfirmBar.Activated() {
		switch a.list.deleteConfirmBar.ActiveIndex() {
		case listDeleteConfirmBtnYes:
			if entry, ok := a.list.selectedEntry(); ok {
				a.list.deleteConfirmBar.cursor = listDeleteConfirmBtnNo
				a.list.sub = listSubMain
				return a, deleteEntryCmd(a.db, entry.ID)
			}
		case listDeleteConfirmBtnNo:
			a.list.deleteConfirmBar.cursor = listDeleteConfirmBtnNo
			a.list.sub = listSubMain
		}
	}

	return a, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (l listScreenModel) View() string {
	title := appTitleStyle.Render("  List / Browse  ")

	switch l.sub {
	case listSubDetail:
		return l.viewDetail(title)
	case listSubEdit:
		return l.viewEdit(title)
	case listSubDeleteConfirm:
		return l.viewDeleteConfirm(title)
	}

	return l.viewMain(title)
}

// ── Main list view ──────────────────────────────────────────────────────────

func (l listScreenModel) viewMain(title string) string {
	if l.loading {
		return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			title, "", appMutedStyle.Render("Loading...")))
	}
	if l.err != nil {
		return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			title, "", appErrorStyle.Render("Error: "+l.err.Error())))
	}

	if len(l.entries) == 0 {
		return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			title, "",
			appMutedStyle.Render("No entries yet. Use Add Entry to get started."),
			"",
			appHelpStyle.Render("esc: back to menu"),
		))
	}

	page := l.currentPageEntries()

	// Cap to visible rows so title stays on screen.
	maxRows := l.visiblePageSize()
	if len(page) > maxRows {
		page = page[:maxRows]
	}

	// Header row
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

		if i == l.cursor {
			rows += listRowSelectedStyle.Render("▶ "+line[2:]) + "\n"
		} else {
			rows += listRowNormalStyle.Render(line) + "\n"
		}
	}

	pagination := appMutedStyle.Render(fmt.Sprintf(
		"  Page %d/%d  (%d entries)",
		l.page+1, l.totalPage, len(l.entries),
	))

	bar := l.mainBar.View()

	var help string
	if l.mainZone == zoneButtons {
		help = appHelpStyle.Render("←/→: choose action • enter: activate • tab: back to list • esc: menu")
	} else {
		help = appHelpStyle.Render("↑/↓: navigate • enter: view • e: edit • d: delete • ←/→: page • tab: actions • esc: menu")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
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

// ── Detail view ─────────────────────────────────────────────────────────────

func (l listScreenModel) viewDetail(title string) string {
	entry, ok := l.selectedEntry()
	if !ok {
		return appDocStyle.Render(title + "\n\nNo entry selected.")
	}

	idLine := lipgloss.JoinHorizontal(lipgloss.Top,
		listRowMutedStyle.Render("Entry #"),
		listRowSelectedStyle.Render(fmt.Sprintf("%d", entry.ID)),
	)

	tagLine := lipgloss.JoinHorizontal(lipgloss.Top,
		listRowMutedStyle.Render("Tag:     "),
		listTagStyle.Render(entry.Tag),
	)

	timeLine := lipgloss.JoinHorizontal(lipgloss.Top,
		listRowMutedStyle.Render("Created: "),
		listRowNormalStyle.Render(humanize.Time(entry.CreatedAt)),
	)

	reviewLine := lipgloss.JoinHorizontal(lipgloss.Top,
		listRowMutedStyle.Render("Reviews: "),
		listRowNormalStyle.Render(fmt.Sprintf("%d  (last: %s)", entry.ReviewCount, humanize.Time(entry.LastReviewedAt))),
	)

	bodyBox := listDetailBoxStyle.Render(entry.Body)

	bar := l.detailBar.View()

	var help string
	if l.detailZone == zoneButtons {
		help = appHelpStyle.Render("←/→: choose action • enter: activate • tab: back to entry • esc: list")
	} else {
		help = appHelpStyle.Render("tab: actions • e: edit • d: delete • esc/enter: back to list")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		idLine,
		tagLine,
		timeLine,
		reviewLine,
		"",
		bodyBox,
		bar,
		help,
	)

	return appDocStyle.Render(content)
}

// ── Edit view ────────────────────────────────────────────────────────────────

func (l listScreenModel) viewEdit(title string) string {
	if l.editForm == nil {
		return appDocStyle.Render(title)
	}

	editTitle := appTitleStyle.Render("  Edit Entry #" + fmt.Sprintf("%d", l.editEntry.ID) + "  ")

	var errDisplay string
	if l.editForm.err != nil {
		errDisplay = appErrorStyle.Render("✗ "+l.editForm.err.Error()) + "\n"
	}

	f := l.editForm
	bodyLabel := "Learning Entry"
	tagLabel := "Tag (required)"
	var bodyView, tagView string

	switch f.focus {
	case editFocusBody:
		bodyLabel = focusedStyle.Render("● " + bodyLabel)
		tagLabel = blurredStyle.Render("  " + tagLabel)
		bodyView = focusedBorderStyle.Render(f.bodyInput.View())
		tagView = blurredBorderStyle.Render(f.tagInput.View())
	case editFocusTag:
		bodyLabel = blurredStyle.Render("  " + bodyLabel)
		tagLabel = focusedStyle.Render("● " + tagLabel)
		bodyView = blurredBorderStyle.Render(f.bodyInput.View())
		tagView = focusedBorderStyle.Render(f.tagInput.View())
	case editFocusButtons:
		bodyLabel = blurredStyle.Render("  " + bodyLabel)
		tagLabel = blurredStyle.Render("  " + tagLabel)
		bodyView = blurredBorderStyle.Render(f.bodyInput.View())
		tagView = blurredBorderStyle.Render(f.tagInput.View())
	}

	var help string
	if f.focus == editFocusButtons {
		help = appHelpStyle.Render("←/→: choose action • enter: activate • tab: back to body")
	} else {
		help = appHelpStyle.Render("tab: next field • enter: save (on tag) • ctrl+s: save • esc: cancel")
	}

	s := lipgloss.JoinVertical(lipgloss.Left,
		editTitle,
		"",
		errDisplay,
		bodyLabel,
		bodyView,
		"",
		tagLabel,
		tagView,
		f.bar.View(),
		help,
	)

	return appDocStyle.Render(s)
}

// ── Delete confirm view ──────────────────────────────────────────────────────

func (l listScreenModel) viewDeleteConfirm(title string) string {
	entry, ok := l.selectedEntry()
	if !ok {
		return appDocStyle.Render(title)
	}

	preview := entry.Body
	if len(preview) > 60 {
		preview = preview[:60] + "…"
	}

	warning := listDeleteWarningStyle.Render("⚠  Delete this entry permanently?")
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

	bar := l.deleteConfirmBar.View()

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")).
		Padding(1, 3).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			warning,
			"",
			idLine,
			tagLine,
			"",
			bodyBox,
			"",
			bar,
		))

	help := appHelpStyle.Render("←/→: choose • enter: confirm • y: yes • esc: cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		box,
		help,
	)

	return appDocStyle.Render(content)
}

// ---------------------------------------------------------------------------
// newEditFormModel — builds the embedded edit_tui model for list editing
// ---------------------------------------------------------------------------

func newEditFormModel(body, tag string) *model {
	ta := textarea.New()
	ta.Placeholder = "What did you learn?"
	ta.SetValue(body)
	ta.Focus()
	ta.CharLimit = 1000
	ta.SetWidth(50)
	ta.SetHeight(5)

	ti := textinput.New()
	ti.Placeholder = "e.g. go, sql, algorithms"
	ti.SetValue(tag)
	ti.CharLimit = 100
	ti.Width = 30

	return &model{
		bodyInput: ta,
		tagInput:  ti,
		focus:     editFocusBody,
		bar:       newEditBar(),
	}
}