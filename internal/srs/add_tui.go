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
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

type entryAddedMsg struct {
	entry database.Entry
	err   error
}

// ---------------------------------------------------------------------------
// addScreenModel — embeds the form fields directly as a TUI page
// ---------------------------------------------------------------------------

// focus zones for the add form
const (
	addFocusBody    = 0
	addFocusTag     = 1
	addFocusButtons = 2
)

// button indices for the add form bar
const (
	addBtnSave   = 0
	addBtnCancel = 1
)

type addScreenModel struct {
	bodyInput textarea.Model
	tagInput  textinput.Model
	focus     int // addFocusBody | addFocusTag | addFocusButtons
	bar       buttonBar
	err       error

	// confirmation state
	savedBody  string
	savedTag   string
	saveErr    error
	confirming bool
}

func newAddScreenModel() addScreenModel {
	ta := textarea.New()
	ta.Placeholder = "What did you learn?"
	ta.CharLimit = 1000
	ta.SetWidth(50)
	ta.SetHeight(5)
	ta.Focus()

	ti := textinput.New()
	ti.Placeholder = "e.g. go, sql, algorithms"
	ti.CharLimit = 100
	ti.Width = 30

	bar := newButtonBar([]string{"Save", "Cancel"})
	bar.SetDanger(addBtnCancel)

	return addScreenModel{
		bodyInput: ta,
		tagInput:  ti,
		focus:     addFocusBody,
		bar:       bar,
	}
}

func (m addScreenModel) validate() error {
	if strings.TrimSpace(m.bodyInput.Value()) == "" {
		return fmt.Errorf("Body cannot be empty")
	}
	if strings.TrimSpace(m.tagInput.Value()) == "" {
		return fmt.Errorf("Tag cannot be empty")
	}
	return nil
}

// resetForm clears fields and returns focus to body.
func resetAddForm() addScreenModel {
	return newAddScreenModel()
}

// ---------------------------------------------------------------------------
// saveEntryCmd — async DB write
// ---------------------------------------------------------------------------

