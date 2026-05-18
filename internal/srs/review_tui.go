package srs

import (
	"context"
	"fmt"

	"github.com/brendreyes/til/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// reviewUpdatedMsg is sent when a DB update finishes
type reviewUpdatedMsg struct{ err error }

type reviewModel struct {
	entries       []database.Entry
	currentIndex  int
	db            *database.Queries
	showAnswer    bool
	err           error
	reviewedCount int
}

func NewReviewModel(entries []database.Entry, db *database.Queries) *reviewModel {
	return &reviewModel{
		entries:      entries,
		db:           db,
		showAnswer:   false,
	}
}

func (m *reviewModel) Init() tea.Cmd {
	return nil
}

func (m *reviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// 1. Reveal Answer
		case "enter", " ":
			if !m.showAnswer {
				m.showAnswer = true
				return m, nil
			}

		// 2. Score Answer (Only if answer is shown)
		case "1", "2", "3", "4":
			if !m.showAnswer {
				return m, nil // Ignore if they haven't revealed the answer
			}

			// Map keypress to SM-2 quality score
			quality := 0
			switch msg.String() {
			case "1": quality = 1 // Again (Failed)
			case "2": quality = 3 // Hard
			case "3": quality = 4 // Good
			case "4": quality = 5 // Easy
			}

			// Calculate the new values
			currentEntry := m.entries[m.currentIndex]
			updateParams := calculateNextReview(currentEntry, quality)

			// Create a command to save to DB in the background
			saveCmd := func() tea.Msg {
				err := m.db.UpdateReview(context.Background(), updateParams)
				return reviewUpdatedMsg{err}
			}

			// Advance the card
			m.reviewedCount++
			m.currentIndex++
			m.showAnswer = false

			// If we are at the end, quit
			if m.currentIndex >= len(m.entries) {
				return m, tea.Batch(saveCmd, tea.Quit)
			}

			// Otherwise, return the model and the save command
			return m, saveCmd
		}

	// 3. Handle Database Error
	case reviewUpdatedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *reviewModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n", m.err)
	}
	if m.currentIndex >= len(m.entries) {
		return "\nAll done!\n"
	}

	entry := m.entries[m.currentIndex]

	// Basic Styling
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62")).MarginBottom(1)
	tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(2)

	// Build View
	view := titleStyle.Render(fmt.Sprintf("Reviewing %d of %d", m.currentIndex+1, len(m.entries))) + "\n"
	view += fmt.Sprintf("ID: %d\n", entry.ID)

	if entry.Tag.Valid {
		view += tagStyle.Render(fmt.Sprintf("Tags: %s\n", entry.Tag.String))
	}

	view += "\n" + entry.Body + "\n"

	// Show Controls
	if !m.showAnswer {
		view += helpStyle.Render("\n[Enter/Space] Continue/Proceed   [q] Quit")
	} else {
		// NOTE: In a real flashcard app, the answer is hidden.
		// Since TIL entries are just bodies of text, we just ask "How well did you know this?"
		view += helpStyle.Render("\nHow well did you recall this?\n[1] Again  [2] Hard  [3] Good  [4] Easy   [q] Quit")
	}

	return view
}