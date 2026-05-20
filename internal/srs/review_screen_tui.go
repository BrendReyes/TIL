package srs

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Review sub-screens
// ---------------------------------------------------------------------------

type reviewSubScreen int

const (
	reviewSubSession    reviewSubScreen = iota // active card review
	reviewSubEmpty                             // nothing due
	reviewSubComplete                          // session finished
)

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	reviewCompleteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true)

	reviewEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true)

	reviewStatBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(1, 4).
				MarginTop(1)
)

// ---------------------------------------------------------------------------
// reviewScreenState — wraps reviewModel + sub-screen state
// Stored in AppModel.review (pointer); nil means not started.
// ---------------------------------------------------------------------------

// We track sub-screen and completion data alongside the reviewModel pointer.
// These live in AppModel directly since review is already a pointer field.

type reviewScreen struct {
	model         *reviewModel
	sub           reviewSubScreen
	reviewedCount int // snapshot at completion
	totalDue      int // total entries when session started
}

// ---------------------------------------------------------------------------
// updateReview — main dispatcher
// ---------------------------------------------------------------------------

func (a AppModel) updateReview(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ── Handle dueEntriesFetchedMsg: initialise session ─────────────────────
	if ev, ok := msg.(dueEntriesFetchedMsg); ok {
		if ev.err != nil || len(ev.entries) == 0 {
			// Nothing due — show empty state
			a.reviewScreen.sub = reviewSubEmpty
			a.current = screenReview
			return a, nil
		}
		a.reviewScreen = reviewScreen{
			model:    NewReviewModel(ev.entries, a.db),
			sub:      reviewSubSession,
			totalDue: len(ev.entries),
		}
		a.current = screenReview
		return a, nil
	}

	switch a.reviewScreen.sub {
	case reviewSubEmpty:
		return a.updateReviewEmpty(msg)
	case reviewSubComplete:
		return a.updateReviewComplete(msg)
	default:
		return a.updateReviewSession(msg)
	}
}

// ── Active session ───────────────────────────────────────────────────────────

func (a AppModel) updateReviewSession(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.reviewScreen.model == nil {
		a.current = screenMenu
		return a, nil
	}

	// Handle reviewUpdatedMsg here to avoid it triggering tea.Quit
	if ev, ok := msg.(reviewUpdatedMsg); ok {
		if ev.err != nil {
			a.reviewScreen.model.err = ev.err
			a.reviewScreen.reviewedCount = a.reviewScreen.model.reviewedCount
			a.reviewScreen.sub = reviewSubComplete
			return a, nil
		}
		// If the model flagged quitting, transition to complete screen
		if a.reviewScreen.model.quitting ||
			a.reviewScreen.model.currentIndex >= len(a.reviewScreen.model.entries) {
			a.reviewScreen.reviewedCount = a.reviewScreen.model.reviewedCount
			a.reviewScreen.sub = reviewSubComplete
			return a, nil
		}
		return a, nil
	}

	// Intercept q/esc/ctrl+c to go to menu directly
	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "q", "esc", "ctrl+c":
			a.reviewScreen.reviewedCount = a.reviewScreen.model.reviewedCount
			a.reviewScreen.sub = reviewSubComplete
			return a, nil
		}
	}

	// Delegate to reviewModel.Update — intercept tea.Quit via quitting flag.
	next, cmd := a.reviewScreen.model.Update(msg)
	rm := next.(*reviewModel)
	a.reviewScreen.model = rm

	if rm.quitting || rm.currentIndex >= len(rm.entries) {
		// Don't quit the program — completion screen lands when reviewUpdatedMsg arrives.
		return a, cmd
	}

	return a, cmd
}

// ── Empty state ──────────────────────────────────────────────────────────────

func (a AppModel) updateReviewEmpty(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "esc", "q", "enter", " ":
			a.current = screenMenu
			return a, nil
		}
	}
	return a, nil
}

// ── Complete screen ──────────────────────────────────────────────────────────

func (a AppModel) updateReviewComplete(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "esc", "q", "enter", " ":
			a.reviewScreen = reviewScreen{} // reset
			a.current = screenMenu
			return a, nil
		}
	}
	return a, nil
}

// ---------------------------------------------------------------------------
// View — dispatched from AppModel.View()
// ---------------------------------------------------------------------------

func (a AppModel) viewReview() string {
	switch a.reviewScreen.sub {
	case reviewSubEmpty:
		return a.viewReviewEmpty()
	case reviewSubComplete:
		return a.viewReviewComplete()
	default:
		if a.reviewScreen.model != nil {
			return a.reviewScreen.model.View()
		}
		return ""
	}
}

func (a AppModel) viewReviewEmpty() string {
	title := appTitleStyle.Render("  Review  ")

	box := reviewStatBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		reviewEmptyStyle.Render("✓  You're all caught up!"),
		"",
		appMutedStyle.Render("Nothing is due for review right now."),
		appMutedStyle.Render("Keep adding entries and come back later."),
	))

	help := appHelpStyle.Render("enter/esc: back to menu")

	return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		title,
		box,
		help,
	))
}

func (a AppModel) viewReviewComplete() string {
	title := appTitleStyle.Render("  Review Complete  ")

	reviewed := a.reviewScreen.reviewedCount
	total := a.reviewScreen.totalDue
	skipped := total - reviewed

	summary := lipgloss.JoinVertical(lipgloss.Left,
		reviewCompleteStyle.Render("✓  Session complete!"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top,
			statsLabelStyle.Render("Reviewed:"),
			statsValueStyle.Render(fmt.Sprintf("%d", reviewed)),
		),
		lipgloss.JoinHorizontal(lipgloss.Top,
			statsLabelStyle.Render("Skipped:"),
			appMutedStyle.Render(fmt.Sprintf("%d", skipped)),
		),
		lipgloss.JoinHorizontal(lipgloss.Top,
			statsLabelStyle.Render("Total due:"),
			appMutedStyle.Render(fmt.Sprintf("%d", total)),
		),
	)

	box := reviewStatBoxStyle.Render(summary)
	help := appHelpStyle.Render("enter/esc: back to menu")

	return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		title,
		box,
		help,
	))
}

// ---------------------------------------------------------------------------
// RunMainTUI — entry point called from cmd/tui.go
// ---------------------------------------------------------------------------

func (s *State) RunMainTUI() error {
	m := NewAppModel(s.DB)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}