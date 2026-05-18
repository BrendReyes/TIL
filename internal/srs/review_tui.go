package srs

import (
	"context"
	"fmt"
	"strings"

	"github.com/brendreyes/til/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	reviewFocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	reviewBlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	reviewBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1).
				Width(60)

	reviewDocStyle  = lipgloss.NewStyle().Margin(1, 2)
	reviewHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	reviewTitleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			MarginBottom(1).
			Bold(true)

	scoreSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("205")).
				Padding(0, 1).
				MarginRight(1).
				Bold(true)

	scoreNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Background(lipgloss.Color("235")).
				Padding(0, 1).
				MarginRight(1)
)

type reviewUpdatedMsg struct{ err error }

type reviewModel struct {
	entries       []database.Entry
	currentIndex  int
	db            *database.Queries
	showAnswer    bool
	selection     int // 0: Again, 1: Hard, 2: Good, 3: Easy
	err           error
	reviewedCount int
	quitting bool
}

func NewReviewModel(entries []database.Entry, db *database.Queries) *reviewModel {
	return &reviewModel{
		entries:    entries,
		db:         db,
		showAnswer: false,
		selection:  2, // Default to "Good"
	}
}

func (m *reviewModel) Init() tea.Cmd {
	return nil
}

func (m *reviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		// Reveal answer, then on second press submit the selected score
		case "enter", " ":
			if !m.showAnswer {
				m.showAnswer = true
				return m, nil
			}
			return m.submitScore()

		case "left", "h", "up":
			if m.showAnswer && m.selection > 0 {
				m.selection--
			}
		case "right", "l", "down":
			if m.showAnswer && m.selection < 3 {
				m.selection++
			}

		// Direct score selection via number keys
		case "1", "2", "3", "4":
			if !m.showAnswer {
				m.showAnswer = true
			}
			switch msg.String() {
			case "1":
				m.selection = 0
			case "2":
				m.selection = 1
			case "3":
				m.selection = 2
			case "4":
				m.selection = 3
			}
			return m.submitScore()
		}

	case reviewUpdatedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}

		if m.quitting {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *reviewModel) submitScore() (tea.Model, tea.Cmd) {
	// Map UI selection to SM-2 quality score
	quality := 0
	switch m.selection {
	case 0:
		quality = 1 // Again (failed)
	case 1:
		quality = 3 // Hard
	case 2:
		quality = 4 // Good
	case 3:
		quality = 5 // Easy
	}

	currentEntry := m.entries[m.currentIndex]
	updateParams := calculateNextReview(currentEntry, quality)

	saveCmd := func() tea.Msg {
		err := m.db.UpdateReview(context.Background(), updateParams)
		return reviewUpdatedMsg{err}
	}

	m.reviewedCount++
	m.currentIndex++
	m.showAnswer = false
	m.selection = 2 // Reset to "Good" for the next card

	if m.currentIndex >= len(m.entries) {
		m.quitting = true
		return m, saveCmd
	}

	return m, saveCmd
}

func (m *reviewModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n", m.err)
	}
	if m.currentIndex >= len(m.entries) {
		return ""
	}

	entry := m.entries[m.currentIndex]

	header := reviewTitleStyle.Render(fmt.Sprintf("Reviewing %d of %d", m.currentIndex+1, len(m.entries)))

	var tagView string
	if entry.Tag != "" {
    tagView = reviewFocusedStyle.Render("● " + entry.Tag)
	} else {
		tagView = reviewBlurredStyle.Render("○ no tags")
	}

	bodyContent := entry.Body
	if len(bodyContent) > 56 {
		bodyContent = ""
		words := strings.Fields(entry.Body)
		line := ""
		for _, word := range words {
			if len(line)+len(word) > 56 {
				bodyContent += line + "\n"
				line = word + " "
			} else {
				line += word + " "
			}
		}
		bodyContent += line
	}
	bodyView := reviewBorderStyle.Render(bodyContent)

	var controls string
	if !m.showAnswer {
		controls = reviewHelpStyle.Render("enter: reveal • q: quit")
	} else {
		scores := []string{"Again", "Hard", "Good", "Easy"}
		var scoreButtons []string
		for i, name := range scores {
			btnText := fmt.Sprintf("%d:%s", i+1, name)
			if m.selection == i {
				scoreButtons = append(scoreButtons, scoreSelectedStyle.Render(btnText))
			} else {
				scoreButtons = append(scoreButtons, scoreNormalStyle.Render(btnText))
			}
		}
		controls = lipgloss.JoinVertical(
			lipgloss.Left,
			reviewHelpStyle.Render("How well did you remember this?"),
			lipgloss.JoinHorizontal(lipgloss.Top, scoreButtons...),
			reviewHelpStyle.Render("\narrows: navigate • enter: confirm • 1-4: direct select • q: quit"),
		)
	}

	s := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tagView,
		"",
		bodyView,
		"",
		controls,
	)

	return reviewDocStyle.Render(s)
}