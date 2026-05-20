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

		case "tab":
			switch s.form.focus {
			case addFocusBody:
				s.form.focus = addFocusTag
				s.form.bodyInput.Blur()
				s.form.bar.focused = false
				return s, s.form.tagInput.Focus()
			case addFocusTag:
				s.form.focus = addFocusButtons
				s.form.tagInput.Blur()
				s.form.bar.focused = true
				return s, nil
			case addFocusButtons:
				s.form.focus = addFocusBody
				s.form.bar.focused = false
				return s, s.form.bodyInput.Focus()
			}

		case "shift+tab":
			switch s.form.focus {
			case addFocusBody:
				s.form.focus = addFocusButtons
				s.form.bodyInput.Blur()
				s.form.bar.focused = true
				return s, nil
			case addFocusTag:
				s.form.focus = addFocusBody
				s.form.tagInput.Blur()
				s.form.bar.focused = false
				return s, s.form.bodyInput.Focus()
			case addFocusButtons:
				s.form.focus = addFocusTag
				s.form.bar.focused = false
				return s, s.form.tagInput.Focus()
			}
		}

		// Button bar active
		if s.form.focus == addFocusButtons {
			s.form.bar.HandleKey(msg.String())
			if s.form.bar.Activated() {
				switch s.form.bar.ActiveIndex() {
				case addBtnSave:
					if err := s.form.validate(); err != nil {
						s.form.err = err
						s.form.focus = addFocusBody
						s.form.bar.focused = false
						return s, s.form.bodyInput.Focus()
					}
					body := strings.TrimSpace(s.form.bodyInput.Value())
					tag := strings.TrimSpace(s.form.tagInput.Value())
					return s, saveEntryCmd(s.db, body, tag)
				case addBtnCancel:
					return s, tea.Quit
				}
			}
			return s, nil
		}

		// enter on tag field saves
		if msg.String() == "enter" && s.form.focus == addFocusTag {
			if err := s.form.validate(); err != nil {
				s.form.err = err
				return s, nil
			}
			body := strings.TrimSpace(s.form.bodyInput.Value())
			tag := strings.TrimSpace(s.form.tagInput.Value())
			return s, saveEntryCmd(s.db, body, tag)
		}
	}

	var cmd tea.Cmd
	if s.form.focus == addFocusBody {
		s.form.bodyInput, cmd = s.form.bodyInput.Update(msg)
	} else if s.form.focus == addFocusTag {
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