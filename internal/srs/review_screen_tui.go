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
	reviewSubSession     reviewSubScreen = iota // active card review
	reviewSubEmpty                              // nothing due
	reviewSubComplete                           // session finished
	reviewSubResetResult                        // reset succeeded
)

// Button indices for the empty-state bar
const (
	reviewBtnBack  = 0
	reviewBtnReset = 1
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

	reviewResetSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true)
)

// ---------------------------------------------------------------------------
// reviewScreen — wraps reviewModel + sub-screen state + reset bar
// ---------------------------------------------------------------------------

type reviewScreen struct {
	model         *reviewModel
	sub           reviewSubScreen
	reviewedCount int       // snapshot at completion
	totalDue      int       // total entries when session started
	emptyBar      buttonBar // [Back] [Reset] on empty state
	resetCount    int64     // entries reset
	resetErr      error
}

func newReviewScreen() reviewScreen {
	bar := newButtonBar([]string{"Back", "Reset All"})
	bar.SetDanger(reviewBtnReset)
	return reviewScreen{emptyBar: bar}
}

// ---------------------------------------------------------------------------
// updateReview — main dispatcher
// ---------------------------------------------------------------------------

func (a AppModel) updateReview(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ── dueEntriesFetchedMsg: initialise session ─────────────────────────────
	if ev, ok := msg.(dueEntriesFetchedMsg); ok {
		rs := newReviewScreen()
		if ev.err != nil || len(ev.entries) == 0 {
			rs.sub = reviewSubEmpty
			a.reviewScreen = rs
			a.current = screenReview
			return a, nil
		}
		rs.model = NewReviewModel(ev.entries, a.db)
		rs.sub = reviewSubSession
		rs.totalDue = len(ev.entries)
		a.reviewScreen = rs
		a.current = screenReview
		return a, nil
	}

	// ── reviewResetMsg: reset completed ──────────────────────────────────────
	if ev, ok := msg.(reviewResetMsg); ok {
		a.reviewScreen.resetCount = ev.count
		a.reviewScreen.resetErr = ev.err
		a.reviewScreen.sub = reviewSubResetResult
		return a, nil
	}

	switch a.reviewScreen.sub {
	case reviewSubEmpty:
		return a.updateReviewEmpty(msg)
	case reviewSubComplete:
		return a.updateReviewComplete(msg)
	case reviewSubResetResult:
		return a.updateReviewResetResult(msg)
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

	// Handle reviewUpdatedMsg — avoid it triggering tea.Quit
	if ev, ok := msg.(reviewUpdatedMsg); ok {
		if ev.err != nil {
			a.reviewScreen.model.err = ev.err
			a.reviewScreen.reviewedCount = a.reviewScreen.model.reviewedCount
			a.reviewScreen.sub = reviewSubComplete
			return a, nil
		}
		if a.reviewScreen.model.quitting ||
			a.reviewScreen.model.currentIndex >= len(a.reviewScreen.model.entries) {
			a.reviewScreen.reviewedCount = a.reviewScreen.model.reviewedCount
			a.reviewScreen.sub = reviewSubComplete
			return a, nil
		}
		return a, nil
	}

	// Intercept q/esc to go to complete screen
	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "q", "esc":
			a.reviewScreen.reviewedCount = a.reviewScreen.model.reviewedCount
			a.reviewScreen.sub = reviewSubComplete
			return a, nil
		}
	}

	// Delegate to reviewModel.Update
	next, cmd := a.reviewScreen.model.Update(msg)
	rm := next.(*reviewModel)
	a.reviewScreen.model = rm

	if rm.quitting || rm.currentIndex >= len(rm.entries) {
		return a, cmd
	}

	return a, cmd
}

// ── Empty state ──────────────────────────────────────────────────────────────

func (a AppModel) updateReviewEmpty(msg tea.Msg) (tea.Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	bar := &a.reviewScreen.emptyBar

	switch kMsg.String() {
	case "r":
    a.confirm = newConfirmModel(
        "Reset ALL review progress?",
        []string{
            "Every entry's interval, ease factor, and review count",
            "will be cleared. This cannot be undone.",
        },
    )
    a.confirmReturn = screenReview
    a.current = screenConfirm
    return a, nil
	
	case "esc", "q":
		a.current = screenMenu
		return a, nil

	case "tab":
		bar.focused = !bar.focused
		return a, nil
	}

	// Delegate to button bar when focused
	if bar.focused {
		bar.HandleKey(kMsg.String())
		if bar.Activated() {
			switch bar.ActiveIndex() {
			case reviewBtnBack:
				a.current = screenMenu
				return a, nil
			case reviewBtnReset:
				// Open confirm modal — return here on cancel
				a.confirm = newConfirmModel(
					"Reset ALL review progress?",
					[]string{
						"Every entry's interval, ease factor, and review count",
						"will be cleared. This cannot be undone.",
					},
				)
				a.confirmReturn = screenReview
				a.current = screenConfirm
				return a, nil
			}
		}
		return a, nil
	}

	// Row-level shortcuts when bar not focused
	switch kMsg.String() {
	case "enter", " ":
		a.current = screenMenu
	}

	return a, nil
}

// ── Complete screen ──────────────────────────────────────────────────────────

func (a AppModel) updateReviewComplete(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "esc", "q", "enter", " ":
			a.reviewScreen = reviewScreen{}
			a.current = screenMenu
			return a, nil
		}
	}
	return a, nil
}

// ── Reset result screen ──────────────────────────────────────────────────────

func (a AppModel) updateReviewResetResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "esc", "q", "enter", " ":
			a.reviewScreen = reviewScreen{}
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
	case reviewSubResetResult:
		return a.viewReviewResetResult()
	default:
		if a.reviewScreen.model != nil {
			return a.reviewScreen.model.View()
		}
		return ""
	}
}

// ── Empty state view ─────────────────────────────────────────────────────────

func (a AppModel) viewReviewEmpty() string {
	title := appTitleStyle.Render("  Review  ")

	box := reviewStatBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		reviewEmptyStyle.Render("✓  You're all caught up!"),
		"",
		appMutedStyle.Render("Nothing is due for review right now."),
		appMutedStyle.Render("Keep adding entries and come back later."),
	))

	bar := a.reviewScreen.emptyBar.View()

	var focusHint string
	if a.reviewScreen.emptyBar.focused {
		focusHint = appHelpStyle.Render("←/→: choose action • enter: select • tab: back to info")
	} else {
		focusHint = appHelpStyle.Render("tab: focus buttons • r: reset all • esc: back to menu")
	}

	return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		title,
		box,
		bar,
		focusHint,
	))
}

// ── Complete view ────────────────────────────────────────────────────────────

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

// ── Reset result view ────────────────────────────────────────────────────────

func (a AppModel) viewReviewResetResult() string {
	title := appTitleStyle.Render("  Review Reset  ")

	var content string
	if a.reviewScreen.resetErr != nil {
		content = appErrorStyle.Render("✗ Error: " + a.reviewScreen.resetErr.Error())
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			reviewResetSuccessStyle.Render("✓  Reset complete!"),
			"",
			lipgloss.JoinHorizontal(lipgloss.Top,
				statsLabelStyle.Render("Entries reset:"),
				statsValueStyle.Render(fmt.Sprintf("%d", a.reviewScreen.resetCount)),
			),
			"",
			appMutedStyle.Render("All review progress cleared — everything is due again."),
		)
	}

	box := reviewStatBoxStyle.Render(content)
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