func saveEntryCmd(db *database.Queries, body, tag string) tea.Cmd {
	return func() tea.Msg {
		entry, err := db.CreateEntry(context.Background(), database.CreateEntryParams{
			Body:           body,
			Tag:            tag,
			CreatedAt:      timeNow(),
			UpdatedAt:      timeNow(),
			LastReviewedAt: timeNow(),
		})
		return entryAddedMsg{entry: entry, err: err}
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (a AppModel) updateAdd(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ── Handle confirmation screen ──────────────────────────────────────────
	if a.add.confirming {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", " ", "esc", "q":
				a.add = resetAddForm()
				return a, a.add.bodyInput.Focus()
			}
		}
		return a, nil
	}

	// ── Handle async save result ────────────────────────────────────────────
	if ev, ok := msg.(entryAddedMsg); ok {
		if ev.err != nil {
			a.add.saveErr = ev.err
			return a, nil
		}
		a.add.confirming = true
		a.add.savedBody = ev.entry.Body
		a.add.savedTag = ev.entry.Tag
		return a, nil
	}

	// ── Form key handling ───────────────────────────────────────────────────
	switch msg := msg.(type) {
	case tea.KeyMsg:
		a.add.err = nil
		a.add.saveErr = nil

		switch msg.String() {
		case "esc":
			a.current = screenMenu
			return a, nil

		case "ctrl+s", "ctrl+enter":
			if err := a.add.validate(); err != nil {
				a.add.err = err
				return a, nil
			}
			body := strings.TrimSpace(a.add.bodyInput.Value())
			tag := strings.TrimSpace(a.add.tagInput.Value())
			return a, saveEntryCmd(a.db, body, tag)

		case "tab":
			switch a.add.focus {
			case addFocusBody:
				a.add.focus = addFocusTag
				a.add.bodyInput.Blur()
				a.add.bar.focused = false
				return a, a.add.tagInput.Focus()
			case addFocusTag:
				a.add.focus = addFocusButtons
				a.add.tagInput.Blur()
				a.add.bar.focused = true
				return a, nil
			case addFocusButtons:
				a.add.focus = addFocusBody
				a.add.bar.focused = false
				return a, a.add.bodyInput.Focus()
			}

		case "shift+tab":
			switch a.add.focus {
			case addFocusBody:
				a.add.focus = addFocusButtons
				a.add.bodyInput.Blur()
				a.add.bar.focused = true
				return a, nil
			case addFocusTag:
				a.add.focus = addFocusBody
				a.add.tagInput.Blur()
				a.add.bar.focused = false
				return a, a.add.bodyInput.Focus()
			case addFocusButtons:
				a.add.focus = addFocusTag
				a.add.bar.focused = false
				return a, a.add.tagInput.Focus()
			}
		}

		// ── Button bar active ────────────────────────────────────────────────
		if a.add.focus == addFocusButtons {
			a.add.bar.HandleKey(msg.String())
			if a.add.bar.Activated() {
				switch a.add.bar.ActiveIndex() {
				case addBtnSave:
					if err := a.add.validate(); err != nil {
						a.add.err = err
						// Return focus to body on validation failure
						a.add.focus = addFocusBody
						a.add.bar.focused = false
						return a, a.add.bodyInput.Focus()
					}
					body := strings.TrimSpace(a.add.bodyInput.Value())
					tag := strings.TrimSpace(a.add.tagInput.Value())
					return a, saveEntryCmd(a.db, body, tag)
				case addBtnCancel:
					a.current = screenMenu
					return a, nil
				}
			}
			return a, nil
		}

		// ── enter on tag field saves ─────────────────────────────────────────
		if msg.String() == "enter" && a.add.focus == addFocusTag {
			if err := a.add.validate(); err != nil {
				a.add.err = err
				return a, nil
			}
			body := strings.TrimSpace(a.add.bodyInput.Value())
			tag := strings.TrimSpace(a.add.tagInput.Value())
			return a, saveEntryCmd(a.db, body, tag)
		}
	}

	// ── Delegate input to focused field ─────────────────────────────────────
	var cmd tea.Cmd
	if a.add.focus == addFocusBody {
		a.add.bodyInput, cmd = a.add.bodyInput.Update(msg)
	} else if a.add.focus == addFocusTag {
		a.add.tagInput, cmd = a.add.tagInput.Update(msg)
	}
	return a, cmd
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (m addScreenModel) View() string {
	title := appTitleStyle.Render("  Add Entry  ")

	// ── Confirmation screen ─────────────────────────────────────────────────
	if m.confirming {
		preview := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("46")).
			Padding(0, 2).
			Width(52).
			Render(m.savedBody)

		tag := appAccentStyle.Render(m.savedTag)

		body := lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			appSuccessStyle.Render("✓ Entry saved!"),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Tag: ")+tag,
			"",
			preview,
			"",
			appHelpStyle.Render("press any key to add another • esc (on form) to go back"),
		)
		return appDocStyle.Render(body)
	}

	// ── Error banner (save-level, not validation) ───────────────────────────
	var saveErrDisplay string
	if m.saveErr != nil {
		saveErrDisplay = appErrorStyle.Render("✗ "+m.saveErr.Error()) + "\n"
	}

	// ── Validation error ────────────────────────────────────────────────────
	var errDisplay string
	if m.err != nil {
		errDisplay = appErrorStyle.Render("✗ "+m.err.Error()) + "\n"
	}

	// ── Body field ──────────────────────────────────────────────────────────
	bodyLabel := "Learning Entry"
	tagLabel := "Tag (required)"

	var bodyView, tagView string
	switch m.focus {
	case addFocusBody:
		bodyLabel = focusedStyle.Render("● " + bodyLabel)
		tagLabel = blurredStyle.Render("  " + tagLabel)
		bodyView = focusedBorderStyle.Render(m.bodyInput.View())
		tagView = blurredBorderStyle.Render(m.tagInput.View())
	case addFocusTag:
		bodyLabel = blurredStyle.Render("  " + bodyLabel)
		tagLabel = focusedStyle.Render("● " + tagLabel)
		bodyView = blurredBorderStyle.Render(m.bodyInput.View())
		tagView = focusedBorderStyle.Render(m.tagInput.View())
	case addFocusButtons:
		bodyLabel = blurredStyle.Render("  " + bodyLabel)
		tagLabel = blurredStyle.Render("  " + tagLabel)
		bodyView = blurredBorderStyle.Render(m.bodyInput.View())
		tagView = blurredBorderStyle.Render(m.tagInput.View())
	}

	bar := m.bar.View()

	var help string
	if m.focus == addFocusButtons {
		help = appHelpStyle.Render("←/→: choose action • enter: activate • tab: back to body")
	} else {
		help = appHelpStyle.Render("tab: next field • enter: save (on tag) • ctrl+s: save • esc: menu")
	}

	s := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		saveErrDisplay,
		errDisplay,
		bodyLabel,
		bodyView,
		"",
		tagLabel,
		tagView,
		bar,
		help,
	)

	return appDocStyle.Render(s)
}