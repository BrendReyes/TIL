package srs

import (
	"fmt"
	"strings"

	"github.com/brendreyes/til/internal/database"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// Standalone add TUI (used by `til add` with no args)
// Wraps addScreenModel in its own minimal program that exits to terminal.
// Uses the shared saveEntryCmd from add_tui.go.
// ---------------------------------------------------------------------------

type standaloneAddModel struct {
	form addScreenModel
	db   *database.Queries
}

func (s standaloneAddModel) Init() tea.Cmd {
	return textarea.Blink
}

func (s standaloneAddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Intercept quit / esc
	if kMsg, ok := msg.(tea.KeyMsg); ok {
		switch kMsg.String() {
		case "ctrl+c":
			return s, tea.Quit

		case "esc":
			if !s.form.confirming {
				return s, tea.Quit
			}
			// esc on confirmation → quit entirely
			return s, tea.Quit
		}

		// On confirmation screen, any non-esc key → fresh form
		if s.form.confirming {
			s.form = resetAddForm()
			return s, textarea.Blink
		}
	}

	// Handle save result
	if ev, ok := msg.(entryAddedMsg); ok {
		if ev.err != nil {
			s.form.saveErr = ev.err
			return s, nil
		}
		s.form.confirming = true
		s.form.savedBody = ev.entry.Body
		s.form.savedTag = ev.entry.Tag
		return s, nil
	}

	// Form key input
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s.form.err = nil
		s.form.saveErr = nil

		switch msg.String() {
		case "ctrl+s", "ctrl+enter":
			if err := s.form.validate(); err != nil {
				s.form.err = err
				return s, nil
			}
			body := strings.TrimSpace(s.form.bodyInput.Value())
			tag := strings.TrimSpace(s.form.tagInput.Value())
			return s, saveEntryCmd(s.db, body, tag)

		case "enter":
			if s.form.focus == 1 {
				if err := s.form.validate(); err != nil {
					s.form.err = err
					return s, nil
				}
				body := strings.TrimSpace(s.form.bodyInput.Value())
				tag := strings.TrimSpace(s.form.tagInput.Value())
				return s, saveEntryCmd(s.db, body, tag)
			}

		case "tab", "shift+tab":
			if s.form.focus == 0 {
				s.form.focus = 1
				s.form.bodyInput.Blur()
				return s, s.form.tagInput.Focus()
			} else {
				s.form.focus = 0
				s.form.tagInput.Blur()
				return s, s.form.bodyInput.Focus()
			}
		}
	}

	var cmd tea.Cmd
	if s.form.focus == 0 {
		s.form.bodyInput, cmd = s.form.bodyInput.Update(msg)
	} else {
		s.form.tagInput, cmd = s.form.tagInput.Update(msg)
	}
	return s, cmd
}

func (s standaloneAddModel) View() string {
	return s.form.View()
}

// RunAddTUI — entry point for `til add` with no args.
func (s *State) RunAddTUI() error {
	m := standaloneAddModel{
		form: newAddScreenModel(),
		db:   s.DB,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return err
	}
	fmt.Println("✓ Done.")
	return nil
}