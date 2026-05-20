package srs

import (
	"fmt"
	"strings"

	"github.com/brendreyes/til/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	statsLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Width(20)

	statsValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Bold(true)

	statsDueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	statsTagBarBg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	statsTagBarFg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true)

	statsSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("62")).
				Bold(true).
				MarginTop(1)

	statsBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 2).
			MarginTop(1)
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type statsScreenModel struct {
	loaded     bool
	loading    bool
	err        error
	total      int64
	reviewed   int64
	unreviewed int64
	due        int64
	byTag      []database.CountEntriesByTagRow
}

func newStatsScreenModel() statsScreenModel {
	return statsScreenModel{loading: true}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (a AppModel) updateStats(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			a.current = screenMenu
			return a, nil
		case "r":
			// refresh
			a.stats = newStatsScreenModel()
			return a, fetchStatsCmd(a.db)
		}

	case statsFetchedMsg:
		if msg.err != nil {
			a.stats.err = msg.err
			a.stats.loading = false
			return a, nil
		}
		a.stats = statsScreenModel{
			loaded:     true,
			loading:    false,
			total:      msg.total,
			reviewed:   msg.reviewed,
			unreviewed: msg.unreviewed,
			due:        msg.due,
			byTag:      msg.byTag,
		}
		return a, nil
	}

	return a, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (s statsScreenModel) View() string {
	title := appTitleStyle.Render("  Stats  ")

	if s.loading {
		return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			title,
			appMutedStyle.Render("\n  Loading..."),
		))
	}

	if s.err != nil {
		return appDocStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			title,
			appErrorStyle.Render("\n  Error: "+s.err.Error()),
		))
	}

	// ── Summary block ────────────────────────────────────────────────────────

	row := func(label string, value string, style lipgloss.Style) string {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			statsLabelStyle.Render(label),
			style.Render(value),
		)
	}

	reviewedPct := 0
	if s.total > 0 {
		reviewedPct = int(float64(s.reviewed) / float64(s.total) * 100)
	}

	progressBar := renderProgressBar(reviewedPct, 30)

	summary := lipgloss.JoinVertical(lipgloss.Left,
		statsSectionStyle.Render("Overview"),
		row("Total entries:", fmt.Sprintf("%d", s.total), statsValueStyle),
		row("Reviewed:", fmt.Sprintf("%d  (%d%%)", s.reviewed, reviewedPct), statsValueStyle),
		row("Unreviewed:", fmt.Sprintf("%d", s.unreviewed), statsValueStyle),
		row("Due today:", fmt.Sprintf("%d", s.due), statsDueStyle),
		"",
		statsLabelStyle.Render("Progress:")+"  "+progressBar,
	)

	summaryBox := statsBoxStyle.Render(summary)

	// ── By tag block ─────────────────────────────────────────────────────────

	var tagRows string
	maxCount := int64(1)
	for _, t := range s.byTag {
		if t.Count > maxCount {
			maxCount = t.Count
		}
	}

	for _, t := range s.byTag {
		barLen := int(float64(t.Count) / float64(maxCount) * 20)
		bar := statsTagBarFg.Render(strings.Repeat("█", barLen)) +
			statsTagBarBg.Render(strings.Repeat("░", 20-barLen))

		tagLabel := lipgloss.NewStyle().Width(18).Foreground(lipgloss.Color("252")).Render(t.Tag)
		count := statsValueStyle.Render(fmt.Sprintf("%d", t.Count))
		tagRows += lipgloss.JoinHorizontal(lipgloss.Top, tagLabel, bar, "  ", count) + "\n"
	}

	var tagsSection string
	if tagRows == "" {
		tagsSection = appMutedStyle.Render("  No tags yet.")
	} else {
		tagsSection = lipgloss.JoinVertical(lipgloss.Left,
			statsSectionStyle.Render("Entries by tag"),
			tagRows,
		)
	}

	tagsBox := statsBoxStyle.Render(tagsSection)

	help := appHelpStyle.Render("r: refresh • esc: back to menu")

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		summaryBox,
		tagsBox,
		help,
	)

	return appDocStyle.Render(body)
}

// ---------------------------------------------------------------------------
// renderProgressBar
// ---------------------------------------------------------------------------

func renderProgressBar(pct int, width int) string {
	filled := int(float64(pct) / 100.0 * float64(width))
	if filled > width {
		filled = width
	}
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("237")).Render(strings.Repeat("░", width-filled))
	return bar + fmt.Sprintf("  %d%%", pct)
}