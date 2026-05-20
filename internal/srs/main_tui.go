package srs

import (
	"github.com/brendreyes/til/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Screen identifiers
// ---------------------------------------------------------------------------

type screen int

const (
	screenMenu screen = iota
	screenAdd
	screenList
	screenReview
	screenDelete
	screenStats
	screenHelp
	screenConfirm
)

// ---------------------------------------------------------------------------
// Shared style tokens (used across all TUI files)
// ---------------------------------------------------------------------------

var (
	appTitleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true)

	appDocStyle = lipgloss.NewStyle().Margin(1, 2)

	appHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	appAccentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	appMutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	appBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	appErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	appSuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
)

// ---------------------------------------------------------------------------
// Root app model
// ---------------------------------------------------------------------------

type AppModel struct {
	db      *database.Queries
	current screen
	height  int // terminal height from tea.WindowSizeMsg

	// sub-models (one per screen)
	menu         menuModel
	add          addScreenModel
	list         listScreenModel
	reviewScreen reviewScreen
	delete       deleteScreenModel
	stats        statsScreenModel
	help         helpScreenModel
	confirm      confirmModel

	// after confirm resolves, where to go back
	confirmReturn screen
}

func NewAppModel(db *database.Queries) AppModel {
	return AppModel{
		db:           db,
		current:      screenMenu,
		menu:         newMenuModel(),
		reviewScreen: newReviewScreen(),
	}
}

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

func (a AppModel) Init() tea.Cmd {
	return nil
}

// ---------------------------------------------------------------------------
// Update — dispatches to the active screen's model
// ---------------------------------------------------------------------------

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global: ctrl+c always quits.
	if kMsg, ok := msg.(tea.KeyMsg); ok && kMsg.String() == "ctrl+c" {
		return a, tea.Quit
	}

	// Global: capture terminal height for dynamic page sizing.
	if sz, ok := msg.(tea.WindowSizeMsg); ok {
		a.height = sz.Height
		return a, nil
	}

	// Global: handle async data messages — route to the currently active screen.
	switch msg.(type) {
	case statsFetchedMsg:
		return a.updateStats(msg)
	case entryAddedMsg:
		return a.updateAdd(msg)
	case entriesFetchedMsg, entryDeletedMsg, entrySavedMsg:
		if a.current == screenDelete {
			return a.updateDelete(msg)
		}
		return a.updateList(msg)
	case allDeletedMsg:
		return a.updateDelete(msg)
	case dueEntriesFetchedMsg, reviewUpdatedMsg, reviewResetMsg:
		return a.updateReview(msg)
	}

	switch a.current {
	case screenMenu:
		return a.updateMenu(msg)
	case screenAdd:
		return a.updateAdd(msg)
	case screenList:
		return a.updateList(msg)
	case screenReview:
		return a.updateReview(msg)
	case screenDelete:
		return a.updateDelete(msg)
	case screenStats:
		return a.updateStats(msg)
	case screenHelp:
		return a.updateHelp(msg)
	case screenConfirm:
		return a.updateConfirm(msg)
	}
	return a, nil
}

// ---------------------------------------------------------------------------
// View — dispatches to the active screen's view
// ---------------------------------------------------------------------------

func (a AppModel) View() string {
	switch a.current {
	case screenMenu:
		return a.menu.View()
	case screenAdd:
		return a.add.View()
	case screenList:
		return a.list.View()
	case screenReview:
		return a.viewReview()
	case screenDelete:
		return a.delete.View()
	case screenStats:
		return a.stats.View()
	case screenHelp:
		return a.help.View()
	case screenConfirm:
		return a.confirm.View()
	}
	return ""
}

// ---------------------------------------------------------------------------
// Navigation helpers
// ---------------------------------------------------------------------------

func (a *AppModel) goTo(s screen) {
	a.current = s
}