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

type addScreenModel struct {
	bodyInput textarea.Model
	tagInput  textinput.Model
	focus     int // 0 = body, 1 = tag
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

	return addScreenModel{
		bodyInput: ta,
		tagInput:  ti,
		focus:     0,
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
				// Any key: back to fresh form
				a.add = resetAddForm()
				return a, textarea.Blink
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

		case "enter":
			if a.add.focus == 1 {
				if err := a.add.validate(); err != nil {
					a.add.err = err
					return a, nil
				}
				body := strings.TrimSpace(a.add.bodyInput.Value())
				tag := strings.TrimSpace(a.add.tagInput.Value())
				return a, saveEntryCmd(a.db, body, tag)
			}

		case "tab", "shift+tab":
			if a.add.focus == 0 {
				a.add.focus = 1
				a.add.bodyInput.Blur()
				return a, a.add.tagInput.Focus()
			} else {
				a.add.focus = 0
				a.add.tagInput.Blur()
				return a, a.add.bodyInput.Focus()
			}
		}
	}

	// ── Delegate input to focused field ─────────────────────────────────────
	var cmd tea.Cmd
	if a.add.focus == 0 {
		a.add.bodyInput, cmd = a.add.bodyInput.Update(msg)
	} else {
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
	if m.focus == 0 {
		bodyLabel = focusedStyle.Render("● " + bodyLabel)
		tagLabel = blurredStyle.Render("  " + tagLabel)
		bodyView = focusedBorderStyle.Render(m.bodyInput.View())
		tagView = blurredBorderStyle.Render(m.tagInput.View())
	} else {
		bodyLabel = blurredStyle.Render("  " + bodyLabel)
		tagLabel = focusedStyle.Render("● " + tagLabel)
		bodyView = blurredBorderStyle.Render(m.bodyInput.View())
		tagView = focusedBorderStyle.Render(m.tagInput.View())
	}

	help := appHelpStyle.Render("tab: switch field • enter: save (on tag) • ctrl+s: save • esc: menu")

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
		help,
	)

	return appDocStyle.Render(s)
}